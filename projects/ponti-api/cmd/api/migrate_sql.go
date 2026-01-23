package main

import (
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"
	"log"
	"net/url"
	"strings"
	"time"

	_ "database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
)

func runMigrations(dbConfig config.DB, migConfig config.Migrations) error {
	schema := strings.TrimSpace(dbConfig.Schema)
	if schema == "" {
		schema = "public"
	}

	// Crear conexión temporal para inicializar schema y adquirir lock
	dsn := buildMigrateDatabaseURL(dbConfig)
	tempDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer tempDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Inicializar schema si no existe
	if err := initializeSchema(ctx, tempDB, dbConfig); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	// Adquirir lock de migración para evitar ejecuciones concurrentes
	unlock, err := acquireMigrationLock(ctx, tempDB, schema)
	if err != nil {
		return fmt.Errorf("failed to acquire migration lock: %w", err)
	}
	defer unlock()

	log.Printf("Migration lock acquired for schema: %s", schema)

	// DSN para migrate: agregar x-migrations-table si no es public
	migrateDSN := dsn
	if schema != "public" {
		if strings.Contains(migrateDSN, "?") {
			migrateDSN += "&"
		} else {
			migrateDSN += "?"
		}
		migrateDSN += fmt.Sprintf("x-migrations-table=%s.schema_migrations", quoteIdentifier(schema))
	}

	m, err := migrate.New(migConfig.Dir, migrateDSN)
	if err != nil {
		return fmt.Errorf("error creating migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("error applying migrations: %w", err)
	}

	log.Printf("Migrations completed successfully for schema: %s", schema)
	return nil
}

func runMigrationsWithInstance(sqlDB *sql.DB, dbConfig config.DB, migConfig config.Migrations) error {
	schema := strings.TrimSpace(dbConfig.Schema)
	if schema == "" {
		schema = "public"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Inicializar schema si no existe
	if err := initializeSchema(ctx, sqlDB, dbConfig); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	// Adquirir lock de migración para evitar ejecuciones concurrentes
	unlock, err := acquireMigrationLock(ctx, sqlDB, schema)
	if err != nil {
		return fmt.Errorf("failed to acquire migration lock: %w", err)
	}
	defer unlock()

	log.Printf("Migration lock acquired for schema: %s", schema)

	// Configurar driver de migrate para usar el schema
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{
		DatabaseName:    dbConfig.Name,
		SchemaName:      schema, // Usar el schema especificado
		MigrationsTable: fmt.Sprintf("%s.schema_migrations", quoteIdentifier(schema)),
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
		return fmt.Errorf("running migrations: %w", err)
	}

	log.Printf("Migrations completed successfully for schema: %s", schema)
	return nil
}

// quoteIdentifier escapa un identificador de PostgreSQL de forma segura
func quoteIdentifier(name string) string {
	escaped := strings.ReplaceAll(name, `"`, `""`)
	return fmt.Sprintf(`"%s"`, escaped)
}

func buildMigrateDatabaseURL(cfg config.DB) string {
	user := strings.TrimSpace(cfg.User)
	pass := strings.TrimSpace(cfg.Password)
	host := strings.TrimSpace(cfg.Host)
	name := strings.TrimSpace(cfg.Name)
	ssl := strings.TrimSpace(cfg.SSLMode)
	schema := strings.TrimSpace(cfg.Schema)
	if schema == "" {
		schema = "public"
	}

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, pass),
		Host:   fmt.Sprintf("%s:%d", host, cfg.Port),
		Path:   "/" + name,
	}

	q := url.Values{}
	if ssl != "" {
		q.Set("sslmode", ssl)
	}
	
	// Agregar search_path usando options=-c (se aplica a TODAS las conexiones)
	// Esto asegura que el driver de migrate use el schema correcto
	if schema != "" && schema != "public" {
		// Escapar el schema name para URL
		escapedSchema := url.QueryEscape(schema)
		q.Set("options", fmt.Sprintf("-csearch_path=%s,public", escapedSchema))
	}
	
	u.RawQuery = q.Encode()

	return u.String()
}

