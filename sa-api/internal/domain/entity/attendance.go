package entity

import "time"

// AttendanceStatus trạng thái chấm công
type AttendanceStatus string

const (
	StatusPresent        AttendanceStatus = "present"          // Có mặt đúng giờ
	StatusLate           AttendanceStatus = "late"             // Đi muộn
	StatusEarlyLeave     AttendanceStatus = "early_leave"      // Về sớm
	StatusLateEarlyLeave AttendanceStatus = "late_early_leave" // Đi muộn + Về sớm
	StatusAbsent         AttendanceStatus = "absent"           // Vắng mặt
	StatusHalfDay        AttendanceStatus = "half_day"         // Nửa ngày (check-in/out chỉ nửa ca)
	StatusLeave          AttendanceStatus = "leave"            // Nghỉ phép (cả ngày)
	StatusHalfDayLeave   AttendanceStatus = "half_day_leave"   // Nghỉ phép nửa ngày
)

// CheckMethod phương thức xác thực vị trí
type CheckMethod string

const (
	CheckMethodWiFi CheckMethod = "wifi" // Xác thực qua WiFi SSID/BSSID
	CheckMethodGPS  CheckMethod = "gps"  // Xác thực qua GPS Geofencing
)

// AttendanceLog lưu bản ghi chấm công của nhân viên
//
// Index strategy (ước tính 5000 nhân viên × 365 ngày ≈ 1.8M rows/năm):
//
//   - uniq_attendance_user_date   : đảm bảo mỗi user chỉ có 1 bản ghi/ngày (business rule)
//     → dùng cho FindByUserAndDate, tránh race condition check-in đồng thời
//
//   - idx_attendance_user_date    : query cá nhân — "lịch sử chấm công của tôi"
//     → dùng cho GetMyToday, GetSummary per user
//
//   - idx_attendance_branch_date  : query quản lý — "hôm nay chi nhánh X có bao nhiêu người?"
//     → dùng cho GetList với filter branch_id, dashboard branch manager
//
//   - idx_attendance_branch_status_date : filter theo trạng thái trong chi nhánh
//     → dùng cho GetList với filter status + branch_id (báo cáo đi muộn, vắng mặt)
//
// Partial index cho fraud (partial index không thể khai báo qua GORM tags):
//   → Xem migrations/001_init_schema.sql:
//     CREATE INDEX idx_attendance_fraud ON attendance_logs (user_id, created_at DESC)
//     WHERE is_fake_gps = TRUE OR is_vpn = TRUE;
type AttendanceLog struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	// uniq_attendance_user_date (priority:1) | idx_attendance_user_date (priority:1)
	// idx_attendance_branch_status_date không dùng user_id → index riêng bên dưới
	UserID uint `gorm:"not null;uniqueIndex:uniq_attendance_user_date,priority:1;index:idx_attendance_user_date,priority:1" json:"user_id"`
	User   User `gorm:"foreignKey:UserID"                                                                                    json:"user,omitempty"`

	// idx_attendance_branch_date (priority:1) | idx_attendance_branch_status_date (priority:1)
	BranchID uint   `gorm:"not null;index:idx_attendance_branch_date,priority:1;index:idx_attendance_branch_status_date,priority:1" json:"branch_id"`
	Branch   Branch `gorm:"foreignKey:BranchID"                                                                                       json:"branch,omitempty"`

	ShiftID *uint  `gorm:"index"               json:"shift_id"`
	Shift   *Shift `gorm:"foreignKey:ShiftID"  json:"shift,omitempty"`

	// uniq_attendance_user_date (priority:2) | idx_attendance_user_date (priority:2)
	// idx_attendance_branch_date (priority:2) | idx_attendance_branch_status_date (priority:3)
	Date time.Time `gorm:"type:date;not null;uniqueIndex:uniq_attendance_user_date,priority:2;index:idx_attendance_user_date,priority:2;index:idx_attendance_branch_date,priority:2;index:idx_attendance_branch_status_date,priority:3" json:"date"`

	// Check-in info
	CheckInTime   *time.Time   `json:"check_in_time"`
	CheckInLat    *float64     `gorm:"type:decimal(10,8)"  json:"check_in_lat"`
	CheckInLng    *float64     `gorm:"type:decimal(11,8)"  json:"check_in_lng"`
	CheckInMethod *CheckMethod `gorm:"type:varchar(10)"    json:"check_in_method"`
	CheckInSSID   string       `gorm:"column:check_in_ssid;size:100"   json:"check_in_ssid"`
	CheckInBSSID  string       `gorm:"column:check_in_bssid;size:50"   json:"check_in_bssid"`

	// Check-out info
	CheckOutTime   *time.Time   `json:"check_out_time"`
	CheckOutLat    *float64     `gorm:"type:decimal(10,8)"  json:"check_out_lat"`
	CheckOutLng    *float64     `gorm:"type:decimal(11,8)"  json:"check_out_lng"`
	CheckOutMethod *CheckMethod `gorm:"type:varchar(10)"    json:"check_out_method"`
	CheckOutSSID   string       `gorm:"column:check_out_ssid;size:100"  json:"check_out_ssid"`
	CheckOutBSSID  string       `gorm:"column:check_out_bssid;size:50"  json:"check_out_bssid"`

	// Device & security info
	DeviceID    string `gorm:"size:200" json:"device_id"`
	DeviceModel string `gorm:"size:100" json:"device_model"`
	IPAddress   string `gorm:"size:45"  json:"ip_address"`
	AppVersion  string `gorm:"size:20"  json:"app_version"`

	// Anti-fraud flags — partial index trong migrations SQL
	IsFakeGPS bool   `gorm:"default:false" json:"is_fake_gps"`
	IsVPN     bool   `gorm:"default:false" json:"is_vpn"`
	FraudNote string `gorm:"size:500"      json:"fraud_note"`

	// Calculated fields
	// idx_attendance_branch_status_date (priority:2) — filter status trong branch
	Status    AttendanceStatus `gorm:"type:varchar(20);not null;default:'present';index:idx_attendance_branch_status_date,priority:2" json:"status"`
	WorkHours float64          `gorm:"type:decimal(5,2);default:0"                                                                    json:"work_hours"`
	Overtime  float64          `gorm:"type:decimal(5,2);default:0"                                                                    json:"overtime"`
	Note      string           `gorm:"size:500"                                                                                       json:"note"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (AttendanceLog) TableName() string { return "attendance_logs" }

// IsCheckedIn kiểm tra nhân viên đã check-in chưa
func (a *AttendanceLog) IsCheckedIn() bool { return a.CheckInTime != nil }

// IsCheckedOut kiểm tra nhân viên đã check-out chưa
func (a *AttendanceLog) IsCheckedOut() bool { return a.CheckOutTime != nil }

// IsSuspicious kiểm tra bản ghi có dấu hiệu gian lận không
func (a *AttendanceLog) IsSuspicious() bool { return a.IsFakeGPS || a.IsVPN }
