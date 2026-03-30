package repository

import (
	"context"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
)

// ShiftRepository định nghĩa contract cho thao tác ca làm việc
type ShiftRepository interface {
	// Create tạo mới ca làm việc
	Create(ctx context.Context, shift *entity.Shift) error

	// Update cập nhật ca làm việc
	Update(ctx context.Context, shift *entity.Shift) error

	// Delete xóa ca làm việc
	Delete(ctx context.Context, id uint) error

	// FindByID tìm ca làm việc theo ID
	FindByID(ctx context.Context, id uint) (*entity.Shift, error)

	// FindByBranch lấy danh sách ca của chi nhánh
	FindByBranch(ctx context.Context, branchID uint) ([]*entity.Shift, error)

	// FindDefault tìm ca mặc định của chi nhánh
	FindDefault(ctx context.Context, branchID uint) (*entity.Shift, error)
}
