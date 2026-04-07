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
						'$2a$10$ZSZnC8n7hO8awy2PHsSrSOY8bfwYHCpF5/yqT7yuCNK8/gcvy0CAW',
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
		// ── 004: Tạo bảng attendance_corrections — Chấm công bù ──
		{
			ID: "20250330000004",
			Migrate: func(tx *gorm.DB) error {
				type AttendanceCorrection struct {
					entity.AttendanceCorrection
				}
				return tx.AutoMigrate(&AttendanceCorrection{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("attendance_corrections")
			},
		},

		// ── 005: Remove latitude/longitude from branches (moved to gps_configs) ──
		{
			ID: "20250331000001",
			Migrate: func(tx *gorm.DB) error {
				sqls := []string{
					`ALTER TABLE branches DROP COLUMN IF EXISTS latitude`,
					`ALTER TABLE branches DROP COLUMN IF EXISTS longitude`,
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
					`ALTER TABLE branches ADD COLUMN IF NOT EXISTS latitude DECIMAL(10,8)`,
					`ALTER TABLE branches ADD COLUMN IF NOT EXISTS longitude DECIMAL(11,8)`,
				}
				for _, sql := range sqls {
					if err := tx.Exec(sql).Error; err != nil {
						return err
					}
				}
				return nil
			},
		},

		// ── 006: Thêm credit_count vào attendance_corrections ──
		{
			ID: "20250406000001",
			Migrate: func(tx *gorm.DB) error {
				return tx.Exec(`ALTER TABLE attendance_corrections ADD COLUMN IF NOT EXISTS credit_count INTEGER NOT NULL DEFAULT 1`).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Exec(`ALTER TABLE attendance_corrections DROP COLUMN IF EXISTS credit_count`).Error
			},
		},
	}
}
