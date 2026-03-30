package entity

import "time"

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

	// idx_shift_branch_default_active priority:2
	IsDefault bool `gorm:"default:false;index:idx_shift_branch_default_active,priority:2" json:"is_default"`

	// idx_shift_branch_default_active priority:3
	IsActive bool `gorm:"default:true;index:idx_shift_branch_default_active,priority:3" json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Shift) TableName() string { return "shifts" }
