package usecase

import (
	"context"
	"time"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	"github.com/hdbank/smart-attendance/internal/domain/repository"
)

// OvertimeCheckInRequest yêu cầu check-in tăng ca từ nhân viên
type OvertimeCheckInRequest struct {
	UserID uint `json:"-"` // Từ JWT
}

// OvertimeCheckOutRequest yêu cầu check-out tăng ca
type OvertimeCheckOutRequest struct {
	UserID     uint `json:"-"`              // Từ JWT
	OvertimeID uint `json:"overtime_id"`    // ID yêu cầu OT (từ check-in)
}

// OvertimeCheckInResponse phản hồi check-in OT kèm thông tin bo tròn
type OvertimeCheckInResponse struct {
	OvertimeRequest *entity.OvertimeRequest `json:"overtime_request"`
	EstimatedStart  *time.Time              `json:"estimated_start"`  // Giờ dự kiến bắt đầu tính OT
	EstimatedEnd    *time.Time              `json:"estimated_end"`    // Giờ dự kiến kết thúc tối đa (22:00)
	Note            string                  `json:"note"`             // Lưu ý quy tắc bo tròn
}

// OvertimeCheckOutResponse phản hồi check-out OT kèm thông tin thời gian dự kiến
type OvertimeCheckOutResponse struct {
	OvertimeRequest *entity.OvertimeRequest `json:"overtime_request"`
	EstimatedStart  *time.Time              `json:"estimated_start"`  // Giờ dự kiến bắt đầu tính OT (sau bo tròn)
	EstimatedEnd    *time.Time              `json:"estimated_end"`    // Giờ dự kiến kết thúc tính OT (sau bo tròn)
	EstimatedHours  float64                 `json:"estimated_hours"`  // Giờ OT dự kiến
	Note            string                  `json:"note"`             // Lưu ý quy tắc bo tròn
}

// ProcessOvertimeRequest yêu cầu duyệt/từ chối từ Manager
type ProcessOvertimeRequest struct {
	OvertimeID    uint                   `json:"-"`            // Từ path param
	ProcessedByID uint                   `json:"-"`            // Từ JWT
	Status        entity.OvertimeStatus  `json:"status"`       // approved hoặc rejected
	ManagerNote   string                 `json:"manager_note"` // Ghi chú từ manager
}

// OvertimeUsecase định nghĩa business logic cho tăng ca
type OvertimeUsecase interface {
	// CheckIn check-in tăng ca (employee) — chỉ cho phép sau 17:00
	CheckIn(ctx context.Context, req OvertimeCheckInRequest) (*OvertimeCheckInResponse, error)

	// CheckOut check-out tăng ca
	CheckOut(ctx context.Context, req OvertimeCheckOutRequest) (*OvertimeCheckOutResponse, error)

	// Process duyệt hoặc từ chối yêu cầu (manager)
	// Khi approved: tính calculated_start, calculated_end, total_hours trong transaction
	Process(ctx context.Context, req ProcessOvertimeRequest) (*entity.OvertimeRequest, error)

	// GetByID lấy chi tiết yêu cầu
	GetByID(ctx context.Context, id uint) (*entity.OvertimeRequest, error)

	// GetList lấy danh sách yêu cầu có phân trang và lọc
	GetList(ctx context.Context, filter repository.OvertimeFilter) ([]*entity.OvertimeRequest, int64, error)

	// GetMyList lấy danh sách yêu cầu của employee
	GetMyList(ctx context.Context, userID uint, status entity.OvertimeStatus, page, limit int) ([]*entity.OvertimeRequest, int64, error)

	// GetMyToday lấy yêu cầu OT hôm nay của employee
	GetMyToday(ctx context.Context, userID uint) (*entity.OvertimeRequest, error)

	// BatchApprove duyệt tất cả yêu cầu PENDING (theo branch nếu có)
	BatchApprove(ctx context.Context, processedByID uint, branchID *uint) (int64, error)

	// AutoRejectExpired tự động reject yêu cầu PENDING của tháng cũ (cron job)
	AutoRejectExpired(ctx context.Context) (int64, error)
}
