package correction

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/hdbank/smart-attendance/config"
	"github.com/hdbank/smart-attendance/internal/domain/entity"
	"github.com/hdbank/smart-attendance/internal/domain/repository"
	"github.com/hdbank/smart-attendance/internal/domain/usecase"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"github.com/hdbank/smart-attendance/pkg/utils"
	"gorm.io/gorm"
)

type correctionUsecase struct {
	correctionRepo            repository.CorrectionRepository
	attendanceRepo            repository.AttendanceRepository
	overtimeRepo              repository.OvertimeRepository
	userRepo                  repository.UserRepository
	shiftRepo                 repository.ShiftRepository
	db                        *gorm.DB
	maxCreditPerMonth         int64
	overtimeMaxCreditPerMonth int64
}

// NewCorrectionUsecase tạo instance CorrectionUsecase
func NewCorrectionUsecase(
	correctionRepo repository.CorrectionRepository,
	attendanceRepo repository.AttendanceRepository,
	overtimeRepo repository.OvertimeRepository,
	userRepo repository.UserRepository,
	shiftRepo repository.ShiftRepository,
	db *gorm.DB,
	correctionCfg config.CorrectionConfig,
) usecase.CorrectionUsecase {
	maxCredit := int64(correctionCfg.MaxPerMonth)
	if maxCredit <= 0 {
		maxCredit = 4
	}
	otMaxCredit := int64(correctionCfg.OvertimeMaxPerMonth)
	if otMaxCredit <= 0 {
		otMaxCredit = 4
	}
	return &correctionUsecase{
		correctionRepo:            correctionRepo,
		attendanceRepo:            attendanceRepo,
		overtimeRepo:              overtimeRepo,
		userRepo:                  userRepo,
		shiftRepo:                 shiftRepo,
		db:                        db,
		maxCreditPerMonth:         maxCredit,
		overtimeMaxCreditPerMonth: otMaxCredit,
	}
}

// Create tạo yêu cầu chấm công bù
//
// Hai loại:
//   - attendance: bù cho ca chính thức (late, early_leave, late_early_leave)
//   - overtime: bù cho tăng ca (missing_checkin, missing_checkout)
func (u *correctionUsecase) Create(ctx context.Context, req usecase.CreateCorrectionRequest) (*entity.AttendanceCorrection, error) {
	logger := slog.With("user_id", req.UserID, "correction_type", req.CorrectionType)

	// Validate description
	if req.Description == "" {
		return nil, apperrors.NewValidationError(map[string]string{
			"description": "Vui lòng nhập lý do xin bù công",
		})
	}

	// Default correction type
	corrType := req.CorrectionType
	if corrType == "" {
		corrType = entity.CorrectionTypeAttendance
	}

	if corrType == entity.CorrectionTypeOvertime {
		return u.createOvertimeCorrection(ctx, req, logger)
	}
	return u.createAttendanceCorrection(ctx, req, logger)
}

// createAttendanceCorrection bù công ca chính thức (logic cũ)
func (u *correctionUsecase) createAttendanceCorrection(ctx context.Context, req usecase.CreateCorrectionRequest, logger *slog.Logger) (*entity.AttendanceCorrection, error) {
	attendLog, err := u.attendanceRepo.FindByID(ctx, req.AttendanceLogID)
	if err != nil {
		return nil, err
	}
	if attendLog.UserID != req.UserID {
		logger.Warn("correction rejected - user does not own attendance log", "attendance_log_id", req.AttendanceLogID)
		return nil, apperrors.ErrForbidden
	}
	if !isCorrectableStatus(attendLog.Status) {
		logger.Warn("correction rejected - invalid status", "status", attendLog.Status)
		return nil, apperrors.ErrCorrectionInvalidStatus
	}

	existing, err := u.correctionRepo.FindByAttendanceLogID(ctx, req.AttendanceLogID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, apperrors.ErrCorrectionAlreadyExists
	}

	creditNeeded := int64(entity.CreditCountForStatus(attendLog.Status))
	now := utils.Now()
	usedCredits, err := u.correctionRepo.CountByUserInMonth(ctx, req.UserID, now, entity.CorrectionTypeAttendance)
	if err != nil {
		return nil, err
	}
	if usedCredits+creditNeeded > u.maxCreditPerMonth {
		logger.Warn("correction rejected - monthly limit exceeded",
			"used", usedCredits, "needed", creditNeeded, "max", u.maxCreditPerMonth)
		return nil, apperrors.ErrCorrectionLimitExceeded
	}

	correction := &entity.AttendanceCorrection{
		CorrectionType:  entity.CorrectionTypeAttendance,
		UserID:          req.UserID,
		BranchID:        attendLog.BranchID,
		AttendanceLogID: &req.AttendanceLogID,
		OriginalStatus:  attendLog.Status,
		CreditCount:     int(creditNeeded),
		Description:     req.Description,
		Status:          entity.CorrectionStatusPending,
	}

	if err := u.correctionRepo.Create(ctx, correction); err != nil {
		return nil, err
	}

	logger.Info("correction request created",
		"correction_id", correction.ID,
		"original_status", attendLog.Status,
	)

	return correction, nil
}

