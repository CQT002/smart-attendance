package entity

import "time"

// GPSConfig định nghĩa vùng địa lý (geofence) cho phép chấm công tại chi nhánh.
// Một chi nhánh có thể có nhiều geofence (tầng 1, tầng 5, bãi xe...).
//
// Index strategy:
//   - idx_gps_branch_active : FindActiveBranch query — "lấy tất cả geofence active của branch X"
//     → hot path: chạy mỗi lần check-in khi WiFi không hợp lệ
type GPSConfig struct {
	ID       uint   `gorm:"primaryKey;autoIncrement"                               json:"id"`
	BranchID uint   `gorm:"not null;index:idx_gps_branch_active,priority:1"        json:"branch_id"`
	Branch   Branch `gorm:"foreignKey:BranchID"                                    json:"branch,omitempty"`

	Name        string  `gorm:"size:100"                                    json:"name"`        // VD: "Văn phòng tầng 3"
	Latitude    float64 `gorm:"type:decimal(10,8);not null"                 json:"latitude"`
	Longitude   float64 `gorm:"type:decimal(11,8);not null"                 json:"longitude"`
	Radius      float64 `gorm:"type:decimal(8,2);not null"                  json:"radius"`      // Bán kính cho phép (mét)
	Description string  `gorm:"size:200"                                    json:"description"`

	// idx_gps_branch_active priority:2
	IsActive bool `gorm:"default:true;index:idx_gps_branch_active,priority:2" json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (GPSConfig) TableName() string { return "gps_configs" }
