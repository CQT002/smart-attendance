package overtime

import (
	"context"
	"log/slog"
	"time"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	"github.com/hdbank/smart-attendance/internal/domain/repository"
	"github.com/hdbank/smart-attendance/internal/domain/usecase"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"github.com/hdbank/smart-attendance/pkg/utils"
	"gorm.io/gorm"
)

// OT time boundaries (HCM timezone)
const (
	otCheckInMinHour  = 17 // Chỉ check-in sau 17:00
	otStartHour       = 18 // Giờ bắt đầu tính OT
	otEndHour         = 22 // Giờ kết thúc tính OT
	otMaxHours        = 4  // Tối đa 4 giờ/ngày
)

type overtimeUsecase struct {
	overtimeRepo   repository.OvertimeRepository
	userRepo       repository.UserRepository
	attendanceRepo repository.AttendanceRepository
	db             *gorm.DB
}

// NewOvertimeUsecase tạo instance OvertimeUsecase
func NewOvertimeUsecase(
	overtimeRepo repository.OvertimeRepository,
	userRepo repository.UserRepository,
	attendanceRepo repository.AttendanceRepository,
	db *gorm.DB,
) usecase.OvertimeUsecase {
	return &overtimeUsecase{
		overtimeRepo:   overtimeRepo,
		userRepo:       userRepo,
		attendanceRepo: attendanceRepo,
		db:             db,
	}
}

// CheckIn check-in tăng ca
//
// Flow:
//  1. Validate thời gian >= 17:00
//  2. Kiểm tra chưa có OT request cho ngày này
//  3. Tạo OvertimeRequest với actual_checkin
//  4. Trả về thông tin dự kiến bo tròn
func (u *overtimeUsecase) CheckIn(ctx context.Context, req usecase.OvertimeCheckInRequest) (*usecase.OvertimeCheckInResponse, error) {
	logger := slog.With("user_id", req.UserID)

	now := utils.Now()

	// 1. Validate thời gian >= 17:00
	if now.Hour() < otCheckInMinHour {
		logger.Warn("overtime check-in rejected - too early", "hour", now.Hour())
		return nil, apperrors.ErrOvertimeCheckInTooEarly
	}

	// Lấy thông tin user để có branch_id
	user, err := u.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if user.BranchID == nil {
		return nil, apperrors.ErrForbidden
	}

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, utils.HCM)

	// 2. Kiểm tra chưa có OT request cho ngày này
	existing, err := u.overtimeRepo.FindByUserAndDate(ctx, req.UserID, today)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if existing.IsCheckedIn() {
			return nil, apperrors.ErrOvertimeAlreadyCheckedIn
		}
		return nil, apperrors.ErrOvertimeAlreadyExists
	}

	// 3. Tạo OvertimeRequest
	ot := &entity.OvertimeRequest{
		UserID:        req.UserID,
		BranchID:      *user.BranchID,
		Date:          today,
		ActualCheckin: &now,
		Status:        entity.OvertimeStatusInit,
	}

	if err := u.overtimeRepo.Create(ctx, ot); err != nil {
		return nil, err
	}

	logger.Info("overtime check-in successful",
		"overtime_id", ot.ID,
		"actual_checkin", now.Format("15:04:05"),
	)

	// 4. Tính thời gian dự kiến
	estimatedStart := clampStart(now, today)
	estimatedEnd := time.Date(today.Year(), today.Month(), today.Day(), otEndHour, 0, 0, 0, utils.HCM)

	return &usecase.OvertimeCheckInResponse{
		OvertimeRequest: ot,
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
	var ot *entity.OvertimeRequest
	var err error

	if req.OvertimeID > 0 {
		ot, err = u.overtimeRepo.FindByID(ctx, req.OvertimeID)
		if err != nil {
			return nil, err
		}
		if ot.UserID != req.UserID {
			return nil, apperrors.ErrForbidden
		}
	} else {
		// Tìm OT init (đã check-in) cho hôm nay
		ot, err = u.overtimeRepo.FindActiveByUserAndDate(ctx, req.UserID, today)
		if err != nil {
			return nil, err
		}
	}

	if ot != nil && ot.IsCheckedOut() {
		return nil, apperrors.ErrOvertimeAlreadyCheckedOut
	}

	if ot != nil && ot.IsCheckedIn() {
		// Kịch bản 1: Đã check-in → cập nhật checkout, chuyển pending
		ot.ActualCheckout = &now
		ot.Status = entity.OvertimeStatusPending

		if err := u.overtimeRepo.Update(ctx, ot); err != nil {
			return nil, err
		}
	} else {
		// Kịch bản 2: Chưa check-in → tạo mới chỉ có checkout, status init
		// Kiểm tra chưa có OT record cho ngày này
		existing, err := u.overtimeRepo.FindByUserAndDate(ctx, req.UserID, today)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, apperrors.ErrOvertimeAlreadyExists
		}

		user, err := u.userRepo.FindByID(ctx, req.UserID)
		if err != nil {
			return nil, err
		}
		if user.BranchID == nil {
			return nil, apperrors.ErrForbidden
		}

		ot = &entity.OvertimeRequest{
			UserID:         req.UserID,
			BranchID:       *user.BranchID,
			Date:           today,
			ActualCheckout: &now,
			Status:         entity.OvertimeStatusInit, // init — cần bù check-in
		}
		if err := u.overtimeRepo.Create(ctx, ot); err != nil {
			return nil, err
		}
	}

	logger.Info("overtime check-out successful",
		"overtime_id", ot.ID,
		"actual_checkout", now.Format("15:04:05"),
		"has_checkin", ot.IsCheckedIn(),
	)

	// Tính thời gian dự kiến
	var estimatedStart, estimatedEnd time.Time
	var estimatedHours float64

	if ot.IsCheckedIn() && ot.IsCheckedOut() {
		estimatedStart = clampStart(*ot.ActualCheckin, ot.Date)
		estimatedEnd = clampEnd(now, ot.Date)
		estimatedHours = estimatedEnd.Sub(estimatedStart).Hours()
		if estimatedHours < 0 {
			estimatedHours = 0
		}
		if estimatedHours > otMaxHours {
			estimatedHours = otMaxHours
		}
	} else {
		// Chỉ có checkout, chưa có checkin → dự kiến từ 18:00
		estimatedStart = time.Date(today.Year(), today.Month(), today.Day(), otStartHour, 0, 0, 0, utils.HCM)
		estimatedEnd = clampEnd(now, ot.Date)
		estimatedHours = estimatedEnd.Sub(estimatedStart).Hours()
		if estimatedHours < 0 {
			estimatedHours = 0
		}
		if estimatedHours > otMaxHours {
			estimatedHours = otMaxHours
		}
	}

	note := "Lưu ý: Giờ tăng ca chỉ bắt đầu tính từ 18:00 đến 22:00 theo quy định"
	if !ot.IsCheckedIn() {
		note = "Bạn chưa check-in tăng ca. Vui lòng tạo yêu cầu chấm công bù để bổ sung check-in."
	}

	return &usecase.OvertimeCheckOutResponse{
		OvertimeRequest: ot,
		EstimatedStart:  &estimatedStart,
		EstimatedEnd:    &estimatedEnd,
		EstimatedHours:  float64(int(estimatedHours*100)) / 100,
		Note:            note,
	}, nil
}

