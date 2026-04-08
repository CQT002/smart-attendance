package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	domainrepo "github.com/hdbank/smart-attendance/internal/domain/repository"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"github.com/hdbank/smart-attendance/pkg/utils"
	"gorm.io/gorm"
)

type attendanceRepository struct {
	db *gorm.DB
}

// NewAttendanceRepository tạo instance AttendanceRepository với PostgreSQL
func NewAttendanceRepository(db *gorm.DB) domainrepo.AttendanceRepository {
	return &attendanceRepository{db: db}
}

func (r *attendanceRepository) Create(ctx context.Context, log *entity.AttendanceLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		slog.Error("attendance create failed", "user_id", log.UserID, "error", err)
		return apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi tạo bản ghi chấm công")
	}
	return nil
}

func (r *attendanceRepository) Update(ctx context.Context, log *entity.AttendanceLog) error {
	if err := r.db.WithContext(ctx).Save(log).Error; err != nil {
		slog.Error("attendance update failed", "id", log.ID, "error", err)
		return apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi cập nhật bản ghi chấm công")
	}
	return nil
}

func (r *attendanceRepository) FindByID(ctx context.Context, id uint) (*entity.AttendanceLog, error) {
	var log entity.AttendanceLog
	err := r.db.WithContext(ctx).
		Preload("User").Preload("Branch").Preload("Shift").
		First(&log, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn chấm công")
	}
	return &log, nil
}

// FindByUserAndDate tìm bản ghi chấm công của user trong ngày
// Index composite (user_id, date) được đánh để query này O(log n)
func (r *attendanceRepository) FindByUserAndDate(ctx context.Context, userID uint, date time.Time) (*entity.AttendanceLog, error) {
	var log entity.AttendanceLog
	dateStr := date.Format("2006-01-02")
	err := r.db.WithContext(ctx).
		Preload("Shift").
		Where("user_id = ? AND DATE(date) = ?", userID, dateStr).
		First(&log).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Không phải lỗi, chỉ là chưa có bản ghi
		}
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn chấm công")
	}
	return &log, nil
}

