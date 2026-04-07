package entity

import "time"

// UserRole định nghĩa vai trò của người dùng trong hệ thống (RBAC 3 cấp)
type UserRole string

const (
	RoleAdmin    UserRole = "admin"    // Quản trị toàn hệ thống — không gắn chi nhánh
	RoleManager  UserRole = "manager"  // Quản lý chi nhánh — chỉ thấy data chi nhánh mình
	RoleEmployee UserRole = "employee" // Nhân viên — chỉ thấy data của bản thân
)

// User đại diện cho người dùng/nhân viên trong hệ thống
//
// Index strategy (5000 employees):
//   - uniq_user_email        : login lookup — O(log n)
//   - uniq_user_code         : employee code lookup
//   - idx_user_branch_role_active : query phổ biến nhất của Manager/Admin —
//     "lấy danh sách nhân viên active theo branch + role"
//   - idx_user_branch_active : Manager xem toàn bộ nhân viên trong chi nhánh
type User struct {
	ID           uint      `gorm:"primaryKey;autoIncrement"                                               json:"id"`
	EmployeeCode string    `gorm:"uniqueIndex:uniq_user_code;size:50;not null"                            json:"employee_code"`
	Name         string    `gorm:"size:200;not null"                                                      json:"name"`
	Email        string    `gorm:"uniqueIndex:uniq_user_email;size:100;not null"                          json:"email"`
	Phone        string    `gorm:"size:20"                                                                json:"phone"`
	Password     string    `gorm:"size:255;not null"                                                      json:"-"`
	Department   string    `gorm:"size:100"                                                               json:"department"`
	Position     string    `gorm:"size:100"                                                               json:"position"`
	AvatarURL    string    `gorm:"size:500"                                                               json:"avatar_url"`
	HiredAt      *time.Time `gorm:"type:date"                                                             json:"hired_at"`
	LastLoginAt  *time.Time `                                                                             json:"last_login_at"`
	CreatedAt    time.Time `                                                                               json:"created_at"`
	UpdatedAt    time.Time `                                                                               json:"updated_at"`

	// BranchID = nil chỉ dành cho RoleAdmin (toàn hệ thống)
	// idx_user_branch_role_active: (branch_id, role, is_active) — RBAC filter chính
	BranchID *uint    `gorm:"index:idx_user_branch_role_active,priority:1;index:idx_user_branch_active,priority:1" json:"branch_id"`
	Branch   *Branch  `gorm:"foreignKey:BranchID"                                                                   json:"branch,omitempty"`

	// idx_user_branch_role_active priority:2
	Role UserRole `gorm:"type:varchar(20);not null;default:'employee';index:idx_user_branch_role_active,priority:2" json:"role"`

	// idx_user_branch_role_active priority:3 | idx_user_branch_active priority:2
	IsActive bool `gorm:"default:true;index:idx_user_branch_role_active,priority:3;index:idx_user_branch_active,priority:2" json:"is_active"`

	// Số ngày phép hiện có (full_day = 1.0, half_day = 0.5)
	LeaveBalance float64 `gorm:"type:decimal(5,1);not null;default:0" json:"leave_balance"`

	// Relations
	AttendanceLogs []AttendanceLog `gorm:"foreignKey:UserID" json:"-"`
}

func (User) TableName() string { return "users" }

// IsAdmin kiểm tra user có phải admin toàn hệ thống không
func (u *User) IsAdmin() bool { return u.Role == RoleAdmin }

// IsManager kiểm tra user có phải quản lý chi nhánh không
func (u *User) IsManager() bool { return u.Role == RoleManager }

// CanAccessBranch kiểm tra user có quyền truy cập vào chi nhánh cụ thể không
func (u *User) CanAccessBranch(branchID uint) bool {
	if u.IsAdmin() {
		return true
	}
	if u.BranchID == nil {
		return false
	}
	return *u.BranchID == branchID
}