// createOvertimeCorrection bù công tăng ca
//
// Kịch bản:
//   - OT có check-in nhưng thiếu check-out → original_status = missing_checkout
//   - OT có check-out nhưng thiếu check-in → original_status = missing_checkin
func (u *correctionUsecase) createOvertimeCorrection(ctx context.Context, req usecase.CreateCorrectionRequest, logger *slog.Logger) (*entity.AttendanceCorrection, error) {
	// Tìm OvertimeRequest gốc
	otReq, err := u.overtimeRepo.FindByID(ctx, req.OvertimeRequestID)
	if err != nil {
		return nil, err
	}
	if otReq.UserID != req.UserID {
		logger.Warn("overtime correction rejected - user does not own overtime request", "overtime_request_id", req.OvertimeRequestID)
		return nil, apperrors.ErrForbidden
	}

	// Chỉ cho phép bù khi OT ở trạng thái init (thiếu 1 trong 2)
	if !otReq.IsInit() {
		return nil, apperrors.NewValidationError(map[string]string{
			"overtime_request_id": "Yêu cầu tăng ca đã đủ check-in/check-out hoặc đã được xử lý",
		})
	}

	// Xác định thiếu gì
	var originalStatus entity.AttendanceStatus
	if otReq.IsCheckedIn() && !otReq.IsCheckedOut() {
		originalStatus = entity.StatusMissingCheckout
	} else if !otReq.IsCheckedIn() && otReq.IsCheckedOut() {
		originalStatus = entity.StatusMissingCheckin
	} else {
		return nil, apperrors.NewValidationError(map[string]string{
			"overtime_request_id": "Yêu cầu tăng ca không có dữ liệu check-in hoặc check-out",
		})
	}

	// Kiểm tra chưa có correction trùng
	existing, err := u.correctionRepo.FindByOvertimeRequestID(ctx, req.OvertimeRequestID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, apperrors.ErrCorrectionAlreadyExists
	}

	// Kiểm tra hạn mức (riêng biệt cho overtime)
	now := utils.Now()
	usedCredits, err := u.correctionRepo.CountByUserInMonth(ctx, req.UserID, now, entity.CorrectionTypeOvertime)
	if err != nil {
		return nil, err
	}
	if usedCredits+1 > u.overtimeMaxCreditPerMonth {
		logger.Warn("overtime correction rejected - monthly limit exceeded",
			"used", usedCredits, "max", u.overtimeMaxCreditPerMonth)
		return nil, apperrors.ErrCorrectionLimitExceeded
	}

	otID := req.OvertimeRequestID
	correction := &entity.AttendanceCorrection{
		CorrectionType:    entity.CorrectionTypeOvertime,
		UserID:            req.UserID,
		BranchID:          otReq.BranchID,
		OvertimeRequestID: &otID,
		OriginalStatus:    originalStatus,
		CreditCount:       1,
		Description:       req.Description,
		Status:            entity.CorrectionStatusPending,
	}

	// Transaction: tạo correction + bổ sung thời gian thiếu + chuyển OT sang pending
	err = u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(correction).Error; err != nil {
			return err
		}

		otDate := otReq.Date
		otUpdates := map[string]interface{}{
			"status": entity.OvertimeStatusPending,
		}

		// Lấy OT config từ shift của branch
		otShift, _ := u.shiftRepo.FindDefault(ctx, otReq.BranchID)
		otStartHour, otEndHour := 18, 22
		if otShift != nil {
			if otShift.OTStartHour > 0 {
				otStartHour = otShift.OTStartHour
			}
			if otShift.OTEndHour > 0 {
				otEndHour = otShift.OTEndHour
			}
		}

		switch originalStatus {
		case entity.StatusMissingCheckout:
			// Có check-in, thiếu check-out → bổ sung OTEndHour
			otEnd := time.Date(otDate.Year(), otDate.Month(), otDate.Day(), otEndHour, 0, 0, 0, utils.HCM)
			otUpdates["actual_checkout"] = otEnd
		case entity.StatusMissingCheckin:
			// Có check-out, thiếu check-in → bổ sung OTStartHour
			otStart := time.Date(otDate.Year(), otDate.Month(), otDate.Day(), otStartHour, 0, 0, 0, utils.HCM)
			otUpdates["actual_checkin"] = otStart
		}

		return tx.Model(&entity.OvertimeRequest{}).
			Where("id = ?", otReq.ID).
			Updates(otUpdates).Error
	})
	if err != nil {
		slog.Error("overtime correction create failed", "error", err)
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi tạo yêu cầu bổ sung công tăng ca")
	}

	logger.Info("overtime correction request created",
		"correction_id", correction.ID,
		"overtime_request_id", req.OvertimeRequestID,
		"original_status", originalStatus,
	)

	return correction, nil
}

