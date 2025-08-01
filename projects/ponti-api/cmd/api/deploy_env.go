package main

import (
	"context"
	"log"

	pkgenv "github.com/alphacodinggroup/ponti-backend/pkg/config/env"
	seed "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/api/seed"
	wire "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/wire"
)

func setDeployEnv(ctx context.Context, deps *wire.Dependencies) {
	platform := pkgenv.GetPlatformFromString(deps.Config.Deploy.Platform)
	env := pkgenv.GetEnvFromString(deps.Config.Deploy.Environment)

	switch platform {
	case pkgenv.Local:
		switch env {
		case pkgenv.Dev:
			if err := runMigrations(deps.Config.DB, deps.Config.Migrations); err != nil {
				log.Fatalf("Failed to run SQL migrations: %v", err)
			}
		}
	case pkgenv.Mix:
		switch env {
		case pkgenv.Dev:
			if err := runGormMigrations(ctx, deps.GormRepo); err != nil {
				log.Fatalf("Failed to run Gorm migrations: %v", err)
			}
			if err := seed.Base(ctx, deps.GormRepo); err != nil {
				log.Fatalf("Failed to run database seeders: %v", err)
			}
		case pkgenv.Stg:
			if err := runMigrationsWithInstance(deps.GormRepo.GetSQLDB(), deps.Config.DB, deps.Config.Migrations); err != nil {
				log.Fatalf("Failed to run SQL migrations: %v", err)
			}
			if err := seed.Base(ctx, deps.GormRepo); err != nil {
				log.Fatalf("Failed to run database seeders: %v", err)
			}

		default:
			log.Fatalf("Unsupported environment: %s", env)
		}

	case pkgenv.GCP:
		switch env {
		case pkgenv.Dev, pkgenv.Stg, pkgenv.Prod:
			deps.Config.DB.Name = "postgres"
			deps.Config.Migrations.Dir = "file://migrations"

			if err := runMigrationsWithInstance(deps.GormRepo.GetSQLDB(), deps.Config.DB, deps.Config.Migrations); err != nil {
				log.Fatalf("Failed to run SQL migrations: %v", err)
			}
		default:
			log.Fatalf("Unsupported environment: %s", env)
		}

	default:
		log.Fatalf("Unsupported platform: %s", platform)
	}
}
