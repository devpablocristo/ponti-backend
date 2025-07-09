package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	env "github.com/alphacodinggroup/ponti-backend/pkg/config/env"
	wire "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/wire"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		<-sigChan
		log.Println("Received termination signal. Shutting down the application...")
		cancel()
	}()

	deps, err := wire.Initialize()
	if err != nil {
		log.Fatalf("Error initializing dependencies: %s", err)
	}

	currentEnv := env.GetFromString(deps.Config.General.Environment)
	switch currentEnv {
	case env.Local:
		// INFO: las vars se cargan desde env.local
		if err := runMigrations(deps.Config.DB, deps.Config.Migrations); err != nil {
			log.Fatalf("Failed to run SQL migrations: %v", err)
		}
	case env.Dev, env.Staging:
		// INFO: las vars se cargan desde env.dev y env.staging
		if err := runGormMigrations(ctx, deps.GormRepo); err != nil {
			log.Fatalf("Failed to run Gorm migrations: %v", err)
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

	// Run the HTTP server
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		if err := runHttpServer(ctx, deps); err != nil {
			log.Fatalf("Error running HTTP server: %v", err)
		}
	}()

	wg.Wait()

	log.Println("Application terminated successfully.")
}
