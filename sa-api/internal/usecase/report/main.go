package report

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/hdbank/smart-attendance/internal/domain/repository"
	"github.com/hdbank/smart-attendance/internal/domain/usecase"
	"github.com/hdbank/smart-attendance/internal/infrastructure/cache"
	"github.com/hdbank/smart-attendance/pkg/utils"
)

type reportUsecase struct {
	attendanceRepo repository.AttendanceRepository
	userRepo       repository.UserRepository
	branchRepo     repository.BranchRepository
	cache          cache.Cache
}

// NewReportUsecase tạo instance ReportUsecase
func NewReportUsecase(
	attendanceRepo repository.AttendanceRepository,
	userRepo repository.UserRepository,
	branchRepo repository.BranchRepository,
	cache cache.Cache,
) usecase.ReportUsecase {
	return &reportUsecase{
		attendanceRepo: attendanceRepo,
		userRepo:       userRepo,
		branchRepo:     branchRepo,
		cache:          cache,
	}
}

// GetTodayBranchStats thống kê chấm công hôm nay theo từng chi nhánh.
// Cache 2 phút — đủ ngắn để dashboard phản ánh gần realtime.
func (u *reportUsecase) GetTodayBranchStats(ctx context.Context, filter usecase.TodayStatsFilter) ([]*repository.BranchTodayStats, int64, error) {
	cacheKey := fmt.Sprintf("dashboard:today:branch:%v:q:%s:p%d:l%d",
		filter.BranchID, filter.Search, filter.Page, filter.Limit)

	type cachedPage struct {
		Items []*repository.BranchTodayStats `json:"items"`
		Total int64                           `json:"total"`
	}
	var cached cachedPage
	if err := u.cache.Get(ctx, cacheKey, &cached); err == nil {
		return cached.Items, cached.Total, nil
	}

	items, total, err := u.attendanceRepo.GetTodayStatsByBranch(ctx, filter.BranchID, filter.Search, filter.Page, filter.Limit)
	if err != nil {
		return nil, 0, err
	}

	u.cache.Set(ctx, cacheKey, cachedPage{Items: items, Total: total}, 2*time.Minute)
	return items, total, nil
}

