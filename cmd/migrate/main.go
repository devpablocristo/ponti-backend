package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/devpablocristo/platform/observability/go"

	config "github.com/devpablocristo/ponti-backend/cmd/config"
	gormRepo "github.com/devpablocristo/ponti-backend/internal/platform/persistence/gorm"
)

func main() {
	logger := observability.NewJSONLogger("ponti-migrate")
	slog.SetDefault(logger)

	useGorm := flag.Bool("gorm", false, "Ejecutar migraciones GORM (AutoMigrate)")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("loading configuration failed", "error", err.Error())
		os.Exit(1)
	}

	if *useGorm {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		repo, err := gormRepo.Bootstrap(cfg.DB.Type, cfg.DB.Host, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode, cfg.DB.Port)
		if err != nil {
			logger.Error("starting GORM repository failed", "error", err.Error())
			os.Exit(1)
		}

		if err := runGormMigrations(ctx, logger, repo); err != nil {
			logger.Error("running GORM migrations failed", "error", err.Error())
			os.Exit(1)
		}
		logger.Info("GORM migrations finished", "event", "migrations_completed", "kind", "gorm")
		return
	}

	if err := runMigrations(logger, cfg.DB, cfg.Migrations); err != nil {
		logger.Error("running SQL migrations failed", "error", err.Error())
		os.Exit(1)
	}
	logger.Info("SQL migrations finished", "event", "migrations_completed", "kind", "sql")
}
