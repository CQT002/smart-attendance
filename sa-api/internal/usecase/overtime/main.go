package overtime

import (
	"context"
	"log/slog"
	"time"

	"github.com/hdbank/smart-attendance/config"
	"github.com/hdbank/smart-attendance/internal/domain/entity"
	"github.com/hdbank/smart-attendance/internal/domain/repository"
	"github.com/hdbank/smart-attendance/internal/domain/usecase"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"github.com/hdbank/smart-attendance/pkg/utils"
	"gorm.io/gorm"
)

type overtimeUsecase struct {
	overtimeRepo   repository.OvertimeRepository
	userRepo       repository.UserRepository
	attendanceRepo repository.AttendanceRepository
	shiftRepo      repository.ShiftRepository
	db             *gorm.DB
	maxHoursPerDay float64
}

// NewOvertimeUsecase tạo instance OvertimeUsecase
func NewOvertimeUsecase(
	overtimeRepo repository.OvertimeRepository,
	userRepo repository.UserRepository,
	attendanceRepo repository.AttendanceRepository,
	shiftRepo repository.ShiftRepository,
	db *gorm.DB,
	otCfg config.OvertimeConfig,
) usecase.OvertimeUsecase {
	maxHours := otCfg.MaxHoursPerDay
	if maxHours <= 0 {
		maxHours = 4
	}
	return &overtimeUsecase{
		overtimeRepo:   overtimeRepo,
		userRepo:       userRepo,
		attendanceRepo: attendanceRepo,
		shiftRepo:      shiftRepo,
		db:             db,
		maxHoursPerDay: maxHours,
	}
}

// otConfig chứa thông số OT lấy từ shift, có fallback mặc định
type otConfig struct {
	minCheckInHour int
	startHour      int
	endHour        int
}

// getOTConfig lấy cấu hình OT từ shift của branch, fallback về default nếu không có
func (u *overtimeUsecase) getOTConfig(ctx context.Context, branchID uint) otConfig {
	cfg := otConfig{minCheckInHour: 17, startHour: 18, endHour: 22}
	shift, err := u.shiftRepo.FindDefault(ctx, branchID)
	if err != nil || shift == nil {
		return cfg
	}
	if shift.OTMinCheckInHour > 0 {
		cfg.minCheckInHour = shift.OTMinCheckInHour
	}
	if shift.OTStartHour > 0 {
		cfg.startHour = shift.OTStartHour
	}
	if shift.OTEndHour > 0 {
		cfg.endHour = shift.OTEndHour
	}
	return cfg
}

