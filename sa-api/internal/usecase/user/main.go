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
		// Không tiết lộ email có tồn tại không
		return nil, apperrors.ErrInvalidCredential
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		slog.Warn("login failed - wrong password", "email", req.Email)
		return nil, apperrors.ErrInvalidCredential
	}

	if !user.IsActive {
		return nil, apperrors.New(403, "ACCOUNT_DISABLED", "Tài khoản đã bị vô hiệu hóa")
	}

	accessToken, err := utils.GenerateToken(user, u.jwtCfg.Secret, u.jwtCfg.ExpireHours)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, u.jwtCfg.Secret, u.jwtCfg.RefreshExpireDays)
	if err != nil {
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
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	slog.Info("user created", "user_id", user.ID, "email", user.Email, "role", user.Role)
	return user, nil
}

// Update cập nhật thông tin cơ bản của người dùng
func (u *userUsecase) Update(ctx context.Context, id uint, req usecase.UpdateUserRequest) (*entity.User, error) {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.Name = req.Name
	user.Phone = req.Phone
	user.Department = req.Department
	user.Position = req.Position
	user.AvatarURL = req.AvatarURL

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	// Xóa cache user
	u.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixUser, fmt.Sprintf("%d", id)))

	return user, nil
}

// Delete vô hiệu hóa người dùng (soft delete)
func (u *userUsecase) Delete(ctx context.Context, id uint) error {
	if err := u.userRepo.Delete(ctx, id); err != nil {
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
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return apperrors.ErrInvalidPassword
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return apperrors.ErrInternal
	}

	user.Password = string(hashedPass)
	return u.userRepo.Update(ctx, user)
}

// ResetPassword đặt lại mật khẩu (admin/manager)
func (u *userUsecase) ResetPassword(ctx context.Context, userID uint, newPassword string) error {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return apperrors.ErrInternal
	}

	user.Password = string(hashedPass)
	if err := u.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// Vô hiệu hóa session hiện tại của user
	u.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixSession, fmt.Sprintf("%d", userID)))

	slog.Info("password reset", "target_user_id", userID)
	return nil
}
