package entity

import "time"

// Branch đại diện cho một chi nhánh trong hệ thống
type Branch struct {
	ID       uint   `gorm:"primaryKey;autoIncrement"                         json:"id"`
	Code     string `gorm:"uniqueIndex:uniq_branch_code;size:20;not null"     json:"code"`
	Name     string `gorm:"size:200;not null"                                json:"name"`
	Address  string `gorm:"size:500"                                         json:"address"`
	City     string `gorm:"size:100;index:idx_branch_city_active,priority:1" json:"city"`
	Province string `gorm:"size:100"                                         json:"province"`
	Phone    string `gorm:"size:20"                                          json:"phone"`
	Email    string `gorm:"size:100"                                         json:"email"`

	// Tọa độ GPS trụ sở chi nhánh (dùng để hiển thị map, khác với gps_configs geofencing)
	Latitude  *float64 `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude *float64 `gorm:"type:decimal(11,8)" json:"longitude"`

	// idx_branch_city_active: dashboard filter theo tỉnh/thành phố + is_active
	IsActive bool `gorm:"default:true;index:idx_branch_city_active,priority:2" json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Users          []User          `gorm:"foreignKey:BranchID" json:"-"`
	WiFiConfigs    []WiFiConfig    `gorm:"foreignKey:BranchID" json:"-"`
	GPSConfigs     []GPSConfig     `gorm:"foreignKey:BranchID" json:"-"`
	Shifts         []Shift         `gorm:"foreignKey:BranchID" json:"-"`
	AttendanceLogs []AttendanceLog `gorm:"foreignKey:BranchID" json:"-"`
	DailySummaries []DailySummary  `gorm:"foreignKey:BranchID" json:"-"`
}

func (Branch) TableName() string { return "branches" }
