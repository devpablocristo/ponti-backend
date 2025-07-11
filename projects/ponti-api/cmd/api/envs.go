package main

import (
	"context"
	"log"

	env "github.com/alphacodinggroup/ponti-backend/pkg/config/env"
	wire "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/wire"
)

func setEnv(ctx context.Context, deps *wire.Dependencies) {
	currentEnv := env.GetFromString(deps.Config.General.Environment)
	switch currentEnv {
	case env.Local:
		// INFO: las vars se cargan desde env.local
		if err := runMigrations(deps.Config.DB, deps.Config.Migrations); err != nil {
			log.Fatalf("Failed to run SQL migrations: %v", err)
		}
	case env.Dev:
		// INFO: las vars se cargan desde env.dev
		if err := runGormMigrations(ctx, deps.GormRepo); err != nil {
			log.Fatalf("Failed to run Gorm migrations: %v", err)
		}
		if err := seedDatabase(ctx, deps.GormRepo); err != nil {
			log.Fatalf("Failed to run database seeders: %v", err)
		}
	case env.Stg:
		// INFO: las vars se cargan desde env.stg
		if err := runMigrationsWithInstance(deps.GormRepo.GetSQLDB(), deps.Config.DB, deps.Config.Migrations); err != nil {
			log.Fatalf("Failed to run SQL migrations: %v", err)
		}
		if err := seedDatabase(ctx, deps.GormRepo); err != nil {
			log.Fatalf("Failed to run database seeders: %v", err)
		}
	case env.Cloud:
		// TODO: las vars deberian configurarse en GCP y cargarse en el paquete config
		// TODO: harcodeo los valores que estaban antes de los cambios
		// TODO: cuando esten configurandas las vars, eliminar el hardcodeo
		deps.Config.DB.Name = "postgres"
		deps.Config.Migrations.Dir = "file://migrations"

		if err := runMigrationsWithInstance(deps.GormRepo.GetSQLDB(), deps.Config.DB, deps.Config.Migrations); err != nil {
			log.Fatalf("Failed to run SQL migrations: %v", err)
		}
	default:
		log.Fatalf("Unsupported environment: %s", currentEnv)
	}
}
