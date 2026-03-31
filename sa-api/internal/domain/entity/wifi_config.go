package entity

import "time"

// WiFiConfig cấu hình WiFi được phép chấm công tại chi nhánh
//
// Index strategy:
//   - idx_wifi_branch_active : ValidateWiFi query — "lấy tất cả WiFi active của branch X"
//     → hot path: chạy mỗi lần check-in
//   - idx_wifi_bssid         : lookup nhanh theo BSSID (MAC address router)
//     → BSSID là định danh chính xác nhất, unique trên thực tế
type WiFiConfig struct {
	ID          uint   `gorm:"primaryKey;autoIncrement"                                                json:"id"`
	BranchID    uint   `gorm:"not null;index:idx_wifi_branch_active,priority:1"                        json:"branch_id"`
	Branch      Branch `gorm:"foreignKey:BranchID"                                                     json:"branch,omitempty"`
	SSID        string `gorm:"column:ssid;size:100;not null"                        json:"ssid"`
	BSSID       string `gorm:"column:bssid;size:50;index:idx_wifi_bssid"            json:"bssid"` // MAC address router
	Description string `gorm:"size:200"                                            json:"description"`

	// idx_wifi_branch_active priority:2
	IsActive bool `gorm:"default:true;index:idx_wifi_branch_active,priority:2" json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (WiFiConfig) TableName() string { return "wifi_configs" }