// FindAll lấy danh sách chấm công với nhiều bộ lọc
// Sử dụng các index: user_id, branch_id, date, status
func (r *attendanceRepository) FindAll(ctx context.Context, filter domainrepo.AttendanceFilter) ([]*entity.AttendanceLog, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.AttendanceLog{})

	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.BranchID != nil {
		query = query.Where("branch_id = ?", *filter.BranchID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.DateFrom != nil {
		query = query.Where("date >= ?", filter.DateFrom.Format("2006-01-02"))
	}
	if filter.DateTo != nil {
		query = query.Where("date <= ?", filter.DateTo.Format("2006-01-02"))
	}
	// Filter theo department: join với bảng users
	if filter.Department != "" {
		query = query.Joins("JOIN users ON users.id = attendance_logs.user_id").
			Where("users.department = ?", filter.Department)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi đếm bản ghi chấm công")
	}

	offset := (filter.Page - 1) * filter.Limit
	var logs []*entity.AttendanceLog
	err := query.Preload("User").Preload("Branch").Preload("Shift").
		Order("date DESC, check_in_time DESC").
		Offset(offset).Limit(filter.Limit).
		Find(&logs).Error
	if err != nil {
		return nil, 0, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn danh sách chấm công")
	}

	return logs, total, nil
}

// GetSummary tổng hợp thống kê chấm công dùng SQL aggregate - hiệu quả hơn load toàn bộ bản ghi
func (r *attendanceRepository) GetSummary(ctx context.Context, userID uint, from, to time.Time) (*domainrepo.AttendanceSummary, error) {
	type Result struct {
		TotalDays       int
		PresentCount    int
		LateCount       int
		EarlyLeaveCount int
		HalfDayCount    int
		AbsentCount     int
		TotalWorkHours  float64
		TotalOvertime   float64
	}

	var result Result
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(*) as total_days,
			COUNT(CASE WHEN status = 'present' THEN 1 END) as present_count,
			COUNT(CASE WHEN status IN ('late', 'late_early_leave') THEN 1 END) as late_count,
			COUNT(CASE WHEN status IN ('early_leave') THEN 1 END) as early_leave_count,
			COUNT(CASE WHEN status = 'half_day' THEN 1 END) as half_day_count,
			COUNT(CASE WHEN status = 'absent' THEN 1 END) as absent_count,
			COALESCE(SUM(work_hours), 0) as total_work_hours,
			COALESCE((SELECT SUM(ot.total_hours) FROM overtime_requests ot
				WHERE ot.user_id = attendance_logs.user_id
				AND ot.date BETWEEN ? AND ?
				AND ot.status = 'approved'
				AND ot.deleted_at IS NULL), 0) as total_overtime
		FROM attendance_logs
		WHERE user_id = ? AND date BETWEEN ? AND ? AND deleted_at IS NULL
	`, from.Format("2006-01-02"), to.Format("2006-01-02"), userID, from.Format("2006-01-02"), to.Format("2006-01-02")).Scan(&result).Error

	if err != nil {
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi tổng hợp chấm công")
	}

	attendanceRate := float64(0)
	onTimeRate := float64(0)
	if result.TotalDays > 0 {
		attendanceRate = float64(result.PresentCount+result.LateCount) / float64(result.TotalDays) * 100
		total := result.PresentCount + result.LateCount
		if total > 0 {
			onTimeRate = float64(result.PresentCount) / float64(total) * 100
		}
	}

	return &domainrepo.AttendanceSummary{
		TotalDays:       result.TotalDays,
		PresentCount:    result.PresentCount,
		LateCount:       result.LateCount,
		EarlyLeaveCount: result.EarlyLeaveCount,
		HalfDayCount:    result.HalfDayCount,
		AbsentCount:     result.AbsentCount,
		TotalWorkHours:  result.TotalWorkHours,
		TotalOvertime:   result.TotalOvertime,
		AttendanceRate:  attendanceRate,
		OnTimeRate:      onTimeRate,
	}, nil
}

// GetBranchSummary tổng hợp chấm công theo chi nhánh dùng một câu SQL duy nhất
// Tránh N+1 query khi có 5000 nhân viên
func (r *attendanceRepository) GetBranchSummary(ctx context.Context, branchID uint, from, to time.Time) ([]*domainrepo.UserAttendanceSummary, error) {
	var results []*domainrepo.UserAttendanceSummary
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			u.id as user_id,
			u.name as user_name,
			u.employee_code,
			u.department,
			COUNT(a.id) as total_days,
			COUNT(CASE WHEN a.status = 'present' THEN 1 END) as present_count,
			COUNT(CASE WHEN a.status IN ('late', 'late_early_leave') THEN 1 END) as late_count,
			COUNT(CASE WHEN a.status IN ('early_leave') THEN 1 END) as early_leave_count,
			COUNT(CASE WHEN a.status = 'half_day' THEN 1 END) as half_day_count,
			COUNT(CASE WHEN a.status = 'absent' THEN 1 END) as absent_count,
			COALESCE(SUM(a.work_hours), 0) as total_work_hours,
			COALESCE((SELECT SUM(ot.total_hours) FROM overtime_requests ot
				WHERE ot.user_id = u.id AND ot.date BETWEEN ? AND ?
				AND ot.status = 'approved' AND ot.deleted_at IS NULL), 0) as total_overtime,
			CASE WHEN COUNT(a.id) > 0
				THEN ROUND(
					(COUNT(CASE WHEN a.status IN ('present','late','early_leave','late_early_leave','half_day') THEN 1 END)::numeric / COUNT(a.id)) * 100, 2
				)
				ELSE 0
			END as attendance_rate,
			CASE WHEN COUNT(CASE WHEN a.status IN ('present','late','early_leave','late_early_leave','half_day') THEN 1 END) > 0
				THEN ROUND(
					(COUNT(CASE WHEN a.status = 'present' THEN 1 END)::numeric
					/ COUNT(CASE WHEN a.status IN ('present','late','early_leave','late_early_leave','half_day') THEN 1 END)) * 100, 2
				)
				ELSE 0
			END as on_time_rate
		FROM users u
		LEFT JOIN attendance_logs a ON a.user_id = u.id
			AND a.date BETWEEN ? AND ? AND a.deleted_at IS NULL
		WHERE u.branch_id = ? AND u.is_active = true AND u.role = 'employee' AND u.deleted_at IS NULL
		GROUP BY u.id, u.name, u.employee_code, u.department
		ORDER BY u.name
	`, from.Format("2006-01-02"), to.Format("2006-01-02"), from.Format("2006-01-02"), to.Format("2006-01-02"), branchID).
		Scan(&results).Error

	if err != nil {
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi tổng hợp chấm công chi nhánh")
	}

	return results, nil
}

func (r *attendanceRepository) FindActiveCheckIn(ctx context.Context, userID uint) (*entity.AttendanceLog, error) {
	today := utils.Now().Format("2006-01-02")
	var log entity.AttendanceLog
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND DATE(date) = ? AND check_in_time IS NOT NULL AND check_out_time IS NULL", userID, today).
		First(&log).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi kiểm tra trạng thái chấm công")
	}
	return &log, nil
}

