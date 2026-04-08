package config

import (
	"log/slog"
	"strings"

	"github.com/spf13/viper"
)

// Config cấu hình toàn bộ ứng dụng
type Config struct {
	App        AppConfig        `mapstructure:"app"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Correction CorrectionConfig `mapstructure:"correction"`
	Attendance AttendanceConfig `mapstructure:"attendance"`
	Overtime   OvertimeConfig   `mapstructure:"overtime"`
}

type AppConfig struct {
	Name    string `mapstructure:"name"`
	Port    string `mapstructure:"port"`
	Env     string `mapstructure:"env"`    // development | production
	Debug   bool   `mapstructure:"debug"`
	Version string `mapstructure:"version"`
}

type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            string `mapstructure:"port"`
	Name            string `mapstructure:"name"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	SSLMode         string `mapstructure:"ssl_mode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"` // seconds
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type JWTConfig struct {
	Secret            string `mapstructure:"secret"`
	ExpireHours       int    `mapstructure:"expire_hours"`
	RefreshExpireDays int    `mapstructure:"refresh_expire_days"`
}

type CorrectionConfig struct {
	MaxPerMonth         int `mapstructure:"max_per_month"`          // Hạn mức bù công ca chính thức tối đa mỗi tháng (tính theo credit)
	OvertimeMaxPerMonth int `mapstructure:"overtime_max_per_month"` // Hạn mức bù công tăng ca tối đa mỗi tháng
}

type AttendanceConfig struct {
	MaxSuspiciousCount int `mapstructure:"max_suspicious_count"` // Số lần vi phạm tối đa trong 7 ngày trước khi block (default: 3)
	SuspiciousWindowDays int `mapstructure:"suspicious_window_days"` // Số ngày kiểm tra lịch sử vi phạm (default: 7)
}

type OvertimeConfig struct {
	MaxHoursPerDay float64 `mapstructure:"max_hours_per_day"` // Số giờ OT tối đa mỗi ngày (default: 4)
}

// Load đọc cấu hình từ config.yaml, sau đó override bằng biến môi trường nếu có
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/app/config") // Docker container path

	// Biến môi trường override config file (dùng APP_PORT thay app.port)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			slog.Warn("config.yaml not found, using environment variables only")
		} else {
			return nil, err
		}
	} else {
		slog.Info("loaded config", "file", viper.ConfigFileUsed())
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
