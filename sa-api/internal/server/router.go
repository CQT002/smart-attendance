package server

import (
	"net/http"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	handlerAdmin "github.com/hdbank/smart-attendance/internal/handler/admin"
	handlerUser "github.com/hdbank/smart-attendance/internal/handler/user"
	"github.com/hdbank/smart-attendance/internal/infrastructure/cache"
	"github.com/hdbank/smart-attendance/internal/middleware"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

// RouterDeps chứa tất cả dependencies cần thiết để đăng ký routes
type RouterDeps struct {
	// User app handlers
	AuthHandler           *handlerUser.AuthHandler
	AttendanceHandler     *handlerUser.AttendanceHandler
	CorrectionHandler     *handlerUser.CorrectionHandler
	LeaveHandler          *handlerUser.LeaveHandler
	// Admin portal handlers
	AdminAuthHandler       *handlerAdmin.AdminAuthHandler
	UserHandler            *handlerAdmin.UserHandler
	BranchHandler          *handlerAdmin.BranchHandler
	AdminAttendanceHandler *handlerAdmin.AttendanceHandler
	AdminCorrectionHandler *handlerAdmin.CorrectionHandler
	AdminLeaveHandler      *handlerAdmin.LeaveHandler
	ReportHandler          *handlerAdmin.ReportHandler
	WiFiConfigHandler      *handlerAdmin.WiFiConfigHandler

	Cache     cache.Cache
	JWTSecret string
}

// SetupRouter cấu hình tất cả routes và middleware
func SetupRouter(e *echo.Echo, deps RouterDeps) {
	// === Global Middleware ===
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORS())
	e.Use(middleware.RequestLogger())
	e.Use(middleware.GlobalRateLimiter(deps.Cache))

	// Health check - không cần auth
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "ok",
			"service": "smart-attendance",
		})
	})

	api := e.Group("/api/v1")

	// === Auth routes (không cần JWT) ===
	auth := api.Group("/auth")
	auth.POST("/login", deps.AuthHandler.Login,
		middleware.LoginRateLimiter(deps.Cache))

	// === Protected routes (cần JWT) ===
	protected := api.Group("", middleware.JWTAuth(deps.JWTSecret))

	// Auth (đã đăng nhập)
	protected.GET("/auth/me", deps.AuthHandler.Me)
	protected.PUT("/auth/change-password", deps.AuthHandler.ChangePassword)

	// ── Employee attendance routes (user app) ──
	attend := protected.Group("/attendance")
	attend.POST("/check-in", deps.AttendanceHandler.CheckIn,
		middleware.CheckInRateLimiter(deps.Cache))
	attend.POST("/check-out", deps.AttendanceHandler.CheckOut,
		middleware.CheckInRateLimiter(deps.Cache))
	attend.GET("/today", deps.AttendanceHandler.GetMyToday)
	attend.GET("/history", deps.AttendanceHandler.GetMyHistory)

	// ── Employee correction routes (chấm công bù) ──
	corrections := protected.Group("/attendance/corrections")
	corrections.POST("", deps.CorrectionHandler.Create)
	corrections.GET("", deps.CorrectionHandler.GetMyList)
	corrections.GET("/:id", deps.CorrectionHandler.GetByID)

	// ── Employee leave routes (nghỉ phép) ──
	leaves := protected.Group("/attendance/leaves")
	leaves.POST("", deps.LeaveHandler.Create)
	leaves.GET("", deps.LeaveHandler.GetMyList)
	leaves.GET("/:id", deps.LeaveHandler.GetByID)

	// ── Admin portal routes ──
	// Admin auth — login không cần JWT, me/change-password cần JWT
	adminAuth := api.Group("/admin/auth")
	adminAuth.POST("/login", deps.AdminAuthHandler.Login,
		middleware.LoginRateLimiter(deps.Cache))
	adminAuth.GET("/me", deps.AdminAuthHandler.Me,
		middleware.JWTAuth(deps.JWTSecret),
		middleware.RequireRole(entity.RoleAdmin, entity.RoleManager))
	adminAuth.PUT("/change-password", deps.AdminAuthHandler.ChangePassword,
		middleware.JWTAuth(deps.JWTSecret),
		middleware.RequireRole(entity.RoleAdmin, entity.RoleManager))

	adminGroup := protected.Group("/admin",
		middleware.RequireRole(entity.RoleAdmin, entity.RoleManager))

	// Admin - Attendance management
	adminAttend := adminGroup.Group("/attendance")
	adminAttend.GET("", deps.AdminAttendanceHandler.GetList)
	adminAttend.GET("/:id", deps.AdminAttendanceHandler.GetByID)
	adminAttend.GET("/summary/:user_id", deps.AdminAttendanceHandler.GetSummary)

	// Admin - Correction management (chấm công bù)
	adminCorrections := adminGroup.Group("/corrections")
	adminCorrections.GET("", deps.AdminCorrectionHandler.GetList)
	adminCorrections.GET("/:id", deps.AdminCorrectionHandler.GetByID)
	adminCorrections.PUT("/:id/process", deps.AdminCorrectionHandler.Process)
	adminCorrections.POST("/batch-approve", deps.AdminCorrectionHandler.BatchApprove)

	// Admin - Leave management (nghỉ phép)
	adminLeaves := adminGroup.Group("/leaves")
	adminLeaves.GET("", deps.AdminLeaveHandler.GetList)
	adminLeaves.GET("/:id", deps.AdminLeaveHandler.GetByID)
	adminLeaves.PUT("/:id/process", deps.AdminLeaveHandler.Process)
	adminLeaves.POST("/batch-approve", deps.AdminLeaveHandler.BatchApprove)

	// Admin - Unified approvals (tổng hợp duyệt chấm công)
	adminGroup.GET("/approvals", deps.AdminLeaveHandler.GetApprovals)
	adminGroup.GET("/approvals/pending", deps.AdminLeaveHandler.GetPendingApprovals) // backward compat

	// Admin - User management
	users := adminGroup.Group("/users")
	users.GET("", deps.UserHandler.GetList)
	users.POST("", deps.UserHandler.Create)
	users.GET("/:id", deps.UserHandler.GetByID)
	users.PUT("/:id", deps.UserHandler.Update)
	users.DELETE("/:id", deps.UserHandler.Delete,
		middleware.RequireRole(entity.RoleAdmin))
	users.POST("/:id/reset-password", deps.UserHandler.ResetPassword)

	// Admin - Branch management
	branches := adminGroup.Group("/branches")
	branches.GET("/active", deps.BranchHandler.GetActive)
	branches.GET("", deps.BranchHandler.GetList)
	branches.POST("", deps.BranchHandler.Create,
		middleware.RequireRole(entity.RoleAdmin))
	branches.GET("/:id", deps.BranchHandler.GetByID)
	branches.PUT("/:id", deps.BranchHandler.Update,
		middleware.RequireRole(entity.RoleAdmin))
	branches.DELETE("/:id", deps.BranchHandler.Delete,
		middleware.RequireRole(entity.RoleAdmin))

	// Admin - WiFi config management (nested under branches)
	wifiConfigs := branches.Group("/:branch_id/wifi-configs")
	wifiConfigs.GET("", deps.WiFiConfigHandler.GetByBranch)
	wifiConfigs.POST("", deps.WiFiConfigHandler.Create,
		middleware.RequireRole(entity.RoleAdmin))
	wifiConfigs.PUT("/:id", deps.WiFiConfigHandler.Update,
		middleware.RequireRole(entity.RoleAdmin))
	wifiConfigs.DELETE("/:id", deps.WiFiConfigHandler.Delete,
		middleware.RequireRole(entity.RoleAdmin))

	// Admin - Reports
	reports := adminGroup.Group("/reports")
	reports.GET("/today", deps.ReportHandler.GetTodayStats)
	reports.GET("/today/employees", deps.ReportHandler.GetTodayEmployees)
	reports.GET("/dashboard", deps.ReportHandler.GetDashboard)
	reports.GET("/attendance", deps.ReportHandler.GetAttendanceReport)
	reports.GET("/branches", deps.ReportHandler.GetBranchReport)
	reports.GET("/users/:user_id", deps.ReportHandler.GetUserReport)
}
