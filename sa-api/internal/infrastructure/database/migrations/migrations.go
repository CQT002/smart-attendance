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

		// ── 007: Tạo bảng leave_requests — Nghỉ phép ──
		{
			ID: "20250407000001",
			Migrate: func(tx *gorm.DB) error {
				type LeaveRequest struct {
					entity.LeaveRequest
				}
				return tx.AutoMigrate(&LeaveRequest{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("leave_requests")
			},
		},

		// ── 008: Thêm leave_balance vào users + seed 4 ngày phép cho user hiện tại ──
		{
			ID: "20250407000002",
			Migrate: func(tx *gorm.DB) error {
				// Thêm column leave_balance vào users
				return tx.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS leave_balance DECIMAL(5,1) NOT NULL DEFAULT 0`).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Exec(`ALTER TABLE users DROP COLUMN IF EXISTS leave_balance`).Error
			},
		},

		// ── 009: Bổ sung index hiệu năng cho bảng lớn ──
		{
			ID: "20250407000003",
			Migrate: func(tx *gorm.DB) error {
				sqls := []string{
					// leave_requests: index cho query ngày nghỉ theo user
					`CREATE INDEX IF NOT EXISTS idx_leave_user_leave_date ON leave_requests (user_id, leave_date)`,

					// attendance_corrections: index created_at cho sort/filter
					`CREATE INDEX IF NOT EXISTS idx_correction_branch_created ON attendance_corrections (branch_id, created_at DESC)`,

					// leave_requests: index cho sort theo thời gian tạo trong branch
					`CREATE INDEX IF NOT EXISTS idx_leave_branch_created ON leave_requests (branch_id, created_at DESC)`,

					// attendance_logs: index cho query status theo user (lịch sử cá nhân filter status)
					`CREATE INDEX IF NOT EXISTS idx_attendance_user_status ON attendance_logs (user_id, status)`,
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
					`DROP INDEX IF EXISTS idx_leave_user_leave_date`,
					`DROP INDEX IF EXISTS idx_correction_branch_created`,
					`DROP INDEX IF EXISTS idx_leave_branch_created`,
					`DROP INDEX IF EXISTS idx_attendance_user_status`,
				}
				for _, sql := range sqls {
					if err := tx.Exec(sql).Error; err != nil {
						return err
					}
				}
				return nil
			},
		},

		// ── 010: Tạo bảng overtime_requests — Tăng ca ──
		{
			ID: "20250407000004",
			Migrate: func(tx *gorm.DB) error {
				type OvertimeRequest struct {
					entity.OvertimeRequest
				}
				return tx.AutoMigrate(&OvertimeRequest{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("overtime_requests")
			},
		},

		// ── 011: Thêm deleted_at vào tất cả bảng + bỏ column overtime + thêm overtime_request_id ──
		{
			ID: "20250407000005",
			Migrate: func(tx *gorm.DB) error {
				tables := []string{
					"branches", "users", "shifts", "wifi_configs", "gps_configs",
					"attendance_logs", "daily_summaries", "attendance_corrections", "leave_requests",
				}
				for _, t := range tables {
					if err := tx.Exec("ALTER TABLE " + t + " ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ").Error; err != nil {
						return err
					}
					if err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_" + t + "_deleted_at ON " + t + " (deleted_at)").Error; err != nil {
						return err
					}
				}

				// Bỏ column overtime trong attendance_logs (track qua overtime_requests)
				if err := tx.Exec("ALTER TABLE attendance_logs DROP COLUMN IF EXISTS overtime").Error; err != nil {
					return err
				}

				// Thêm overtime_request_id vào attendance_logs (FK tới overtime_requests đã tạo ở migration 010)
				sqls := []string{
					"ALTER TABLE attendance_logs ADD COLUMN IF NOT EXISTS overtime_request_id BIGINT REFERENCES overtime_requests(id)",
					"CREATE INDEX IF NOT EXISTS idx_attendance_overtime_request ON attendance_logs (overtime_request_id)",
				}
				for _, sql := range sqls {
					if err := tx.Exec(sql).Error; err != nil {
						return err
					}
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				tables := []string{
					"branches", "users", "shifts", "wifi_configs", "gps_configs",
					"attendance_logs", "daily_summaries", "attendance_corrections", "leave_requests",
				}
				for _, t := range tables {
					tx.Exec("DROP INDEX IF EXISTS idx_" + t + "_deleted_at")
					tx.Exec("ALTER TABLE " + t + " DROP COLUMN IF EXISTS deleted_at")
				}
				tx.Exec("ALTER TABLE attendance_logs ADD COLUMN IF NOT EXISTS overtime DECIMAL(5,2) DEFAULT 0")
				tx.Exec("DROP INDEX IF EXISTS idx_attendance_overtime_request")
				tx.Exec("ALTER TABLE attendance_logs DROP COLUMN IF EXISTS overtime_request_id")
				return nil
			},
		},

		// ── 012: Bổ sung correction_type, overtime_request_id cho attendance_corrections ──
		// Cho phép AttendanceLogID nullable (overtime correction không cần attendance log)
		{
			ID: "20250407000006",
			Migrate: func(tx *gorm.DB) error {
				sqls := []string{
					// Thêm correction_type (default 'attendance' cho data cũ)
					"ALTER TABLE attendance_corrections ADD COLUMN IF NOT EXISTS correction_type VARCHAR(20) NOT NULL DEFAULT 'attendance'",
					// Thêm overtime_request_id FK
					"ALTER TABLE attendance_corrections ADD COLUMN IF NOT EXISTS overtime_request_id BIGINT REFERENCES overtime_requests(id)",
					"CREATE INDEX IF NOT EXISTS idx_correction_overtime_request ON attendance_corrections (overtime_request_id)",
					// Cho phép attendance_log_id nullable (overtime corrections không dùng)
					"ALTER TABLE attendance_corrections ALTER COLUMN attendance_log_id DROP NOT NULL",
					// Drop unique index cũ trên attendance_log_id (vì giờ nullable)
					"DROP INDEX IF EXISTS uniq_correction_log",
					// Tạo partial unique index: chỉ enforce unique khi attendance_log_id NOT NULL và chưa soft-delete
					"CREATE UNIQUE INDEX IF NOT EXISTS uniq_correction_log ON attendance_corrections (attendance_log_id) WHERE attendance_log_id IS NOT NULL AND deleted_at IS NULL",
					// Unique: mỗi overtime_request chỉ có 1 correction (chưa soft-delete)
					"CREATE UNIQUE INDEX IF NOT EXISTS uniq_correction_overtime ON attendance_corrections (overtime_request_id) WHERE overtime_request_id IS NOT NULL AND deleted_at IS NULL",
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
					"DROP INDEX IF EXISTS uniq_correction_overtime",
					"DROP INDEX IF EXISTS idx_correction_overtime_request",
					"ALTER TABLE attendance_corrections DROP COLUMN IF EXISTS overtime_request_id",
					"ALTER TABLE attendance_corrections DROP COLUMN IF EXISTS correction_type",
				}
				for _, sql := range sqls {
					tx.Exec(sql)
				}
				return nil
			},
		},

		// ── 013: Thêm fields vào shifts + tạo bảng schedulers ──
		// Shift: morning_end, afternoon_start, ot_min_checkin_hour, ot_start_hour, ot_end_hour
		// Schedulers: bảng quản lý tác vụ định kỳ từ DB
		{
			ID: "20250408000001",
			Migrate: func(tx *gorm.DB) error {
				// Thêm fields mới vào shifts
				shiftCols := []string{
					"ALTER TABLE shifts ADD COLUMN IF NOT EXISTS morning_end VARCHAR(5) NOT NULL DEFAULT '12:00'",
					"ALTER TABLE shifts ADD COLUMN IF NOT EXISTS afternoon_start VARCHAR(5) NOT NULL DEFAULT '13:00'",
					"ALTER TABLE shifts ADD COLUMN IF NOT EXISTS ot_min_checkin_hour INTEGER NOT NULL DEFAULT 17",
					"ALTER TABLE shifts ADD COLUMN IF NOT EXISTS ot_start_hour INTEGER NOT NULL DEFAULT 18",
					"ALTER TABLE shifts ADD COLUMN IF NOT EXISTS ot_end_hour INTEGER NOT NULL DEFAULT 22",
				}
				for _, sql := range shiftCols {
					if err := tx.Exec(sql).Error; err != nil {
						return err
					}
				}

				// Tạo bảng schedulers
				type Scheduler struct {
					entity.Scheduler
				}
				if err := tx.AutoMigrate(&Scheduler{}); err != nil {
					return err
				}

				// Seed schedulers mặc định
				seeds := []string{
					`INSERT INTO schedulers (name, description, cron_expr, is_active, timeout_sec, created_at, updated_at)
					 VALUES ('leave_accrual', 'Cộng 1 ngày phép cho tất cả user active vào đầu tháng', '30 0 1 * *', true, 30, NOW(), NOW())
					 ON CONFLICT DO NOTHING`,
					`INSERT INTO schedulers (name, description, cron_expr, is_active, timeout_sec, created_at, updated_at)
					 VALUES ('correction_auto_reject', 'Tự động từ chối yêu cầu chấm công bù PENDING của tháng trước', '5 0 1 * *', true, 30, NOW(), NOW())
					 ON CONFLICT DO NOTHING`,
					`INSERT INTO schedulers (name, description, cron_expr, is_active, timeout_sec, created_at, updated_at)
					 VALUES ('leave_auto_reject', 'Tự động từ chối yêu cầu nghỉ phép PENDING của tháng trước', '5 0 1 * *', true, 30, NOW(), NOW())
					 ON CONFLICT DO NOTHING`,
				}
				for _, sql := range seeds {
					if err := tx.Exec(sql).Error; err != nil {
						return err
					}
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				tx.Exec("DROP TABLE IF EXISTS schedulers")
				cols := []string{"morning_end", "afternoon_start", "ot_min_checkin_hour", "ot_start_hour", "ot_end_hour"}
				for _, col := range cols {
					tx.Exec("ALTER TABLE shifts DROP COLUMN IF EXISTS " + col)
				}
				return nil
			},
		},
		// ── 014: Thêm regular_end_day, regular_end_time vào shifts ──
		// Cấu hình khung giờ làm việc chính thức: T2→RegularEndDay tại RegularEndTime
		{
			ID: "20260408000001",
			Migrate: func(tx *gorm.DB) error {
				sqls := []string{
					"ALTER TABLE shifts ADD COLUMN IF NOT EXISTS regular_end_day INTEGER NOT NULL DEFAULT 6",
					"ALTER TABLE shifts ADD COLUMN IF NOT EXISTS regular_end_time VARCHAR(5) NOT NULL DEFAULT '12:00'",
				}
				for _, sql := range sqls {
					if err := tx.Exec(sql).Error; err != nil {
						return err
					}
				}
				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				tx.Exec("ALTER TABLE shifts DROP COLUMN IF EXISTS regular_end_day")
				tx.Exec("ALTER TABLE shifts DROP COLUMN IF EXISTS regular_end_time")
				return nil
			},
		},

		// ── 015: Fix shift records có OT fields = 0 (seeded trước migration 013) ──
		{
			ID: "20260409000001",
			Migrate: func(tx *gorm.DB) error {
				sqls := []string{
					"UPDATE shifts SET morning_end = '12:00' WHERE morning_end IS NULL OR morning_end = ''",
					"UPDATE shifts SET afternoon_start = '13:00' WHERE afternoon_start IS NULL OR afternoon_start = ''",
					"UPDATE shifts SET ot_min_checkin_hour = 17 WHERE ot_min_checkin_hour = 0",
					"UPDATE shifts SET ot_start_hour = 18 WHERE ot_start_hour = 0",
					"UPDATE shifts SET ot_end_hour = 22 WHERE ot_end_hour = 0",
					"UPDATE shifts SET regular_end_day = 6 WHERE regular_end_day = 0 AND regular_end_time = '12:00'",
					"UPDATE shifts SET regular_end_time = '12:00' WHERE regular_end_time IS NULL OR regular_end_time = ''",
				}
				for _, sql := range sqls {
					if err := tx.Exec(sql).Error; err != nil {
						return err
					}
				}
				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				return nil // data fix, không rollback
			},
		},

		// ── 016: Tạo bảng holidays + partial unique index + seed VN 2026 ──
		{
			ID: "20260413000001",
			Migrate: func(tx *gorm.DB) error {
				type Holiday struct {
					entity.Holiday
				}
				if err := tx.AutoMigrate(&Holiday{}); err != nil {
					return err
				}

				// Partial unique index: 1 ngày chỉ có 1 holiday active
				// (soft-deleted rows bị loại bỏ — cho phép tạo lại holiday mới trên cùng date sau khi xoá)
				if err := tx.Exec(`
					CREATE UNIQUE INDEX IF NOT EXISTS uniq_holiday_date
						ON holidays (date)
						WHERE deleted_at IS NULL
				`).Error; err != nil {
					return err
				}

				// Seed ngày lễ Việt Nam 2026 (coefficient 3.00 = 300% theo Luật LĐ 2019)
				seeds := []string{
					`INSERT INTO holidays (name, date, year, coefficient, type, is_compensated, description, created_at, updated_at)
					 VALUES ('Tết Dương lịch', '2026-01-01', 2026, 3.00, 'national', false, 'Tết Dương lịch', NOW(), NOW())
					 ON CONFLICT DO NOTHING`,
					`INSERT INTO holidays (name, date, year, coefficient, type, is_compensated, description, created_at, updated_at)
					 VALUES ('Giao thừa Tết Nguyên đán', '2026-02-16', 2026, 3.00, 'national', false, 'Âm 29 tháng Chạp', NOW(), NOW())
					 ON CONFLICT DO NOTHING`,
					`INSERT INTO holidays (name, date, year, coefficient, type, is_compensated, description, created_at, updated_at)
					 VALUES ('Mùng 1 Tết Bính Ngọ', '2026-02-17', 2026, 3.00, 'national', false, 'Âm mùng 1 tháng Giêng', NOW(), NOW())
					 ON CONFLICT DO NOTHING`,
					`INSERT INTO holidays (name, date, year, coefficient, type, is_compensated, description, created_at, updated_at)
					 VALUES ('Mùng 2 Tết Bính Ngọ', '2026-02-18', 2026, 3.00, 'national', false, 'Âm mùng 2 tháng Giêng', NOW(), NOW())
					 ON CONFLICT DO NOTHING`,
					`INSERT INTO holidays (name, date, year, coefficient, type, is_compensated, description, created_at, updated_at)
					 VALUES ('Mùng 3 Tết Bính Ngọ', '2026-02-19', 2026, 3.00, 'national', false, 'Âm mùng 3 tháng Giêng', NOW(), NOW())
					 ON CONFLICT DO NOTHING`,
					`INSERT INTO holidays (name, date, year, coefficient, type, is_compensated, description, created_at, updated_at)
					 VALUES ('Mùng 4 Tết Bính Ngọ', '2026-02-20', 2026, 3.00, 'national', false, 'Âm mùng 4 tháng Giêng', NOW(), NOW())
					 ON CONFLICT DO NOTHING`,
					`INSERT INTO holidays (name, date, year, coefficient, type, is_compensated, description, created_at, updated_at)
					 VALUES ('Giỗ Tổ Hùng Vương', '2026-04-26', 2026, 3.00, 'national', false, 'Âm 10 tháng 3', NOW(), NOW())
					 ON CONFLICT DO NOTHING`,
					`INSERT INTO holidays (name, date, year, coefficient, type, is_compensated, description, created_at, updated_at)
					 VALUES ('Giải phóng miền Nam 30/4', '2026-04-30', 2026, 3.00, 'national', false, 'Ngày Giải phóng miền Nam, thống nhất đất nước', NOW(), NOW())
					 ON CONFLICT DO NOTHING`,
					`INSERT INTO holidays (name, date, year, coefficient, type, is_compensated, description, created_at, updated_at)
					 VALUES ('Quốc tế Lao động 1/5', '2026-05-01', 2026, 3.00, 'national', false, 'Ngày Quốc tế Lao động', NOW(), NOW())
					 ON CONFLICT DO NOTHING`,
					`INSERT INTO holidays (name, date, year, coefficient, type, is_compensated, description, created_at, updated_at)
					 VALUES ('Quốc khánh 2/9', '2026-09-02', 2026, 3.00, 'national', false, 'Ngày Quốc khánh', NOW(), NOW())
					 ON CONFLICT DO NOTHING`,
					`INSERT INTO holidays (name, date, year, coefficient, type, is_compensated, description, created_at, updated_at)
					 VALUES ('Quốc khánh — nghỉ bù', '2026-09-03', 2026, 3.00, 'national', true, 'Nghỉ bù cho Quốc khánh rơi vào Thứ 4', NOW(), NOW())
					 ON CONFLICT DO NOTHING`,
				}
				// Fix is_compensated row — set compensate_for column
				for _, s := range seeds {
					if err := tx.Exec(s).Error; err != nil {
						return err
					}
				}
				// Update compensate_for cho bản ghi nghỉ bù 2026-09-03
				if err := tx.Exec(`UPDATE holidays SET compensate_for = '2026-09-02' WHERE date = '2026-09-03'`).Error; err != nil {
					return err
				}
				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				tx.Exec(`DROP INDEX IF EXISTS uniq_holiday_date`)
				return tx.Migrator().DropTable("holidays")
			},
		},
	}
}
