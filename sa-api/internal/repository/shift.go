package repository

import (
	"context"
	"errors"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	domainrepo "github.com/hdbank/smart-attendance/internal/domain/repository"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"gorm.io/gorm"
)

type shiftRepository struct {
	db *gorm.DB
}

// NewShiftRepository tạo instance ShiftRepository
func NewShiftRepository(db *gorm.DB) domainrepo.ShiftRepository {
	return &shiftRepository{db: db}
}

func (r *shiftRepository) Create(ctx context.Context, shift *entity.Shift) error {
	return r.db.WithContext(ctx).Create(shift).Error
}

func (r *shiftRepository) Update(ctx context.Context, shift *entity.Shift) error {
	return r.db.WithContext(ctx).
		Model(shift).
		Select("*").
		Omit("Branch", "CreatedAt").
		Updates(shift).Error
}

func (r *shiftRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Shift{}, id).Error
}

func (r *shiftRepository) FindByID(ctx context.Context, id uint) (*entity.Shift, error) {
	var shift entity.Shift
	err := r.db.WithContext(ctx).First(&shift, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn ca làm việc")
	}
	return &shift, nil
}

func (r *shiftRepository) FindByBranch(ctx context.Context, branchID uint) ([]*entity.Shift, error) {
	var shifts []*entity.Shift
	err := r.db.WithContext(ctx).
		Where("branch_id = ? AND is_active = true", branchID).
		Order("start_time").Find(&shifts).Error
	return shifts, err
}

func (r *shiftRepository) FindDefault(ctx context.Context, branchID uint) (*entity.Shift, error) {
	var shift entity.Shift
	err := r.db.WithContext(ctx).
		Where("branch_id = ? AND is_default = true AND is_active = true", branchID).
		First(&shift).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn ca mặc định")
	}
	return &shift, nil
}
