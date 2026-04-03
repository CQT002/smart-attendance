package repository

import (
	"context"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
)

// BranchFilter bộ lọc tìm kiếm chi nhánh
type BranchFilter struct {
	BranchID *uint
	Search   string
	IsActive *bool
	Page     int
	Limit    int
}

// BranchRepository định nghĩa contract cho thao tác dữ liệu chi nhánh
type BranchRepository interface {
	// Create tạo mới chi nhánh
	Create(ctx context.Context, branch *entity.Branch) error

	// Update cập nhật thông tin chi nhánh
	Update(ctx context.Context, branch *entity.Branch) error

	// Delete xóa mềm chi nhánh
	Delete(ctx context.Context, id uint) error

	// FindByID tìm chi nhánh theo ID
	FindByID(ctx context.Context, id uint) (*entity.Branch, error)

	// FindByCode tìm chi nhánh theo mã
	FindByCode(ctx context.Context, code string) (*entity.Branch, error)

	// FindAll lấy danh sách chi nhánh có phân trang và lọc
	FindAll(ctx context.Context, filter BranchFilter) ([]*entity.Branch, int64, error)

	// FindActive lấy danh sách chi nhánh đang hoạt động
	FindActive(ctx context.Context) ([]*entity.Branch, error)
}