// CountSuspicious đếm số lần bị gắn flag gian lận trong N ngày gần đây
// Dùng để block user có nhiều lần vi phạm
func (r *attendanceRepository) CountSuspicious(ctx context.Context, userID uint, from time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.AttendanceLog{}).
		Where("user_id = ? AND (is_fake_gps = true OR is_vpn = true) AND created_at >= ?", userID, from).
		Count(&count).Error
	return count, err
}

// GetTodayStatsByBranch thống kê chấm công hôm nay per chi nhánh bằng một cặp SQL tối ưu.
//
// Dùng CTE để tính toán song song employee_counts và today_attendance,
// tránh N+1 query khi có 100 chi nhánh × 5000 nhân viên.
func (r *attendanceRepository) GetTodayStatsByBranch(ctx context.Context, branchID *uint, search string, page, limit int) ([]*domainrepo.BranchTodayStats, int64, error) {
	branchFilter := ""
	var args []interface{}

	if branchID != nil {
		branchFilter += " AND b.id = ?"
		args = append(args, *branchID)
	}

	if search != "" {
		branchFilter += " AND (b.name ILIKE ? OR b.code ILIKE ?)"
		args = append(args, "%"+search+"%", "%"+search+"%")
	}

	// ── Count tổng số chi nhánh thoả filter (để phân trang) ──
	countSQL := fmt.Sprintf(
		`SELECT COUNT(DISTINCT b.id) FROM branches b WHERE b.is_active = true AND b.deleted_at IS NULL %s`,
		branchFilter,
	)
	var total int64
	countQuery := r.db.WithContext(ctx).Raw(countSQL, args...)
	if err := countQuery.Scan(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi đếm chi nhánh")
	}

	if total == 0 {
		return []*domainrepo.BranchTodayStats{}, 0, nil
	}

	// ── Fetch dữ liệu có phân trang ──
	offset := (page - 1) * limit
	dataSQL := fmt.Sprintf(`
		WITH employee_counts AS (
			SELECT
				b.id        AS branch_id,
				b.name      AS branch_name,
				b.code      AS branch_code,
				COUNT(u.id) AS total_employees
			FROM branches b
			JOIN users u ON u.branch_id = b.id AND u.is_active = true AND u.role = 'employee' AND u.deleted_at IS NULL
			WHERE b.is_active = true AND b.deleted_at IS NULL %s
			GROUP BY b.id, b.name, b.code
		),
		today_attendance AS (
			SELECT
				a.branch_id,
				COUNT(CASE WHEN a.status = 'present'     THEN 1 END) AS present_count,
				COUNT(CASE WHEN a.status IN ('late', 'late_early_leave') THEN 1 END) AS late_count,
				COUNT(CASE WHEN a.status IN ('early_leave') THEN 1 END) AS early_leave_count,
				COUNT(CASE WHEN a.status = 'half_day'    THEN 1 END) AS half_day_count,
				COUNT(CASE WHEN a.is_fake_gps = true OR a.is_vpn = true THEN 1 END) AS suspicious_count
			FROM attendance_logs a
			WHERE a.date = CURRENT_DATE AND a.deleted_at IS NULL
			GROUP BY a.branch_id
		)
		SELECT
			ec.branch_id,
			ec.branch_name,
			ec.branch_code,
			ec.total_employees,
			COALESCE(ta.present_count,     0) AS present_count,
			COALESCE(ta.late_count,        0) AS late_count,
			COALESCE(ta.early_leave_count, 0) AS early_leave_count,
			COALESCE(ta.half_day_count,    0) AS half_day_count,
			COALESCE(ta.suspicious_count,  0) AS suspicious_count,
			ec.total_employees
				- COALESCE(ta.present_count + ta.late_count + ta.early_leave_count + ta.half_day_count, 0)
				AS absent_count,
			CASE WHEN ec.total_employees > 0
				THEN ROUND(
					COALESCE(ta.present_count + ta.late_count + ta.early_leave_count + ta.half_day_count, 0)::numeric
					/ ec.total_employees * 100, 2)
				ELSE 0
			END AS attendance_rate,
			CASE WHEN COALESCE(ta.present_count + ta.late_count + ta.early_leave_count + ta.half_day_count, 0) > 0
				THEN ROUND(
					COALESCE(ta.present_count, 0)::numeric
					/ (ta.present_count + ta.late_count + ta.early_leave_count + ta.half_day_count) * 100, 2)
				ELSE 0
			END AS on_time_rate
		FROM employee_counts ec
		LEFT JOIN today_attendance ta ON ta.branch_id = ec.branch_id
		ORDER BY ec.branch_name
		LIMIT ? OFFSET ?
	`, branchFilter)

	dataArgs := append(args, limit, offset)
	var results []*domainrepo.BranchTodayStats
	if err := r.db.WithContext(ctx).Raw(dataSQL, dataArgs...).Scan(&results).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi thống kê chấm công hôm nay")
	}

	return results, total, nil
}

