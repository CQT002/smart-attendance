package leave

import (
	"context"
	"log/slog"
	"sort"
	"strconv"
	"time"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	"github.com/hdbank/smart-attendance/internal/domain/repository"
	"github.com/hdbank/smart-attendance/internal/domain/usecase"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"github.com/hdbank/smart-attendance/pkg/utils"
	"gorm.io/gorm"
)

type leaveUsecase struct {
	leaveRepo      repository.LeaveRepository
	correctionRepo repository.CorrectionRepository
	overtimeRepo   repository.OvertimeRepository
	attendanceRepo repository.AttendanceRepository
	userRepo       repository.UserRepository
	shiftRepo      repository.ShiftRepository
	db             *gorm.DB
}

// NewLeaveUsecase tạo instance LeaveUsecase
func NewLeaveUsecase(
	leaveRepo repository.LeaveRepository,
	correctionRepo repository.CorrectionRepository,
	overtimeRepo repository.OvertimeRepository,
	attendanceRepo repository.AttendanceRepository,
	userRepo repository.UserRepository,
	shiftRepo repository.ShiftRepository,
	db *gorm.DB,
) usecase.LeaveUsecase {
	return &leaveUsecase{
		leaveRepo:      leaveRepo,
		correctionRepo: correctionRepo,
		overtimeRepo:   overtimeRepo,
		attendanceRepo: attendanceRepo,
		userRepo:       userRepo,
		shiftRepo:      shiftRepo,
		db:             db,
	}
}

// Create tạo yêu cầu nghỉ phép
//
// Flow:
//  1. Validate input (description, leave_date, leave_type)
//  2. Xác định ngày quá khứ hay tương lai
//  3. Ngày quá khứ: validate attendance status (absent → full_day, half_day → half_day bù)
//  4. Ngày tương lai: validate leave_type hợp lệ
//  5. Kiểm tra chưa có yêu cầu trùng
//  6. Tạo yêu cầu
func (u *leaveUsecase) Create(ctx context.Context, req usecase.CreateLeaveRequest) (*entity.LeaveRequest, error) {
	logger := slog.With("user_id", req.UserID, "leave_date", req.LeaveDate)

	// 1. Validate description
	if req.Description == "" {
		return nil, apperrors.NewValidationError(map[string]string{
			"description": "Vui lòng nhập lý do xin nghỉ phép",
		})
	}

	// 2. Parse và validate ngày
	leaveDate, err := utils.ParseDateHCM(req.LeaveDate)
	if err != nil {
		return nil, apperrors.ErrLeaveInvalidDate
	}

	// Lấy thông tin user để có branch_id
	user, err := u.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	// Lấy shift để biết khung giờ chuẩn
	shift, _ := u.shiftRepo.FindDefault(ctx, *user.BranchID)
	if shift == nil {
		shift = &entity.Shift{StartTime: "08:00", EndTime: "17:00", WorkHours: 8}
	}

	now := utils.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, utils.HCM)
	isPassedDate := leaveDate.Before(today)

	var leaveType entity.LeaveType
	var timeFrom, timeTo string
	var originalStatus entity.AttendanceStatus

	if isPassedDate {
		// 3. Ngày quá khứ — kiểm tra attendance log
		attendLog, err := u.attendanceRepo.FindByUserAndDate(ctx, req.UserID, leaveDate)
		if err != nil {
			return nil, err
		}

		if attendLog == nil {
			// Không có log → vắng mặt cả ngày → nghỉ phép full day
			originalStatus = entity.StatusAbsent
			leaveType = entity.LeaveTypeFullDay
			timeFrom = shift.StartTime
			timeTo = shift.EndTime
		} else if attendLog.Status == entity.StatusHalfDay {
			// Nửa ngày → nghỉ phép cho nửa ngày còn lại
			originalStatus = entity.StatusHalfDay
			if attendLog.CheckInTime != nil {
				checkInHour := attendLog.CheckInTime.In(utils.HCM).Hour()
				if checkInHour < 12 {
					// Đã làm buổi sáng → nghỉ buổi chiều
					leaveType = entity.LeaveTypeHalfDayAfternoon
					timeFrom = "13:00"
					timeTo = shift.EndTime
				} else {
					// Đã làm buổi chiều → nghỉ buổi sáng
					leaveType = entity.LeaveTypeHalfDayMorning
					timeFrom = shift.StartTime
					timeTo = "12:00"
				}
			} else {
				leaveType = entity.LeaveTypeHalfDayAfternoon
				timeFrom = "13:00"
				timeTo = shift.EndTime
			}
		} else {
			// Ngày đã có log với status khác (present, late, ...) → không cho phép nghỉ phép
			logger.Warn("leave rejected - day already has valid attendance", "status", attendLog.Status)
			return nil, apperrors.ErrLeaveInvalidStatus
		}
	} else {
		// 4. Ngày hiện tại & tương lai — validate leave_type từ request
		if !isValidLeaveType(req.LeaveType) {
			return nil, apperrors.ErrLeaveInvalidType
		}
		leaveType = req.LeaveType

		switch leaveType {
		case entity.LeaveTypeFullDay:
			timeFrom = shift.StartTime
			timeTo = shift.EndTime
		case entity.LeaveTypeHalfDayMorning:
			timeFrom = shift.StartTime
			timeTo = "12:00"
		case entity.LeaveTypeHalfDayAfternoon:
			timeFrom = "13:00"
			timeTo = shift.EndTime
		}
	}

	// 5. Kiểm tra số ngày phép còn đủ
	deductDays := leaveCost(leaveType)
	if user.LeaveBalance < deductDays {
		logger.Warn("leave rejected - insufficient balance",
			"balance", user.LeaveBalance, "needed", deductDays)
		return nil, apperrors.ErrLeaveInsufficientBalance
	}

	// 6. Kiểm tra chưa có yêu cầu trùng
	existing, err := u.leaveRepo.FindByUserAndDate(ctx, req.UserID, leaveDate)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, apperrors.ErrLeaveAlreadyExists
	}

	// 6. Tạo yêu cầu
	leave := &entity.LeaveRequest{
		UserID:         req.UserID,
		BranchID:       *user.BranchID,
		LeaveDate:      leaveDate,
		LeaveType:      leaveType,
		TimeFrom:       timeFrom,
		TimeTo:         timeTo,
		OriginalStatus: originalStatus,
		Description:    req.Description,
		Status:         entity.LeaveStatusPending,
	}

	if err := u.leaveRepo.Create(ctx, leave); err != nil {
		return nil, err
	}

	logger.Info("leave request created",
		"leave_id", leave.ID,
		"leave_type", leaveType,
		"is_past_date", isPassedDate,
	)

	return leave, nil
}

