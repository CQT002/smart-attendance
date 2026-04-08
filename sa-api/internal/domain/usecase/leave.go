package usecase

import (
	"context"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	"github.com/hdbank/smart-attendance/internal/domain/repository"
)

// CreateLeaveRequest yêu cầu tạo nghỉ phép từ nhân viên
type CreateLeaveRequest struct {
	UserID      uint             `json:"-"`           // Từ JWT
	LeaveDate   string           `json:"leave_date"`  // YYYY-MM-DD
	LeaveType   entity.LeaveType `json:"leave_type"`  // full_day, half_day_morning, half_day_afternoon
	Description string           `json:"description"` // Lý do xin nghỉ phép
}

// ProcessLeaveRequest yêu cầu duyệt/từ chối từ Manager
type ProcessLeaveRequest struct {
	LeaveID       uint               `json:"-"`            // Từ path param
	ProcessedByID uint               `json:"-"`            // Từ JWT
	Status        entity.LeaveStatus `json:"status"`       // approved hoặc rejected
	ManagerNote   string             `json:"manager_note"` // Ghi chú từ manager
}

// PendingApprovalItem item chung cho dashboard duyệt (correction + leave)
type PendingApprovalItem struct {
	ID           uint   `json:"id"`
	Type         string `json:"type"` // "correction" hoặc "leave"
	UserID       uint   `json:"user_id"`
	UserName     string `json:"user_name"`
	EmployeeCode string `json:"employee_code"`
	Department   string `json:"department"`
	BranchID     uint   `json:"branch_id"`
	Date         string `json:"date"`          // Ngày liên quan (YYYY-MM-DD)
	Description  string `json:"description"`   // Lý do
	Detail       string `json:"detail"`        // Chi tiết bổ sung (loại nghỉ phép / trạng thái gốc)
	Status       string `json:"status"`        // pending, approved, rejected
	CreatedAt    string `json:"created_at"`

	// Audit fields — có giá trị khi status != pending
	ProcessedByName string  `json:"processed_by_name,omitempty"`
	ProcessedAt     *string `json:"processed_at,omitempty"`
	ManagerNote     string  `json:"manager_note,omitempty"`

	// Extra fields cho từng loại
	CheckInTime  *string `json:"check_in_time,omitempty"`  // correction only
	CheckOutTime *string `json:"check_out_time,omitempty"` // correction only
	LeaveType    string  `json:"leave_type,omitempty"`     // leave only: full_day, half_day_morning, ...
	TimeFrom     string  `json:"time_from,omitempty"`      // leave only
	TimeTo       string  `json:"time_to,omitempty"`        // leave only

	// Overtime-specific fields
	ActualCheckin   *string  `json:"actual_checkin,omitempty"`   // overtime only
	ActualCheckout  *string  `json:"actual_checkout,omitempty"`  // overtime only
	CalculatedStart *string  `json:"calculated_start,omitempty"` // overtime only (sau khi duyệt)
	CalculatedEnd   *string  `json:"calculated_end,omitempty"`   // overtime only (sau khi duyệt)
	TotalHours      *float64 `json:"total_hours,omitempty"`      // overtime only (sau khi duyệt)
}

// LeaveUsecase định nghĩa business logic cho nghỉ phép
type LeaveUsecase interface {
	// Create tạo yêu cầu nghỉ phép (employee)
	Create(ctx context.Context, req CreateLeaveRequest) (*entity.LeaveRequest, error)

	// Process duyệt hoặc từ chối yêu cầu (manager)
	// Khi approved: tạo/cập nhật AttendanceLog với status=leave trong transaction
	Process(ctx context.Context, req ProcessLeaveRequest) (*entity.LeaveRequest, error)

	// GetByID lấy chi tiết yêu cầu
	GetByID(ctx context.Context, id uint) (*entity.LeaveRequest, error)

	// GetList lấy danh sách yêu cầu có phân trang và lọc
	GetList(ctx context.Context, filter repository.LeaveFilter) ([]*entity.LeaveRequest, int64, error)

	// GetMyList lấy danh sách yêu cầu của employee
	GetMyList(ctx context.Context, userID uint, status entity.LeaveStatus, page, limit int) ([]*entity.LeaveRequest, int64, error)

	// BatchApprove duyệt tất cả yêu cầu nghỉ phép PENDING (theo branch nếu có)
	BatchApprove(ctx context.Context, processedByID uint, branchID *uint) (int64, error)

	// GetPendingApprovals lấy danh sách tổng hợp chờ duyệt (correction + leave)
	GetPendingApprovals(ctx context.Context, branchID *uint, page, limit int) ([]PendingApprovalItem, int64, error)

	// GetApprovals lấy danh sách tổng hợp duyệt chấm công (correction + leave) — hỗ trợ lọc theo status
	GetApprovals(ctx context.Context, branchID *uint, status string, page, limit int) ([]PendingApprovalItem, int64, error)

	// AutoRejectExpired tự động reject yêu cầu PENDING của tháng cũ
	AutoRejectExpired(ctx context.Context) (int64, error)

	// AccrueMonthlyLeave cộng 1 ngày phép cho tất cả user active (chạy đầu tháng)
	AccrueMonthlyLeave(ctx context.Context) (int64, error)
}
