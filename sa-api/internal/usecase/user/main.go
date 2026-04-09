package user

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/hdbank/smart-attendance/config"
	"github.com/hdbank/smart-attendance/internal/domain/entity"
	"github.com/hdbank/smart-attendance/internal/domain/repository"
	"github.com/hdbank/smart-attendance/internal/domain/usecase"
	"github.com/hdbank/smart-attendance/internal/infrastructure/cache"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"github.com/hdbank/smart-attendance/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

type userUsecase struct {
	userRepo repository.UserRepository
	cache    cache.Cache
	jwtCfg   config.JWTConfig
}

// NewUserUsecase tạo instance UserUsecase
func NewUserUsecase(userRepo repository.UserRepository, cache cache.Cache, jwtCfg config.JWTConfig) usecase.UserUsecase {
	return &userUsecase{
		userRepo: userRepo,
		cache:    cache,
		jwtCfg:   jwtCfg,
	}
}

// Login xác thực người dùng và tạo JWT token
func (u *userUsecase) Login(ctx context.Context, req usecase.LoginRequest) (*usecase.LoginResponse, error) {
	user, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		// Không tiết lộ email có tồn tại không — nhưng vẫn log để debug
		slog.Warn("login failed - user not found or db error", "email", req.Email, "error", err)
		return nil, apperrors.ErrInvalidCredential
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		slog.Warn("login failed - wrong password", "email", req.Email)
		return nil, apperrors.ErrInvalidCredential
	}

	if !user.IsActive {
		slog.Warn("login failed - account disabled", "user_id", user.ID, "email", req.Email)
		return nil, apperrors.New(403, "ACCOUNT_DISABLED", "Tài khoản đã bị vô hiệu hóa")
	}

	accessToken, err := utils.GenerateToken(user, u.jwtCfg.Secret, u.jwtCfg.ExpireHours)
	if err != nil {
		slog.Error("generate access token failed", "user_id", user.ID, "error", err)
		return nil, apperrors.ErrInternal
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, u.jwtCfg.Secret, u.jwtCfg.RefreshExpireDays)
	if err != nil {
		slog.Error("generate refresh token failed", "user_id", user.ID, "error", err)
		return nil, apperrors.ErrInternal
	}

	// Cập nhật last login không đồng bộ để không block response
	go func() {
		bgCtx := context.Background()
		if err := u.userRepo.UpdateLastLogin(bgCtx, user.ID); err != nil {
			slog.Error("update last login failed", "user_id", user.ID, "error", err)
		}
	}()

	slog.Info("user logged in", "user_id", user.ID, "email", user.Email, "role", user.Role)

	return &usecase.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

// Create tạo mới người dùng với mật khẩu mặc định Admin@123
func (u *userUsecase) Create(ctx context.Context, req usecase.CreateUserRequest) (*entity.User, error) {
	const defaultPassword = "Admin@123"
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("hash default password failed", "error", err)
		return nil, apperrors.ErrInternal
	}

	user := &entity.User{
		BranchID:     req.BranchID,
		EmployeeCode: strings.ToUpper(req.EmployeeCode),
		Name:         req.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		Password:     string(hashedPass),
		Role:         req.Role,
		Department:   req.Department,
		Position:     req.Position,
		IsActive:     true,
		LeaveBalance: 1,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		slog.Error("failed to create user", "email", user.Email, "error", err)
		return nil, err
	}

	slog.Info("user created", "user_id", user.ID, "email", user.Email, "role", user.Role)
	return user, nil
}

// Update cập nhật thông tin cơ bản của người dùng
func (u *userUsecase) Update(ctx context.Context, id uint, req usecase.UpdateUserRequest) (*entity.User, error) {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		slog.Error("failed to find user for update", "user_id", id, "error", err)
		return nil, err
	}

	user.Name = req.Name
	user.Phone = req.Phone
	user.Department = req.Department
	user.Position = req.Position
	user.AvatarURL = req.AvatarURL

	if err := u.userRepo.Update(ctx, user); err != nil {
		slog.Error("failed to update user", "user_id", id, "error", err)
		return nil, err
	}

	// Xóa cache user
	u.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixUser, fmt.Sprintf("%d", id)))

	return user, nil
}

// Delete vô hiệu hóa người dùng (soft delete)
func (u *userUsecase) Delete(ctx context.Context, id uint) error {
	if err := u.userRepo.Delete(ctx, id); err != nil {
		slog.Error("failed to delete user", "user_id", id, "error", err)
		return err
	}
	// Xóa cache và session của user này
	u.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixUser, fmt.Sprintf("%d", id)))
	u.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixSession, fmt.Sprintf("%d", id)))
	return nil
}

// GetByID lấy thông tin user, ưu tiên từ cache
func (u *userUsecase) GetByID(ctx context.Context, id uint) (*entity.User, error) {
	cacheKey := cache.BuildKey(cache.KeyPrefixUser, fmt.Sprintf("%d", id))

	var cached entity.User
	if err := u.cache.Get(ctx, cacheKey, &cached); err == nil {
		return &cached, nil
	}

	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		slog.Error("failed to find user", "user_id", id, "error", err)
		return nil, err
	}

	// Cache 10 phút
	u.cache.Set(ctx, cacheKey, user, 10*time.Minute)
	return user, nil
}

func (u *userUsecase) GetList(ctx context.Context, filter repository.UserFilter) ([]*entity.User, int64, error) {
	return u.userRepo.FindAll(ctx, filter)
}

// ChangePassword đổi mật khẩu sau khi xác thực mật khẩu cũ
func (u *userUsecase) ChangePassword(ctx context.Context, userID uint, req usecase.ChangePasswordRequest) error {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		slog.Error("failed to find user for password change", "user_id", userID, "error", err)
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		slog.Warn("change password failed - wrong old password", "user_id", userID)
		return apperrors.ErrInvalidPassword
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("hash new password failed", "user_id", userID, "error", err)
		return apperrors.ErrInternal
	}

	user.Password = string(hashedPass)
	if err := u.userRepo.Update(ctx, user); err != nil {
		slog.Error("failed to update password", "user_id", userID, "error", err)
		return err
	}
	return nil
}

// ResetPassword đặt lại mật khẩu (admin/manager)
func (u *userUsecase) ResetPassword(ctx context.Context, userID uint, newPassword string) error {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		slog.Error("failed to find user for password reset", "user_id", userID, "error", err)
		return err
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("hash reset password failed", "user_id", userID, "error", err)
		return apperrors.ErrInternal
	}

	user.Password = string(hashedPass)
	if err := u.userRepo.Update(ctx, user); err != nil {
		slog.Error("failed to update reset password", "user_id", userID, "error", err)
		return err
	}

	// Vô hiệu hóa session hiện tại của user
	u.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixSession, fmt.Sprintf("%d", userID)))

	slog.Info("password reset", "target_user_id", userID)
	return nil
}
