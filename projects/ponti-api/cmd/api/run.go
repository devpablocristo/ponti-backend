package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
	wire "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/wire"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// runHttpServer registers routes in the Gin router and starts the HTTP server.
func runHttpServer(ctx context.Context, deps *wire.Dependencies) error {
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

func runMigrations(dbConfig config.DB, migConfig config.Migrations) error {
	m, err := migrate.New(
		"file://migrations",               //<--- migConfig.Dir, directorio de migraciones
		buildMigrateDatabaseURL(dbConfig), //<--- dbConfig, variables de entorno de la db
	)

	fmt.Println(migConfig.Dir)
	fmt.Println(migConfig.Dir)
	fmt.Println(migConfig.Dir)
	fmt.Println(migConfig.Dir)
	fmt.Println(migConfig.Dir)
	fmt.Println(migConfig.Dir)
	fmt.Println(migConfig.Dir)
	fmt.Println(migConfig.Dir)

	// m, err := migrate.New(
	// 	migConfig.Dir,
	// 	buildMigrateDatabaseURL(dbConfig),
	// )
	if err != nil {
		return fmt.Errorf("error creating migrate instance: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("error applying migrations: %w", err)
	}
	return nil
}

func runMigrationsWithInstance(sqlDB *sql.DB, dbConfig config.DB, migConfig config.Migrations) error {
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("creating postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations", //<--- migConfig.Dir, directorio de migraciones
		"postgres",          //<--- dbConfig.Name, nombre de la db
		driver,
	)

	fmt.Println(migConfig.Dir)
	fmt.Println(migConfig.Dir)
	fmt.Println(migConfig.Dir)
	fmt.Println(migConfig.Dir)
	fmt.Println(migConfig.Dir)
	fmt.Println(migConfig.Dir)
	fmt.Println(migConfig.Dir)
	fmt.Println(migConfig.Dir)

	// m, err := migrate.NewWithDatabaseInstance(
	// 	migConfig.Dir,
	// 	dbConfig.Name,
	// 	driver,
	// )
	if err != nil {
		return fmt.Errorf("creating migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("running migrations: %w", err)
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