// GetTodayEmployeeDetails danh sách chi tiết nhân viên theo trạng thái hôm nay.
// Cache 2 phút — đủ ngắn để phản ánh gần realtime.
func (u *reportUsecase) GetTodayEmployeeDetails(ctx context.Context, filter repository.TodayEmployeeFilter) ([]*repository.EmployeeTodayDetail, int64, error) {
	cacheKey := fmt.Sprintf("dashboard:today:employees:branch:%v:status:%s:p%d:l%d",
		filter.BranchID, filter.Status, filter.Page, filter.Limit)

	type cachedPage struct {
		Items []*repository.EmployeeTodayDetail `json:"items"`
		Total int64                              `json:"total"`
	}
	var cached cachedPage
	if err := u.cache.Get(ctx, cacheKey, &cached); err == nil {
		return cached.Items, cached.Total, nil
	}

	items, total, err := u.attendanceRepo.GetTodayEmployeeDetails(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	u.cache.Set(ctx, cacheKey, cachedPage{Items: items, Total: total}, 2*time.Minute)
	return items, total, nil
}

// GetDashboardStats lấy thống kê tổng quan cho admin
// Cache 5 phút để tránh query nặng liên tục
func (u *reportUsecase) GetDashboardStats(ctx context.Context) (*usecase.DashboardStats, error) {
	cacheKey := "dashboard:admin:stats"
	var cached usecase.DashboardStats
	if err := u.cache.Get(ctx, cacheKey, &cached); err == nil {
		return &cached, nil
	}

	stats, err := u.computeDashboardStats(ctx, nil)
	if err != nil {
		return nil, err
	}

	u.cache.Set(ctx, cacheKey, stats, 5*time.Minute)
	return stats, nil
}

// GetBranchDashboard lấy thống kê cho manager theo chi nhánh
func (u *reportUsecase) GetBranchDashboard(ctx context.Context, branchID uint) (*usecase.DashboardStats, error) {
	cacheKey := cache.BuildKey("dashboard:branch:", string(rune(branchID)))
	var cached usecase.DashboardStats
	if err := u.cache.Get(ctx, cacheKey, &cached); err == nil {
		return &cached, nil
	}

	stats, err := u.computeDashboardStats(ctx, &branchID)
	if err != nil {
		return nil, err
	}

	u.cache.Set(ctx, cacheKey, stats, 5*time.Minute)
	return stats, nil
}

// computeDashboardStats tính toán dashboard stats từ GetTodayStatsByBranch (đã filter employee)
func (u *reportUsecase) computeDashboardStats(ctx context.Context, branchID *uint) (*usecase.DashboardStats, error) {
	// Lấy thống kê hôm nay per branch — SQL đã JOIN users WHERE role='employee'
	summaries, _, err := u.attendanceRepo.GetTodayStatsByBranch(ctx, branchID, "", 1, 1000)
	if err != nil {
		return nil, err
	}

	var totalEmployees, totalPresent, presentCount, lateEarlyCount int64
	for _, s := range summaries {
		totalEmployees += int64(s.TotalEmployees)
		totalPresent += int64(s.PresentCount + s.LateCount + s.EarlyLeaveCount + s.HalfDayCount)
		presentCount += int64(s.PresentCount)
		lateEarlyCount += int64(s.LateCount + s.EarlyLeaveCount + s.HalfDayCount) // Gom Đi trễ + Về sớm
	}

	totalBranches := int64(len(summaries))

	absent := totalEmployees - totalPresent
	if absent < 0 {
		absent = 0
	}

	attendanceRate := float64(0)
	onTimeRate := float64(0)
	if totalEmployees > 0 {
		attendanceRate = float64(totalPresent) / float64(totalEmployees) * 100
		if totalPresent > 0 {
			onTimeRate = float64(presentCount) / float64(totalPresent) * 100
		}
	}

	return &usecase.DashboardStats{
		TotalBranches:  totalBranches,
		TotalEmployees: totalEmployees,
		PresentToday:   totalPresent,
		AbsentToday:    absent,
		LateToday:      lateEarlyCount,
		AttendanceRate: roundToTwoDecimal(attendanceRate),
		OnTimeRate:     roundToTwoDecimal(onTimeRate),
	}, nil
}

// GetAttendanceReport lấy báo cáo chấm công chi tiết theo user
func (u *reportUsecase) GetAttendanceReport(ctx context.Context, filter usecase.ReportFilter) ([]*repository.UserAttendanceSummary, int64, error) {
	from, to := u.resolveDateRange(filter)

	if filter.BranchID != nil {
		results, err := u.attendanceRepo.GetBranchSummary(ctx, *filter.BranchID, from, to)
		if err != nil {
			return nil, 0, err
		}
		// Phân trang thủ công trên kết quả aggregate
		total := int64(len(results))
		start := (filter.Page - 1) * filter.Limit
		end := start + filter.Limit
		if start >= len(results) {
			return []*repository.UserAttendanceSummary{}, total, nil
		}
		if end > len(results) {
			end = len(results)
		}
		return results[start:end], total, nil
	}

	return []*repository.UserAttendanceSummary{}, 0, nil
}

// GetBranchReport lấy báo cáo tổng hợp theo từng chi nhánh
func (u *reportUsecase) GetBranchReport(ctx context.Context, filter usecase.ReportFilter) ([]*usecase.BranchAttendanceReport, error) {
	from, to := u.resolveDateRange(filter)

	branches, err := u.branchRepo.FindActive(ctx)
	if err != nil {
		return nil, err
	}

	var reports []*usecase.BranchAttendanceReport
	for _, branch := range branches {
		// Nếu filter theo chi nhánh cụ thể (manager), bỏ qua các chi nhánh khác
		if filter.BranchID != nil && branch.ID != *filter.BranchID {
			continue
		}

		summaries, err := u.attendanceRepo.GetBranchSummary(ctx, branch.ID, from, to)
		if err != nil {
			slog.Error("failed to get branch summary for report", "branch_id", branch.ID, "branch_code", branch.Code, "error", err)
			continue
		}

		report := &usecase.BranchAttendanceReport{
			BranchID:   branch.ID,
			BranchName: branch.Name,
			BranchCode: branch.Code,
		}

		// Tổng hợp từ tất cả nhân viên
		for _, s := range summaries {
			report.TotalDays += s.TotalDays
			report.PresentCount += s.PresentCount
			report.LateCount += s.LateCount
			report.EarlyLeaveCount += s.EarlyLeaveCount
			report.HalfDayCount += s.HalfDayCount
			report.AbsentCount += s.AbsentCount
			report.TotalWorkHours += s.TotalWorkHours
			report.TotalOvertime += s.TotalOvertime
		}

		// Tính tỷ lệ
		if report.TotalDays > 0 {
			total := report.PresentCount + report.LateCount
			report.AttendanceRate = roundToTwoDecimal(float64(total) / float64(report.TotalDays) * 100)
			if total > 0 {
				report.OnTimeRate = roundToTwoDecimal(float64(report.PresentCount) / float64(total) * 100)
			}
		}

		if filter.UserID == nil {
			report.Employees = summaries
		}
		reports = append(reports, report)
	}

	return reports, nil
}

// GetUserReport lấy báo cáo chi tiết của một nhân viên
func (u *reportUsecase) GetUserReport(ctx context.Context, userID uint, from, to time.Time) (*repository.UserAttendanceSummary, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	summary, err := u.attendanceRepo.GetSummary(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}

	return &repository.UserAttendanceSummary{
		UserID:            user.ID,
		UserName:          user.Name,
		EmployeeCode:      user.EmployeeCode,
		Department:        user.Department,
		AttendanceSummary: *summary,
	}, nil
}

// resolveDateRange tính toán khoảng thời gian dựa trên ReportPeriod
func (u *reportUsecase) resolveDateRange(filter usecase.ReportFilter) (time.Time, time.Time) {
	now := utils.Now()
	switch filter.Period {
	case usecase.PeriodDaily:
		today := utils.StartOfDay(now)
		return today, today
	case usecase.PeriodWeekly:
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		monday := now.AddDate(0, 0, -(weekday - 1))
		from := utils.StartOfDay(monday)
		return from, now
	case usecase.PeriodMonthly:
		from := utils.DateInHCM(now.Year(), now.Month(), 1, 0, 0, 0)
		return from, now
	default:
		return filter.DateFrom, filter.DateTo
	}
}

func roundToTwoDecimal(f float64) float64 {
	return float64(int(f*100)) / 100
}