// Process duyệt hoặc từ chối yêu cầu nghỉ phép
//
// Khi approved — chạy trong transaction:
//  1. Tạo/cập nhật AttendanceLog với status=leave
//  2. Ghi audit log: processed_by_id, processed_at, manager_note
func (u *leaveUsecase) Process(ctx context.Context, req usecase.ProcessLeaveRequest) (*entity.LeaveRequest, error) {
	logger := slog.With("leave_id", req.LeaveID, "processed_by", req.ProcessedByID)

	// Validate status input
	if req.Status != entity.LeaveStatusApproved && req.Status != entity.LeaveStatusRejected {
		return nil, apperrors.NewValidationError(map[string]string{
			"status": "Trạng thái phải là approved hoặc rejected",
		})
	}

	// Tìm leave request
	leave, err := u.leaveRepo.FindByID(ctx, req.LeaveID)
	if err != nil {
		return nil, err
	}

	// Chỉ xử lý yêu cầu PENDING
	if !leave.IsPending() {
		return nil, apperrors.ErrLeaveNotPending
	}

	// Manager không được tự duyệt cho chính mình
	if leave.UserID == req.ProcessedByID {
		logger.Warn("self-approval attempt blocked")
		return nil, apperrors.ErrLeaveSelfApprove
	}

	now := utils.Now()

	if req.Status == entity.LeaveStatusApproved {
		// Lấy shift để tính giờ
		shift, _ := u.shiftRepo.FindDefault(ctx, leave.BranchID)
		if shift == nil {
			shift = &entity.Shift{StartTime: "08:00", EndTime: "17:00", WorkHours: 8}
		}

		// Tính số ngày phép cần trừ
		cost := leaveCost(leave.LeaveType)

		// Kiểm tra số ngày phép còn đủ trước khi duyệt
		leaveUser, err := u.userRepo.FindByID(ctx, leave.UserID)
		if err != nil {
			return nil, err
		}
		if leaveUser.LeaveBalance < cost {
			return nil, apperrors.ErrLeaveInsufficientBalance
		}

		// Transaction: tạo/cập nhật attendance_log + trừ phép + cập nhật leave request
		err = u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			// Trừ số ngày phép
			if err := tx.Model(&entity.User{}).
				Where("id = ? AND leave_balance >= ?", leave.UserID, cost).
				Update("leave_balance", gorm.Expr("leave_balance - ?", cost)).Error; err != nil {
				return err
			}

			// Tìm attendance log hiện có cho ngày này
			var existingLog entity.AttendanceLog
			findErr := tx.Where("user_id = ? AND date = ?", leave.UserID, leave.LeaveDate).
				First(&existingLog).Error

			fromH, fromM := parseTime(leave.TimeFrom)
			toH, toM := parseTime(leave.TimeTo)
			checkIn := time.Date(leave.LeaveDate.Year(), leave.LeaveDate.Month(), leave.LeaveDate.Day(), fromH, fromM, 0, 0, utils.HCM)
			checkOut := time.Date(leave.LeaveDate.Year(), leave.LeaveDate.Month(), leave.LeaveDate.Day(), toH, toM, 0, 0, utils.HCM)

			if findErr != nil && findErr == gorm.ErrRecordNotFound {
				// Không có log (vắng mặt hoặc ngày tương lai) → tạo mới với status=leave
				newLog := entity.AttendanceLog{
					UserID:       leave.UserID,
					BranchID:     leave.BranchID,
					Date:         leave.LeaveDate,
					CheckInTime:  &checkIn,
					CheckOutTime: &checkOut,
					Status:       entity.StatusLeave,
					WorkHours:    shift.WorkHours,
					Note:         "Nghỉ phép - " + leave.Description,
				}
				if shift.ID != 0 {
					newLog.ShiftID = &shift.ID
				}
				if err := tx.Create(&newLog).Error; err != nil {
					return err
				}
			} else if findErr == nil {
				// Có log (half_day) → cập nhật status thành half_day_leave, work_hours = full shift
				updates := map[string]interface{}{
					"status":     entity.StatusHalfDayLeave,
					"work_hours": shift.WorkHours,
					"note":       "Nghỉ phép nửa ngày - " + leave.Description,
				}
				if err := tx.Model(&entity.AttendanceLog{}).
					Where("id = ?", existingLog.ID).
					Updates(updates).Error; err != nil {
					return err
				}
			} else {
				return findErr
			}

			// Ghi audit log và cập nhật leave request
			leave.Status = entity.LeaveStatusApproved
			leave.ProcessedByID = &req.ProcessedByID
			leave.ProcessedAt = &now
			leave.ManagerNote = req.ManagerNote

			return tx.Save(leave).Error
		})
		if err != nil {
			slog.Error("leave approval transaction failed", "error", err)
			return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi duyệt yêu cầu nghỉ phép")
		}

		logger.Info("leave approved",
			"leave_date", leave.LeaveDate.Format("2006-01-02"),
			"leave_type", leave.LeaveType,
		)
	} else {
		// Rejected
		leave.Status = entity.LeaveStatusRejected
		leave.ProcessedByID = &req.ProcessedByID
		leave.ProcessedAt = &now
		leave.ManagerNote = req.ManagerNote

		if err := u.leaveRepo.Update(ctx, leave); err != nil {
			return nil, err
		}

		logger.Info("leave rejected", "leave_date", leave.LeaveDate.Format("2006-01-02"))
	}

	// Reload để có đầy đủ relations
	return u.leaveRepo.FindByID(ctx, leave.ID)
}

