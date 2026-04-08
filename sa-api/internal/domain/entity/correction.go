package entity

import (
	"time"

	"gorm.io/gorm"
)

// CorrectionType loại chấm công bù
type CorrectionType string

const (
	CorrectionTypeAttendance CorrectionType = "attendance" // Bù công ca chính thức
	CorrectionTypeOvertime   CorrectionType = "overtime"   // Bù công tăng ca
)

// CorrectionStatus trạng thái yêu cầu chấm công bù
type CorrectionStatus string

const (
	CorrectionStatusPending  CorrectionStatus = "pending"  // Chờ duyệt
	CorrectionStatusApproved CorrectionStatus = "approved" // Đã duyệt
	CorrectionStatusRejected CorrectionStatus = "rejected" // Từ chối
)

// AttendanceCorrection yêu cầu chấm công bù
//
// Business rules:
//   - Mỗi nhân viên tối đa 4 lần/tháng (đếm theo SUM(credit_count)) — riêng biệt cho attendance và overtime
//   - Attendance: late/early_leave = 1 credit, late_early_leave = 2 credits
//   - Overtime: thiếu check-in hoặc check-out = 1 credit
//   - Manager chi nhánh là người duyệt, không được tự duyệt cho mình
//   - Khi duyệt correction loại overtime → đồng thời approve OvertimeRequest
//   - Auto-reject: PENDING của tháng cũ bị reject lúc 00:05 ngày 1 hàng tháng
//
// Index strategy:
//   - idx_correction_user_status: query "yêu cầu của tôi" filter theo status
//   - idx_correction_branch_status: query "danh sách chờ duyệt" của manager
//   - idx_correction_created_at: đếm hạn mức 4 lần/tháng
type AttendanceCorrection struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	// Loại bù công: attendance (ca chính thức) hoặc overtime (tăng ca)
	CorrectionType CorrectionType `gorm:"type:varchar(20);not null;default:'attendance'" json:"correction_type"`

	// Nhân viên yêu cầu bù
	UserID uint `gorm:"not null;index:idx_correction_user_status,priority:1;index:idx_correction_created_at,priority:1" json:"user_id"`
	User   User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// Chi nhánh — denormalize để manager query nhanh
	BranchID uint   `gorm:"not null;index:idx_correction_branch_status,priority:1" json:"branch_id"`
	Branch   Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`

	// Attendance log gốc cần bù (dùng cho correction_type = attendance)
	AttendanceLogID *uint         `gorm:"uniqueIndex:uniq_correction_log" json:"attendance_log_id"`
	AttendanceLog   AttendanceLog `gorm:"foreignKey:AttendanceLogID" json:"attendance_log,omitempty"`

	// Overtime request gốc cần bù (dùng cho correction_type = overtime)
	OvertimeRequestID *uint            `gorm:"index" json:"overtime_request_id"`
	OvertimeRequest   *OvertimeRequest `gorm:"foreignKey:OvertimeRequestID" json:"overtime_request,omitempty"`

	// Trạng thái gốc của ngày cần bù (late, early_leave, late_early_leave cho attendance;
	// missing_checkin, missing_checkout cho overtime)
	OriginalStatus AttendanceStatus `gorm:"type:varchar(20);not null" json:"original_status"`

	// Số lần tính hạn mức
	CreditCount int `gorm:"not null;default:1" json:"credit_count"`

	// Lý do xin bù công từ nhân viên
	Description string `gorm:"type:text;not null" json:"description"`

	// Trạng thái duyệt
	Status CorrectionStatus `gorm:"type:varchar(20);not null;default:'pending';index:idx_correction_user_status,priority:2;index:idx_correction_branch_status,priority:2" json:"status"`

	// === Audit Log ===
	// Người duyệt (Manager hoặc System)
	ProcessedByID *uint `json:"processed_by_id"`
	ProcessedBy   *User `gorm:"foreignKey:ProcessedByID" json:"processed_by,omitempty"`

	// Thời điểm duyệt
	ProcessedAt *time.Time `json:"processed_at"`

	// Ghi chú từ manager khi duyệt/từ chối
	ManagerNote string `gorm:"type:text" json:"manager_note"`

	// Index cho đếm hạn mức theo tháng
	CreatedAt time.Time  `gorm:"index:idx_correction_created_at,priority:2" json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AttendanceCorrection) TableName() string { return "attendance_corrections" }

// IsOvertime kiểm tra có phải bù công tăng ca không
func (c *AttendanceCorrection) IsOvertime() bool { return c.CorrectionType == CorrectionTypeOvertime }

// IsPending kiểm tra yêu cầu còn chờ duyệt không
func (c *AttendanceCorrection) IsPending() bool { return c.Status == CorrectionStatusPending }

// IsProcessed kiểm tra yêu cầu đã được xử lý chưa
func (c *AttendanceCorrection) IsProcessed() bool {
	return c.Status == CorrectionStatusApproved || c.Status == CorrectionStatusRejected
}

// CreditCountForStatus trả về số lần tính hạn mức theo trạng thái gốc
func CreditCountForStatus(status AttendanceStatus) int {
	if status == StatusLateEarlyLeave {
		return 2 // Đi trễ + Về sớm = 2 lần
	}
	return 1 // late hoặc early_leave = 1 lần
}
