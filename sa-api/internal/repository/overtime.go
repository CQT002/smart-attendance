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

type overtimeRepository struct {
	db *gorm.DB
}

// NewOvertimeRepository tạo instance OvertimeRepository với PostgreSQL
func NewOvertimeRepository(db *gorm.DB) domainrepo.OvertimeRepository {
	return &overtimeRepository{db: db}
}

func (r *overtimeRepository) Create(ctx context.Context, overtime *entity.OvertimeRequest) error {
	if err := r.db.WithContext(ctx).Create(overtime).Error; err != nil {
		slog.Error("overtime create failed", "user_id", overtime.UserID, "error", err)
		return apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi tạo yêu cầu tăng ca")
	}
	return nil
}

func (r *overtimeRepository) Update(ctx context.Context, overtime *entity.OvertimeRequest) error {
	if err := r.db.WithContext(ctx).Save(overtime).Error; err != nil {
		slog.Error("overtime update failed", "id", overtime.ID, "error", err)
		return apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi cập nhật yêu cầu tăng ca")
	}
	return nil
}

func (r *overtimeRepository) FindByID(ctx context.Context, id uint) (*entity.OvertimeRequest, error) {
	var overtime entity.OvertimeRequest
	err := r.db.WithContext(ctx).
		Preload("User").Preload("Branch").Preload("ProcessedBy").
		First(&overtime, "overtime_requests.id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn yêu cầu tăng ca")
	}
	return &overtime, nil
}

func (r *overtimeRepository) FindAll(ctx context.Context, filter domainrepo.OvertimeFilter) ([]*entity.OvertimeRequest, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.OvertimeRequest{})

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
		query = query.Where("date >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		query = query.Where("date <= ?", *filter.DateTo)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi đếm yêu cầu tăng ca")
	}

	offset := (filter.Page - 1) * filter.Limit
	var overtimes []*entity.OvertimeRequest
	err := query.
		Preload("User").Preload("Branch").Preload("ProcessedBy").
		Order("created_at DESC").
		Offset(offset).Limit(filter.Limit).
		Find(&overtimes).Error
	if err != nil {
		return nil, 0, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn danh sách yêu cầu tăng ca")
	}

	return overtimes, total, nil
}

func (r *overtimeRepository) FindByUserAndDate(ctx context.Context, userID uint, date time.Time) (*entity.OvertimeRequest, error) {
	var overtime entity.OvertimeRequest
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND date = ?", userID, date).
		First(&overtime).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn yêu cầu tăng ca")
	}
	return &overtime, nil
}

func (r *overtimeRepository) FindActiveByUserAndDate(ctx context.Context, userID uint, date time.Time) (*entity.OvertimeRequest, error) {
	var overtime entity.OvertimeRequest
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND date = ? AND actual_checkin IS NOT NULL AND actual_checkout IS NULL", userID, date).
		First(&overtime).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn yêu cầu tăng ca active")
	}
	return &overtime, nil
}

// AutoRejectExpired chuyển PENDING sang REJECTED bằng single UPDATE
func (r *overtimeRepository) AutoRejectExpired(ctx context.Context, beforeMonth time.Time, note string) (int64, error) {
	startOfMonth := time.Date(beforeMonth.Year(), beforeMonth.Month(), 1, 0, 0, 0, 0, beforeMonth.Location())

	result := r.db.WithContext(ctx).
		Model(&entity.OvertimeRequest{}).
		Where("status IN (?, ?) AND created_at < ?", entity.OvertimeStatusInit, entity.OvertimeStatusPending, startOfMonth).
		Updates(map[string]interface{}{
			"status":       entity.OvertimeStatusRejected,
			"manager_note": note,
			"processed_at": time.Now(),
		})

	if result.Error != nil {
		return 0, apperrors.Wrap(result.Error, 500, "DB_ERROR", "Lỗi auto-reject yêu cầu tăng ca hết hạn")
	}

	return result.RowsAffected, nil
}
