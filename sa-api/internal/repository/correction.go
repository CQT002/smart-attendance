package repository

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	domainrepo "github.com/hdbank/smart-attendance/internal/domain/repository"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"gorm.io/gorm"
)

type correctionRepository struct {
	db *gorm.DB
}

// NewCorrectionRepository tạo instance CorrectionRepository với PostgreSQL
func NewCorrectionRepository(db *gorm.DB) domainrepo.CorrectionRepository {
	return &correctionRepository{db: db}
}

func (r *correctionRepository) Create(ctx context.Context, correction *entity.AttendanceCorrection) error {
	if err := r.db.WithContext(ctx).Create(correction).Error; err != nil {
		slog.Error("correction create failed", "user_id", correction.UserID, "error", err)
		return apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi tạo yêu cầu chấm công bù")
	}
	return nil
}

func (r *correctionRepository) Update(ctx context.Context, correction *entity.AttendanceCorrection) error {
	if err := r.db.WithContext(ctx).Save(correction).Error; err != nil {
		slog.Error("correction update failed", "id", correction.ID, "error", err)
		return apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi cập nhật yêu cầu chấm công bù")
	}
	return nil
}

func (r *correctionRepository) FindByID(ctx context.Context, id uint) (*entity.AttendanceCorrection, error) {
	var correction entity.AttendanceCorrection
	err := r.db.WithContext(ctx).
		Preload("User").Preload("Branch").
		Preload("AttendanceLog").Preload("OvertimeRequest").
		Preload("ProcessedBy").
		First(&correction, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn yêu cầu chấm công bù")
	}
	return &correction, nil
}

func (r *correctionRepository) FindAll(ctx context.Context, filter domainrepo.CorrectionFilter) ([]*entity.AttendanceCorrection, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.AttendanceCorrection{})

	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.BranchID != nil {
		query = query.Where("branch_id = ?", *filter.BranchID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.CorrectionType != "" {
		query = query.Where("correction_type = ?", filter.CorrectionType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi đếm yêu cầu chấm công bù")
	}

	offset := (filter.Page - 1) * filter.Limit
	var corrections []*entity.AttendanceCorrection
	err := query.
		Preload("User").Preload("Branch").
		Preload("AttendanceLog").Preload("OvertimeRequest").
		Preload("ProcessedBy").
		Order("created_at DESC").
		Offset(offset).Limit(filter.Limit).
		Find(&corrections).Error
	if err != nil {
		return nil, 0, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn danh sách yêu cầu chấm công bù")
	}

	return corrections, total, nil
}

// CountByUserInMonth đếm tổng credit (SUM) trong tháng, lọc theo loại correction
// correctionType rỗng = đếm tất cả
func (r *correctionRepository) CountByUserInMonth(ctx context.Context, userID uint, month time.Time, correctionType entity.CorrectionType) (int64, error) {
	startOfMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	query := r.db.WithContext(ctx).
		Model(&entity.AttendanceCorrection{}).
		Select("COALESCE(SUM(credit_count), 0)").
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startOfMonth, endOfMonth)

	if correctionType != "" {
		query = query.Where("correction_type = ?", correctionType)
	}

	var total int64
	err := query.Scan(&total).Error
	if err != nil {
		return 0, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi đếm hạn mức chấm công bù")
	}
	return total, nil
}

func (r *correctionRepository) FindByAttendanceLogID(ctx context.Context, logID uint) (*entity.AttendanceCorrection, error) {
	var correction entity.AttendanceCorrection
	err := r.db.WithContext(ctx).
		Where("attendance_log_id = ?", logID).
		First(&correction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn yêu cầu chấm công bù")
	}
	return &correction, nil
}

func (r *correctionRepository) FindByOvertimeRequestID(ctx context.Context, otID uint) (*entity.AttendanceCorrection, error) {
	var correction entity.AttendanceCorrection
	err := r.db.WithContext(ctx).
		Where("overtime_request_id = ?", otID).
		First(&correction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn yêu cầu chấm công bù")
	}
	return &correction, nil
}

// AutoRejectExpired chuyển PENDING sang REJECTED bằng single UPDATE — tối ưu cho cron job
func (r *correctionRepository) AutoRejectExpired(ctx context.Context, beforeMonth time.Time, note string) (int64, error) {
	startOfMonth := time.Date(beforeMonth.Year(), beforeMonth.Month(), 1, 0, 0, 0, 0, beforeMonth.Location())

	result := r.db.WithContext(ctx).
		Model(&entity.AttendanceCorrection{}).
		Where("status = ? AND created_at < ?", entity.CorrectionStatusPending, startOfMonth).
		Updates(map[string]interface{}{
			"status":       entity.CorrectionStatusRejected,
			"manager_note": note,
			"processed_at": time.Now(),
		})

	if result.Error != nil {
		return 0, apperrors.Wrap(result.Error, 500, "DB_ERROR", "Lỗi auto-reject yêu cầu hết hạn")
	}

	return result.RowsAffected, nil
}
