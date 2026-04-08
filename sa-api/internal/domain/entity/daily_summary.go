package entity

import (
	"time"

	"gorm.io/gorm"
)

// DailySummary là bảng tổng hợp chấm công pre-computed theo chi nhánh mỗi ngày.
//
// Mục đích: thay thế aggregate query nặng trên attendance_logs cho dashboard/báo cáo.
// Với 5000 nhân viên, việc COUNT/GROUP BY trực tiếp trên attendance_logs mỗi lần
// load dashboard sẽ scan hàng nghìn rows — DailySummary giải quyết điều này bằng
// cách lưu kết quả đã tính toán sẵn, cập nhật cuối mỗi ngày (hoặc theo sự kiện).
//
// Index strategy:
//   - uniq_daily_branch_date     : mỗi chi nhánh chỉ có 1 bản ghi/ngày (primary key logic)
//     → dùng cho Upsert hàng ngày (INSERT ... ON CONFLICT DO UPDATE)
//   - idx_daily_branch_date_range: range scan theo ngày — "30 ngày gần nhất của branch X"
//     → dùng cho báo cáo xu hướng, biểu đồ attendance rate
//   - idx_daily_date             : admin xem toàn hệ thống theo ngày
//     → dùng cho GetDashboardStats — aggregator tất cả branches ngày hôm nay
type DailySummary struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	// uniq_daily_branch_date (priority:1) | idx_daily_branch_date_range (priority:1)
	BranchID uint   `gorm:"not null;uniqueIndex:uniq_daily_branch_date,priority:1;index:idx_daily_branch_date_range,priority:1" json:"branch_id"`
	Branch   Branch `gorm:"foreignKey:BranchID"                                                                                json:"-"`

	// uniq_daily_branch_date (priority:2) | idx_daily_branch_date_range (priority:2) | idx_daily_date
	Date time.Time `gorm:"type:date;not null;uniqueIndex:uniq_daily_branch_date,priority:2;index:idx_daily_branch_date_range,priority:2;index:idx_daily_date" json:"date"`

	// === Headcount ===
	TotalEmployees  int `gorm:"not null;default:0" json:"total_employees"`  // Tổng nhân viên active tại chi nhánh
	PresentCount    int `gorm:"not null;default:0" json:"present_count"`    // Số người có mặt đúng giờ
	LateCount       int `gorm:"not null;default:0" json:"late_count"`       // Số người đi muộn
	EarlyLeaveCount int `gorm:"not null;default:0" json:"early_leave_count"` // Số người về sớm
	HalfDayCount    int `gorm:"not null;default:0" json:"half_day_count"`   // Số người làm nửa ngày
	AbsentCount     int `gorm:"not null;default:0" json:"absent_count"`     // Số người vắng mặt (= TotalEmployees - checkin count)

	// === Giờ làm ===
	TotalWorkHours float64 `gorm:"type:decimal(10,2);default:0" json:"total_work_hours"` // Tổng giờ làm thực tế
	TotalOvertime  float64 `gorm:"type:decimal(10,2);default:0" json:"total_overtime"`   // Tổng giờ làm thêm

	// === Anti-fraud ===
	FraudCount int `gorm:"not null;default:0" json:"fraud_count"` // Số bản ghi bị đánh dấu gian lận

	// === KPI ===
	// AttendanceRate = (PresentCount + LateCount + EarlyLeaveCount + HalfDayCount) / TotalEmployees * 100
	AttendanceRate float64 `gorm:"type:decimal(5,2);default:0" json:"attendance_rate"` // 0.00 - 100.00 (%)
	// OnTimeRate = PresentCount / (TotalEmployees - AbsentCount) * 100
	OnTimeRate float64 `gorm:"type:decimal(5,2);default:0" json:"on_time_rate"` // 0.00 - 100.00 (%)

	// Thời điểm tính toán lần cuối — dùng để biết data có stale không
	ComputedAt time.Time `gorm:"not null" json:"computed_at"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DailySummary) TableName() string { return "daily_summaries" }

// IsStale kiểm tra bản ghi đã quá 1 giờ kể từ lần tính toán gần nhất
func (d *DailySummary) IsStale() bool {
	return time.Since(d.ComputedAt) > time.Hour
}

// CheckedInCount trả về tổng số người đã chấm công (không tính absent)
func (d *DailySummary) CheckedInCount() int {
	return d.PresentCount + d.LateCount + d.EarlyLeaveCount + d.HalfDayCount
}
