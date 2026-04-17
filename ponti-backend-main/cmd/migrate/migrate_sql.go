package main

import (
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"
	"log"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	config "github.com/alphacodinggroup/ponti-backend/cmd/config"
)

func runMigrations(dbConfig config.DB, migConfig config.Migrations) error {
	// Nota: el backend no conoce estrategias de deploy. Las migraciones corren siempre en public.
	dsn := buildPostgresDSN(dbConfig)
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer func() { _ = sqlDB.Close() }()

	return runMigrationsWithInstance(sqlDB, dbConfig, migConfig)
}

func runMigrationsWithInstance(sqlDB *sql.DB, dbConfig config.DB, migConfig config.Migrations) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	return runMigrationsWithContext(ctx, sqlDB, dbConfig, migConfig)
}

func runMigrationsWithContext(ctx context.Context, sqlDB *sql.DB, dbConfig config.DB, migConfig config.Migrations) error {
	// Adquirir lock de migración para evitar ejecuciones concurrentes.
	unlock, err := acquireMigrationLock(ctx, sqlDB, dbConfig.Name)
	if err != nil {
		return fmt.Errorf("failed to acquire migration lock: %w", err)
	}
	defer unlock()

	log.Printf("Migration lock acquired for database: %s", dbConfig.Name)

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{
		DatabaseName: dbConfig.Name,
	})
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
		if strings.Contains(err.Error(), "Dirty database version") || strings.Contains(err.Error(), "dirty") {
			return fmt.Errorf("dirty migration state - reset the database or fix and force version manually: %w", err)
		}
		if strings.Contains(err.Error(), "cannot drop columns from view") {
			return fmt.Errorf("migration failed due to incompatible view shape (cannot drop columns from view) - reset the database or adjust migrations: %w", err)
		}
		return fmt.Errorf("running migrations: %w", err)
	}

	log.Printf("Migrations completed successfully for database: %s", dbConfig.Name)
	return nil
}

func buildPostgresDSN(cfg config.DB) string {
	host := strings.TrimSpace(cfg.Host)
	user := strings.TrimSpace(cfg.User)
	pass := strings.TrimSpace(cfg.Password)
	name := strings.TrimSpace(cfg.Name)
	ssl := strings.TrimSpace(cfg.SSLMode)

	// DSN key/value (lib/pq) soporta host por unix socket (/cloudsql/...) y port separado.
	if ssl == "" {
		return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d", host, user, pass, name, cfg.Port)
	}
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s", host, user, pass, name, cfg.Port, ssl)
}

// acquireMigrationLock adquiere un lock de migración usando pg_advisory_lock.
// El lock ID se deriva del nombre de la base de datos para evitar ejecuciones concurrentes
// de migraciones dentro de la misma DB.
func acquireMigrationLock(ctx context.Context, sqlDB *sql.DB, databaseName string) (func(), error) {
	return acquireMigrationLockWithTimeout(ctx, sqlDB, databaseName, 5*time.Minute)
}

// acquireMigrationLockWithTimeout adquiere un lock con timeout configurable
func acquireMigrationLockWithTimeout(ctx context.Context, sqlDB *sql.DB, databaseName string, timeout time.Duration) (func(), error) {
	lockID := hashDatabaseName(databaseName)

	log.Printf("Attempting to acquire migration lock for database: %s (lock_id: %d)", databaseName, lockID)

	startTime := time.Now()
	deadline := startTime.Add(timeout)

	// Loop con pg_try_advisory_lock hasta adquirir el lock o timeout
	for {
		// Verificar timeout
		if time.Now().After(deadline) {
			waitTime := time.Since(startTime)
			return nil, fmt.Errorf("timeout waiting for migration lock (database: %s, lock_id: %d, waited: %v)",
				databaseName, lockID, waitTime)
		}

		// Intentar adquirir el lock de forma no bloqueante
		var acquired bool
		err := sqlDB.QueryRowContext(ctx,
			"SELECT pg_try_advisory_lock($1)", lockID).Scan(&acquired)
		if err != nil {
			return nil, fmt.Errorf("failed to try acquire migration lock (database: %s, lock_id: %d): %w",
				databaseName, lockID, err)
		}

		if acquired {
			waitTime := time.Since(startTime)
			if waitTime > 100*time.Millisecond {
				log.Printf("✅ Migration lock acquired for database: %s (lock_id: %d, waited: %v)",
					databaseName, lockID, waitTime)
			} else {
				log.Printf("✅ Migration lock acquired immediately for database: %s (lock_id: %d)",
					databaseName, lockID)
			}

			// Retornar función de unlock con defer garantizado
			unlock := func() {
				_, _ = sqlDB.ExecContext(context.Background(), "SELECT pg_advisory_unlock($1)", lockID)
				log.Printf("🔓 Migration lock released for database: %s (lock_id: %d)", databaseName, lockID)
			}

			return unlock, nil
		}

		// Lock no disponible, esperar un poco antes de reintentar
		waitTime := time.Since(startTime)
		if waitTime < 1*time.Second {
			// Primer segundo: esperar 100ms
			time.Sleep(100 * time.Millisecond)
		} else if waitTime < 10*time.Second {
			// Primeros 10 segundos: esperar 500ms
			time.Sleep(500 * time.Millisecond)
		} else {
			// Después de 10 segundos: esperar 1s y log cada 5s
			time.Sleep(1 * time.Second)
			if int(waitTime.Seconds())%5 == 0 {
				log.Printf("⏳ Waiting for migration lock (database: %s, lock_id: %d, waited: %v)...",
					databaseName, lockID, waitTime)
			}
		}
	}
}

// hashDatabaseName convierte un database name a un int64 para usar como lock ID en pg_advisory_lock
// Usa FNV-1a hash para generar un hash determinístico del database name
func hashDatabaseName(databaseName string) int64 {
	h := fnv.New64a()
	h.Write([]byte(databaseName))
	return int64(h.Sum64() >> 1)
}