// CheckIn check-in tăng ca
//
// Flow:
//  1. Validate thời gian >= OTMinCheckInHour (từ shift)
//  2. Kiểm tra chưa có OT request cho ngày này
//  3. Tạo OvertimeRequest với actual_checkin
//  4. Trả về thông tin dự kiến bo tròn
func (u *overtimeUsecase) CheckIn(ctx context.Context, req usecase.OvertimeCheckInRequest) (*usecase.OvertimeCheckInResponse, error) {
	logger := slog.With("user_id", req.UserID)

	now := utils.Now()

	// Lấy thông tin user để có branch_id
	user, err := u.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if user.BranchID == nil {
		slog.Error("overtime check-in failed - user has no branch", "user_id", req.UserID)
		return nil, apperrors.ErrForbidden
	}

	// Lấy OT config từ shift
	ot := u.getOTConfig(ctx, *user.BranchID)

	// 1. Validate thời gian >= OTMinCheckInHour
	if now.Hour() < ot.minCheckInHour {
		logger.Warn("overtime check-in rejected - too early", "hour", now.Hour(), "min_hour", ot.minCheckInHour)
		return nil, apperrors.ErrOvertimeCheckInTooEarly
	}

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, utils.HCM)

	// 2. Kiểm tra chưa có OT request cho ngày này
	existing, err := u.overtimeRepo.FindByUserAndDate(ctx, req.UserID, today)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if existing.IsCheckedIn() {
			logger.Warn("overtime already checked in", "existing_id", existing.ID)
			return nil, apperrors.ErrOvertimeAlreadyCheckedIn
		}
		logger.Warn("overtime request already exists", "existing_id", existing.ID)
		return nil, apperrors.ErrOvertimeAlreadyExists
	}

	// 3. Tạo OvertimeRequest
	otReq := &entity.OvertimeRequest{
		UserID:        req.UserID,
		BranchID:      *user.BranchID,
		Date:          today,
		ActualCheckin: &now,
		Status:        entity.OvertimeStatusInit,
	}

	if err := u.overtimeRepo.Create(ctx, otReq); err != nil {
		slog.Error("failed to create overtime request", "user_id", req.UserID, "error", err)
		return nil, err
	}

	logger.Info("overtime check-in successful",
		"overtime_id", otReq.ID,
		"actual_checkin", now.Format("15:04:05"),
	)

	// 4. Tính thời gian dự kiến
	estimatedStart := clampStart(now, today, ot.startHour)
	estimatedEnd := time.Date(today.Year(), today.Month(), today.Day(), ot.endHour, 0, 0, 0, utils.HCM)

	return &usecase.OvertimeCheckInResponse{
		OvertimeRequest: otReq,
		EstimatedStart:  &estimatedStart,
		EstimatedEnd:    &estimatedEnd,
		Note:            "Lưu ý: Giờ tăng ca chỉ bắt đầu tính từ 18:00 đến 22:00 theo quy định",
	}, nil
}

