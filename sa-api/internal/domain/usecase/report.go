package usecase

import (
	"context"
	"time"

	"github.com/hdbank/smart-attendance/internal/domain/repository"
)

// ReportPeriod chu kỳ báo cáo
type ReportPeriod string

const (
	PeriodDaily   ReportPeriod = "daily"
	PeriodWeekly  ReportPeriod = "weekly"
	PeriodMonthly ReportPeriod = "monthly"
	PeriodCustom  ReportPeriod = "custom"
)

// TodayStatsFilter bộ lọc cho API thống kê dashboard hôm nay
type TodayStatsFilter struct {
	// BranchID = nil → admin xem toàn bộ; BranchID != nil → filter theo chi nhánh
	BranchID *uint
	Search   string // Tìm kiếm theo tên chi nhánh
	Page     int
	Limit    int
}

// ReportFilter bộ lọc cho báo cáo
type ReportFilter struct {
	BranchID   *uint
	UserID     *uint
	Department string
	Period     ReportPeriod
	DateFrom   time.Time
	DateTo     time.Time
	Page       int
	Limit      int
}

// DashboardStats thống kê tổng quan cho dashboard
type DashboardStats struct {
	TotalBranches    int64   `json:"total_branches"`
	TotalEmployees   int64   `json:"total_employees"`
	PresentToday     int64   `json:"present_today"`
	AbsentToday      int64   `json:"absent_today"`
	LateToday        int64   `json:"late_today"`
	AttendanceRate   float64 `json:"attendance_rate"`   // % có mặt
	OnTimeRate       float64 `json:"on_time_rate"`      // % đúng giờ
}

// BranchAttendanceReport báo cáo chấm công theo chi nhánh
type BranchAttendanceReport struct {
	BranchID   uint   `json:"branch_id"`
	BranchName string `json:"branch_name"`
	BranchCode string `json:"branch_code"`
	repository.AttendanceSummary
	Employees []*repository.UserAttendanceSummary `json:"employees,omitempty"`
}

// ReportUsecase định nghĩa business logic cho báo cáo và dashboard
type ReportUsecase interface {
	// GetTodayBranchStats thống kê chấm công hôm nay theo từng chi nhánh (dashboard admin)
	// RBAC được xử lý ở handler: manager → BranchID bắt buộc = chi nhánh mình
	GetTodayBranchStats(ctx context.Context, filter TodayStatsFilter) ([]*repository.BranchTodayStats, int64, error)

	// GetTodayEmployeeDetails danh sách chi tiết nhân viên theo trạng thái hôm nay, có phân trang
	// Bao gồm cả nhân viên "absent" (chưa chấm công)
	GetTodayEmployeeDetails(ctx context.Context, filter repository.TodayEmployeeFilter) ([]*repository.EmployeeTodayDetail, int64, error)

	// GetDashboardStats lấy thống kê tổng quan cho admin dashboard
	GetDashboardStats(ctx context.Context) (*DashboardStats, error)

	// GetBranchDashboard lấy thống kê cho manager dashboard của chi nhánh
	GetBranchDashboard(ctx context.Context, branchID uint) (*DashboardStats, error)

	// GetAttendanceReport lấy báo cáo chấm công với filter và phân trang
	GetAttendanceReport(ctx context.Context, filter ReportFilter) ([]*repository.UserAttendanceSummary, int64, error)

	// GetBranchReport lấy báo cáo theo từng chi nhánh
	GetBranchReport(ctx context.Context, filter ReportFilter) ([]*BranchAttendanceReport, error)

	// GetUserReport lấy báo cáo của một nhân viên cụ thể
	GetUserReport(ctx context.Context, userID uint, from, to time.Time) (*repository.UserAttendanceSummary, error)
}
