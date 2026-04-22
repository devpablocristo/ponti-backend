package stock

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	corectx "github.com/devpablocristo/core/security/go/contextkeys"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	migpostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	config "github.com/devpablocristo/ponti-backend/cmd/config"
	projectmod "github.com/devpablocristo/ponti-backend/internal/project"
)

type stockE2EGormEngine struct {
	client *gorm.DB
}

func (e *stockE2EGormEngine) Client() *gorm.DB { return e.client }

type stockE2EGinEngine struct {
	r *gin.Engine
}

func (e *stockE2EGinEngine) GetRouter() *gin.Engine          { return e.r }
func (e *stockE2EGinEngine) RunServer(context.Context) error { return nil }

type stockE2EConfig struct{}

func (stockE2EConfig) APIVersion() string { return "v1" }
func (stockE2EConfig) APIBaseURL() string { return "/api/v1" }

type stockE2EMiddlewares struct{}

func (stockE2EMiddlewares) GetGlobal() []gin.HandlerFunc { return nil }
func (stockE2EMiddlewares) GetProtected() []gin.HandlerFunc {
	return nil
}
func (stockE2EMiddlewares) GetValidation() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), corectx.Actor, "auditor@ponti.test")
			c.Request = c.Request.WithContext(ctx)
			c.Next()
		},
	}
}