// CheckOut check-out tăng ca
//
// Hai kịch bản:
//  1. Đã check-in (có OT record init) → cập nhật checkout, chuyển status → pending
//  2. Chưa check-in (quên check-in, chỉ check-out) → tạo OT record mới với chỉ checkout, status = init
//     Trường hợp 2 cần chấm công bù để bổ sung check-in
func (u *overtimeUsecase) CheckOut(ctx context.Context, req usecase.OvertimeCheckOutRequest) (*usecase.OvertimeCheckOutResponse, error) {
	logger := slog.With("user_id", req.UserID, "overtime_id", req.OvertimeID)

	now := utils.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, utils.HCM)

	// Tìm OT request active
	var otReq *entity.OvertimeRequest
	var err error

	if req.OvertimeID > 0 {
		otReq, err = u.overtimeRepo.FindByID(ctx, req.OvertimeID)
		if err != nil {
			return nil, err
		}
		if otReq.UserID != req.UserID {
			logger.Warn("overtime check-out forbidden - user mismatch", "ot_user_id", otReq.UserID)
			return nil, apperrors.ErrForbidden
		}
	} else {
		// Tìm OT init (đã check-in) cho hôm nay
		otReq, err = u.overtimeRepo.FindActiveByUserAndDate(ctx, req.UserID, today)
		if err != nil {
			return nil, err
		}
	}

	if otReq != nil && otReq.IsCheckedOut() {
		logger.Warn("overtime already checked out", "overtime_id", otReq.ID)
		return nil, apperrors.ErrOvertimeAlreadyCheckedOut
	}

	// Lấy branchID để query OT config
	var branchID uint
	if otReq != nil {
		branchID = otReq.BranchID
	} else {
		user, err := u.userRepo.FindByID(ctx, req.UserID)
		if err != nil {
			return nil, err
		}
		if user.BranchID == nil {
			slog.Error("overtime check-out failed - user has no branch", "user_id", req.UserID)
			return nil, apperrors.ErrForbidden
		}
		branchID = *user.BranchID
	}

	ot := u.getOTConfig(ctx, branchID)

	// Validate thời gian >= OTStartHour (18:00)
	if now.Hour() < ot.startHour {
		logger.Warn("overtime check-out rejected - too early", "hour", now.Hour(), "start_hour", ot.startHour)
		return nil, apperrors.ErrOvertimeCheckOutTooEarly
	}

	if otReq != nil && otReq.IsCheckedIn() {
		// Kịch bản 1: Đã check-in → cập nhật checkout, chuyển pending
		otReq.ActualCheckout = &now
		otReq.Status = entity.OvertimeStatusPending

		if err := u.overtimeRepo.Update(ctx, otReq); err != nil {
			slog.Error("failed to update overtime checkout", "overtime_id", otReq.ID, "error", err)
			return nil, err
		}
	} else {
		if otReq != nil {
			logger.Warn("overtime request already exists", "overtime_id", otReq.ID)
			return nil, apperrors.ErrOvertimeAlreadyExists
		}
		// Kịch bản 2: Chưa check-in → tạo mới chỉ có checkout, status init
		existing, err := u.overtimeRepo.FindByUserAndDate(ctx, req.UserID, today)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			logger.Warn("overtime request already exists for today", "existing_id", existing.ID)
			return nil, apperrors.ErrOvertimeAlreadyExists
		}

		otReq = &entity.OvertimeRequest{
			UserID:         req.UserID,
			BranchID:       branchID,
			Date:           today,
			ActualCheckout: &now,
			Status:         entity.OvertimeStatusInit, // init — cần bù check-in
		}
		if err := u.overtimeRepo.Create(ctx, otReq); err != nil {
			slog.Error("failed to create overtime request (checkout only)", "user_id", req.UserID, "error", err)
			return nil, err
		}
	}

	logger.Info("overtime check-out successful",
		"overtime_id", otReq.ID,
		"actual_checkout", now.Format("15:04:05"),
		"has_checkin", otReq.IsCheckedIn(),
	)

	// Tính thời gian dự kiến
	var estimatedStart, estimatedEnd time.Time
	var estimatedHours float64

	if otReq.IsCheckedIn() && otReq.IsCheckedOut() {
		estimatedStart = clampStart(*otReq.ActualCheckin, otReq.Date, ot.startHour)
		estimatedEnd = clampEnd(now, otReq.Date, ot.endHour)
		estimatedHours = estimatedEnd.Sub(estimatedStart).Hours()
		if estimatedHours < 0 {
			estimatedHours = 0
		}
		if estimatedHours > u.maxHoursPerDay {
			estimatedHours = u.maxHoursPerDay
		}
	} else {
		// Chỉ có checkout, chưa có checkin → dự kiến từ OTStartHour
		estimatedStart = time.Date(today.Year(), today.Month(), today.Day(), ot.startHour, 0, 0, 0, utils.HCM)
		estimatedEnd = clampEnd(now, otReq.Date, ot.endHour)
		estimatedHours = estimatedEnd.Sub(estimatedStart).Hours()
		if estimatedHours < 0 {
			estimatedHours = 0
		}
		if estimatedHours > u.maxHoursPerDay {
			estimatedHours = u.maxHoursPerDay
		}
	}

	note := "Lưu ý: Giờ tăng ca chỉ bắt đầu tính từ 18:00 đến 22:00 theo quy định"
	if !otReq.IsCheckedIn() {
		note = "Bạn chưa check-in tăng ca. Vui lòng tạo yêu cầu chấm công bù để bổ sung check-in."
	}

	return &usecase.OvertimeCheckOutResponse{
		OvertimeRequest: otReq,
		EstimatedStart:  &estimatedStart,
		EstimatedEnd:    &estimatedEnd,
		EstimatedHours:  float64(int(estimatedHours*100)) / 100,
		Note:            note,
	}, nil
}