// Process duyệt hoặc từ chối yêu cầu chấm công bù
//
// Phân nhánh theo correction_type:
//   - attendance: cập nhật AttendanceLog gốc
//   - overtime: bổ sung thời gian thiếu cho OT và approve OvertimeRequest
func (u *correctionUsecase) Process(ctx context.Context, req usecase.ProcessCorrectionRequest) (*entity.AttendanceCorrection, error) {
	logger := slog.With("correction_id", req.CorrectionID, "processed_by", req.ProcessedByID)

	if req.Status != entity.CorrectionStatusApproved && req.Status != entity.CorrectionStatusRejected {
		return nil, apperrors.NewValidationError(map[string]string{
			"status": "Trạng thái phải là approved hoặc rejected",
		})
	}

	correction, err := u.correctionRepo.FindByID(ctx, req.CorrectionID)
	if err != nil {
		return nil, err
	}
	if !correction.IsPending() {
		return nil, apperrors.ErrCorrectionNotPending
	}
	if correction.UserID == req.ProcessedByID {
		logger.Warn("self-approval attempt blocked")
		return nil, apperrors.ErrCorrectionSelfApprove
	}

	now := utils.Now()

	if req.Status == entity.CorrectionStatusRejected {
		correction.Status = entity.CorrectionStatusRejected
		correction.ProcessedByID = &req.ProcessedByID
		correction.ProcessedAt = &now
		correction.ManagerNote = req.ManagerNote

		if err := u.correctionRepo.Update(ctx, correction); err != nil {
			return nil, err
		}
		logger.Info("correction rejected", "correction_type", correction.CorrectionType)
		return u.correctionRepo.FindByID(ctx, correction.ID)
	}

	// === APPROVED ===
	if correction.IsOvertime() {
		err = u.processOvertimeCorrection(ctx, correction, req.ProcessedByID, req.ManagerNote, now)
	} else {
		err = u.processAttendanceCorrection(ctx, correction, req.ProcessedByID, req.ManagerNote, now)
	}
	if err != nil {
		return nil, err
	}

	return u.correctionRepo.FindByID(ctx, correction.ID)
}

