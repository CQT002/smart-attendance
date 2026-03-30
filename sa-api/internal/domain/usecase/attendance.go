package usecase

import (
	"context"
	"time"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	"github.com/hdbank/smart-attendance/internal/domain/repository"
)

// CheckInRequest dữ liệu yêu cầu check-in từ mobile app
type CheckInRequest struct {
	UserID      uint    `json:"user_id"`
	BranchID    uint    `json:"branch_id"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	SSID        string  `json:"ssid"`
	BSSID       string  `json:"bssid"`
	DeviceID    string  `json:"device_id"`
	DeviceModel string  `json:"device_model"`
	IPAddress   string  `json:"ip_address"`
	AppVersion  string  `json:"app_version"`
	IsFakeGPS   bool    `json:"is_fake_gps"`  // Flag từ mobile SDK phát hiện GPS giả
	IsVPN       bool    `json:"is_vpn"`       // Flag từ mobile SDK phát hiện VPN
}

// CheckOutRequest dữ liệu yêu cầu check-out
type CheckOutRequest struct {
	AttendanceID uint    `json:"attendance_id"`
	UserID       uint    `json:"user_id"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	SSID         string  `json:"ssid"`
	BSSID        string  `json:"bssid"`
	DeviceID     string  `json:"device_id"`
	IPAddress    string  `json:"ip_address"`
	IsFakeGPS    bool    `json:"is_fake_gps"`
	IsVPN        bool    `json:"is_vpn"`
}

// AttendanceUsecase định nghĩa business logic cho chấm công
type AttendanceUsecase interface {
	// CheckIn xử lý nghiệp vụ check-in với xác thực vị trí và chống gian lận
	CheckIn(ctx context.Context, req CheckInRequest) (*entity.AttendanceLog, error)

	// CheckOut xử lý nghiệp vụ check-out và tính toán giờ làm
	CheckOut(ctx context.Context, req CheckOutRequest) (*entity.AttendanceLog, error)

	// GetByID lấy thông tin một bản ghi chấm công
	GetByID(ctx context.Context, id uint) (*entity.AttendanceLog, error)

	// GetList lấy danh sách chấm công với phân trang và lọc
	GetList(ctx context.Context, filter repository.AttendanceFilter) ([]*entity.AttendanceLog, int64, error)

	// GetMyToday lấy thông tin chấm công của user trong ngày hôm nay
	GetMyToday(ctx context.Context, userID uint) (*entity.AttendanceLog, error)

	// GetSummary lấy tổng hợp thống kê chấm công
	GetSummary(ctx context.Context, userID uint, from, to time.Time) (*repository.AttendanceSummary, error)
}