// Process duyệt hoặc từ chối yêu cầu tăng ca
//
// Khi approved — chạy trong transaction:
//  1. Tính calculated_start = Max(actual_checkin, OTStartHour)
//  2. Tính calculated_end = Min(actual_checkout, OTEndHour)
//  3. Tính total_hours = calculated_end - calculated_start
//  4. Ghi audit log
func (u *overtimeUsecase) Process(ctx context.Context, req usecase.ProcessOvertimeRequest) (*entity.OvertimeRequest, error) {
	logger := slog.With("overtime_id", req.OvertimeID, "processed_by", req.ProcessedByID)

	// Validate status input
	if req.Status != entity.OvertimeStatusApproved && req.Status != entity.OvertimeStatusRejected {
		return nil, apperrors.NewValidationError(map[string]string{
			"status": "Trạng thái phải là approved hoặc rejected",
		})
	}

	otReq, err := u.overtimeRepo.FindByID(ctx, req.OvertimeID)
	if err != nil {
		return nil, err
	}

	if !otReq.IsPending() {
		logger.Warn("overtime not in pending status", "current_status", otReq.Status)
		return nil, apperrors.ErrOvertimeNotPending
	}

	// Manager không được tự duyệt cho chính mình
	if otReq.UserID == req.ProcessedByID {
		logger.Warn("self-approval attempt blocked")
		return nil, apperrors.ErrOvertimeSelfApprove
	}

	// Check-out phải hoàn tất trước khi duyệt
	if !otReq.IsCheckedOut() {
		logger.Warn("overtime not completed - missing checkout")
		return nil, apperrors.ErrOvertimeNotCompleted
	}

	now := utils.Now()

	if req.Status == entity.OvertimeStatusApproved {
		// Lấy OT config từ shift của branch
		ot := u.getOTConfig(ctx, otReq.BranchID)

		// Transaction: tính toán và cập nhật
		err = u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			// Tính bo tròn
			calcStart := clampStart(*otReq.ActualCheckin, otReq.Date, ot.startHour)
			calcEnd := clampEnd(*otReq.ActualCheckout, otReq.Date, ot.endHour)
			totalHours := calcEnd.Sub(calcStart).Hours()
			if totalHours < 0 {
				totalHours = 0
			}
			if totalHours > u.maxHoursPerDay {
				totalHours = u.maxHoursPerDay
			}
			totalHours = float64(int(totalHours*100)) / 100

			otReq.CalculatedStart = &calcStart
			otReq.CalculatedEnd = &calcEnd
			otReq.TotalHours = totalHours
			otReq.Status = entity.OvertimeStatusApproved
			otReq.ManagerID = &req.ProcessedByID
			otReq.ProcessedAt = &now
			otReq.ManagerNote = req.ManagerNote

			if err := tx.Save(otReq).Error; err != nil {
				return err
			}

			// Link OT vào attendance_log nếu có
			today := otReq.Date
			var attendLog entity.AttendanceLog
			if findErr := tx.Where("user_id = ? AND date = ?", otReq.UserID, today).
				First(&attendLog).Error; findErr == nil {
				tx.Model(&entity.AttendanceLog{}).
					Where("id = ?", attendLog.ID).
					Update("overtime_request_id", otReq.ID)
			}

			return nil
		})
		if err != nil {
			slog.Error("overtime approval transaction failed", "error", err)
			return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi duyệt yêu cầu tăng ca")
		}

		logger.Info("overtime approved",
			"total_hours", otReq.TotalHours,
			"calculated_start", otReq.CalculatedStart,
			"calculated_end", otReq.CalculatedEnd,
		)
	} else {
		// Rejected
		otReq.Status = entity.OvertimeStatusRejected
		otReq.ManagerID = &req.ProcessedByID
		otReq.ProcessedAt = &now
		otReq.ManagerNote = req.ManagerNote

		if err := u.overtimeRepo.Update(ctx, otReq); err != nil {
			slog.Error("failed to update overtime rejection", "overtime_id", otReq.ID, "error", err)
			return nil, err
		}

		logger.Info("overtime rejected", "overtime_id", otReq.ID)
	}

	return u.overtimeRepo.FindByID(ctx, otReq.ID)
}