// processAttendanceCorrection duyệt bù công ca chính thức (logic cũ)
func (u *correctionUsecase) processAttendanceCorrection(
	ctx context.Context,
	correction *entity.AttendanceCorrection,
	processedByID uint,
	managerNote string,
	now time.Time,
) error {
	if correction.AttendanceLogID == nil {
		slog.Error("correction missing attendance_log_id", "correction_id", correction.ID)
		return apperrors.ErrNotFound
	}

	attendLog, err := u.attendanceRepo.FindByID(ctx, *correction.AttendanceLogID)
	if err != nil {
		return err
	}

	shift, _ := u.shiftRepo.FindDefault(ctx, correction.BranchID)
	if shift == nil {
		shift = &entity.Shift{StartTime: "08:00", EndTime: "17:00", WorkHours: 8, MorningEnd: "12:00", AfternoonStart: "13:00"}
	}

	err = u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"status": entity.StatusPresent,
		}

		startH, startM := parseTime(shift.StartTime)
		endH, endM := parseTime(shift.EndTime)
		shiftStart := time.Date(attendLog.Date.Year(), attendLog.Date.Month(), attendLog.Date.Day(), startH, startM, 0, 0, utils.HCM)
		shiftEnd := time.Date(attendLog.Date.Year(), attendLog.Date.Month(), attendLog.Date.Day(), endH, endM, 0, 0, utils.HCM)

		switch correction.OriginalStatus {
		case entity.StatusLate:
			updates["check_in_time"] = shiftStart
		case entity.StatusEarlyLeave:
			updates["check_out_time"] = shiftEnd
		case entity.StatusLateEarlyLeave:
			updates["check_in_time"] = shiftStart
			updates["check_out_time"] = shiftEnd
		}

		checkIn := shiftStart
		checkOut := shiftEnd
		if correction.OriginalStatus == entity.StatusLate {
			if attendLog.CheckOutTime != nil {
				checkOut = *attendLog.CheckOutTime
			}
		} else if correction.OriginalStatus == entity.StatusEarlyLeave {
			if attendLog.CheckInTime != nil {
				checkIn = *attendLog.CheckInTime
			}
		}
		workHours := checkOut.Sub(checkIn).Hours()
		if workHours < 0 {
			workHours = 0
		}
		updates["work_hours"] = float64(int(workHours*100)) / 100

		if err := tx.Model(&entity.AttendanceLog{}).
			Where("id = ?", *correction.AttendanceLogID).
			Updates(updates).Error; err != nil {
			return err
		}

		correction.Status = entity.CorrectionStatusApproved
		correction.ProcessedByID = &processedByID
		correction.ProcessedAt = &now
		correction.ManagerNote = managerNote

		return tx.Save(correction).Error
	})
	if err != nil {
		slog.Error("attendance correction approval failed", "error", err)
		return apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi duyệt yêu cầu chấm công bù")
	}

	slog.Info("attendance correction approved",
		"correction_id", correction.ID,
		"attendance_log_id", correction.AttendanceLogID,
	)
	return nil
}

// processOvertimeCorrection duyệt bù công tăng ca
//
// Thời gian thiếu đã được bổ sung lúc tạo correction (createOvertimeCorrection).
// Process chỉ cần: tính bo tròn → approve OT.
func (u *correctionUsecase) processOvertimeCorrection(
	ctx context.Context,
	correction *entity.AttendanceCorrection,
	processedByID uint,
	managerNote string,
	now time.Time,
) error {
	if correction.OvertimeRequestID == nil {
		slog.Error("correction missing overtime_request_id", "correction_id", correction.ID)
		return apperrors.ErrNotFound
	}

	// Reload OT (đã có đủ actual_checkin + actual_checkout từ lúc tạo correction)
	otReq, err := u.overtimeRepo.FindByID(ctx, *correction.OvertimeRequestID)
	if err != nil {
		return err
	}

	if otReq.ActualCheckin == nil || otReq.ActualCheckout == nil {
		slog.Error("overtime request missing checkin/checkout after correction",
			"correction_id", correction.ID, "overtime_request_id", *correction.OvertimeRequestID)
		return apperrors.ErrOvertimeNotCompleted
	}

	// Lấy OT config từ shift của branch
	otShift, _ := u.shiftRepo.FindDefault(ctx, correction.BranchID)
	otStartHour, otEndHour := 18, 22
	maxOTHours := 4.0
	if otShift != nil {
		if otShift.OTStartHour > 0 {
			otStartHour = otShift.OTStartHour
		}
		if otShift.OTEndHour > 0 {
			otEndHour = otShift.OTEndHour
		}
	}

	err = u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Tính bo tròn
		otDate := otReq.Date
		calcStart := clampOTStart(*otReq.ActualCheckin, otDate, otStartHour)
		calcEnd := clampOTEnd(*otReq.ActualCheckout, otDate, otEndHour)
		totalHours := calcEnd.Sub(calcStart).Hours()
		if totalHours < 0 {
			totalHours = 0
		}
		if totalHours > maxOTHours {
			totalHours = maxOTHours
		}
		totalHours = float64(int(totalHours*100)) / 100

		// Approve OT request
		otReq.CalculatedStart = &calcStart
		otReq.CalculatedEnd = &calcEnd
		otReq.TotalHours = totalHours
		otReq.Status = entity.OvertimeStatusApproved
		otReq.ManagerID = &processedByID
		otReq.ProcessedAt = &now
		otReq.ManagerNote = managerNote

		if err := tx.Save(otReq).Error; err != nil {
			return err
		}

		// Approve correction
		correction.Status = entity.CorrectionStatusApproved
		correction.ProcessedByID = &processedByID
		correction.ProcessedAt = &now
		correction.ManagerNote = managerNote

		return tx.Save(correction).Error
	})
	if err != nil {
		slog.Error("overtime correction approval failed", "error", err)
		return apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi duyệt yêu cầu bù công tăng ca")
	}

	slog.Info("overtime correction approved",
		"correction_id", correction.ID,
		"overtime_request_id", correction.OvertimeRequestID,
		"total_hours", otReq.TotalHours,
	)
	return nil
}

