package main

import (
	"context"
	"flag"
	"log"
	"time"

	config "github.com/devpablocristo/ponti-backend/cmd/config"
	gormRepo "github.com/devpablocristo/ponti-backend/pkg/databases/sql/gorm"
)

func main() {
	useGorm := flag.Bool("gorm", false, "Ejecutar migraciones GORM (AutoMigrate)")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error cargando configuración: %v", err)
	}

	if *useGorm {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		repo, err := gormRepo.Bootstrap(cfg.DB.Type, cfg.DB.Host, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode, cfg.DB.Port)
		if err != nil {
			log.Fatalf("Error iniciando repositorio GORM: %v", err)
		}

		if err := runGormMigrations(ctx, repo); err != nil {
			log.Fatalf("Error ejecutando migraciones GORM: %v", err)
		}
		log.Println("Migraciones GORM finalizadas.")
		return
	}

	if err := runMigrations(cfg.DB, cfg.Migrations); err != nil {
		log.Fatalf("Error ejecutando migraciones SQL: %v", err)
	}
	log.Println("Migraciones SQL finalizadas.")
}
