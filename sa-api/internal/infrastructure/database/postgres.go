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