func TestStockE2E_SummaryAndCountFlow(t *testing.T) {
	db, cleanup := setupStockE2EDatabase(t)
	defer cleanup()

	seedStockE2EScenario(t, db)

	router := setupStockE2ERouter(t, db)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/1/stocks/summary?cutoff_date=2026-04-18", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var summaryResp struct {
		Items []struct {
			SupplyID          int64   `json:"supply_id"`
			EntryStock        string  `json:"entry_stock"`
			OutStock          string  `json:"out_stock"`
			Consumed          string  `json:"consumed"`
			StockUnits        string  `json:"stock_units"`
			RealStockUnits    string  `json:"real_stock_units"`
			LastCountAt       *string `json:"last_count_at"`
			HasRealStockCount bool    `json:"has_real_stock_count"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &summaryResp))
	require.Len(t, summaryResp.Items, 1)
	require.Equal(t, int64(1), summaryResp.Items[0].SupplyID)
	require.Equal(t, "103.00", summaryResp.Items[0].EntryStock)
	require.Equal(t, "10.00", summaryResp.Items[0].OutStock)
	require.Equal(t, "7.00", summaryResp.Items[0].Consumed)
	require.Equal(t, "86.00", summaryResp.Items[0].StockUnits)
	require.Equal(t, "80.00", summaryResp.Items[0].RealStockUnits)
	require.NotNil(t, summaryResp.Items[0].LastCountAt)
	assertStockE2ETimeEqual(t, *summaryResp.Items[0].LastCountAt, time.Date(2026, 4, 10, 10, 0, 0, 0, time.UTC))
	require.True(t, summaryResp.Items[0].HasRealStockCount)

	postBody := bytes.NewBufferString(`{"counted_units":"70","counted_at":"2026-04-21T12:00:00Z","note":"Conteo nuevo"}`)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/projects/1/supplies/1/stock-counts", postBody)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	var createResp struct {
		ID      int64  `json:"id"`
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &createResp))
	require.NotZero(t, createResp.ID)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/projects/1/stocks/summary", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &summaryResp))
	require.Len(t, summaryResp.Items, 1)
	require.Equal(t, "86.00", summaryResp.Items[0].StockUnits)
	require.Equal(t, "70.00", summaryResp.Items[0].RealStockUnits)
	require.NotNil(t, summaryResp.Items[0].LastCountAt)
	assertStockE2ETimeEqual(t, *summaryResp.Items[0].LastCountAt, time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC))

	var countRows int
	var createdBy string
	err := db.Raw(`
		SELECT COUNT(*)
		FROM public.supply_stock_counts
		WHERE supply_id = 1
	`).Row().Scan(&countRows)
	require.NoError(t, err)
	require.Equal(t, 2, countRows)

	err = db.Raw(`
		SELECT COALESCE(created_by, '')
		FROM public.supply_stock_counts
		WHERE supply_id = 1
		ORDER BY id DESC
		LIMIT 1
	`).Row().Scan(&createdBy)
	require.NoError(t, err)
	require.Equal(t, "auditor@ponti.test", createdBy)
}

func TestStockE2E_CreateCountRejectsNegativeUnits(t *testing.T) {
	db, cleanup := setupStockE2EDatabase(t)
	defer cleanup()

	seedStockE2EScenario(t, db)
	router := setupStockE2ERouter(t, db)

	postBody := bytes.NewBufferString(`{"counted_units":"-1","counted_at":"2026-04-21T12:00:00Z"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/1/supplies/1/stock-counts", postBody)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), "counted_units")

	var countRows int
	err := db.Raw(`SELECT COUNT(*) FROM public.supply_stock_counts WHERE supply_id = 1`).Row().Scan(&countRows)
	require.NoError(t, err)
	require.Equal(t, 1, countRows)
}

func setupStockE2ERouter(t *testing.T, db *gorm.DB) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	engine := &stockE2EGormEngine{client: db}
	projectRepo := projectmod.NewRepository(engine)
	projectUC := projectmod.NewUseCases(projectRepo, nil)
	stockRepo := NewRepository(engine)
	stockUC := NewUseCases(stockRepo, nil, projectUC)

	h := NewHandler(
		stockUC,
		&stockE2EGinEngine{r: router},
		stockE2EConfig{},
		stockE2EMiddlewares{},
	)
	h.Routes()

	return router
}

func setupStockE2EDatabase(t *testing.T) (*gorm.DB, func()) {
	t.Helper()

	adminCfg, ok := stockE2EDBConfig()
	if !ok {
		t.Skip("Skipping integration test: TEST_DB_HOST not set")
	}

	adminDB, err := sql.Open("postgres", stockE2EPostgresDSN(adminCfg))
	require.NoError(t, err)
	defer func() { _ = adminDB.Close() }()

	tempName := fmt.Sprintf("ponti_stock_e2e_%d", time.Now().UnixNano())
	_, err = adminDB.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, tempName))
	require.NoError(t, err)

	dbCfg := adminCfg
	dbCfg.Name = tempName

	migrateStockE2EDatabase(t, dbCfg, 214)

	gormDB, err := gorm.Open(gormpostgres.Open(stockE2EPostgresDSN(dbCfg)), &gorm.Config{})
	require.NoError(t, err)

	cleanup := func() {
		sqlDB, err := gormDB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}

		cleanupDB, err := sql.Open("postgres", stockE2EPostgresDSN(adminCfg))
		if err == nil {
			defer func() { _ = cleanupDB.Close() }()
			_, _ = cleanupDB.Exec(fmt.Sprintf(`
				SELECT pg_terminate_backend(pid)
				FROM pg_stat_activity
				WHERE datname = '%s' AND pid <> pg_backend_pid()
			`, tempName))
			_, _ = cleanupDB.Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS "%s"`, tempName))
		}
	}

	return gormDB, cleanup
}

func stockE2EDBConfig() (config.DB, bool) {
	if os.Getenv("TEST_DB_HOST") == "" {
		return config.DB{}, false
	}

	port := 5432
	if p := os.Getenv("TEST_DB_PORT"); p != "" {
		port, _ = strconv.Atoi(p)
	}

	return config.DB{
		Type:     "postgres",
		Host:     stockE2EEnvOrDefault("TEST_DB_HOST", "localhost"),
		User:     stockE2EEnvOrDefault("TEST_DB_USER", "postgres"),
		Password: stockE2EEnvOrDefault("TEST_DB_PASSWORD", "postgres"),
		Name:     stockE2EEnvOrDefault("TEST_DB_ADMIN_NAME", "postgres"),
		SSLMode:  stockE2EEnvOrDefault("TEST_DB_SSL_MODE", "disable"),
		Port:     port,
	}, true
}

func migrateStockE2EDatabase(t *testing.T, cfg config.DB, version uint) {
	t.Helper()

	sqlDB, err := sql.Open("postgres", stockE2EPostgresDSN(cfg))
	require.NoError(t, err)
	defer func() { _ = sqlDB.Close() }()

	driver, err := migpostgres.WithInstance(sqlDB, &migpostgres.Config{DatabaseName: cfg.Name})
	require.NoError(t, err)

	sourceURL := "file://" + stockE2EMigrationsDir(t)
	m, err := migrate.NewWithDatabaseInstance(sourceURL, cfg.Name, driver)
	require.NoError(t, err)
	defer func() {
		sourceErr, dbErr := m.Close()
		require.NoError(t, sourceErr)
		require.NoError(t, dbErr)
	}()

	err = m.Migrate(version)
	if err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err)
	}
}

func stockE2EMigrationsDir(t *testing.T) string {
	t.Helper()

	dir, err := filepath.Abs("../../migrations_v4")
	require.NoError(t, err)
	return dir
}

func stockE2EPostgresDSN(cfg config.DB) string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.SSLMode,
	)
}

func stockE2EEnvOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func seedStockE2EScenario(t *testing.T, db *gorm.DB) {
	t.Helper()

	mustExecStockE2E(t, db, `INSERT INTO public.customers (id, name) VALUES (1, 'Customer test')`)
	mustExecStockE2E(t, db, `INSERT INTO public.campaigns (id, name) VALUES (1, 'Campaign test')`)
	mustExecStockE2E(t, db, `INSERT INTO public.projects (id, name, customer_id, campaign_id, admin_cost) VALUES (1, 'Project stock', 1, 1, 0)`)
	mustExecStockE2E(t, db, `INSERT INTO public.investors (id, name) VALUES (1, 'Investor principal')`)
	mustExecStockE2E(t, db, `INSERT INTO public.types (id, name) VALUES (1, 'Supply type')`)
	mustExecStockE2E(t, db, `INSERT INTO public.categories (id, name, type_id) VALUES (1, 'Fertilizantes', 1)`)
	mustExecStockE2E(t, db, `INSERT INTO public.providers (id, name) VALUES (1, 'Proveedor uno')`)
	mustExecStockE2E(t, db, `
		INSERT INTO public.supplies (id, project_id, name, price, unit_id, category_id, type_id)
		VALUES (1, 1, 'Urea', 4, 1, 1, 1)
	`)
	mustExecStockE2E(t, db, `
		INSERT INTO public.supply_stock_counts (
			supply_id,
			counted_units,
			counted_at,
			note,
			created_by,
			updated_by
		) VALUES (
			1,
			80,
			TIMESTAMP WITH TIME ZONE '2026-04-10 10:00:00+00',
			'Backfill inicial',
			'seed@test',
			'seed@test'
		)
	`)
	mustExecStockE2E(t, db, `
		INSERT INTO public.supply_movements (
			id,
			stock_id,
			quantity,
			movement_type,
			movement_date,
			reference_number,
			is_entry,
			project_id,
			project_destination_id,
			supply_id,
			investor_id,
			provider_id
		) VALUES
		(1, NULL, 100, 'Remito oficial', TIMESTAMP '2026-04-15 09:00:00', 'REM-1', TRUE, 1, 1, 1, 1, 1),
		(2, NULL, -10, 'Movimiento interno', TIMESTAMP '2026-04-17 08:00:00', 'INT-OUT-1', FALSE, 1, 1, 1, 1, 1),
		(3, NULL, 3, 'Movimiento interno entrada', TIMESTAMP '2026-04-18 07:30:00', 'INT-IN-1', TRUE, 1, 1, 1, 1, 1)
	`)
	mustExecStockE2E(t, db, `INSERT INTO public.lease_types (id, name) VALUES (1, 'Propio')`)
	mustExecStockE2E(t, db, `INSERT INTO public.crops (id, name) VALUES (1, 'Maíz'), (2, 'Soja')`)
	mustExecStockE2E(t, db, `INSERT INTO public.labor_types (id, name) VALUES (1, 'Aplicación')`)
	mustExecStockE2E(t, db, `INSERT INTO public.labor_categories (id, name, type_id) VALUES (1, 'Pulverización', 1)`)
	mustExecStockE2E(t, db, `
		INSERT INTO public.labors (id, project_id, name, category_id, price, contractor_name)
		VALUES (1, 1, 'Labor test', 1, 1, 'Contratista')
	`)
	mustExecStockE2E(t, db, `
		INSERT INTO public.fields (id, name, project_id, lease_type_id)
		VALUES (1, 'Campo 1', 1, 1)
	`)
	mustExecStockE2E(t, db, `
		INSERT INTO public.lots (
			id,
			name,
			field_id,
			hectares,
			previous_crop_id,
			current_crop_id,
			season
		) VALUES (
			1,
			'Lote 1',
			1,
			10,
			1,
			2,
			'2025/2026'
		)
	`)
	mustExecStockE2E(t, db, `
		INSERT INTO public.workorders (
			id,
			number,
			project_id,
			field_id,
			lot_id,
			crop_id,
			labor_id,
			date,
			investor_id,
			effective_area
		) VALUES (
			1,
			'WO-1',
			1,
			1,
			1,
			2,
			1,
			DATE '2026-04-18',
			1,
			10
		)
	`)
	mustExecStockE2E(t, db, `
		INSERT INTO public.workorder_items (
			id,
			workorder_id,
			supply_id,
			total_used,
			final_dose,
			supply_name
		) VALUES (
			1,
			1,
			1,
			7,
			0.7,
			'Urea'
		)
	`)
}

func mustExecStockE2E(t *testing.T, db *gorm.DB, query string, args ...any) {
	t.Helper()
	require.NoError(t, db.Exec(query, args...).Error)
}

func assertStockE2ETimeEqual(t *testing.T, actual string, expected time.Time) {
	t.Helper()

	parsed, err := time.Parse(time.RFC3339, actual)
	require.NoError(t, err)
	require.True(t, parsed.UTC().Equal(expected.UTC()), "expected %s, got %s", expected.UTC(), parsed.UTC())
}
