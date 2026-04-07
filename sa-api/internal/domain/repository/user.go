package repository

import (
	"context"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
)

// UserFilter bộ lọc tìm kiếm người dùng
type UserFilter struct {
	BranchID   *uint
	Role       entity.UserRole
	Department string
	Search     string
	IsActive   *bool
	Page       int
	Limit      int
}

// UserRepository định nghĩa contract cho thao tác dữ liệu người dùng
type UserRepository interface {
	// Create tạo mới người dùng
	Create(ctx context.Context, user *entity.User) error

	// Update cập nhật thông tin người dùng
	Update(ctx context.Context, user *entity.User) error

	// Delete xóa mềm người dùng
	Delete(ctx context.Context, id uint) error

	// FindByID tìm người dùng theo ID
	FindByID(ctx context.Context, id uint) (*entity.User, error)

	// FindByEmail tìm người dùng theo email (dùng cho login)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)

	// FindByEmployeeCode tìm theo mã nhân viên
	FindByEmployeeCode(ctx context.Context, code string) (*entity.User, error)

	// FindAll lấy danh sách người dùng có phân trang và lọc
	FindAll(ctx context.Context, filter UserFilter) ([]*entity.User, int64, error)

	// FindByBranch lấy danh sách nhân viên theo chi nhánh
	FindByBranch(ctx context.Context, branchID uint) ([]*entity.User, error)

	// UpdateLastLogin cập nhật thời gian đăng nhập cuối
	UpdateLastLogin(ctx context.Context, userID uint) error

	// CountByBranch đếm số nhân viên theo chi nhánh
	CountByBranch(ctx context.Context, branchID uint) (int64, error)

	// AccrueLeaveBalance cộng thêm ngày phép cho tất cả user active
	// Trả về số user được cộng phép
	AccrueLeaveBalance(ctx context.Context, days float64) (int64, error)
}
