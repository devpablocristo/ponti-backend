package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
)

// TestSchemaIsolation verifica que los schemas están completamente aislados
// entre sí y que no afectan al schema public
func TestSchemaIsolation(t *testing.T) {
	// Skip si no hay configuración de DB de test
	if os.Getenv("TEST_DB_HOST") == "" {
		t.Skip("Skipping integration test: TEST_DB_HOST not set")
	}

	// Configurar DB de test
	port := 5432
	if p := os.Getenv("TEST_DB_PORT"); p != "" {
		port, _ = strconv.Atoi(p)
	}
	dbConfig := config.DB{
		Type:     "postgres",
		Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		User:     getEnvOrDefault("TEST_DB_USER", "admin"),
		Password: getEnvOrDefault("TEST_DB_PASSWORD", "admin"),
		Name:     getEnvOrDefault("TEST_DB_NAME", "ponti_api_db"),
		SSLMode:  getEnvOrDefault("TEST_DB_SSL_MODE", "disable"),
		Port:     port,
	}

	migConfig := config.Migrations{
		Dir: getEnvOrDefault("TEST_MIGRATIONS_DIR", "file://migrations"),
	}

	ctx := context.Background()

	// Conectar a la DB para verificaciones
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		dbConfig.Host, dbConfig.User, dbConfig.Password, dbConfig.Name, dbConfig.Port, dbConfig.SSLMode)
	
	checkDB, err := sql.Open("postgres", dsn)
	require.NoError(t, err, "Failed to connect to test database")
	defer checkDB.Close()

	// Limpiar schemas de test antes de empezar
	cleanupTestSchemas(t, checkDB, []string{"pr_1", "pr_2"})

	// Guardar estado inicial de public
	publicTablesBefore := getTableCount(t, checkDB, "public")

	t.Run("Migrate pr_1 and verify isolation", func(t *testing.T) {
		// Configurar schema pr_1
		dbConfig1 := dbConfig
		dbConfig1.Schema = "pr_1"

		// Ejecutar migraciones en pr_1
		err := runMigrations(dbConfig1, migConfig)
		require.NoError(t, err, "Failed to run migrations for pr_1")

		// Verificar que pr_1 tiene tablas
		pr1Tables := getTableCount(t, checkDB, "pr_1")
		assert.Greater(t, pr1Tables, 0, "pr_1 should have tables after migrations")

		// Verificar que pr_2 está vacío
		pr2Tables := getTableCount(t, checkDB, "pr_2")
		assert.Equal(t, 0, pr2Tables, "pr_2 should be empty")

		// Verificar que public no cambió
		publicTablesAfter := getTableCount(t, checkDB, "public")
		assert.Equal(t, publicTablesBefore, publicTablesAfter, "public schema should not change")
	})

	t.Run("Migrate pr_2 and verify pr_1 unchanged", func(t *testing.T) {
		// Guardar estado de pr_1 antes de migrar pr_2
		pr1TablesBefore := getTableCount(t, checkDB, "pr_1")

		// Configurar schema pr_2
		dbConfig2 := dbConfig
		dbConfig2.Schema = "pr_2"

		// Ejecutar migraciones en pr_2
		err := runMigrations(dbConfig2, migConfig)
		require.NoError(t, err, "Failed to run migrations for pr_2")

		// Verificar que pr_2 tiene tablas
		pr2Tables := getTableCount(t, checkDB, "pr_2")
		assert.Greater(t, pr2Tables, 0, "pr_2 should have tables after migrations")

		// Verificar que pr_1 no cambió
		pr1TablesAfter := getTableCount(t, checkDB, "pr_1")
		assert.Equal(t, pr1TablesBefore, pr1TablesAfter, "pr_1 should not change when migrating pr_2")

		// Verificar que ambas tienen el mismo número de tablas (mismas migraciones)
		assert.Equal(t, pr1TablesBefore, pr2Tables, "Both schemas should have the same number of tables")
	})

	t.Run("Verify schema_migrations table isolation", func(t *testing.T) {
		// Verificar que cada schema tiene su propia tabla schema_migrations
		pr1Migrations := getMigrationVersion(t, checkDB, "pr_1")
		pr2Migrations := getMigrationVersion(t, checkDB, "pr_2")

		assert.Greater(t, pr1Migrations, 0, "pr_1 should have migration version")
		assert.Greater(t, pr2Migrations, 0, "pr_2 should have migration version")
		assert.Equal(t, pr1Migrations, pr2Migrations, "Both schemas should have the same migration version")
	})

	t.Run("Verify data isolation", func(t *testing.T) {
		// Crear una tabla de prueba en pr_1
		_, err := checkDB.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS pr_1.test_isolation (id SERIAL PRIMARY KEY, data TEXT)`)
		require.NoError(t, err)

		// Insertar datos en pr_1
		_, err = checkDB.ExecContext(ctx, `INSERT INTO pr_1.test_isolation (data) VALUES ('pr_1_data')`)
		require.NoError(t, err)

		// Verificar que pr_2 no tiene la tabla
		var pr2TableExists bool
		err = checkDB.QueryRowContext(ctx,
			`SELECT EXISTS (
				SELECT 1 FROM information_schema.tables 
				WHERE table_schema = 'pr_2' AND table_name = 'test_isolation'
			)`).Scan(&pr2TableExists)
		require.NoError(t, err)
		assert.False(t, pr2TableExists, "pr_2 should not have test_isolation table")

		// Verificar que los datos están solo en pr_1
		var count int
		err = checkDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM pr_1.test_isolation`).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "pr_1 should have 1 row")

		// Verificar que public no tiene la tabla
		var publicTableExists bool
		err = checkDB.QueryRowContext(ctx,
			`SELECT EXISTS (
				SELECT 1 FROM information_schema.tables 
				WHERE table_schema = 'public' AND table_name = 'test_isolation'
			)`).Scan(&publicTableExists)
		require.NoError(t, err)
		assert.False(t, publicTableExists, "public should not have test_isolation table")
	})

	// Limpiar después del test
	cleanupTestSchemas(t, checkDB, []string{"pr_1", "pr_2"})
}

