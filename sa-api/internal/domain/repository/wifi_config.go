package repository

import (
	"context"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
)

// WiFiConfigRepository định nghĩa contract cho thao tác cấu hình WiFi
type WiFiConfigRepository interface {
	// Create tạo mới cấu hình WiFi
	Create(ctx context.Context, config *entity.WiFiConfig) error

	// Update cập nhật cấu hình WiFi
	Update(ctx context.Context, config *entity.WiFiConfig) error

	// Delete xóa cấu hình WiFi
	Delete(ctx context.Context, id uint) error

	// FindByID tìm cấu hình theo ID
	FindByID(ctx context.Context, id uint) (*entity.WiFiConfig, error)

	// FindByBranch lấy danh sách WiFi config của chi nhánh
	FindByBranch(ctx context.Context, branchID uint) ([]*entity.WiFiConfig, error)

	// FindActiveBranch lấy danh sách WiFi config đang active của chi nhánh
	FindActiveBranch(ctx context.Context, branchID uint) ([]*entity.WiFiConfig, error)

	// FindByBSSID tìm config theo địa chỉ MAC của router
	FindByBSSID(ctx context.Context, bssid string) (*entity.WiFiConfig, error)

	// ValidateWiFi kiểm tra xem SSID/BSSID có hợp lệ cho chi nhánh không
	ValidateWiFi(ctx context.Context, branchID uint, ssid, bssid string) (bool, error)
}
