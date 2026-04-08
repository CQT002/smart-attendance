package repository

import (
	"context"
	"errors"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	domainrepo "github.com/hdbank/smart-attendance/internal/domain/repository"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"gorm.io/gorm"
)

type schedulerRepository struct {
	db *gorm.DB
}

// NewSchedulerRepository tạo instance SchedulerRepository
func NewSchedulerRepository(db *gorm.DB) domainrepo.SchedulerRepository {
	return &schedulerRepository{db: db}
}

func (r *schedulerRepository) FindByName(ctx context.Context, name string) (*entity.Scheduler, error) {
	var s entity.Scheduler
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn scheduler")
	}
	return &s, nil
}

func (r *schedulerRepository) FindAllActive(ctx context.Context) ([]*entity.Scheduler, error) {
	var schedulers []*entity.Scheduler
	err := r.db.WithContext(ctx).Where("is_active = true").Find(&schedulers).Error
	if err != nil {
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn schedulers")
	}
	return schedulers, nil
}

func (r *schedulerRepository) FindAll(ctx context.Context) ([]*entity.Scheduler, error) {
	var schedulers []*entity.Scheduler
	err := r.db.WithContext(ctx).Find(&schedulers).Error
	if err != nil {
		return nil, apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn schedulers")
	}
	return schedulers, nil
}

func (r *schedulerRepository) Update(ctx context.Context, scheduler *entity.Scheduler) error {
	return r.db.WithContext(ctx).Save(scheduler).Error
}