// GetTodayEmployeeDetails lấy danh sách nhân viên với trạng thái chấm công hôm nay.
//
// Dùng CTE để tính derived_status:
//   - "absent"     → nhân viên không có bản ghi attendance_logs hôm nay
//   - "suspicious" → có bản ghi nhưng is_fake_gps=true hoặc is_vpn=true
//   - còn lại      → attendance_logs.status gốc
//
// Tất cả truy vấn thực hiện bằng một cặp SQL (count + data) — không N+1.
func (r *attendanceRepository) GetTodayEmployeeDetails(ctx context.Context, filter domainrepo.TodayEmployeeFilter) ([]*domainrepo.EmployeeTodayDetail, int64, error) {
	branchClause := ""
	if filter.BranchID != nil {
		branchClause = "AND u.branch_id = ?"
	}

	statusClause := todayStatusWhereClause(filter.Status)

	// ── CTE dùng chung cho count và data ──
	cteSQL := fmt.Sprintf(`
		WITH employee_today AS (
			SELECT
				u.id                                                          AS user_id,
				u.employee_code,
				u.name,
				COALESCE(u.department, '')                                    AS department,
				u.branch_id,
				b.name                                                        AS branch_name,
				CASE
					WHEN a.id IS NULL                                         THEN 'absent'
					WHEN a.is_fake_gps = true OR a.is_vpn = true             THEN 'suspicious'
					ELSE a.status::text
				END                                                           AS status,
				TO_CHAR(a.check_in_time,  'YYYY-MM-DD"T"HH24:MI:SS"Z"')     AS check_in_time,
				TO_CHAR(a.check_out_time, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')     AS check_out_time,
				COALESCE(a.work_hours, 0)                                     AS work_hours,
				COALESCE(a.is_fake_gps, false)                               AS is_fake_gps,
				COALESCE(a.is_vpn, false)                                    AS is_vpn,
				COALESCE(a.fraud_note, '')                                   AS fraud_note
			FROM users u
			JOIN branches b ON b.id = u.branch_id AND b.is_active = true AND b.deleted_at IS NULL
			LEFT JOIN attendance_logs a ON a.user_id = u.id AND a.date = CURRENT_DATE AND a.deleted_at IS NULL
			WHERE u.is_active = true AND u.role = 'employee' AND u.deleted_at IS NULL %s
		)
	`, branchClause)

	// ── Build args ──
	var baseArgs []interface{}
	if filter.BranchID != nil {
		baseArgs = append(baseArgs, *filter.BranchID)
	}

	// Count
	var total int64
	countSQL := cteSQL + fmt.Sprintf("SELECT COUNT(*) FROM employee_today %s", statusClause)
	if err := r.db.WithContext(ctx).Raw(countSQL, baseArgs...).Scan(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi đếm nhân viên hôm nay")
	}

	if total == 0 {
		return []*domainrepo.EmployeeTodayDetail{}, 0, nil
	}

	// Data
	offset := (filter.Page - 1) * filter.Limit
	dataArgs := append(baseArgs, filter.Limit, offset)
	dataSQL := cteSQL + fmt.Sprintf(`
		SELECT * FROM employee_today %s
		ORDER BY branch_name, name
		LIMIT ? OFFSET ?
	`, statusClause)

	var results []*domainrepo.EmployeeTodayDetail
	if err := r.db.WithContext(ctx).Raw(dataSQL, dataArgs...).Scan(&results).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi lấy danh sách nhân viên hôm nay")
	}

	return results, total, nil
}

// todayStatusWhereClause trả về mệnh đề WHERE cho derived_status.
// Chỉ được gọi với giá trị đã được whitelist ở handler — an toàn để embed vào SQL.
func todayStatusWhereClause(status domainrepo.TodayEmployeeStatus) string {
	switch status {
	case domainrepo.TodayStatusAbsent:
		return "WHERE status = 'absent'"
	case domainrepo.TodayStatusSuspicious:
		return "WHERE status = 'suspicious'"
	case domainrepo.TodayStatusPresent:
		return "WHERE status = 'present'"
	case domainrepo.TodayStatusLate:
		return "WHERE status = 'late'"
	case domainrepo.TodayStatusEarlyLeave:
		return "WHERE status = 'early_leave'"
	case domainrepo.TodayStatusHalfDay:
		return "WHERE status = 'half_day'"
	default:
		return "" // TodayStatusAll — không filter
	}
}
