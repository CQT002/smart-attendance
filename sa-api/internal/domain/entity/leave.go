package entity

import "time"

// LeaveType loại nghỉ phép
type LeaveType string

const (
	LeaveTypeFullDay          LeaveType = "full_day"           // Nghỉ cả ngày (08:00 - 17:00)
	LeaveTypeHalfDayMorning   LeaveType = "half_day_morning"   // Nghỉ buổi sáng (08:00 - 12:00)
	LeaveTypeHalfDayAfternoon LeaveType = "half_day_afternoon" // Nghỉ buổi chiều (13:00 - 17:00)
)

// LeaveStatus trạng thái yêu cầu nghỉ phép (reuse CorrectionStatus values)
type LeaveStatus string

const (
	LeaveStatusPending  LeaveStatus = "pending"
	LeaveStatusApproved LeaveStatus = "approved"
	LeaveStatusRejected LeaveStatus = "rejected"
)

// LeaveRequest yêu cầu nghỉ phép
//
// Business rules:
//   - Cho phép chọn bất kỳ ngày nào trong tháng (quá khứ, hiện tại, tương lai)
//   - Ngày quá khứ status=absent → mặc định full_day
//   - Ngày quá khứ status=half_day → nghỉ nửa ngày còn lại
//   - Ngày hiện tại & tương lai → chọn khung giờ: full_day, half_day_morning, half_day_afternoon
//   - Manager chi nhánh duyệt, không được tự duyệt
//   - Khi approved: tạo/cập nhật attendance_log với status=leave
//   - Auto-reject PENDING tháng cũ vào ngày 1 hàng tháng
//
// Index strategy:
//   - idx_leave_user_status: query "yêu cầu của tôi" filter theo status
//   - idx_leave_branch_status: query "danh sách chờ duyệt" của manager
//   - idx_leave_user_date: unique constraint — mỗi user chỉ 1 đơn nghỉ phép/ngày
//   - idx_leave_created_at: sort và filter theo thời gian tạo
type LeaveRequest struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	// Nhân viên yêu cầu nghỉ phép
	UserID uint `gorm:"not null;index:idx_leave_user_status,priority:1;index:idx_leave_user_date,priority:1" json:"user_id"`
	User   User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// Chi nhánh — denormalize để manager query nhanh
	BranchID uint   `gorm:"not null;index:idx_leave_branch_status,priority:1" json:"branch_id"`
	Branch   Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`

	// Ngày nghỉ phép
	LeaveDate time.Time `gorm:"type:date;not null;uniqueIndex:idx_leave_user_date,priority:2" json:"leave_date"`

	// Loại nghỉ phép
	LeaveType LeaveType `gorm:"type:varchar(30);not null" json:"leave_type"`

	// Khung giờ nghỉ
	TimeFrom string `gorm:"type:varchar(5);not null" json:"time_from"` // HH:MM format
	TimeTo   string `gorm:"type:varchar(5);not null" json:"time_to"`   // HH:MM format

	// Trạng thái gốc của ngày (nếu là ngày quá khứ: absent, half_day; nếu tương lai: rỗng)
	OriginalStatus AttendanceStatus `gorm:"type:varchar(20)" json:"original_status"`

	// Lý do xin nghỉ phép
	Description string `gorm:"type:text;not null" json:"description"`

	// Trạng thái duyệt
	Status LeaveStatus `gorm:"type:varchar(20);not null;default:'pending';index:idx_leave_user_status,priority:2;index:idx_leave_branch_status,priority:2" json:"status"`

	// === Audit Log ===
	ProcessedByID *uint `json:"processed_by_id"`
	ProcessedBy   *User `gorm:"foreignKey:ProcessedByID" json:"processed_by,omitempty"`

	ProcessedAt *time.Time `json:"processed_at"`

	ManagerNote string `gorm:"type:text" json:"manager_note"`

	CreatedAt time.Time `gorm:"index:idx_leave_created_at" json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (LeaveRequest) TableName() string { return "leave_requests" }

// IsPending kiểm tra yêu cầu còn chờ duyệt không
func (l *LeaveRequest) IsPending() bool { return l.Status == LeaveStatusPending }

// IsProcessed kiểm tra yêu cầu đã được xử lý chưa
func (l *LeaveRequest) IsProcessed() bool {
	return l.Status == LeaveStatusApproved || l.Status == LeaveStatusRejected
}

// IsLeavableStatus kiểm tra trạng thái attendance có được phép xin nghỉ phép cho ngày quá khứ
func IsLeavableStatus(status AttendanceStatus) bool {
	return status == StatusAbsent || status == StatusHalfDay
}
