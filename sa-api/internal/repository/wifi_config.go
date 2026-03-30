package repository

import (
	"context"
	"errors"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	domainrepo "github.com/hdbank/smart-attendance/internal/domain/repository"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"gorm.io/gorm"
)

type wifiConfigRepository struct {
	db *gorm.DB
}

// NewWiFiConfigRepository tạo instance WiFiConfigRepository
func NewWiFiConfigRepository(db *gorm.DB) domainrepo.WiFiConfigRepository {
	return &wifiConfigRepository{db: db}
}

func (r *wifiConfigRepository) Create(ctx context.Context, config *entity.WiFiConfig) error {
	if err := r.db.WithContext(ctx).Create(config).Error; err != nil {
		if isDuplicateError(err) {
			return apperrors.ErrDuplicate
		}
		return apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi tạo cấu hình WiFi")
	}
	return nil
}

func (r *wifiConfigRepository) Update(ctx context.Context, config *entity.WiFiConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

func (r *wifiConfigRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.WiFiConfig{}, id).Error
}

func (r *wifiConfigRepository) FindByID(ctx context.Context, id uint) (*entity.WiFiConfig, error) {
	var config entity.WiFiConfig
	err := r.db.WithContext(ctx).First(&config, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn WiFi config")
	}
	return &config, nil
}

func (r *wifiConfigRepository) FindByBranch(ctx context.Context, branchID uint) ([]*entity.WiFiConfig, error) {
	var configs []*entity.WiFiConfig
	err := r.db.WithContext(ctx).Where("branch_id = ?", branchID).Find(&configs).Error
	return configs, err
}

func (r *wifiConfigRepository) FindActiveBranch(ctx context.Context, branchID uint) ([]*entity.WiFiConfig, error) {
	var configs []*entity.WiFiConfig
	// Index (branch_id, is_active) cho phép query này rất nhanh
	err := r.db.WithContext(ctx).
		Where("branch_id = ? AND is_active = true", branchID).
		Find(&configs).Error
	return configs, err
}

func (r *wifiConfigRepository) FindByBSSID(ctx context.Context, bssid string) (*entity.WiFiConfig, error) {
	var config entity.WiFiConfig
	// Index trên bssid để lookup nhanh
	err := r.db.WithContext(ctx).Where("bssid = ? AND is_active = true", bssid).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn WiFi config")
	}
	return &config, nil
}

// ValidateWiFi kiểm tra SSID/BSSID có được phép chấm công tại chi nhánh không
// Ưu tiên match theo BSSID (chính xác hơn) sau đó mới theo SSID
func (r *wifiConfigRepository) ValidateWiFi(ctx context.Context, branchID uint, ssid, bssid string) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&entity.WiFiConfig{}).
		Where("branch_id = ? AND is_active = true", branchID)

	if bssid != "" {
		// BSSID là MAC address - match chính xác nhất
		query = query.Where("bssid = ? OR ssid = ?", bssid, ssid)
	} else {
		query = query.Where("ssid = ?", ssid)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi xác thực WiFi")
	}
	return count > 0, nil
}