// Process duyệt hoặc từ chối yêu cầu tăng ca
//
// Khi approved — chạy trong transaction:
//  1. Tính calculated_start = Max(actual_checkin, 18:00)
//  2. Tính calculated_end = Min(actual_checkout, 22:00)
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

	ot, err := u.overtimeRepo.FindByID(ctx, req.OvertimeID)
	if err != nil {
		return nil, err
	}

	if !ot.IsPending() {
		return nil, apperrors.ErrOvertimeNotPending
	}

	// Manager không được tự duyệt cho chính mình
	if ot.UserID == req.ProcessedByID {
		logger.Warn("self-approval attempt blocked")
		return nil, apperrors.ErrOvertimeSelfApprove
	}

	// Check-out phải hoàn tất trước khi duyệt
	if !ot.IsCheckedOut() {
		return nil, apperrors.ErrOvertimeNotCompleted
	}

	now := utils.Now()

	if req.Status == entity.OvertimeStatusApproved {
		// Transaction: tính toán và cập nhật
		err = u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			// Tính bo tròn
			calcStart := clampStart(*ot.ActualCheckin, ot.Date)
			calcEnd := clampEnd(*ot.ActualCheckout, ot.Date)
			totalHours := calcEnd.Sub(calcStart).Hours()
			if totalHours < 0 {
				totalHours = 0
			}
			if totalHours > otMaxHours {
				totalHours = otMaxHours
			}
			totalHours = float64(int(totalHours*100)) / 100

			ot.CalculatedStart = &calcStart
			ot.CalculatedEnd = &calcEnd
			ot.TotalHours = totalHours
			ot.Status = entity.OvertimeStatusApproved
			ot.ManagerID = &req.ProcessedByID
			ot.ProcessedAt = &now
			ot.ManagerNote = req.ManagerNote

			if err := tx.Save(ot).Error; err != nil {
				return err
			}

			// Link OT vào attendance_log nếu có
			today := ot.Date
			var attendLog entity.AttendanceLog
			if findErr := tx.Where("user_id = ? AND date = ?", ot.UserID, today).
				First(&attendLog).Error; findErr == nil {
				tx.Model(&entity.AttendanceLog{}).
					Where("id = ?", attendLog.ID).
					Update("overtime_request_id", ot.ID)
			}

			return nil
		})
		if err != nil {
			slog.Error("overtime approval transaction failed", "error", err)
			return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi duyệt yêu cầu tăng ca")
		}

		logger.Info("overtime approved",
			"total_hours", ot.TotalHours,
			"calculated_start", ot.CalculatedStart,
			"calculated_end", ot.CalculatedEnd,
		)
	} else {
		// Rejected
		ot.Status = entity.OvertimeStatusRejected
		ot.ManagerID = &req.ProcessedByID
		ot.ProcessedAt = &now
		ot.ManagerNote = req.ManagerNote

		if err := u.overtimeRepo.Update(ctx, ot); err != nil {
			return nil, err
		}

		logger.Info("overtime rejected", "overtime_id", ot.ID)
	}

	return u.overtimeRepo.FindByID(ctx, ot.ID)
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
	ot, err := u.overtimeRepo.FindByUserAndDate(ctx, userID, today)
	if err != nil {
		return nil, err
	}
	if ot == nil {
		return nil, apperrors.ErrNotFound
	}
	return ot, nil
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

// clampStart trả về Max(actualCheckin, 18:00 ngày date)
func clampStart(actualCheckin time.Time, date time.Time) time.Time {
	otStart := time.Date(date.Year(), date.Month(), date.Day(), otStartHour, 0, 0, 0, utils.HCM)
	if actualCheckin.Before(otStart) {
		return otStart
	}
	return actualCheckin
}

// clampEnd trả về Min(actualCheckout, 22:00 ngày date)
func clampEnd(actualCheckout time.Time, date time.Time) time.Time {
	otEnd := time.Date(date.Year(), date.Month(), date.Day(), otEndHour, 0, 0, 0, utils.HCM)
	if actualCheckout.After(otEnd) {
		return otEnd
	}
	return actualCheckout
}