func (u *leaveUsecase) GetByID(ctx context.Context, id uint) (*entity.LeaveRequest, error) {
	return u.leaveRepo.FindByID(ctx, id)
}

func (u *leaveUsecase) GetList(ctx context.Context, filter repository.LeaveFilter) ([]*entity.LeaveRequest, int64, error) {
	return u.leaveRepo.FindAll(ctx, filter)
}

func (u *leaveUsecase) GetMyList(ctx context.Context, userID uint, status entity.LeaveStatus, page, limit int) ([]*entity.LeaveRequest, int64, error) {
	filter := repository.LeaveFilter{
		UserID: &userID,
		Status: status,
		Page:   page,
		Limit:  limit,
	}
	return u.leaveRepo.FindAll(ctx, filter)
}

// GetPendingApprovals lấy danh sách tổng hợp chờ duyệt từ corrections, leave requests và overtime
func (u *leaveUsecase) GetPendingApprovals(ctx context.Context, branchID *uint, page, limit int) ([]usecase.PendingApprovalItem, int64, error) {
	// Lấy pending corrections
	corrFilter := repository.CorrectionFilter{
		BranchID: branchID,
		Status:   entity.CorrectionStatusPending,
		Page:     1,
		Limit:    1000,
	}
	corrections, corrTotal, err := u.correctionRepo.FindAll(ctx, corrFilter)
	if err != nil {
		return nil, 0, err
	}

	// Lấy pending leaves
	leaveFilter := repository.LeaveFilter{
		BranchID: branchID,
		Status:   entity.LeaveStatusPending,
		Page:     1,
		Limit:    1000,
	}
	leaves, leaveTotal, err := u.leaveRepo.FindAll(ctx, leaveFilter)
	if err != nil {
		return nil, 0, err
	}

	// Lấy pending overtimes
	otFilter := repository.OvertimeFilter{
		BranchID: branchID,
		Status:   entity.OvertimeStatusPending,
		Page:     1,
		Limit:    1000,
	}
	overtimes, otTotal, err := u.overtimeRepo.FindAll(ctx, otFilter)
	if err != nil {
		return nil, 0, err
	}

	total := corrTotal + leaveTotal + otTotal

	items := make([]usecase.PendingApprovalItem, 0, len(corrections)+len(leaves)+len(overtimes))

	for _, c := range corrections {
		detail := string(c.OriginalStatus)
		date := ""
		if c.AttendanceLog.ID != 0 {
			date = c.AttendanceLog.Date.Format("2006-01-02")
		} else if c.OvertimeRequest != nil && c.OvertimeRequest.ID != 0 {
			date = c.OvertimeRequest.Date.Format("2006-01-02")
		}
		items = append(items, usecase.PendingApprovalItem{
			ID:           c.ID,
			Type:         "correction",
			UserID:       c.UserID,
			UserName:     c.User.Name,
			EmployeeCode: c.User.EmployeeCode,
			Department:   c.User.Department,
			BranchID:     c.BranchID,
			Date:         date,
			Description:  c.Description,
			Detail:       detail,
			CreatedAt:    c.CreatedAt.Format(time.RFC3339),
		})
	}

	for _, l := range leaves {
		items = append(items, usecase.PendingApprovalItem{
			ID:           l.ID,
			Type:         "leave",
			UserID:       l.UserID,
			UserName:     l.User.Name,
			EmployeeCode: l.User.EmployeeCode,
			Department:   l.User.Department,
			BranchID:     l.BranchID,
			Date:         l.LeaveDate.Format("2006-01-02"),
			Description:  l.Description,
			Detail:       string(l.LeaveType),
			CreatedAt:    l.CreatedAt.Format(time.RFC3339),
		})
	}

	items = appendOvertimeItems(items, overtimes)

	// Sort by created_at DESC
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt > items[j].CreatedAt
	})

	// Paginate
	start := (page - 1) * limit
	if start >= len(items) {
		return []usecase.PendingApprovalItem{}, total, nil
	}
	end := start + limit
	if end > len(items) {
		end = len(items)
	}

	return items[start:end], total, nil
}

