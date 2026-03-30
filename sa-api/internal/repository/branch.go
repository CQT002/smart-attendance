package repository

import (
	"context"
	"errors"
	"log/slog"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	domainrepo "github.com/hdbank/smart-attendance/internal/domain/repository"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"gorm.io/gorm"
)

type branchRepository struct {
	db *gorm.DB
}

// NewBranchRepository tạo instance BranchRepository với PostgreSQL
func NewBranchRepository(db *gorm.DB) domainrepo.BranchRepository {
	return &branchRepository{db: db}
}

func (r *branchRepository) Create(ctx context.Context, branch *entity.Branch) error {
	if err := r.db.WithContext(ctx).Create(branch).Error; err != nil {
		slog.Error("branch create failed", "error", err)
		if isDuplicateError(err) {
			return apperrors.ErrCodeDuplicate
		}
		return apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi tạo chi nhánh")
	}
	return nil
}

func (r *branchRepository) Update(ctx context.Context, branch *entity.Branch) error {
	if err := r.db.WithContext(ctx).Save(branch).Error; err != nil {
		slog.Error("branch update failed", "branch_id", branch.ID, "error", err)
		return apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi cập nhật chi nhánh")
	}
	return nil
}

func (r *branchRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Model(&entity.Branch{}).Where("id = ?", id).Update("is_active", false)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, 500, "DB_ERROR", "Lỗi xóa chi nhánh")
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrBranchNotFound
	}
	return nil
}

func (r *branchRepository) FindByID(ctx context.Context, id uint) (*entity.Branch, error) {
	var branch entity.Branch
	err := r.db.WithContext(ctx).First(&branch, "id = ? AND is_active = true", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrBranchNotFound
		}
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn chi nhánh")
	}
	return &branch, nil
}

func (r *branchRepository) FindByCode(ctx context.Context, code string) (*entity.Branch, error) {
	var branch entity.Branch
	err := r.db.WithContext(ctx).First(&branch, "code = ?", code).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrBranchNotFound
		}
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn chi nhánh")
	}
	return &branch, nil
}

// FindAll lấy danh sách chi nhánh với filter và phân trang
// Sử dụng index trên is_active để tối ưu query
func (r *branchRepository) FindAll(ctx context.Context, filter domainrepo.BranchFilter) ([]*entity.Branch, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.Branch{})

	if filter.Search != "" {
		query = query.Where("name ILIKE ? OR code ILIKE ?",
			"%"+filter.Search+"%", "%"+filter.Search+"%")
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi đếm chi nhánh")
	}

	offset := (filter.Page - 1) * filter.Limit
	var branches []*entity.Branch
	err := query.Order("created_at DESC").
		Offset(offset).Limit(filter.Limit).
		Find(&branches).Error
	if err != nil {
		return nil, 0, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn danh sách chi nhánh")
	}

	return branches, total, nil
}

func (r *branchRepository) FindActive(ctx context.Context) ([]*entity.Branch, error) {
	var branches []*entity.Branch
	// Index trên is_active giúp query này nhanh ngay cả với 100 chi nhánh
	err := r.db.WithContext(ctx).Where("is_active = true").Order("name").Find(&branches).Error
	if err != nil {
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn chi nhánh")
	}
	return branches, nil
}
