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

	config "github.com/alphacodinggroup/ponti-backend/cmd/config"
)

// TestMigrationLock verifica que el lock de migraciones funciona correctamente.
// Nota: ya no se testea "schema por rama" porque la estrategia de deploy manual
// pasó a ser "DB por rama" y el backend no implementa lógica de despliegue.
func TestMigrationLock(t *testing.T) {
	if os.Getenv("TEST_DB_HOST") == "" {
		t.Skip("Skipping integration test: TEST_DB_HOST not set")
	}

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

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		dbConfig.Host, dbConfig.User, dbConfig.Password, dbConfig.Name, dbConfig.Port, dbConfig.SSLMode)

	sqlDB, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer sqlDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Adquirir lock
	unlock1, err := acquireMigrationLock(ctx, sqlDB, dbConfig.Name)
	require.NoError(t, err)

	// Intentar adquirir el mismo lock desde otra "conexión" (debería fallar)
	lockID := hashDatabaseName(dbConfig.Name)
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
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