// clampOTStart trả về Max(actualCheckin, startHour:00)
func clampOTStart(actual time.Time, date time.Time, startHour int) time.Time {
	otStart := time.Date(date.Year(), date.Month(), date.Day(), startHour, 0, 0, 0, utils.HCM)
	if actual.Before(otStart) {
		return otStart
	}
	return actual
}

// clampOTEnd trả về Min(actualCheckout, endHour:00)
func clampOTEnd(actual time.Time, date time.Time, endHour int) time.Time {
	otEnd := time.Date(date.Year(), date.Month(), date.Day(), endHour, 0, 0, 0, utils.HCM)
	if actual.After(otEnd) {
		return otEnd
	}
	return actual
}

func (u *correctionUsecase) GetByID(ctx context.Context, id uint) (*entity.AttendanceCorrection, error) {
	return u.correctionRepo.FindByID(ctx, id)
}

func (u *correctionUsecase) GetList(ctx context.Context, filter repository.CorrectionFilter) ([]*entity.AttendanceCorrection, int64, error) {
	return u.correctionRepo.FindAll(ctx, filter)
}

func (u *correctionUsecase) GetMyList(ctx context.Context, userID uint, status entity.CorrectionStatus, page, limit int) ([]*entity.AttendanceCorrection, int64, error) {
	filter := repository.CorrectionFilter{
		UserID: &userID,
		Status: status,
		Page:   page,
		Limit:  limit,
	}
	return u.correctionRepo.FindAll(ctx, filter)
}

// BatchApprove duyệt tất cả yêu cầu PENDING — gọi Process tuần tự cho từng yêu cầu
func (u *correctionUsecase) BatchApprove(ctx context.Context, processedByID uint, branchID *uint) (int64, error) {
	filter := repository.CorrectionFilter{
		BranchID: branchID,
		Status:   entity.CorrectionStatusPending,
		Page:     1,
		Limit:    1000,
	}
	corrections, _, err := u.correctionRepo.FindAll(ctx, filter)
	if err != nil {
		return 0, err
	}

	var approved int64
	for _, c := range corrections {
		// Bỏ qua yêu cầu của chính mình (self-approve)
		if c.UserID == processedByID {
			continue
		}
		req := usecase.ProcessCorrectionRequest{
			CorrectionID:  c.ID,
			ProcessedByID: processedByID,
			Status:        entity.CorrectionStatusApproved,
			ManagerNote:   "Duyệt hàng loạt",
		}
		if _, err := u.Process(ctx, req); err != nil {
			slog.Warn("batch approve correction skipped", "id", c.ID, "error", err)
			continue
		}
		approved++
	}

	slog.Info("batch approve corrections completed", "approved", approved, "total_pending", len(corrections))
	return approved, nil
}

// AutoRejectExpired tự động reject yêu cầu PENDING của tháng trước
func (u *correctionUsecase) AutoRejectExpired(ctx context.Context) (int64, error) {
	now := utils.Now()
	note := "Hệ thống tự động từ chối do quá hạn chốt lương"

	count, err := u.correctionRepo.AutoRejectExpired(ctx, now, note)
	if err != nil {
		slog.Error("auto-reject expired corrections failed", "error", err)
		return 0, err
	}

	if count > 0 {
		slog.Info("auto-reject expired corrections completed", "rejected_count", count)
	}

	return count, nil
}

// isCorrectableStatus kiểm tra trạng thái có được phép đăng ký bù công không
func isCorrectableStatus(status entity.AttendanceStatus) bool {
	return status == entity.StatusLate ||
		status == entity.StatusEarlyLeave ||
		status == entity.StatusLateEarlyLeave
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
