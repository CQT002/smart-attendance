package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/hdbank/smart-attendance/config"
	"github.com/hdbank/smart-attendance/internal/infrastructure/database"
)

// Sửa status cho các attendance records check-in sau 08:15 mà vẫn ghi "present"
// Chạy: go run cmd/fixdata/main.go
func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	db, err := database.NewPostgres(&cfg.Database)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}

	// Fix 1: check-in sau 08:15 mà status = 'present' → 'late'
	r1 := db.Exec(`
		UPDATE attendance_logs
		SET status = 'late', updated_at = NOW()
		WHERE status = 'present'
		  AND check_in_time IS NOT NULL
		  AND (EXTRACT(HOUR FROM check_in_time AT TIME ZONE 'Asia/Ho_Chi_Minh') * 60
		     + EXTRACT(MINUTE FROM check_in_time AT TIME ZONE 'Asia/Ho_Chi_Minh')) > (8 * 60 + 15)
	`)
	if r1.Error != nil {
		slog.Error("fix 1 failed", "error", r1.Error)
	} else {
		fmt.Printf("Fix 1: %d records present → late\n", r1.RowsAffected)
	}

	// Fix 2: đi trễ + về sớm (checkout trước 16:45) → 'late_early_leave'
	r2 := db.Exec(`
		UPDATE attendance_logs
		SET status = 'late_early_leave', updated_at = NOW()
		WHERE status = 'late'
		  AND check_out_time IS NOT NULL
		  AND (EXTRACT(HOUR FROM check_out_time AT TIME ZONE 'Asia/Ho_Chi_Minh') * 60
		     + EXTRACT(MINUTE FROM check_out_time AT TIME ZONE 'Asia/Ho_Chi_Minh')) < (16 * 60 + 45)
	`)
	if r2.Error != nil {
		slog.Error("fix 2 failed", "error", r2.Error)
	} else {
		fmt.Printf("Fix 2: %d records late → late_early_leave\n", r2.RowsAffected)
	}

	// Fix 3: đúng giờ + về sớm → 'early_leave'
	r3 := db.Exec(`
		UPDATE attendance_logs
		SET status = 'early_leave', updated_at = NOW()
		WHERE status = 'present'
		  AND check_out_time IS NOT NULL
		  AND (EXTRACT(HOUR FROM check_out_time AT TIME ZONE 'Asia/Ho_Chi_Minh') * 60
		     + EXTRACT(MINUTE FROM check_out_time AT TIME ZONE 'Asia/Ho_Chi_Minh')) < (16 * 60 + 45)
	`)
	if r3.Error != nil {
		slog.Error("fix 3 failed", "error", r3.Error)
	} else {
		fmt.Printf("Fix 3: %d records present → early_leave\n", r3.RowsAffected)
	}

	// Fix 4: early_leave nhưng check-in sau 08:15 → thực ra là late_early_leave
	r4 := db.Exec(`
		UPDATE attendance_logs
		SET status = 'late_early_leave', updated_at = NOW()
		WHERE status = 'early_leave'
		  AND check_in_time IS NOT NULL
		  AND (EXTRACT(HOUR FROM check_in_time AT TIME ZONE 'Asia/Ho_Chi_Minh') * 60
		     + EXTRACT(MINUTE FROM check_in_time AT TIME ZONE 'Asia/Ho_Chi_Minh')) > (8 * 60 + 15)
	`)
	if r4.Error != nil {
		slog.Error("fix 4 failed", "error", r4.Error)
	} else {
		fmt.Printf("Fix 4: %d records early_leave → late_early_leave (check-in after 08:15)\n", r4.RowsAffected)
	}
}
