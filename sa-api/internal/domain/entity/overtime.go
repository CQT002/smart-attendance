package entity

import (
	"time"

	"gorm.io/gorm"
)

// OvertimeStatus trạng thái yêu cầu tăng ca
type OvertimeStatus string

const (
	OvertimeStatusInit     OvertimeStatus = "init"     // Đã check-in hoặc check-out (thiếu 1 trong 2)
	OvertimeStatusPending  OvertimeStatus = "pending"  // Đã có đủ check-in + check-out, chờ duyệt
	OvertimeStatusApproved OvertimeStatus = "approved" // Đã duyệt
	OvertimeStatusRejected OvertimeStatus = "rejected" // Từ chối
)

// OvertimeRequest yêu cầu tăng ca (OT)
//
// Business rules:
//   - Check-in OT chỉ được phép sau 17:00
//   - Giờ tính OT bắt đầu từ 18:00, kết thúc tối đa 22:00
//   - Thời gian tối đa: 4 giờ/ngày (18:00 - 22:00)
//   - Bo tròn bắt đầu: check-in trong [17:00-18:00] → tính từ 18:00
//   - Bo tròn kết thúc: check-out sau 22:00 → tính đến 22:00
//   - Giờ OT chỉ cộng vào quỹ lương khi Manager Approve
//   - Khi Approve, backend tính: calculated_start, calculated_end, total_hours
//
// Index strategy:
//   - idx_overtime_user_status: query "yêu cầu OT của tôi" filter theo status
//   - idx_overtime_branch_status: query "danh sách chờ duyệt" của manager
//   - idx_overtime_user_date: unique constraint — mỗi user chỉ 1 đơn OT/ngày
//   - idx_overtime_created_at: sort và filter theo thời gian tạo
type OvertimeRequest struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	// Nhân viên yêu cầu tăng ca
	UserID uint `gorm:"not null;index:idx_overtime_user_status,priority:1;index:idx_overtime_user_date,priority:1" json:"user_id"`
	User   User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// Chi nhánh — denormalize để manager query nhanh
	BranchID uint   `gorm:"not null;index:idx_overtime_branch_status,priority:1" json:"branch_id"`
	Branch   Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`

	// Ngày tăng ca
	Date time.Time `gorm:"type:date;not null;uniqueIndex:idx_overtime_user_date,priority:2" json:"date"`

	// Thời gian thực tế check-in/out OT (timestamptz)
	ActualCheckin  *time.Time `gorm:"type:timestamptz" json:"actual_checkin"`
	ActualCheckout *time.Time `gorm:"type:timestamptz" json:"actual_checkout"`

	// Thời gian hệ thống tính sau khi bo tròn (tính khi Manager duyệt)
	CalculatedStart *time.Time `gorm:"type:timestamptz" json:"calculated_start"`
	CalculatedEnd   *time.Time `gorm:"type:timestamptz" json:"calculated_end"`

	// Tổng giờ OT (calculated_end - calculated_start), dạng decimal
	TotalHours float64 `gorm:"type:decimal(5,2);default:0" json:"total_hours"`

	// Trạng thái: init → pending → approved/rejected
	Status OvertimeStatus `gorm:"type:varchar(20);not null;default:'init';index:idx_overtime_user_status,priority:2;index:idx_overtime_branch_status,priority:2" json:"status"`

	// === Audit Log ===
	// Người duyệt (Manager)
	ManagerID   *uint `json:"manager_id"`
	ProcessedBy *User `gorm:"foreignKey:ManagerID" json:"processed_by,omitempty"`

	// Thời điểm duyệt
	ProcessedAt *time.Time `json:"processed_at"`

	// Ghi chú từ manager khi duyệt/từ chối
	ManagerNote string `gorm:"type:text" json:"manager_note"`

	CreatedAt time.Time `gorm:"index:idx_overtime_created_at" json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (OvertimeRequest) TableName() string { return "overtime_requests" }

// IsInit kiểm tra đang ở trạng thái init (đã check-in, chưa check-out)
func (o *OvertimeRequest) IsInit() bool { return o.Status == OvertimeStatusInit }

// IsPending kiểm tra yêu cầu còn chờ duyệt không
func (o *OvertimeRequest) IsPending() bool { return o.Status == OvertimeStatusPending }

// IsProcessed kiểm tra yêu cầu đã được xử lý chưa
func (o *OvertimeRequest) IsProcessed() bool {
	return o.Status == OvertimeStatusApproved || o.Status == OvertimeStatusRejected
}

// IsCheckedIn kiểm tra đã check-in OT chưa
func (o *OvertimeRequest) IsCheckedIn() bool { return o.ActualCheckin != nil }

// IsCheckedOut kiểm tra đã check-out OT chưa
func (o *OvertimeRequest) IsCheckedOut() bool { return o.ActualCheckout != nil }
