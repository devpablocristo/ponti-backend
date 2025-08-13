package main

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
)

func runMigrations(dbConfig config.DB, migConfig config.Migrations) error {
	m, err := migrate.New(
		migConfig.Dir,
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

func runMigrationsWithInstance(sqlDB *sql.DB, dbConfig config.DB, migConfig config.Migrations) error {
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("creating postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		migConfig.Dir,
		dbConfig.Name,
		driver,
	)
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
