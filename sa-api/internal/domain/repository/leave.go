package repository

import (
	"context"
	"time"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
)

// LeaveFilter bộ lọc yêu cầu nghỉ phép
type LeaveFilter struct {
	UserID   *uint
	BranchID *uint
	Status   entity.LeaveStatus
	Page     int
	Limit    int
}

// LeaveRepository định nghĩa contract cho thao tác dữ liệu nghỉ phép
type LeaveRepository interface {
	// Create tạo yêu cầu nghỉ phép mới
	Create(ctx context.Context, leave *entity.LeaveRequest) error

	// Update cập nhật yêu cầu nghỉ phép
	Update(ctx context.Context, leave *entity.LeaveRequest) error

	// FindByID tìm yêu cầu theo ID (preload User, Branch, ProcessedBy)
	FindByID(ctx context.Context, id uint) (*entity.LeaveRequest, error)

	// FindAll lấy danh sách yêu cầu có phân trang và lọc
	FindAll(ctx context.Context, filter LeaveFilter) ([]*entity.LeaveRequest, int64, error)

	// FindByUserAndDate tìm yêu cầu nghỉ phép theo user và ngày (unique constraint)
	FindByUserAndDate(ctx context.Context, userID uint, date time.Time) (*entity.LeaveRequest, error)

	// CountPendingByBranch đếm số yêu cầu pending theo chi nhánh
	CountPendingByBranch(ctx context.Context, branchID uint) (int64, error)

	// AutoRejectExpired chuyển toàn bộ PENDING của tháng cũ sang REJECTED
	AutoRejectExpired(ctx context.Context, beforeMonth time.Time, note string) (int64, error)
}
