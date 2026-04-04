package database

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/hdbank/smart-attendance/config"
	"github.com/hdbank/smart-attendance/internal/domain/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// EnsureDatabase kết nối vào database mặc định "postgres" để tạo database nếu chưa tồn tại.
// Giúp user mới chỉ cần chạy `make run` mà không cần tạo database thủ công.
func EnsureDatabase(cfg *config.DatabaseConfig) {
	dbName := cfg.Name
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		slog.Warn("ensure database: cannot connect to postgres, skipping auto-create", "error", err)
		return
	}

	// Kiểm tra database đã tồn tại chưa
	var exists bool
	db.Raw("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = ?)", dbName).Scan(&exists)
	if exists {
		slog.Info("ensure database: already exists", "database", dbName)
	} else {
		// Tạo database — không dùng parameterized query vì CREATE DATABASE không hỗ trợ
		if err := db.Exec(fmt.Sprintf("CREATE DATABASE %q", dbName)).Error; err != nil {
			slog.Warn("ensure database: failed to create", "database", dbName, "error", err)
		} else {
			slog.Info("ensure database: created successfully", "database", dbName)
		}
	}

	// Đóng connection tạm
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
}

// NewPostgres khởi tạo kết nối PostgreSQL với connection pool tối ưu
func NewPostgres(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Ho_Chi_Minh",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	logLevel := logger.Silent
	// Chỉ bật log SQL ở môi trường development
	// Ở production, log SQL sẽ gây overhead lớn với 5000 user
	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		// Tắt auto-create nếu dùng migration riêng
		DisableForeignKeyConstraintWhenMigrating: false,
	}

	db, err := gorm.Open(postgres.Open(dsn), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Lấy underlying sql.DB để cấu hình connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Connection pool tối ưu cho 5000 users giờ cao điểm
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// Kiểm tra kết nối
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("database connected", "host", cfg.Host, "port", cfg.Port, "db", cfg.Name)
	return db, nil
}

// AutoMigrate tự động tạo/cập nhật schema (chỉ dùng cho development)
func AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&entity.Branch{},
		&entity.User{},
		&entity.WiFiConfig{},
		&entity.GPSConfig{},
		&entity.Shift{},
		&entity.AttendanceLog{},
		&entity.DailySummary{},
	)
	if err != nil {
		return fmt.Errorf("auto migrate failed: %w", err)
	}
	slog.Info("database migration completed")
	return nil
}
