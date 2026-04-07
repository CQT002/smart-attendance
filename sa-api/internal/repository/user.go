package repository

import (
	"context"
	"errors"
	"log/slog"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	domainrepo "github.com/hdbank/smart-attendance/internal/domain/repository"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"github.com/hdbank/smart-attendance/pkg/utils"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository tạo instance UserRepository với PostgreSQL
func NewUserRepository(db *gorm.DB) domainrepo.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		slog.Error("user create failed", "email", user.Email, "error", err)
		if isDuplicateError(err) {
			return apperrors.ErrEmailDuplicate
		}
		return apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi tạo người dùng")
	}
	return nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		slog.Error("user update failed", "user_id", user.ID, "error", err)
		return apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi cập nhật người dùng")
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Model(&entity.User{}).
		Where("id = ?", id).Update("is_active", false)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, 500, "DB_ERROR", "Lỗi xóa người dùng")
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrUserNotFound
	}
	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*entity.User, error) {
	var user entity.User
	// Preload Branch để tránh N+1 query
	err := r.db.WithContext(ctx).Preload("Branch").
		First(&user, "id = ? AND is_active = true", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn người dùng")
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	// Index trên email giúp query này O(log n)
	err := r.db.WithContext(ctx).Preload("Branch").
		First(&user, "email = ? AND is_active = true", email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn người dùng")
	}
	return &user, nil
}

func (r *userRepository) FindByEmployeeCode(ctx context.Context, code string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, "UPPER(employee_code) = UPPER(?)", code).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn người dùng")
	}
	return &user, nil
}

// FindAll lấy danh sách người dùng với nhiều điều kiện lọc
// Composite index (branch_id, is_active, role) tối ưu cho query phổ biến nhất
func (r *userRepository) FindAll(ctx context.Context, filter domainrepo.UserFilter) ([]*entity.User, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.User{})

	if filter.BranchID != nil {
		query = query.Where("branch_id = ?", *filter.BranchID)
	}
	if filter.Role != "" {
		query = query.Where("role = ?", filter.Role)
	}
	if filter.Department != "" {
		query = query.Where("department = ?", filter.Department)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.Search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ? OR employee_code ILIKE ?",
			"%"+filter.Search+"%", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi đếm người dùng")
	}

	offset := (filter.Page - 1) * filter.Limit
	var users []*entity.User
	err := query.Preload("Branch").
		Order("id ASC").
		Offset(offset).Limit(filter.Limit).
		Find(&users).Error
	if err != nil {
		return nil, 0, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn danh sách người dùng")
	}

	return users, total, nil
}

func (r *userRepository) FindByBranch(ctx context.Context, branchID uint) ([]*entity.User, error) {
	var users []*entity.User
	err := r.db.WithContext(ctx).
		Where("branch_id = ? AND is_active = true", branchID).
		Order("name ASC").Find(&users).Error
	if err != nil {
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn nhân viên chi nhánh")
	}
	return users, nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uint) error {
	now := utils.Now()
	return r.db.WithContext(ctx).Model(&entity.User{}).
		Where("id = ?", userID).
		Update("last_login_at", now).Error
}

func (r *userRepository) CountByBranch(ctx context.Context, branchID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).
		Where("branch_id = ? AND is_active = true", branchID).
		Count(&count).Error
	return count, err
}

func (r *userRepository) AccrueLeaveBalance(ctx context.Context, days float64) (int64, error) {
	result := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("is_active = true").
		Update("leave_balance", gorm.Expr("leave_balance + ?", days))
	if result.Error != nil {
		return 0, apperrors.Wrap(result.Error, 500, "DB_ERROR", "Lỗi cộng ngày phép hàng tháng")
	}
	return result.RowsAffected, nil
}
