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
	correctionRepo    repository.CorrectionRepository
	attendanceRepo    repository.AttendanceRepository
	userRepo          repository.UserRepository
	shiftRepo         repository.ShiftRepository
	db                *gorm.DB
	maxCreditPerMonth int64
}

// NewCorrectionUsecase tạo instance CorrectionUsecase
func NewCorrectionUsecase(
	correctionRepo repository.CorrectionRepository,
	attendanceRepo repository.AttendanceRepository,
	userRepo repository.UserRepository,
	shiftRepo repository.ShiftRepository,
	db *gorm.DB,
	correctionCfg config.CorrectionConfig,
) usecase.CorrectionUsecase {
	maxCredit := int64(correctionCfg.MaxPerMonth)
	if maxCredit <= 0 {
		maxCredit = 4 // fallback default
	}
	return &correctionUsecase{
		correctionRepo:    correctionRepo,
		attendanceRepo:    attendanceRepo,
		userRepo:          userRepo,
		shiftRepo:         shiftRepo,
		db:                db,
		maxCreditPerMonth: maxCredit,
	}
}

// Create tạo yêu cầu chấm công bù
//
// Flow:
//  1. Validate description không rỗng
//  2. Tìm AttendanceLog gốc, kiểm tra trạng thái hợp lệ (late, early_leave, late_early_leave)
//  3. Kiểm tra chưa có yêu cầu trùng cho cùng attendance_log
//  4. Kiểm tra hạn mức 4 lần/tháng
//  5. Tạo yêu cầu
func (u *correctionUsecase) Create(ctx context.Context, req usecase.CreateCorrectionRequest) (*entity.AttendanceCorrection, error) {
	logger := slog.With("user_id", req.UserID, "attendance_log_id", req.AttendanceLogID)

	// 1. Validate description
	if req.Description == "" {
		return nil, apperrors.NewValidationError(map[string]string{
			"description": "Vui lòng nhập lý do xin bù công",
		})
	}

	// 2. Tìm và validate AttendanceLog gốc
	attendLog, err := u.attendanceRepo.FindByID(ctx, req.AttendanceLogID)
	if err != nil {
		return nil, err
	}

	// Kiểm tra log thuộc về user này
	if attendLog.UserID != req.UserID {
		return nil, apperrors.ErrForbidden
	}

	// Chỉ cho phép bù cho status: late, early_leave, late_early_leave
	if !isCorrectableStatus(attendLog.Status) {
		logger.Warn("correction rejected - invalid status", "status", attendLog.Status)
		return nil, apperrors.ErrCorrectionInvalidStatus
	}

	// 3. Kiểm tra chưa có yêu cầu trùng
	existing, err := u.correctionRepo.FindByAttendanceLogID(ctx, req.AttendanceLogID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, apperrors.ErrCorrectionAlreadyExists
	}

	// 4. Kiểm tra hạn mức 4 lần/tháng (SUM credit_count)
	creditNeeded := int64(entity.CreditCountForStatus(attendLog.Status))
	now := utils.Now()
	usedCredits, err := u.correctionRepo.CountByUserInMonth(ctx, req.UserID, now)
	if err != nil {
		return nil, err
	}
	if usedCredits+creditNeeded > u.maxCreditPerMonth {
		logger.Warn("correction rejected - monthly limit exceeded",
			"used", usedCredits, "needed", creditNeeded, "max", u.maxCreditPerMonth)
		return nil, apperrors.ErrCorrectionLimitExceeded
	}

	// 5. Tạo yêu cầu
	correction := &entity.AttendanceCorrection{
		UserID:          req.UserID,
		BranchID:        attendLog.BranchID,
		AttendanceLogID: req.AttendanceLogID,
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

// Process duyệt hoặc từ chối yêu cầu chấm công bù
//
// Khi approved — chạy trong transaction:
//  1. Cập nhật AttendanceLog gốc: status → "present" (VALIDATED)
//  2. Ghi audit log: processed_by_id, processed_at, manager_note
//  3. Cập nhật correction status → approved
func (u *correctionUsecase) Process(ctx context.Context, req usecase.ProcessCorrectionRequest) (*entity.AttendanceCorrection, error) {
	logger := slog.With("correction_id", req.CorrectionID, "processed_by", req.ProcessedByID)

	// Validate status input
	if req.Status != entity.CorrectionStatusApproved && req.Status != entity.CorrectionStatusRejected {
		return nil, apperrors.NewValidationError(map[string]string{
			"status": "Trạng thái phải là approved hoặc rejected",
		})
	}

	// Tìm correction
	correction, err := u.correctionRepo.FindByID(ctx, req.CorrectionID)
	if err != nil {
		return nil, err
	}

	// Chỉ xử lý yêu cầu PENDING
	if !correction.IsPending() {
		return nil, apperrors.ErrCorrectionNotPending
	}

	// Manager không được tự duyệt cho chính mình
	if correction.UserID == req.ProcessedByID {
		logger.Warn("self-approval attempt blocked")
		return nil, apperrors.ErrCorrectionSelfApprove
	}

	now := utils.Now()

	if req.Status == entity.CorrectionStatusApproved {
		// Lấy attendance log gốc để update
		attendLog, err := u.attendanceRepo.FindByID(ctx, correction.AttendanceLogID)
		if err != nil {
			return nil, err
		}

		// Lấy shift để biết giờ chuẩn
		shift, _ := u.shiftRepo.FindDefault(ctx, correction.BranchID)
		if shift == nil {
			shift = &entity.Shift{StartTime: "08:00", EndTime: "17:00", WorkHours: 8}
		}

		// Transaction: cập nhật attendance_log + correction
		err = u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			updates := map[string]interface{}{
				"status": entity.StatusPresent,
			}

			// Tính giờ chuẩn theo shift cho ngày đó
			startH, startM := parseTime(shift.StartTime)
			endH, endM := parseTime(shift.EndTime)
			shiftStart := time.Date(attendLog.Date.Year(), attendLog.Date.Month(), attendLog.Date.Day(), startH, startM, 0, 0, utils.HCM)
			shiftEnd := time.Date(attendLog.Date.Year(), attendLog.Date.Month(), attendLog.Date.Day(), endH, endM, 0, 0, utils.HCM)

			switch correction.OriginalStatus {
			case entity.StatusLate:
				// Check-in trễ, check-out đúng giờ → fix check-in về giờ shift start
				updates["check_in_time"] = shiftStart
			case entity.StatusEarlyLeave:
				// Check-in đúng giờ, check-out sớm → fix check-out về giờ shift end
				updates["check_out_time"] = shiftEnd
			case entity.StatusLateEarlyLeave:
				// Cả hai trễ + sớm → fix cả check-in và check-out
				updates["check_in_time"] = shiftStart
				updates["check_out_time"] = shiftEnd
			}

			// Recalculate work hours
			checkIn := shiftStart
			checkOut := shiftEnd
			if correction.OriginalStatus == entity.StatusLate {
				// Chỉ fix check-in, giữ check-out gốc
				if attendLog.CheckOutTime != nil {
					checkOut = *attendLog.CheckOutTime
				}
			} else if correction.OriginalStatus == entity.StatusEarlyLeave {
				// Chỉ fix check-out, giữ check-in gốc
				if attendLog.CheckInTime != nil {
					checkIn = *attendLog.CheckInTime
				}
			}
			workHours := checkOut.Sub(checkIn).Hours()
			if workHours < 0 {
				workHours = 0
			}
			updates["work_hours"] = float64(int(workHours*100)) / 100

			// Recalculate overtime
			if workHours > shift.WorkHours {
				updates["overtime"] = float64(int((workHours-shift.WorkHours)*100)) / 100
			} else {
				updates["overtime"] = 0
			}

			if err := tx.Model(&entity.AttendanceLog{}).
				Where("id = ?", correction.AttendanceLogID).
				Updates(updates).Error; err != nil {
				return err
			}

			// Ghi audit log và cập nhật correction
			correction.Status = entity.CorrectionStatusApproved
			correction.ProcessedByID = &req.ProcessedByID
			correction.ProcessedAt = &now
			correction.ManagerNote = req.ManagerNote

			return tx.Save(correction).Error
		})
		if err != nil {
			slog.Error("correction approval transaction failed", "error", err)
			return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi duyệt yêu cầu chấm công bù")
		}

		logger.Info("correction approved",
			"attendance_log_id", correction.AttendanceLogID,
			"original_status", correction.OriginalStatus,
		)
	} else {
		// Rejected — không cần transaction phức tạp
		correction.Status = entity.CorrectionStatusRejected
		correction.ProcessedByID = &req.ProcessedByID
		correction.ProcessedAt = &now
		correction.ManagerNote = req.ManagerNote

		if err := u.correctionRepo.Update(ctx, correction); err != nil {
			return nil, err
		}

		logger.Info("correction rejected", "attendance_log_id", correction.AttendanceLogID)
	}

	// Reload để có đầy đủ relations
	return u.correctionRepo.FindByID(ctx, correction.ID)
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
