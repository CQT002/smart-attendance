package repository

import (
	"context"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
)

// SchedulerRepository định nghĩa contract cho thao tác bảng schedulers
type SchedulerRepository interface {
	// FindByName tìm scheduler theo tên (VD: "leave_accrual")
	FindByName(ctx context.Context, name string) (*entity.Scheduler, error)

	// FindAllActive lấy tất cả scheduler đang bật
	FindAllActive(ctx context.Context) ([]*entity.Scheduler, error)

	// FindAll lấy tất cả scheduler
	FindAll(ctx context.Context) ([]*entity.Scheduler, error)

	// Update cập nhật scheduler (đổi cron_expr, bật/tắt, ghi trạng thái chạy)
	Update(ctx context.Context, scheduler *entity.Scheduler) error
}
