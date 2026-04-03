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

	// Populate GPS from gps_configs
	type gpsInfo struct {
		Latitude  float64
		Longitude float64
		Radius    float64
	}
	var gps gpsInfo
	if err := r.db.WithContext(ctx).Raw(
		"SELECT latitude, longitude, radius FROM gps_configs WHERE branch_id = ? AND is_active = true ORDER BY id ASC LIMIT 1", id,
	).Scan(&gps).Error; err == nil && gps.Latitude != 0 {
		branch.Latitude = &gps.Latitude
		branch.Longitude = &gps.Longitude
		branch.GPSRadius = &gps.Radius
	}

	return &branch, nil
}

func (r *branchRepository) FindByCode(ctx context.Context, code string) (*entity.Branch, error) {
	var branch entity.Branch
	err := r.db.WithContext(ctx).First(&branch, "UPPER(code) = UPPER(?)", code).Error
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

	if filter.BranchID != nil {
		query = query.Where("id = ?", *filter.BranchID)
	}
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

	// Populate computed fields (gorm:"-") from related tables
	if len(branches) > 0 {
		var branchIDs []uint
		for _, b := range branches {
			branchIDs = append(branchIDs, b.ID)
		}

		// WiFi count
		type wifiCount struct {
			BranchID uint
			Count    int64
		}
		var wifiCounts []wifiCount
		r.db.WithContext(ctx).Raw(
			"SELECT branch_id, COUNT(*) as count FROM wifi_configs WHERE branch_id IN ? AND is_active = true GROUP BY branch_id",
			branchIDs,
		).Scan(&wifiCounts)
		wifiMap := make(map[uint]int64)
		for _, c := range wifiCounts {
			wifiMap[c.BranchID] = c.Count
		}

		// GPS config (first active per branch)
		type gpsInfo struct {
			BranchID  uint
			Latitude  float64
			Longitude float64
			Radius    float64
		}
		var gpsInfos []gpsInfo
		r.db.WithContext(ctx).Raw(
			`SELECT DISTINCT ON (branch_id) branch_id, latitude, longitude, radius
			 FROM gps_configs WHERE branch_id IN ? AND is_active = true
			 ORDER BY branch_id, id ASC`,
			branchIDs,
		).Scan(&gpsInfos)
		gpsMap := make(map[uint]gpsInfo)
		for _, g := range gpsInfos {
			gpsMap[g.BranchID] = g
		}

		for _, b := range branches {
			b.WifiCount = wifiMap[b.ID]
			if gps, ok := gpsMap[b.ID]; ok {
				b.Latitude = &gps.Latitude
				b.Longitude = &gps.Longitude
				b.GPSRadius = &gps.Radius
			}
		}
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
