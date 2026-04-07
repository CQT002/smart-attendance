package repository

import (
	"context"
	"time"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
)

// CorrectionFilter bộ lọc yêu cầu chấm công bù
type CorrectionFilter struct {
	UserID   *uint
	BranchID *uint
	Status   entity.CorrectionStatus
	Page     int
	Limit    int
}

// CorrectionRepository định nghĩa contract cho thao tác dữ liệu chấm công bù
type CorrectionRepository interface {
	// Create tạo yêu cầu chấm công bù mới
	Create(ctx context.Context, correction *entity.AttendanceCorrection) error

	// Update cập nhật yêu cầu chấm công bù
	Update(ctx context.Context, correction *entity.AttendanceCorrection) error

	// FindByID tìm yêu cầu theo ID (preload User, AttendanceLog, ProcessedBy)
	FindByID(ctx context.Context, id uint) (*entity.AttendanceCorrection, error)

	// FindAll lấy danh sách yêu cầu có phân trang và lọc
	FindAll(ctx context.Context, filter CorrectionFilter) ([]*entity.AttendanceCorrection, int64, error)

	// CountByUserInMonth đếm số yêu cầu của user trong tháng hiện tại (dựa trên created_at)
	CountByUserInMonth(ctx context.Context, userID uint, month time.Time) (int64, error)

	// FindByAttendanceLogID tìm yêu cầu theo attendance_log_id (unique constraint)
	FindByAttendanceLogID(ctx context.Context, logID uint) (*entity.AttendanceCorrection, error)

	// AutoRejectExpired chuyển toàn bộ PENDING của tháng cũ sang REJECTED
	// Trả về số bản ghi bị ảnh hưởng
	AutoRejectExpired(ctx context.Context, beforeMonth time.Time, note string) (int64, error)
}