// TestMigrationLock verifica que el lock de migraciones funciona correctamente
func TestMigrationLock(t *testing.T) {
	if os.Getenv("TEST_DB_HOST") == "" {
		t.Skip("Skipping integration test: TEST_DB_HOST not set")
	}

	dbConfig := config.DB{
		Type:     "postgres",
		Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		User:     getEnvOrDefault("TEST_DB_USER", "admin"),
		Password: getEnvOrDefault("TEST_DB_PASSWORD", "admin"),
		Name:     getEnvOrDefault("TEST_DB_NAME", "ponti_api_db"),
		SSLMode:  getEnvOrDefault("TEST_DB_SSL_MODE", "disable"),
		Port:     5432,
		Schema:   "pr_lock_test",
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		dbConfig.Host, dbConfig.User, dbConfig.Password, dbConfig.Name, dbConfig.Port, dbConfig.SSLMode)
	
	sqlDB, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer sqlDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Limpiar schema de test
	_, _ = sqlDB.ExecContext(ctx, `DROP SCHEMA IF EXISTS pr_lock_test CASCADE`)

	// Crear schema
	_, err = sqlDB.ExecContext(ctx, `CREATE SCHEMA pr_lock_test`)
	require.NoError(t, err)

	// Adquirir lock
	unlock1, err := acquireMigrationLock(ctx, sqlDB, "pr_lock_test")
	require.NoError(t, err)

	// Intentar adquirir el mismo lock desde otra "conexión" (debería esperar o fallar)
	// En un test real, esto simularía dos instancias intentando migrar simultáneamente
	lockID := hashSchemaName("pr_lock_test")
	
	var acquired bool
	err = sqlDB.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", lockID).Scan(&acquired)
	require.NoError(t, err)
	assert.False(t, acquired, "Second lock attempt should fail (lock already held)")

	// Liberar lock
	unlock1()

	// Ahora debería poder adquirir el lock
	err = sqlDB.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", lockID).Scan(&acquired)
	require.NoError(t, err)
	assert.True(t, acquired, "Should be able to acquire lock after release")

	// Liberar
	_, _ = sqlDB.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", lockID)

	// Limpiar
	_, _ = sqlDB.ExecContext(ctx, `DROP SCHEMA IF EXISTS pr_lock_test CASCADE`)
}

// TestValidateSchemaName (unitario, no requiere DB)
func TestValidateSchemaName(t *testing.T) {
	tests := []struct {
		name    string
		schema  string
		wantErr bool
	}{
		{"ok pr_1", "pr_1", false},
		{"ok public", "public", false},
		{"ok branch_feat_abc123", "branch_feat_abc123", false},
		{"empty", "", true},
		{"reserved pg_catalog", "pg_catalog", true},
		{"invalid char", "pr-1;drop", true},
		{"starts with number", "1pr", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSchemaName(tt.schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSchemaName(%q) err=%v wantErr=%v", tt.schema, err, tt.wantErr)
			}
		})
	}
}

// TestQuoteIdentifier (unitario)
func TestQuoteIdentifier(t *testing.T) {
	assert.Equal(t, `"pr_1"`, quoteIdentifier("pr_1"))
	assert.Equal(t, `"x""y"`, quoteIdentifier(`x"y`))
}

// TestHashSchemaName (unitario)
func TestHashSchemaName(t *testing.T) {
	a := hashSchemaName("pr_1")
	b := hashSchemaName("pr_1")
	assert.Equal(t, a, b)
	assert.NotEqual(t, hashSchemaName("pr_1"), hashSchemaName("pr_2"))
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getTableCount(t *testing.T, db *sql.DB, schema string) int {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = $1 
		AND table_type = 'BASE TABLE'
		AND table_name != 'schema_migrations'
	`
	err := db.QueryRow(query, schema).Scan(&count)
	require.NoError(t, err, "Failed to get table count for schema %s", schema)
	return count
}

func getMigrationVersion(t *testing.T, db *sql.DB, schema string) int {
	var version int
	query := fmt.Sprintf(`SELECT version FROM %s.schema_migrations LIMIT 1`, quoteIdentifier(schema))
	err := db.QueryRow(query).Scan(&version)
	if err == sql.ErrNoRows {
		return 0
	}
	require.NoError(t, err, "Failed to get migration version for schema %s", schema)
	return version
}

func cleanupTestSchemas(t *testing.T, db *sql.DB, schemas []string) {
	ctx := context.Background()
	for _, schema := range schemas {
		_, err := db.ExecContext(ctx, fmt.Sprintf(`DROP SCHEMA IF EXISTS %s CASCADE`, quoteIdentifier(schema)))
		if err != nil {
			t.Logf("Warning: Failed to cleanup schema %s: %v", schema, err)
		}
	}
}