// initializeSchema crea el schema si no existe
// NO setea search_path aquí porque se configura en el DSN
func initializeSchema(ctx context.Context, sqlDB *sql.DB, dbConfig config.DB) error {
	schema := strings.TrimSpace(dbConfig.Schema)
	if schema == "" {
		schema = "public"
	}

	// Validar nombre de schema (seguridad básica)
	if err := validateSchemaName(schema); err != nil {
		return fmt.Errorf("invalid schema name: %w", err)
	}

	log.Printf("Initializing schema: %s", schema)

	// Crear schema si no existe
	createSchemaSQL := fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, quoteIdentifier(schema))
	if _, err := sqlDB.ExecContext(ctx, createSchemaSQL); err != nil {
		return fmt.Errorf("failed to create schema %s: %w", schema, err)
	}

	log.Printf("Schema %s initialized (search_path configured via DSN)", schema)
	return nil
}

// acquireMigrationLock adquiere un lock de migración usando pg_advisory_lock
// El lock ID se deriva del schema name para que cada schema tenga su propio lock
// Esto previene ejecuciones concurrentes de migraciones en el mismo schema
func acquireMigrationLock(ctx context.Context, sqlDB *sql.DB, schema string) (func(), error) {
	if schema == "" || schema == "public" {
		// No necesitamos lock para public (asumimos que solo se migra una vez)
		return func() {}, nil
	}

	// Calcular hash del schema name para usar como lock ID
	lockID := hashSchemaName(schema)

	// Intentar adquirir el lock de forma no bloqueante primero
	var acquired bool
	err := sqlDB.QueryRowContext(ctx, 
		"SELECT pg_try_advisory_lock($1)", lockID).Scan(&acquired)
	if err != nil {
		return nil, fmt.Errorf("failed to try acquire migration lock: %w", err)
	}

	if !acquired {
		// Lock ya está tomado, esperar con timeout bloqueante
		log.Printf("Migration lock for schema %s is held, waiting...", schema)
		
		// Usar un contexto con timeout para evitar esperar indefinidamente
		lockCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()

		// pg_advisory_lock bloquea hasta obtener el lock o timeout
		var result interface{}
		err := sqlDB.QueryRowContext(lockCtx,
			"SELECT pg_advisory_lock($1)", lockID).Scan(&result)
		if err != nil && err != sql.ErrNoRows {
			if lockCtx.Err() == context.DeadlineExceeded {
				return nil, fmt.Errorf("timeout waiting for migration lock (schema: %s)", schema)
			}
			return nil, fmt.Errorf("failed to acquire migration lock: %w", err)
		}
		
		log.Printf("Migration lock acquired for schema %s after waiting", schema)
	} else {
		log.Printf("Migration lock acquired immediately for schema %s", schema)
	}

	// Retornar función de unlock
	unlock := func() {
		_, _ = sqlDB.ExecContext(context.Background(), "SELECT pg_advisory_unlock($1)", lockID)
		log.Printf("Migration lock released for schema %s", schema)
	}

	return unlock, nil
}

// hashSchemaName convierte un schema name a un int64 para usar como lock ID en pg_advisory_lock
// Usa FNV-1a hash para generar un hash determinístico del schema name
func hashSchemaName(schema string) int64 {
	h := fnv.New64a()
	h.Write([]byte(schema))
	// pg_advisory_lock usa bigint (int64), convertir el hash
	// Usamos los 63 bits superiores para evitar problemas con signo
	return int64(h.Sum64() >> 1)
}

// validateSchemaName valida que el nombre del schema sea seguro
func validateSchemaName(schema string) error {
	if schema == "" {
		return fmt.Errorf("schema name cannot be empty")
	}

	// Nombres reservados de PostgreSQL
	reserved := []string{"pg_catalog", "pg_toast", "information_schema", "pg_temp", "pg_toast_temp"}
	for _, r := range reserved {
		if strings.EqualFold(schema, r) {
			return fmt.Errorf("schema name '%s' is reserved", schema)
		}
	}

	// Validar caracteres (solo alfanuméricos, guiones bajos y guiones)
	for _, r := range schema {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-') {
			return fmt.Errorf("schema name contains invalid character: %c", r)
		}
	}

	// No puede empezar con número
	if len(schema) > 0 && schema[0] >= '0' && schema[0] <= '9' {
		return fmt.Errorf("schema name cannot start with a number")
	}

	return nil
}
