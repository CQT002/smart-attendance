package migrations

import (
	"github.com/hdbank/smart-attendance/internal/domain/entity"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// GetMigrations trả về tất cả migration theo thứ tự thời gian.
// ID format: YYYYMMDDHHMMSS — gormigrate chỉ chạy migration chưa có trong bảng "migrations".
func GetMigrations() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		// ── 001: Khởi tạo 6 bảng core ──
		{
			ID: "20250330000001",
			Migrate: func(tx *gorm.DB) error {
				type Branch struct {
					entity.Branch
				}
				type User struct {
					entity.User
				}
				type WiFiConfig struct {
					entity.WiFiConfig
				}
				type GPSConfig struct {
					entity.GPSConfig
				}
				type Shift struct {
					entity.Shift
				}
				type AttendanceLog struct {
					entity.AttendanceLog
				}

				if err := tx.AutoMigrate(
					&Branch{},
					&User{},
					&WiFiConfig{},
					&GPSConfig{},
					&Shift{},
					&AttendanceLog{},
				); err != nil {
					return err
				}

				// Seed: tài khoản admin mặc định — mật khẩu: Admin@123 (bcrypt cost 10)
				return tx.Exec(`
					INSERT INTO users (employee_code, name, email, password, role, is_active, created_at, updated_at)
					VALUES (
						'ADMIN001',
						'System Administrator',
						'admin@hdbank.com.vn',
						'$2a$10$Zf7d7h4Cq8LkGq2v0hHY4OqT3JZhP8FtWZ6DkUQD5IYNJOvLm4vTm',
						'admin',
						true,
						NOW(), NOW()
					) ON CONFLICT (email) DO NOTHING
				`).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable(
					"attendance_logs",
					"shifts",
					"gps_configs",
					"wifi_configs",
					"users",
					"branches",
				)
			},
		},

		// ── 002: Tạo bảng daily_summaries ──
		{
			ID: "20250330000002",
			Migrate: func(tx *gorm.DB) error {
				type DailySummary struct {
					entity.DailySummary
				}
				return tx.AutoMigrate(&DailySummary{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("daily_summaries")
			},
		},

		// ── 003: Partial index cho fraud detection ──
		// GORM tags không hỗ trợ partial index → phải tạo bằng raw SQL
		{
			ID: "20250330000003",
			Migrate: func(tx *gorm.DB) error {
				sqls := []string{
					// Partial index: chỉ index bản ghi nghi ngờ gian lận (nhỏ hơn, nhanh hơn full index)
					`CREATE INDEX IF NOT EXISTS idx_attendance_fraud
						ON attendance_logs (user_id, created_at DESC)
						WHERE is_fake_gps = TRUE OR is_vpn = TRUE`,

					// Partial index: WiFi BSSID lookup — bỏ qua rows NULL/empty
					`CREATE INDEX IF NOT EXISTS idx_wifi_bssid_partial
						ON wifi_configs (bssid)
						WHERE bssid IS NOT NULL AND bssid != ''`,
				}
				for _, sql := range sqls {
					if err := tx.Exec(sql).Error; err != nil {
						return err
					}
				}
				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				sqls := []string{
					`DROP INDEX IF EXISTS idx_attendance_fraud`,
					`DROP INDEX IF EXISTS idx_wifi_bssid_partial`,
				}
				for _, sql := range sqls {
					if err := tx.Exec(sql).Error; err != nil {
						return err
					}
				}
				return nil
			},
		},
	}
}
