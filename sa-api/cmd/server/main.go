package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hdbank/smart-attendance/config"
	"github.com/hdbank/smart-attendance/internal/infrastructure/cache"
	"github.com/hdbank/smart-attendance/internal/infrastructure/database"
	"github.com/hdbank/smart-attendance/internal/infrastructure/database/migrations"
	"github.com/hdbank/smart-attendance/internal/infrastructure/scheduler"

	"github.com/go-gormigrate/gormigrate/v2"
	applogger "github.com/hdbank/smart-attendance/internal/infrastructure/logger"
	"github.com/hdbank/smart-attendance/internal/repository"

	handlerAdmin "github.com/hdbank/smart-attendance/internal/handler/admin"
	handlerUser "github.com/hdbank/smart-attendance/internal/handler/user"

	"github.com/hdbank/smart-attendance/internal/server"

	ucAttendance "github.com/hdbank/smart-attendance/internal/usecase/attendance"
	ucBranch "github.com/hdbank/smart-attendance/internal/usecase/branch"
	ucCorrection "github.com/hdbank/smart-attendance/internal/usecase/correction"
	ucLeave "github.com/hdbank/smart-attendance/internal/usecase/leave"
	ucReport "github.com/hdbank/smart-attendance/internal/usecase/report"
	ucUser "github.com/hdbank/smart-attendance/internal/usecase/user"

	"github.com/labstack/echo/v4"
)

func main() {
	// ── 1. Load config ──
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// ── 2. Setup logger ──
	applogger.Setup(cfg.App.Env, cfg.App.Debug)
	slog.Info("starting smart-attendance API",
		"version", cfg.App.Version,
		"env", cfg.App.Env,
		"port", cfg.App.Port,
	)

	// ── 3. Connect Database ──
	// Tự động tạo database nếu chưa tồn tại (user mới chỉ cần make run)
	database.EnsureDatabase(&cfg.Database)

	db, err := database.NewPostgres(&cfg.Database)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}

	// Chạy migration tự động — mọi môi trường, chỉ apply migration chưa chạy
	m := gormigrate.New(db, gormigrate.DefaultOptions, migrations.GetMigrations())
	if err := m.Migrate(); err != nil {
		slog.Error("database migration failed", "error", err)
		os.Exit(1)
	}
	slog.Info("database migrations completed")

	// Chạy seeder — chỉ seed nếu database chưa có dữ liệu (lần build đầu tiên)
	if err := database.RunSeeder(db); err != nil {
		slog.Error("database seeding failed", "error", err)
		os.Exit(1)
	}



	// ── 4. Connect Redis (graceful degradation nếu không kết nối được) ──
	redisCache, err := cache.NewRedisCache(&cfg.Redis)
	if err != nil {
		slog.Warn("redis connection failed, using no-op cache (rate limiting and caching disabled)", "error", err)
		redisCache = cache.NewNoOpCache()
	}

	// ── 5. Init Repositories ──
	branchRepo := repository.NewBranchRepository(db)
	userRepo := repository.NewUserRepository(db)
	attendanceRepo := repository.NewAttendanceRepository(db)
	wifiConfigRepo := repository.NewWiFiConfigRepository(db)
	gpsConfigRepo := repository.NewGPSConfigRepository(db)
	shiftRepo := repository.NewShiftRepository(db)
	correctionRepo := repository.NewCorrectionRepository(db)
	leaveRepo := repository.NewLeaveRepository(db)

	// ── 6. Init Usecases ──
	userUC := ucUser.NewUserUsecase(userRepo, redisCache, cfg.JWT)
	branchUC := ucBranch.NewBranchUsecase(branchRepo, gpsConfigRepo, redisCache)
	attendanceUC := ucAttendance.NewAttendanceUsecase(
		attendanceRepo, userRepo, wifiConfigRepo, gpsConfigRepo, shiftRepo, redisCache,
	)
	reportUC := ucReport.NewReportUsecase(attendanceRepo, userRepo, branchRepo, redisCache)
	correctionUC := ucCorrection.NewCorrectionUsecase(correctionRepo, attendanceRepo, userRepo, shiftRepo, db, cfg.Correction)
	leaveUC := ucLeave.NewLeaveUsecase(leaveRepo, correctionRepo, attendanceRepo, userRepo, shiftRepo, db)

	// ── 7. Init Handlers ──
	// User app handlers
	authH := handlerUser.NewAuthHandler(userUC)
	attendanceH := handlerUser.NewAttendanceHandler(attendanceUC)
	correctionH := handlerUser.NewCorrectionHandler(correctionUC)
	leaveH := handlerUser.NewLeaveHandler(leaveUC)
	// Admin portal handlers
	adminAuthH := handlerAdmin.NewAdminAuthHandler(userUC)
	userH := handlerAdmin.NewUserHandler(userUC)
	branchH := handlerAdmin.NewBranchHandler(branchUC)
	adminAttendanceH := handlerAdmin.NewAttendanceHandler(attendanceUC)
	adminCorrectionH := handlerAdmin.NewCorrectionHandler(correctionUC)
	adminLeaveH := handlerAdmin.NewLeaveHandler(leaveUC)
	reportH := handlerAdmin.NewReportHandler(reportUC)
	wifiConfigH := handlerAdmin.NewWiFiConfigHandler(wifiConfigRepo)

	// ── 8. Start background schedulers ──
	scheduler.StartCorrectionAutoReject(correctionUC)
	scheduler.StartLeaveAutoReject(leaveUC)
	scheduler.StartLeaveAccrual(leaveUC)

	// ── 9. Setup Echo Router ──
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	server.SetupRouter(e, server.RouterDeps{
		AuthHandler:            authH,
		AttendanceHandler:      attendanceH,
		CorrectionHandler:      correctionH,
		LeaveHandler:           leaveH,
		AdminAuthHandler:       adminAuthH,
		UserHandler:            userH,
		BranchHandler:          branchH,
		AdminAttendanceHandler: adminAttendanceH,
		AdminCorrectionHandler: adminCorrectionH,
		AdminLeaveHandler:      adminLeaveH,
		ReportHandler:          reportH,
		WiFiConfigHandler:      wifiConfigH,
		Cache:                  redisCache,
		JWTSecret:              cfg.JWT.Secret,
	})

	// ── 9. Graceful Shutdown ──
	go func() {
		addr := ":" + cfg.App.Port
		slog.Info("server listening", "addr", addr)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Chờ signal để shutdown gracefully
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("server stopped")
}
