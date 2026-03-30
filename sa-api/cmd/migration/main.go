package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/hdbank/smart-attendance/config"
	"github.com/hdbank/smart-attendance/internal/infrastructure/database"
	"github.com/hdbank/smart-attendance/internal/infrastructure/database/migrations"

	"github.com/go-gormigrate/gormigrate/v2"
)

// Chạy migration độc lập:
//
//	go run ./cmd/migration -cmd up              # migrate tất cả (default)
//	go run ./cmd/migration -cmd down            # rollback migration cuối
//	go run ./cmd/migration -cmd rollback-to -id 20250330000001
//	go run ./cmd/migration -cmd reset           # rollback toàn bộ
func main() {
	var command string
	var migrationID string

	flag.StringVar(&command, "cmd", "up", "Migration command: up, down, rollback-to, reset")
	flag.StringVar(&migrationID, "id", "", "Migration ID for rollback-to command")
	flag.Parse()

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

	m := gormigrate.New(db, gormigrate.DefaultOptions, migrations.GetMigrations())

	switch command {
	case "up":
		if err := m.Migrate(); err != nil {
			slog.Error("migration failed", "error", err)
			os.Exit(1)
		}
		slog.Info("migrations applied successfully")

	case "down":
		allMigrations := migrations.GetMigrations()
		if len(allMigrations) == 0 {
			slog.Error("no migrations to rollback")
			os.Exit(1)
		}
		lastMigration := allMigrations[len(allMigrations)-1]
		if err := m.RollbackTo(lastMigration.ID); err != nil {
			slog.Error("rollback failed", "error", err)
			os.Exit(1)
		}
		slog.Info("last migration rolled back successfully")

	case "rollback-to":
		if migrationID == "" {
			slog.Error("migration ID is required for rollback-to command, use -id flag")
			os.Exit(1)
		}
		if err := m.RollbackTo(migrationID); err != nil {
			slog.Error("rollback failed", "migration_id", migrationID, "error", err)
			os.Exit(1)
		}
		slog.Info("rolled back to migration", "migration_id", migrationID)

	case "reset":
		if err := m.RollbackTo("0"); err != nil {
			slog.Error("reset failed", "error", err)
			os.Exit(1)
		}
		slog.Info("all migrations rolled back successfully")

	default:
		slog.Error("unknown command", "command", command)
		slog.Info("available commands: up, down, rollback-to, reset")
		os.Exit(1)
	}
}
