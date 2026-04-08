package entity

import (
	"time"

	"gorm.io/gorm"
)

// Scheduler định nghĩa một tác vụ lập lịch chạy định kỳ
//
// Cho phép admin quản lý (bật/tắt, đổi lịch) mà không cần redeploy.
// CronExpr dùng chuẩn cron 5 field: minute hour day month weekday
//
// Index strategy:
//   - idx_scheduler_name: unique lookup theo tên
//   - idx_scheduler_active: lọc nhanh scheduler đang bật
type Scheduler struct {
	ID          uint   `gorm:"primaryKey;autoIncrement"                        json:"id"`
	Name        string `gorm:"size:100;not null;uniqueIndex:idx_scheduler_name" json:"name"`        // VD: "leave_accrual", "correction_auto_reject"
	Description string `gorm:"size:500"                                        json:"description"`  // Mô tả ngắn gọn
	CronExpr    string `gorm:"size:100;not null"                               json:"cron_expr"`    // Cron expression 5-field: "30 0 1 * *"
	IsActive    bool   `gorm:"default:true;index:idx_scheduler_active"         json:"is_active"`    // Bật/tắt scheduler
	TimeoutSec  int    `gorm:"default:30"                                      json:"timeout_sec"`  // Context timeout (giây)

	// Trạng thái chạy gần nhất
	LastRunAt  *time.Time `json:"last_run_at"`
	LastStatus string     `gorm:"size:20" json:"last_status"` // "success", "failed"
	LastError  string     `gorm:"type:text" json:"last_error"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Scheduler) TableName() string { return "schedulers" }
