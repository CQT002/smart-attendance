package repository

import (
	"context"
	"time"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
)

// OvertimeFilter bộ lọc yêu cầu tăng ca
type OvertimeFilter struct {
	UserID   *uint
	BranchID *uint
	Status   entity.OvertimeStatus
	DateFrom *time.Time
	DateTo   *time.Time
	Page     int
	Limit    int
}

// OvertimeRepository định nghĩa contract cho thao tác dữ liệu tăng ca
type OvertimeRepository interface {
	// Create tạo yêu cầu tăng ca mới
	Create(ctx context.Context, overtime *entity.OvertimeRequest) error

	// Update cập nhật yêu cầu tăng ca
	Update(ctx context.Context, overtime *entity.OvertimeRequest) error

	// FindByID tìm yêu cầu theo ID (preload User, Branch, ProcessedBy)
	FindByID(ctx context.Context, id uint) (*entity.OvertimeRequest, error)

	// FindAll lấy danh sách yêu cầu có phân trang và lọc
	FindAll(ctx context.Context, filter OvertimeFilter) ([]*entity.OvertimeRequest, int64, error)

	// FindByUserAndDate tìm yêu cầu tăng ca theo user và ngày (unique constraint)
	FindByUserAndDate(ctx context.Context, userID uint, date time.Time) (*entity.OvertimeRequest, error)

	// FindActiveByUserAndDate tìm yêu cầu OT đang active (đã check-in, chưa check-out) cho hôm nay
	FindActiveByUserAndDate(ctx context.Context, userID uint, date time.Time) (*entity.OvertimeRequest, error)

	// AutoRejectExpired chuyển toàn bộ PENDING của tháng cũ sang REJECTED
	AutoRejectExpired(ctx context.Context, beforeMonth time.Time, note string) (int64, error)
}
