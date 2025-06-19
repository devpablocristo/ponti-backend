package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
	wire "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/wire"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunHttpServer registers routes in the Gin router and starts the HTTP server.
func RunHttpServer(ctx context.Context, deps *wire.Dependencies) error {
	if deps == nil {
		return errors.New("dependencies cannot be nil")
	}

	log.Println("Registering HTTP routes...")

	// Set up the Gin router with global middlewares
	deps.GinEngine.GetRouter().Use(deps.Middlewares.GetGlobal()...)

	// Register all application routes.
	log.Println("Starting HTTP Server...")
	registerHttpRoutes(deps)

	// Start the HTTP server (e.g., on port 8080).
	return deps.GinEngine.RunServer(ctx)
}

// registerHttpRoutes registers all application routes in the Gin router.
func registerHttpRoutes(deps *wire.Dependencies) {
	deps.LotHandler.Routes()
	deps.CustomerHandler.Routes()
	deps.CampaignHandler.Routes()
	deps.InvestorHandler.Routes()
	deps.FieldHandler.Routes()
	deps.ProjectHandler.Routes()
	deps.CropHandler.Routes()
	deps.ManagerHandler.Routes()
	deps.LeaseTypeHandler.Routes()
}

func runMigrations(dbConfig config.DB) error {
	m, err := migrate.New(
		"file://migrations",
		buildMigrateDatabaseURL(dbConfig),
	)
	if err != nil {
		return fmt.Errorf("error creating migrate instance: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("error applying migrations: %w", err)
	}
	return nil
}

func buildMigrateDatabaseURL(config config.DB) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Name,
		config.SSLMode,
	)
}
