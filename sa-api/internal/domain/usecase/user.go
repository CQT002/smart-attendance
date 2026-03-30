package usecase

import (
	"context"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	"github.com/hdbank/smart-attendance/internal/domain/repository"
)

// LoginRequest yêu cầu đăng nhập
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse kết quả đăng nhập
type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         *entity.User `json:"user"`
}

// CreateUserRequest yêu cầu tạo người dùng mới
type CreateUserRequest struct {
	BranchID     *uint            `json:"branch_id"`
	EmployeeCode string           `json:"employee_code"`
	Name         string           `json:"name"`
	Email        string           `json:"email"`
	Phone        string           `json:"phone"`
	Password     string           `json:"password"`
	Role         entity.UserRole  `json:"role"`
	Department   string           `json:"department"`
	Position     string           `json:"position"`
}

// UpdateUserRequest yêu cầu cập nhật người dùng
type UpdateUserRequest struct {
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Department string `json:"department"`
	Position   string `json:"position"`
	AvatarURL  string `json:"avatar_url"`
}

// ChangePasswordRequest yêu cầu đổi mật khẩu
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// UserUsecase định nghĩa business logic cho người dùng
type UserUsecase interface {
	// Login xử lý đăng nhập, trả về JWT token
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)

	// Create tạo mới người dùng
	Create(ctx context.Context, req CreateUserRequest) (*entity.User, error)

	// Update cập nhật thông tin người dùng
	Update(ctx context.Context, id uint, req UpdateUserRequest) (*entity.User, error)

	// Delete vô hiệu hóa người dùng
	Delete(ctx context.Context, id uint) error

	// GetByID lấy thông tin người dùng theo ID
	GetByID(ctx context.Context, id uint) (*entity.User, error)

	// GetList lấy danh sách người dùng với phân trang và lọc
	GetList(ctx context.Context, filter repository.UserFilter) ([]*entity.User, int64, error)

	// ChangePassword đổi mật khẩu người dùng
	ChangePassword(ctx context.Context, userID uint, req ChangePasswordRequest) error

	// ResetPassword đặt lại mật khẩu (dành cho admin/manager)
	ResetPassword(ctx context.Context, userID uint, newPassword string) error
}
