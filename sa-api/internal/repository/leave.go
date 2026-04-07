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

type leaveRepository struct {
	db *gorm.DB
}

// NewLeaveRepository tạo instance LeaveRepository với PostgreSQL
func NewLeaveRepository(db *gorm.DB) domainrepo.LeaveRepository {
	return &leaveRepository{db: db}
}

func (r *leaveRepository) Create(ctx context.Context, leave *entity.LeaveRequest) error {
	if err := r.db.WithContext(ctx).Create(leave).Error; err != nil {
		slog.Error("leave create failed", "user_id", leave.UserID, "error", err)
		return apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi tạo yêu cầu nghỉ phép")
	}
	return nil
}

func (r *leaveRepository) Update(ctx context.Context, leave *entity.LeaveRequest) error {
	if err := r.db.WithContext(ctx).Save(leave).Error; err != nil {
		slog.Error("leave update failed", "id", leave.ID, "error", err)
		return apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi cập nhật yêu cầu nghỉ phép")
	}
	return nil
}

func (r *leaveRepository) FindByID(ctx context.Context, id uint) (*entity.LeaveRequest, error) {
	var leave entity.LeaveRequest
	err := r.db.WithContext(ctx).
		Preload("User").Preload("Branch").
		Preload("ProcessedBy").
		First(&leave, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn yêu cầu nghỉ phép")
	}
	return &leave, nil
}

func (r *leaveRepository) FindAll(ctx context.Context, filter domainrepo.LeaveFilter) ([]*entity.LeaveRequest, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.LeaveRequest{})

	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.BranchID != nil {
		query = query.Where("branch_id = ?", *filter.BranchID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi đếm yêu cầu nghỉ phép")
	}

	offset := (filter.Page - 1) * filter.Limit
	var leaves []*entity.LeaveRequest
	err := query.
		Preload("User").Preload("Branch").
		Preload("ProcessedBy").
		Order("created_at DESC").
		Offset(offset).Limit(filter.Limit).
		Find(&leaves).Error
	if err != nil {
		return nil, 0, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn danh sách yêu cầu nghỉ phép")
	}

	return leaves, total, nil
}

func (r *leaveRepository) FindByUserAndDate(ctx context.Context, userID uint, date time.Time) (*entity.LeaveRequest, error) {
	var leave entity.LeaveRequest
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND leave_date = ?", userID, date).
		First(&leave).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn yêu cầu nghỉ phép")
	}
	return &leave, nil
}

func (r *leaveRepository) CountPendingByBranch(ctx context.Context, branchID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.LeaveRequest{}).
		Where("branch_id = ? AND status = ?", branchID, entity.LeaveStatusPending).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi đếm yêu cầu nghỉ phép pending")
	}
	return count, nil
}

// AutoRejectExpired chuyển PENDING sang REJECTED bằng single UPDATE
func (r *leaveRepository) AutoRejectExpired(ctx context.Context, beforeMonth time.Time, note string) (int64, error) {
	startOfMonth := time.Date(beforeMonth.Year(), beforeMonth.Month(), 1, 0, 0, 0, 0, beforeMonth.Location())

	result := r.db.WithContext(ctx).
		Model(&entity.LeaveRequest{}).
		Where("status = ? AND created_at < ?", entity.LeaveStatusPending, startOfMonth).
		Updates(map[string]interface{}{
			"status":       entity.LeaveStatusRejected,
			"manager_note": note,
			"processed_at": time.Now(),
		})

	if result.Error != nil {
		return 0, apperrors.Wrap(result.Error, 500, "DB_ERROR", "Lỗi auto-reject yêu cầu nghỉ phép hết hạn")
	}

	return result.RowsAffected, nil
}
