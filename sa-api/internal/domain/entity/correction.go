package entity

import "time"

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
//   - Mỗi nhân viên tối đa 4 lần/tháng (đếm theo SUM(credit_count))
//   - late hoặc early_leave → credit_count = 1
//   - late_early_leave (đi trễ + về sớm) → credit_count = 2
//   - Chỉ được đăng ký bù cho ngày có status: late, early_leave, late_early_leave
//   - Manager chi nhánh là người duyệt, không được tự duyệt cho mình
//   - Auto-reject: PENDING của tháng cũ bị reject lúc 00:05 ngày 1 hàng tháng
//
// Index strategy:
//   - idx_correction_user_status: query "yêu cầu của tôi" filter theo status
//   - idx_correction_branch_status: query "danh sách chờ duyệt" của manager
//   - idx_correction_created_at: đếm hạn mức 4 lần/tháng
type AttendanceCorrection struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	// Nhân viên yêu cầu bù
	UserID uint `gorm:"not null;index:idx_correction_user_status,priority:1;index:idx_correction_created_at,priority:1" json:"user_id"`
	User   User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// Chi nhánh — denormalize để manager query nhanh
	BranchID uint   `gorm:"not null;index:idx_correction_branch_status,priority:1" json:"branch_id"`
	Branch   Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`

	// Attendance log gốc cần bù
	AttendanceLogID uint          `gorm:"not null;uniqueIndex:uniq_correction_log" json:"attendance_log_id"`
	AttendanceLog   AttendanceLog `gorm:"foreignKey:AttendanceLogID" json:"attendance_log,omitempty"`

	// Trạng thái gốc của ngày cần bù (late, early_leave, late_early_leave)
	OriginalStatus AttendanceStatus `gorm:"type:varchar(20);not null" json:"original_status"`

	// Số lần tính hạn mức: late/early_leave = 1, late_early_leave = 2
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
	CreatedAt time.Time `gorm:"index:idx_correction_created_at,priority:2" json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (AttendanceCorrection) TableName() string { return "attendance_corrections" }

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