// GetApprovals lấy danh sách tổng hợp duyệt chấm công (correction + leave + overtime) — hỗ trợ lọc theo status
func (u *leaveUsecase) GetApprovals(ctx context.Context, branchID *uint, status string, page, limit int) ([]usecase.PendingApprovalItem, int64, error) {
	// Build filters
	corrFilter := repository.CorrectionFilter{
		BranchID: branchID,
		Page:     1,
		Limit:    1000,
	}
	leaveFilter := repository.LeaveFilter{
		BranchID: branchID,
		Page:     1,
		Limit:    1000,
	}
	otFilter := repository.OvertimeFilter{
		BranchID: branchID,
		Page:     1,
		Limit:    1000,
	}
	if status != "" {
		corrFilter.Status = entity.CorrectionStatus(status)
		leaveFilter.Status = entity.LeaveStatus(status)
		otFilter.Status = entity.OvertimeStatus(status)
	}

	corrections, corrTotal, err := u.correctionRepo.FindAll(ctx, corrFilter)
	if err != nil {
		return nil, 0, err
	}

	leaves, leaveTotal, err := u.leaveRepo.FindAll(ctx, leaveFilter)
	if err != nil {
		return nil, 0, err
	}

	overtimes, otTotal, err := u.overtimeRepo.FindAll(ctx, otFilter)
	if err != nil {
		return nil, 0, err
	}

	total := corrTotal + leaveTotal + otTotal

	items := make([]usecase.PendingApprovalItem, 0, len(corrections)+len(leaves)+len(overtimes))

	for _, c := range corrections {
		detail := string(c.OriginalStatus)
		date := ""
		var checkIn, checkOut *string
		if c.AttendanceLog.ID != 0 {
			date = c.AttendanceLog.Date.Format("2006-01-02")
			if c.AttendanceLog.CheckInTime != nil {
				s := c.AttendanceLog.CheckInTime.In(utils.HCM).Format("15:04")
				checkIn = &s
			}
			if c.AttendanceLog.CheckOutTime != nil {
				s := c.AttendanceLog.CheckOutTime.In(utils.HCM).Format("15:04")
				checkOut = &s
			}
		} else if c.OvertimeRequest != nil && c.OvertimeRequest.ID != 0 {
			// Overtime correction — lấy date và times từ OvertimeRequest
			date = c.OvertimeRequest.Date.Format("2006-01-02")
			if c.OvertimeRequest.ActualCheckin != nil {
				s := c.OvertimeRequest.ActualCheckin.In(utils.HCM).Format("15:04")
				checkIn = &s
			}
			if c.OvertimeRequest.ActualCheckout != nil {
				s := c.OvertimeRequest.ActualCheckout.In(utils.HCM).Format("15:04")
				checkOut = &s
			}
		}

		item := usecase.PendingApprovalItem{
			ID:           c.ID,
			Type:         "correction",
			UserID:       c.UserID,
			UserName:     c.User.Name,
			EmployeeCode: c.User.EmployeeCode,
			Department:   c.User.Department,
			BranchID:     c.BranchID,
			Date:         date,
			Description:  c.Description,
			Detail:       detail,
			Status:       string(c.Status),
			CreatedAt:    c.CreatedAt.Format(time.RFC3339),
			CheckInTime:  checkIn,
			CheckOutTime: checkOut,
		}
		if c.ProcessedBy != nil {
			item.ProcessedByName = c.ProcessedBy.Name
		}
		if c.ProcessedAt != nil {
			s := c.ProcessedAt.Format(time.RFC3339)
			item.ProcessedAt = &s
		}
		item.ManagerNote = c.ManagerNote
		items = append(items, item)
	}

	for _, l := range leaves {
		item := usecase.PendingApprovalItem{
			ID:           l.ID,
			Type:         "leave",
			UserID:       l.UserID,
			UserName:     l.User.Name,
			EmployeeCode: l.User.EmployeeCode,
			Department:   l.User.Department,
			BranchID:     l.BranchID,
			Date:         l.LeaveDate.Format("2006-01-02"),
			Description:  l.Description,
			Detail:       string(l.LeaveType),
			Status:       string(l.Status),
			CreatedAt:    l.CreatedAt.Format(time.RFC3339),
			LeaveType:    string(l.LeaveType),
			TimeFrom:     l.TimeFrom,
			TimeTo:       l.TimeTo,
		}
		if l.ProcessedBy != nil {
			item.ProcessedByName = l.ProcessedBy.Name
		}
		if l.ProcessedAt != nil {
			s := l.ProcessedAt.Format(time.RFC3339)
			item.ProcessedAt = &s
		}
		item.ManagerNote = l.ManagerNote
		items = append(items, item)
	}

	items = appendOvertimeItems(items, overtimes)

	// Sort by created_at DESC
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt > items[j].CreatedAt
	})

	// Paginate
	start := (page - 1) * limit
	if start >= len(items) {
		return []usecase.PendingApprovalItem{}, total, nil
	}
	end := start + limit
	if end > len(items) {
		end = len(items)
	}

	return items[start:end], total, nil
}

