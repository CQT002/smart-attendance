package repository

import (
	"context"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
)

// GPSConfigRepository định nghĩa contract cho thao tác cấu hình GPS Geofencing
type GPSConfigRepository interface {
	// Create tạo mới cấu hình GPS
	Create(ctx context.Context, config *entity.GPSConfig) error

	// Update cập nhật cấu hình GPS
	Update(ctx context.Context, config *entity.GPSConfig) error

	// Delete xóa cấu hình GPS
	Delete(ctx context.Context, id uint) error

	// FindByID tìm cấu hình theo ID
	FindByID(ctx context.Context, id uint) (*entity.GPSConfig, error)

	// FindByBranch lấy danh sách GPS config của chi nhánh
	FindByBranch(ctx context.Context, branchID uint) ([]*entity.GPSConfig, error)

	// FindActiveBranch lấy danh sách GPS config đang active của chi nhánh
	FindActiveBranch(ctx context.Context, branchID uint) ([]*entity.GPSConfig, error)
}
