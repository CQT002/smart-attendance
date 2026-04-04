package database

import (
	"embed"
	"log/slog"

	"gorm.io/gorm"
)

//go:embed sql/seed_data.sql
//go:embed sql/gen_bulk_data.sql
var sqlFiles embed.FS

// RunSeeder chạy seed_data.sql rồi gen_bulk_data.sql nếu database chưa có dữ liệu.
// Kiểm tra bằng cách đếm số bản ghi trong bảng branches — nếu = 0 thì seed.
func RunSeeder(db *gorm.DB) error {
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM branches").Scan(&count).Error; err != nil {
		slog.Warn("seeder: cannot check branches count, skipping seed", "error", err)
		return nil
	}

	if count > 0 {
		slog.Info("seeder: database already has data, skipping seed", "branches_count", count)
		return nil
	}

	// Bước 1: Chạy seed_data.sql — dữ liệu gốc thực tế
	seedSQL, err := sqlFiles.ReadFile("sql/seed_data.sql")
	if err != nil {
		return err
	}

	slog.Info("seeder: running seed_data.sql...")
	if err := db.Exec(string(seedSQL)).Error; err != nil {
		return err
	}
	slog.Info("seeder: seed_data.sql completed")

	// Bước 2: Chạy gen_bulk_data.sql — sinh thêm 99 chi nhánh + ~4995 nhân viên
	bulkSQL, err := sqlFiles.ReadFile("sql/gen_bulk_data.sql")
	if err != nil {
		return err
	}

	slog.Info("seeder: running gen_bulk_data.sql...")
	if err := db.Exec(string(bulkSQL)).Error; err != nil {
		return err
	}
	slog.Info("seeder: gen_bulk_data.sql completed")

	return nil
}