func (u *overtimeUsecase) GetByID(ctx context.Context, id uint) (*entity.OvertimeRequest, error) {
	return u.overtimeRepo.FindByID(ctx, id)
}

func (u *overtimeUsecase) GetList(ctx context.Context, filter repository.OvertimeFilter) ([]*entity.OvertimeRequest, int64, error) {
	return u.overtimeRepo.FindAll(ctx, filter)
}

func (u *overtimeUsecase) GetMyList(ctx context.Context, userID uint, status entity.OvertimeStatus, page, limit int) ([]*entity.OvertimeRequest, int64, error) {
	filter := repository.OvertimeFilter{
		UserID: &userID,
		Status: status,
		Page:   page,
		Limit:  limit,
	}
	return u.overtimeRepo.FindAll(ctx, filter)
}

func (u *overtimeUsecase) GetMyToday(ctx context.Context, userID uint) (*entity.OvertimeRequest, error) {
	now := utils.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, utils.HCM)
	otReq, err := u.overtimeRepo.FindByUserAndDate(ctx, userID, today)
	if err != nil {
		return nil, err
	}
	if otReq == nil {
		return nil, apperrors.ErrNotFound
	}
	return otReq, nil
}

// BatchApprove duyệt tất cả yêu cầu PENDING
func (u *overtimeUsecase) BatchApprove(ctx context.Context, processedByID uint, branchID *uint) (int64, error) {
	filter := repository.OvertimeFilter{
		BranchID: branchID,
		Status:   entity.OvertimeStatusPending,
		Page:     1,
		Limit:    1000,
	}
	overtimes, _, err := u.overtimeRepo.FindAll(ctx, filter)
	if err != nil {
		return 0, err
	}

	var approved int64
	for _, ot := range overtimes {
		if ot.UserID == processedByID {
			continue
		}
		if !ot.IsCheckedOut() {
			continue
		}
		req := usecase.ProcessOvertimeRequest{
			OvertimeID:    ot.ID,
			ProcessedByID: processedByID,
			Status:        entity.OvertimeStatusApproved,
			ManagerNote:   "Duyệt hàng loạt",
		}
		if _, err := u.Process(ctx, req); err != nil {
			slog.Warn("batch approve overtime skipped", "id", ot.ID, "error", err)
			continue
		}
		approved++
	}

	slog.Info("batch approve overtimes completed", "approved", approved, "total_pending", len(overtimes))
	return approved, nil
}

// AutoRejectExpired tự động reject yêu cầu PENDING của tháng trước
func (u *overtimeUsecase) AutoRejectExpired(ctx context.Context) (int64, error) {
	now := utils.Now()
	note := "Hệ thống tự động từ chối do quá hạn chốt lương"

	count, err := u.overtimeRepo.AutoRejectExpired(ctx, now, note)
	if err != nil {
		slog.Error("auto-reject expired overtimes failed", "error", err)
		return 0, err
	}

	if count > 0 {
		slog.Info("auto-reject expired overtimes completed", "rejected_count", count)
	}

	return count, nil
}

// clampStart trả về Max(actualCheckin, startHour:00 ngày date)
func clampStart(actualCheckin time.Time, date time.Time, startHour int) time.Time {
	otStart := time.Date(date.Year(), date.Month(), date.Day(), startHour, 0, 0, 0, utils.HCM)
	if actualCheckin.Before(otStart) {
		return otStart
	}
	return actualCheckin
}

// clampEnd trả về Min(actualCheckout, endHour:00 ngày date)
func clampEnd(actualCheckout time.Time, date time.Time, endHour int) time.Time {
	otEnd := time.Date(date.Year(), date.Month(), date.Day(), endHour, 0, 0, 0, utils.HCM)
	if actualCheckout.After(otEnd) {
		return otEnd
	}
	return actualCheckout
}
