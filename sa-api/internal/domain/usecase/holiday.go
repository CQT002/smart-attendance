package usecase

import (
	"context"
	"time"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	"github.com/hdbank/smart-attendance/internal/domain/repository"
)

// CreateHolidayRequest payload tạo ngày lễ
type CreateHolidayRequest struct {
	Name          string             `json:"name"`
	Date          string             `json:"date"`           // YYYY-MM-DD
	Coefficient   float64            `json:"coefficient"`    // optional — nếu 0 sẽ dùng default theo type
	Type          entity.HolidayType `json:"type"`           // national | company
	IsCompensated bool               `json:"is_compensated"`
	CompensateFor string             `json:"compensate_for"` // YYYY-MM-DD — optional
	Description   string             `json:"description"`
	CreatedByID   uint               `json:"-"`
}

// UpdateHolidayRequest payload cập nhật ngày lễ
type UpdateHolidayRequest struct {
	ID            uint
	Name          string             `json:"name"`
	Date          string             `json:"date"`
	Coefficient   float64            `json:"coefficient"`
	Type          entity.HolidayType `json:"type"`
	IsCompensated bool               `json:"is_compensated"`
	CompensateFor string             `json:"compensate_for"`
	Description   string             `json:"description"`
}

// HolidayUsecase định nghĩa business logic cho ngày lễ
type HolidayUsecase interface {
	Create(ctx context.Context, req CreateHolidayRequest) (*entity.Holiday, error)
	Update(ctx context.Context, req UpdateHolidayRequest) (*entity.Holiday, error)
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*entity.Holiday, error)
	GetList(ctx context.Context, filter repository.HolidayFilter) ([]*entity.Holiday, int64, error)

	// GetCalendar lấy danh sách ngày lễ trong khoảng (dùng cho mobile calendar).
	// Có cache Redis, TTL 24h.
	GetCalendar(ctx context.Context, from, to time.Time) ([]*entity.Holiday, error)

	// GetByDate tra nhanh 1 ngày có phải holiday không (cached).
	GetByDate(ctx context.Context, date time.Time) (*entity.Holiday, error)
}
