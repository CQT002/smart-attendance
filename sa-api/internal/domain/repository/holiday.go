package repository

import (
	"context"
	"time"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
)

// HolidayFilter bộ lọc truy vấn ngày lễ
type HolidayFilter struct {
	Year     *int                // lọc theo năm (ưu tiên)
	DateFrom *time.Time          // range from (inclusive)
	DateTo   *time.Time          // range to (inclusive)
	Type     entity.HolidayType  // lọc theo loại (empty = tất cả)
	Page     int
	Limit    int
}

// HolidayRepository contract thao tác dữ liệu ngày lễ
type HolidayRepository interface {
	Create(ctx context.Context, h *entity.Holiday) error
	Update(ctx context.Context, h *entity.Holiday) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*entity.Holiday, error)

	// FindByDate tìm holiday active cho một ngày cụ thể (nil nếu không có)
	FindByDate(ctx context.Context, date time.Time) (*entity.Holiday, error)

	// FindByDateRange trả về tất cả holiday trong khoảng thời gian — dùng cho summary calc
	FindByDateRange(ctx context.Context, from, to time.Time) ([]*entity.Holiday, error)

	// FindAll list có phân trang + filter
	FindAll(ctx context.Context, filter HolidayFilter) ([]*entity.Holiday, int64, error)

	// ExistsByDate kiểm tra ngày đã có holiday chưa (loại trừ id hiện tại khi update)
	ExistsByDate(ctx context.Context, date time.Time, excludeID uint) (bool, error)
}
