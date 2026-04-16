package repository

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	domainrepo "github.com/hdbank/smart-attendance/internal/domain/repository"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"gorm.io/gorm"
)

type holidayRepository struct {
	db *gorm.DB
}

// NewHolidayRepository tạo instance HolidayRepository với PostgreSQL
func NewHolidayRepository(db *gorm.DB) domainrepo.HolidayRepository {
	return &holidayRepository{db: db}
}

func (r *holidayRepository) Create(ctx context.Context, h *entity.Holiday) error {
	if err := r.db.WithContext(ctx).Create(h).Error; err != nil {
		slog.Error("holiday create failed", "date", h.Date.Format("2006-01-02"), "error", err)
		return apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi tạo ngày lễ")
	}
	return nil
}

func (r *holidayRepository) Update(ctx context.Context, h *entity.Holiday) error {
	if err := r.db.WithContext(ctx).Save(h).Error; err != nil {
		slog.Error("holiday update failed", "id", h.ID, "error", err)
		return apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi cập nhật ngày lễ")
	}
	return nil
}

func (r *holidayRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&entity.Holiday{}, id).Error; err != nil {
		slog.Error("holiday delete failed", "id", id, "error", err)
		return apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi xoá ngày lễ")
	}
	return nil
}

func (r *holidayRepository) FindByID(ctx context.Context, id uint) (*entity.Holiday, error) {
	var h entity.Holiday
	err := r.db.WithContext(ctx).
		Preload("CreatedBy").
		First(&h, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrHolidayNotFound
		}
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn ngày lễ")
	}
	return &h, nil
}

func (r *holidayRepository) FindByDate(ctx context.Context, date time.Time) (*entity.Holiday, error) {
	var h entity.Holiday
	err := r.db.WithContext(ctx).
		Where("date = ?", date.Format("2006-01-02")).
		First(&h).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn ngày lễ")
	}
	return &h, nil
}

func (r *holidayRepository) FindByDateRange(ctx context.Context, from, to time.Time) ([]*entity.Holiday, error) {
	var holidays []*entity.Holiday
	err := r.db.WithContext(ctx).
		Where("date BETWEEN ? AND ?", from.Format("2006-01-02"), to.Format("2006-01-02")).
		Order("date ASC").
		Find(&holidays).Error
	if err != nil {
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn danh sách ngày lễ theo khoảng")
	}
	return holidays, nil
}

func (r *holidayRepository) FindAll(ctx context.Context, filter domainrepo.HolidayFilter) ([]*entity.Holiday, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.Holiday{})

	if filter.Year != nil {
		query = query.Where("year = ?", *filter.Year)
	}
	if filter.DateFrom != nil {
		query = query.Where("date >= ?", filter.DateFrom.Format("2006-01-02"))
	}
	if filter.DateTo != nil {
		query = query.Where("date <= ?", filter.DateTo.Format("2006-01-02"))
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi đếm ngày lễ")
	}

	var holidays []*entity.Holiday
	q := query.Preload("CreatedBy").Order("date ASC")
	if filter.Limit > 0 {
		offset := (filter.Page - 1) * filter.Limit
		q = q.Offset(offset).Limit(filter.Limit)
	}
	if err := q.Find(&holidays).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn danh sách ngày lễ")
	}

	return holidays, total, nil
}

func (r *holidayRepository) ExistsByDate(ctx context.Context, date time.Time, excludeID uint) (bool, error) {
	q := r.db.WithContext(ctx).Model(&entity.Holiday{}).
		Where("date = ?", date.Format("2006-01-02"))
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi kiểm tra trùng ngày lễ")
	}
	return count > 0, nil
}
