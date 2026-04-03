package repository

import (
	"context"
	"time"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
)

// AttendanceFilter bộ lọc tìm kiếm log chấm công
type AttendanceFilter struct {
	UserID     *uint
	BranchID   *uint
	Department string
	Status     entity.AttendanceStatus
	DateFrom   *time.Time
	DateTo     *time.Time
	Page       int
	Limit      int
}

// AttendanceSummary tổng hợp thống kê chấm công
type AttendanceSummary struct {
	TotalDays      int     `json:"total_days"`
	PresentCount   int     `json:"present_count"`
	LateCount      int     `json:"late_count"`
	EarlyLeaveCount int    `json:"early_leave_count"`
	HalfDayCount   int     `json:"half_day_count"`
	AbsentCount    int     `json:"absent_count"`
	TotalWorkHours float64 `json:"total_work_hours"`
	TotalOvertime  float64 `json:"total_overtime"`
	AttendanceRate float64 `json:"attendance_rate"`
	OnTimeRate     float64 `json:"on_time_rate"`
}

// AttendanceRepository định nghĩa contract cho thao tác dữ liệu chấm công
type AttendanceRepository interface {
	// Create tạo mới bản ghi chấm công
	Create(ctx context.Context, log *entity.AttendanceLog) error

	// Update cập nhật bản ghi chấm công
	Update(ctx context.Context, log *entity.AttendanceLog) error

	// FindByID tìm bản ghi theo ID
	FindByID(ctx context.Context, id uint) (*entity.AttendanceLog, error)

	// FindByUserAndDate tìm bản ghi chấm công của user trong ngày cụ thể
	FindByUserAndDate(ctx context.Context, userID uint, date time.Time) (*entity.AttendanceLog, error)

	// FindAll lấy danh sách bản ghi có phân trang và lọc
	FindAll(ctx context.Context, filter AttendanceFilter) ([]*entity.AttendanceLog, int64, error)

	// GetSummary lấy tổng hợp thống kê chấm công theo user và khoảng thời gian
	GetSummary(ctx context.Context, userID uint, from, to time.Time) (*AttendanceSummary, error)

	// GetBranchSummary lấy tổng hợp theo chi nhánh
	GetBranchSummary(ctx context.Context, branchID uint, from, to time.Time) ([]*UserAttendanceSummary, error)

	// FindActiveCheckIn tìm bản ghi check-in chưa check-out của user
	FindActiveCheckIn(ctx context.Context, userID uint) (*entity.AttendanceLog, error)

	// CountSuspicious đếm số bản ghi nghi ngờ gian lận theo user
	CountSuspicious(ctx context.Context, userID uint, from time.Time) (int64, error)

	// GetTodayStatsByBranch lấy thống kê chấm công hôm nay theo từng chi nhánh, có phân trang
	// branchID = nil → tất cả chi nhánh active (dành cho admin tổng)
	// branchID != nil → chỉ chi nhánh đó (dành cho manager hoặc admin lọc)
	GetTodayStatsByBranch(ctx context.Context, branchID *uint, search string, page, limit int) ([]*BranchTodayStats, int64, error)

	// GetTodayEmployeeDetails lấy danh sách chi tiết nhân viên với trạng thái chấm công hôm nay.
	// Trả về cả nhân viên "absent" (chưa chấm công) — khác với FindAll chỉ trả về attendance_logs có sẵn.
	GetTodayEmployeeDetails(ctx context.Context, filter TodayEmployeeFilter) ([]*EmployeeTodayDetail, int64, error)
}

// UserAttendanceSummary tổng hợp chấm công theo từng user
type UserAttendanceSummary struct {
	UserID       uint    `json:"user_id"`
	UserName     string  `json:"user_name"`
	EmployeeCode string  `json:"employee_code"`
	Department   string  `json:"department"`
	AttendanceSummary
}

// TodayEmployeeStatus các giá trị hợp lệ cho filter status của GetTodayEmployeeDetails
// (mở rộng entity.AttendanceStatus với "absent" và "suspicious")
type TodayEmployeeStatus string

const (
	TodayStatusAll        TodayEmployeeStatus = ""           // Tất cả
	TodayStatusPresent    TodayEmployeeStatus = "present"
	TodayStatusLate       TodayEmployeeStatus = "late"
	TodayStatusEarlyLeave TodayEmployeeStatus = "early_leave"
	TodayStatusHalfDay    TodayEmployeeStatus = "half_day"
	TodayStatusAbsent     TodayEmployeeStatus = "absent"     // Chưa chấm công
	TodayStatusSuspicious TodayEmployeeStatus = "suspicious" // Có flag gian lận (fake GPS hoặc VPN)
)

// TodayEmployeeFilter bộ lọc cho GetTodayEmployeeDetails
type TodayEmployeeFilter struct {
	BranchID *uint
	Status   TodayEmployeeStatus
	Page     int
	Limit    int
}

// EmployeeTodayDetail thông tin chấm công hôm nay của một nhân viên — dùng cho danh sách chi tiết
//
// Status là derived status:
//   - "absent"     → nhân viên chưa chấm công hôm nay
//   - "suspicious" → đã chấm công nhưng bị flag is_fake_gps=true hoặc is_vpn=true
//   - các giá trị khác → attendance_logs.status gốc (present, late, early_leave, half_day)
type EmployeeTodayDetail struct {
	UserID       uint    `json:"user_id"`
	EmployeeCode string  `json:"employee_code"`
	Name         string  `json:"name"`
	Department   string  `json:"department"`
	BranchID     uint    `json:"branch_id"`
	BranchName   string  `json:"branch_name"`
	Status       string  `json:"status"`
	CheckInTime  *string `json:"check_in_time"`  // RFC3339, nil nếu absent
	CheckOutTime *string `json:"check_out_time"` // RFC3339, nil nếu chưa checkout
	WorkHours    float64 `json:"work_hours"`
	IsFakeGPS    bool    `json:"is_fake_gps"`
	IsVPN        bool    `json:"is_vpn"`
	FraudNote    string  `json:"fraud_note,omitempty"`
}

// BranchTodayStats thống kê chấm công hôm nay của một chi nhánh — dùng cho dashboard admin
type BranchTodayStats struct {
	BranchID        uint    `json:"branch_id"`
	BranchName      string  `json:"branch_name"`
	BranchCode      string  `json:"branch_code"`
	TotalEmployees  int64   `json:"total_employees"`   // Tổng nhân viên active của chi nhánh
	PresentCount    int64   `json:"present_count"`     // Đúng giờ
	LateCount       int64   `json:"late_count"`        // Đi muộn
	EarlyLeaveCount int64   `json:"early_leave_count"` // Về sớm
	HalfDayCount    int64   `json:"half_day_count"`    // Nửa ngày
	AbsentCount     int64   `json:"absent_count"`      // Vắng mặt (chưa chấm công)
	SuspiciousCount int64   `json:"suspicious_count"`  // Nghi ngờ gian lận (fake GPS hoặc VPN)
	AttendanceRate  float64 `json:"attendance_rate"`   // % có mặt = checked_in / total * 100
	OnTimeRate      float64 `json:"on_time_rate"`      // % đúng giờ = present / checked_in * 100
}
