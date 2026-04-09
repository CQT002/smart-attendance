package entity

import (
	"time"

	"gorm.io/gorm"
)

// Shift định nghĩa ca làm việc tại chi nhánh
//
// Index strategy:
//   - idx_shift_branch_default_active : FindDefault query — "lấy ca mặc định active của branch X"
//     → chạy mỗi lần check-in để xác định ca + tính trạng thái đi muộn/về sớm
type Shift struct {
	ID       uint   `gorm:"primaryKey;autoIncrement"                                              json:"id"`
	BranchID uint   `gorm:"not null;index:idx_shift_branch_default_active,priority:1"             json:"branch_id"`
	Branch   Branch `gorm:"foreignKey:BranchID"                                                   json:"branch,omitempty"`

	Name        string  `gorm:"size:100;not null" json:"name"`       // VD: "Ca sáng", "Ca chiều"
	StartTime   string  `gorm:"size:5;not null"   json:"start_time"` // HH:MM — VD: "08:00"
	EndTime     string  `gorm:"size:5;not null"   json:"end_time"`   // HH:MM — VD: "17:00"
	LateAfter   int     `gorm:"default:15"        json:"late_after"`  // Phút sau giờ vào mới tính là muộn
	EarlyBefore int     `gorm:"default:15"        json:"early_before"` // Phút trước giờ ra mới tính là về sớm
	WorkHours   float64 `gorm:"type:decimal(5,2)" json:"work_hours"`   // Số giờ làm chuẩn của ca

	// Khung giờ nghỉ trưa — dùng để phân chia buổi sáng/chiều
	MorningEnd     string `gorm:"size:5;default:'12:00'" json:"morning_end"`     // HH:MM — kết thúc buổi sáng
	AfternoonStart string `gorm:"size:5;default:'13:00'" json:"afternoon_start"` // HH:MM — bắt đầu buổi chiều

	// Cấu hình khung giờ làm việc chính thức trong tuần
	// Chấm công chính thức được phép từ Thứ 2 (00:00) → RegularEndDay tại RegularEndTime
	// Encoding giống Go time.Weekday: 0=CN, 1=T2, 2=T3, 3=T4, 4=T5, 5=T6, 6=T7
	RegularEndDay  int    `gorm:"default:6"              json:"regular_end_day"`  // Ngày cuối tuần làm việc (default: 6=T7)
	RegularEndTime string `gorm:"size:5;default:'12:00'" json:"regular_end_time"` // Giờ kết thúc ngày cuối (default: 12:00)

	// Cấu hình tăng ca (OT) theo ca — cho phép khác nhau giữa các ca/chi nhánh
	OTMinCheckInHour int `gorm:"column:ot_min_checkin_hour;default:17" json:"ot_min_checkin_hour"` // Giờ sớm nhất check-in OT (VD: 17 = 17:00)
	OTStartHour      int `gorm:"column:ot_start_hour;default:18"       json:"ot_start_hour"`      // Giờ bắt đầu tính OT (VD: 18 = 18:00)
	OTEndHour        int `gorm:"column:ot_end_hour;default:22"         json:"ot_end_hour"`        // Giờ kết thúc tính OT (VD: 22 = 22:00)

	// idx_shift_branch_default_active priority:2
	IsDefault bool `gorm:"default:false;index:idx_shift_branch_default_active,priority:2" json:"is_default"`

	// idx_shift_branch_default_active priority:3
	IsActive bool `gorm:"default:true;index:idx_shift_branch_default_active,priority:3" json:"is_active"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Shift) TableName() string { return "shifts" }
