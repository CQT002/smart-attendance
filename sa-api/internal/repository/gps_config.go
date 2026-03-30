package repository

import (
	"context"
	"errors"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	domainrepo "github.com/hdbank/smart-attendance/internal/domain/repository"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"gorm.io/gorm"
)

type gpsConfigRepository struct {
	db *gorm.DB
}

// NewGPSConfigRepository tạo instance GPSConfigRepository
func NewGPSConfigRepository(db *gorm.DB) domainrepo.GPSConfigRepository {
	return &gpsConfigRepository{db: db}
}

func (r *gpsConfigRepository) Create(ctx context.Context, config *entity.GPSConfig) error {
	if err := r.db.WithContext(ctx).Create(config).Error; err != nil {
		return apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi tạo cấu hình GPS")
	}
	return nil
}

func (r *gpsConfigRepository) Update(ctx context.Context, config *entity.GPSConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

func (r *gpsConfigRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.GPSConfig{}, id).Error
}

func (r *gpsConfigRepository) FindByID(ctx context.Context, id uint) (*entity.GPSConfig, error) {
	var config entity.GPSConfig
	err := r.db.WithContext(ctx).First(&config, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn GPS config")
	}
	return &config, nil
}

func (r *gpsConfigRepository) FindByBranch(ctx context.Context, branchID uint) ([]*entity.GPSConfig, error) {
	var configs []*entity.GPSConfig
	err := r.db.WithContext(ctx).Where("branch_id = ?", branchID).Find(&configs).Error
	return configs, err
}

func (r *gpsConfigRepository) FindActiveBranch(ctx context.Context, branchID uint) ([]*entity.GPSConfig, error) {
	var configs []*entity.GPSConfig
	err := r.db.WithContext(ctx).
		Where("branch_id = ? AND is_active = true", branchID).
		Find(&configs).Error
	return configs, err
}
