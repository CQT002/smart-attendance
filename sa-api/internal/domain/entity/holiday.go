package entity

import (
	"time"

	"gorm.io/gorm"
)

// HolidayType loại ngày lễ
type HolidayType string

const (
	HolidayTypeNational HolidayType = "national" // Ngày lễ quốc gia (Tết, 30/4, 1/5, 2/9...)
	HolidayTypeCompany  HolidayType = "company"  // Ngày nghỉ nội bộ công ty
)

// Default coefficient theo loại — admin có thể override mỗi holiday.
// Căn cứ Luật Lao động VN 2019: làm việc ngày lễ quốc gia hưởng 300%,
// ngày thường của công ty thường là 200%.
const (
	DefaultCoefficientNational = 3.00 // 300%
	DefaultCoefficientCompany  = 2.00 // 200%
)

// Holiday ngày lễ hệ thống
//
// Business rules:
//   - Mỗi ngày chỉ có 1 bản ghi holiday active (unique index trên date)
//   - Coefficient áp dụng cho nhân viên làm việc trong ngày lễ (Lương ngày lễ = Lương ngày thường × coefficient)
//   - Nhân viên không làm việc vào ngày lễ → được hưởng "nghỉ lễ hưởng lương" (paid_holiday)
//   - is_compensated = true: đây là ngày nghỉ bù khi lễ gốc rơi vào cuối tuần; compensate_for trỏ đến lễ gốc
//
// Index strategy:
//   - uniq_holiday_date      : một ngày chỉ có 1 holiday active (partial: WHERE deleted_at IS NULL)
//   - idx_holiday_year       : list theo năm (GET /admin/holidays?year=2026)
//   - idx_holiday_year_date  : range query theo khoảng thời gian trong năm
type Holiday struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	// Tên ngày lễ — VD: "Tết Dương lịch", "Giỗ Tổ Hùng Vương"
	Name string `gorm:"size:200;not null" json:"name"`

	// Ngày lễ (date only, không có giờ)
	Date time.Time `gorm:"type:date;not null;index:idx_holiday_year_date,priority:2" json:"date"`

	// Năm — denormalize từ date để query nhanh theo năm
	Year int `gorm:"not null;index:idx_holiday_year;index:idx_holiday_year_date,priority:1" json:"year"`

	// Hệ số lương cho nhân viên làm việc ngày lễ — VD 2.00 = 200%, 3.00 = 300%, 4.00 = 400%
	Coefficient float64 `gorm:"type:decimal(4,2);not null;default:3.00" json:"coefficient"`

	// Loại lễ
	Type HolidayType `gorm:"type:varchar(20);not null;default:'national'" json:"type"`

	// Cờ đánh dấu đây là ngày nghỉ bù (khi lễ gốc rơi vào Thứ 7 / Chủ nhật)
	IsCompensated bool `gorm:"default:false" json:"is_compensated"`

	// Ngày gốc đang được nghỉ bù — bắt buộc khi is_compensated = true
	CompensateFor *time.Time `gorm:"type:date" json:"compensate_for,omitempty"`

	// Mô tả thêm
	Description string `gorm:"size:500" json:"description,omitempty"`

	// Audit — ai tạo (null với bản ghi seed hệ thống)
	CreatedByID *uint `json:"created_by_id,omitempty"`
	CreatedBy   *User `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Holiday) TableName() string { return "holidays" }

// Valid kiểm tra holiday có hợp lệ không
func (h *Holiday) Valid() bool {
	if h.Name == "" || h.Coefficient <= 0 || h.Coefficient > 10 {
		return false
	}
	if h.Type != HolidayTypeNational && h.Type != HolidayTypeCompany {
		return false
	}
	if h.IsCompensated && h.CompensateFor == nil {
		return false
	}
	return true
}

// DefaultCoefficientFor trả về hệ số mặc định theo loại lễ
func DefaultCoefficientFor(t HolidayType) float64 {
	switch t {
	case HolidayTypeCompany:
		return DefaultCoefficientCompany
	default:
		return DefaultCoefficientNational
	}
}