// AutoRejectExpired tự động reject yêu cầu PENDING của tháng trước
func (u *leaveUsecase) AutoRejectExpired(ctx context.Context) (int64, error) {
	now := utils.Now()
	note := "Hệ thống tự động từ chối do quá hạn chốt lương"

	count, err := u.leaveRepo.AutoRejectExpired(ctx, now, note)
	if err != nil {
		slog.Error("auto-reject expired leaves failed", "error", err)
		return 0, err
	}

	if count > 0 {
		slog.Info("auto-reject expired leaves completed", "rejected_count", count)
	}

	return count, nil
}

// BatchApprove duyệt tất cả yêu cầu nghỉ phép PENDING — gọi Process tuần tự
func (u *leaveUsecase) BatchApprove(ctx context.Context, processedByID uint, branchID *uint) (int64, error) {
	filter := repository.LeaveFilter{
		BranchID: branchID,
		Status:   entity.LeaveStatusPending,
		Page:     1,
		Limit:    1000,
	}
	leaves, _, err := u.leaveRepo.FindAll(ctx, filter)
	if err != nil {
		return 0, err
	}

	var approved int64
	for _, l := range leaves {
		if l.UserID == processedByID {
			continue
		}
		req := usecase.ProcessLeaveRequest{
			LeaveID:       l.ID,
			ProcessedByID: processedByID,
			Status:        entity.LeaveStatusApproved,
			ManagerNote:   "Duyệt hàng loạt",
		}
		if _, err := u.Process(ctx, req); err != nil {
			slog.Warn("batch approve leave skipped", "id", l.ID, "error", err)
			continue
		}
		approved++
	}

	slog.Info("batch approve leaves completed", "approved", approved, "total_pending", len(leaves))
	return approved, nil
}

// leaveCost trả về số ngày phép cần trừ theo loại nghỉ
func leaveCost(lt entity.LeaveType) float64 {
	if lt == entity.LeaveTypeFullDay {
		return 1.0
	}
	return 0.5 // half_day_morning hoặc half_day_afternoon
}

// AccrueMonthlyLeave cộng 1 ngày phép cho tất cả user active
func (u *leaveUsecase) AccrueMonthlyLeave(ctx context.Context) (int64, error) {
	count, err := u.userRepo.AccrueLeaveBalance(ctx, 1.0)
	if err != nil {
		slog.Error("monthly leave accrual failed", "error", err)
		return 0, err
	}

	if count > 0 {
		slog.Info("monthly leave accrual completed", "users_updated", count)
	}

	return count, nil
}

func isValidLeaveType(lt entity.LeaveType) bool {
	return lt == entity.LeaveTypeFullDay ||
		lt == entity.LeaveTypeHalfDayMorning ||
		lt == entity.LeaveTypeHalfDayAfternoon
}

// parseTime parse "HH:MM" → (hour, minute)
func parseTime(t string) (int, int) {
	if len(t) != 5 {
		return 8, 0
	}
	hour, _ := strconv.Atoi(t[:2])
	min, _ := strconv.Atoi(t[3:])
	return hour, min
}

// appendOvertimeItems chuyển đổi danh sách OvertimeRequest thành PendingApprovalItem và append vào items
func appendOvertimeItems(items []usecase.PendingApprovalItem, overtimes []*entity.OvertimeRequest) []usecase.PendingApprovalItem {
	for _, ot := range overtimes {
		item := usecase.PendingApprovalItem{
			ID:           ot.ID,
			Type:         "overtime",
			UserID:       ot.UserID,
			UserName:     ot.User.Name,
			EmployeeCode: ot.User.EmployeeCode,
			Department:   ot.User.Department,
			BranchID:     ot.BranchID,
			Date:         ot.Date.Format("2006-01-02"),
			Description:  "Tăng ca",
			Detail:       "overtime",
			Status:       string(ot.Status),
			CreatedAt:    ot.CreatedAt.Format(time.RFC3339),
		}
		if ot.ActualCheckin != nil {
			s := ot.ActualCheckin.In(utils.HCM).Format("15:04")
			item.ActualCheckin = &s
		}
		if ot.ActualCheckout != nil {
			s := ot.ActualCheckout.In(utils.HCM).Format("15:04")
			item.ActualCheckout = &s
		}
		if ot.CalculatedStart != nil {
			s := ot.CalculatedStart.In(utils.HCM).Format("15:04")
			item.CalculatedStart = &s
		}
		if ot.CalculatedEnd != nil {
			s := ot.CalculatedEnd.In(utils.HCM).Format("15:04")
			item.CalculatedEnd = &s
		}
		if ot.TotalHours > 0 {
			h := ot.TotalHours
			item.TotalHours = &h
		}
		if ot.ProcessedBy != nil {
			item.ProcessedByName = ot.ProcessedBy.Name
		}
		if ot.ProcessedAt != nil {
			s := ot.ProcessedAt.Format(time.RFC3339)
			item.ProcessedAt = &s
		}
		item.ManagerNote = ot.ManagerNote
		items = append(items, item)
	}
	return items
}
