package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	config "github.com/devpablocristo/ponti-backend/cmd/config"
)

func TestContinuousStockMigration_BackfillsSingleLegacyRowAndBuildsSummary(t *testing.T) {
	adminCfg, ok := integrationTestDBConfig()
	if !ok {
		t.Skip("Skipping integration test: TEST_DB_HOST not set")
	}

	dbCfg, cleanup := createTemporaryDatabase(t, adminCfg)
	defer cleanup()

	db := openDBForTest(t, dbCfg)
	defer func() { _ = db.Close() }()

	migrateDatabaseToVersion(t, dbCfg, 213)
	seedSingleLegacyStockScenario(t, db)

	migrateDatabaseToVersion(t, dbCfg, 214)

	var backfillCount int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM public.supply_stock_counts
		WHERE supply_id = 1
		  AND note = 'Backfill desde legacy stocks row 1'
	`).Scan(&backfillCount)
	require.NoError(t, err)
	require.Equal(t, 1, backfillCount)

	var backfilledUnits string
	var backfilledAt time.Time
	err = db.QueryRow(`
		SELECT counted_units::text, counted_at
		FROM public.supply_stock_counts
		WHERE supply_id = 1
		ORDER BY counted_at ASC, id ASC
		LIMIT 1
	`).Scan(&backfilledUnits, &backfilledAt)
	require.NoError(t, err)
	require.Equal(t, "80.000", backfilledUnits)
	require.Equal(t, "2026-04-10", backfilledAt.UTC().Format("2006-01-02"))

	seedContinuousStockPost214Data(t, db)

	var movementIn, movementOut, consumed, systemStock string
	err = db.QueryRow(`
		SELECT
			v4_ssot.movement_in_units_for_supply_as_of(1, 1, NULL)::text,
			v4_ssot.movement_out_units_for_supply_as_of(1, 1, NULL)::text,
			v4_ssot.consumed_units_for_supply_as_of(1, 1, NULL)::text,
			v4_ssot.system_stock_units_for_supply_as_of(1, 1, NULL)::text
	`).Scan(&movementIn, &movementOut, &consumed, &systemStock)
	require.NoError(t, err)
	require.Equal(t, "103.000", movementIn)
	require.Equal(t, "30.000", movementOut)
	require.Equal(t, "7.000000", consumed)
	require.Equal(t, "66.000000", systemStock)

	var latestUnits string
	var latestAt time.Time
	err = db.QueryRow(`
		SELECT counted_units::text, counted_at
		FROM v4_ssot.latest_stock_count_for_supply_as_of(1, NULL)
	`).Scan(&latestUnits, &latestAt)
	require.NoError(t, err)
	require.Equal(t, "70.000", latestUnits)
	require.Equal(t, "2026-04-21", latestAt.UTC().Format("2006-01-02"))

	var summary struct {
		SupplyID          int64
		EntryStock        string
		OutStock          string
		Consumed          string
		StockUnits        string
		RealStockUnits    string
		HasRealStockCount bool
		LastCountAt       time.Time
	}
	err = db.QueryRow(`
		SELECT supply_id, entry_stock::text, out_stock::text, consumed::text,
		       stock_units::text, real_stock_units::text, has_real_stock_count, last_count_at
		FROM v4_report.stock_summary_for_project_as_of(1, NULL)
		WHERE supply_id = 1
	`).Scan(
		&summary.SupplyID,
		&summary.EntryStock,
		&summary.OutStock,
		&summary.Consumed,
		&summary.StockUnits,
		&summary.RealStockUnits,
		&summary.HasRealStockCount,
		&summary.LastCountAt,
	)
	require.NoError(t, err)
	require.Equal(t, int64(1), summary.SupplyID)
	require.Equal(t, "103.000", summary.EntryStock)
	require.Equal(t, "30.000", summary.OutStock)
	require.Equal(t, "7.000000", summary.Consumed)
	require.Equal(t, "66.000000", summary.StockUnits)
	require.Equal(t, "70.000", summary.RealStockUnits)
	require.True(t, summary.HasRealStockCount)
	require.Equal(t, "2026-04-21", summary.LastCountAt.UTC().Format("2006-01-02"))

	var cutoffSystem string
	var cutoffReal string
	var cutoffLast time.Time
	err = db.QueryRow(`
		SELECT stock_units::text, real_stock_units::text, last_count_at
		FROM v4_report.stock_summary_for_project_as_of(1, DATE '2026-04-18')
		WHERE supply_id = 1
	`).Scan(&cutoffSystem, &cutoffReal, &cutoffLast)
	require.NoError(t, err)
	require.Equal(t, "86.000000", cutoffSystem)
	require.Equal(t, "80.000", cutoffReal)
	require.Equal(t, "2026-04-10", cutoffLast.UTC().Format("2006-01-02"))

	var lastCountDate string
	var stockValue string
	err = db.QueryRow(`
		SELECT
			v4_ssot.last_stock_count_date_for_project(1)::text,
			v4_ssot.stock_value_for_project(1)::text
	`).Scan(&lastCountDate, &stockValue)
	require.NoError(t, err)
	require.Equal(t, "2026-04-21", lastCountDate)
	require.Equal(t, "264.000000000000", stockValue)

	_, err = db.Exec(`
		INSERT INTO public.supply_movements (
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
		) VALUES (
			1,
			'Stock',
			TIMESTAMP '2026-04-22 09:00:00',
			'FORBIDDEN',
			TRUE,
			1,
			1,
			1,
			1,
			1
		)
	`)
	require.Error(t, err)
	require.Contains(t, err.Error(), "chk_supply_movements_movement_type")
}

func TestContinuousStockMigration_RecordsBackfillConflictForMultipleLegacyRows(t *testing.T) {
	adminCfg, ok := integrationTestDBConfig()
	if !ok {
		t.Skip("Skipping integration test: TEST_DB_HOST not set")
	}

	dbCfg, cleanup := createTemporaryDatabase(t, adminCfg)
	defer cleanup()

	db := openDBForTest(t, dbCfg)
	defer func() { _ = db.Close() }()

	migrateDatabaseToVersion(t, dbCfg, 213)
	seedConflictingLegacyStockScenario(t, db)

	migrateDatabaseToVersion(t, dbCfg, 214)

	var countRows int
	err := db.QueryRow(`SELECT COUNT(*) FROM public.supply_stock_counts WHERE supply_id = 1`).Scan(&countRows)
	require.NoError(t, err)
	require.Equal(t, 0, countRows)

	var conflictProjectID int64
	var conflictSupplyID int64
	var conflictingIDs string
	err = db.QueryRow(`
		SELECT project_id, supply_id, conflicting_stock_ids::text
		FROM public.supply_stock_count_backfill_conflicts
		WHERE project_id = 1 AND supply_id = 1
	`).Scan(&conflictProjectID, &conflictSupplyID, &conflictingIDs)
	require.NoError(t, err)
	require.Equal(t, int64(1), conflictProjectID)
	require.Equal(t, int64(1), conflictSupplyID)
	require.Equal(t, "{1,2}", conflictingIDs)
}

func integrationTestDBConfig() (config.DB, bool) {
	if os.Getenv("TEST_DB_HOST") == "" {
		return config.DB{}, false
	}

	port := 5432
	if p := os.Getenv("TEST_DB_PORT"); p != "" {
		port, _ = strconv.Atoi(p)
	}

	return config.DB{
		Type:     "postgres",
		Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		User:     getEnvOrDefault("TEST_DB_USER", "postgres"),
		Password: getEnvOrDefault("TEST_DB_PASSWORD", "postgres"),
		Name:     getEnvOrDefault("TEST_DB_ADMIN_NAME", "postgres"),
		SSLMode:  getEnvOrDefault("TEST_DB_SSL_MODE", "disable"),
		Port:     port,
	}, true
}

func createTemporaryDatabase(t *testing.T, adminCfg config.DB) (config.DB, func()) {
	t.Helper()

	adminDB := openDBForTest(t, adminCfg)
	defer func() { _ = adminDB.Close() }()

	tempName := fmt.Sprintf("ponti_stock_test_%d", time.Now().UnixNano())
	mustExec(t, adminDB, fmt.Sprintf(`CREATE DATABASE "%s"`, tempName))

	dbCfg := adminCfg
	dbCfg.Name = tempName

	cleanup := func() {
		adminCleanupDB := openDBForTest(t, adminCfg)
		defer func() { _ = adminCleanupDB.Close() }()

		_, _ = adminCleanupDB.Exec(fmt.Sprintf(`
			SELECT pg_terminate_backend(pid)
			FROM pg_stat_activity
			WHERE datname = '%s' AND pid <> pg_backend_pid()
		`, tempName))
		_, _ = adminCleanupDB.Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS "%s"`, tempName))
	}

	return dbCfg, cleanup
}

func openDBForTest(t *testing.T, cfg config.DB) *sql.DB {
	t.Helper()

	db, err := sql.Open("postgres", buildPostgresDSN(cfg))
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	return db
}

func migrateDatabaseToVersion(t *testing.T, cfg config.DB, version uint) {
	t.Helper()

	db := openDBForTest(t, cfg)
	defer func() { _ = db.Close() }()

	driver, err := postgres.WithInstance(db, &postgres.Config{DatabaseName: cfg.Name})
	require.NoError(t, err)

	sourceURL := "file://" + migrationsDir(t)
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

func migrationsDir(t *testing.T) string {
	t.Helper()

	dir, err := filepath.Abs("../../migrations_v4")
	require.NoError(t, err)
	return dir
}

func mustExec(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	_, err := db.Exec(query, args...)
	require.NoError(t, err)
}

func seedSingleLegacyStockScenario(t *testing.T, db *sql.DB) {
	t.Helper()

	seedCommonLegacyStockData(t, db)
	mustExec(t, db, `
		INSERT INTO public.stocks (
			id,
			project_id,
			supply_id,
			investor_id,
			close_date,
			real_stock_units,
			initial_units,
			year_period,
			month_period,
			created_at,
			updated_at,
			units_entered,
			units_consumed,
			has_real_stock_count
		) VALUES (
			1,
			1,
			1,
			1,
			NULL,
			80,
			0,
			2026,
			4,
			TIMESTAMP WITH TIME ZONE '2026-04-10 10:00:00+00',
			TIMESTAMP WITH TIME ZONE '2026-04-10 10:00:00+00',
			0,
			0,
			TRUE
		)
	`)
}

func seedConflictingLegacyStockScenario(t *testing.T, db *sql.DB) {
	t.Helper()

	seedCommonLegacyStockData(t, db)
	mustExec(t, db, `INSERT INTO public.investors (id, name) VALUES (2, 'Investor conflict')`)
	mustExec(t, db, `
		INSERT INTO public.stocks (
			id,
			project_id,
			supply_id,
			investor_id,
			close_date,
			real_stock_units,
			initial_units,
			year_period,
			month_period,
			created_at,
			updated_at,
			units_entered,
			units_consumed,
			has_real_stock_count
		) VALUES
		(
			1,
			1,
			1,
			1,
			NULL,
			80,
			0,
			2026,
			4,
			TIMESTAMP WITH TIME ZONE '2026-04-10 10:00:00+00',
			TIMESTAMP WITH TIME ZONE '2026-04-10 10:00:00+00',
			0,
			0,
			TRUE
		),
		(
			2,
			1,
			1,
			2,
			NULL,
			60,
			0,
			2026,
			4,
			TIMESTAMP WITH TIME ZONE '2026-04-11 10:00:00+00',
			TIMESTAMP WITH TIME ZONE '2026-04-11 10:00:00+00',
			0,
			0,
			TRUE
		)
	`)
}

func seedCommonLegacyStockData(t *testing.T, db *sql.DB) {
	t.Helper()

	mustExec(t, db, `INSERT INTO public.customers (id, name) VALUES (1, 'Customer test')`)
	mustExec(t, db, `INSERT INTO public.campaigns (id, name) VALUES (1, 'Campaign test')`)
	mustExec(t, db, `INSERT INTO public.projects (id, name, customer_id, campaign_id, admin_cost) VALUES (1, 'Project stock', 1, 1, 0)`)
	mustExec(t, db, `INSERT INTO public.investors (id, name) VALUES (1, 'Investor principal')`)
	mustExec(t, db, `INSERT INTO public.types (id, name) VALUES (1, 'Supply type')`)
	mustExec(t, db, `INSERT INTO public.categories (id, name, type_id) VALUES (1, 'Fertilizantes', 1)`)
	mustExec(t, db, `INSERT INTO public.providers (id, name) VALUES (1, 'Proveedor uno')`)
	mustExec(t, db, `
		INSERT INTO public.supplies (
			id,
			project_id,
			name,
			price,
			unit_id,
			category_id,
			type_id
		) VALUES (
			1,
			1,
			'Urea',
			4,
			1,
			1,
			1
		)
	`)
}

func seedContinuousStockPost214Data(t *testing.T, db *sql.DB) {
	t.Helper()

	mustExec(t, db, `
		INSERT INTO public.supply_stock_counts (
			supply_id,
			counted_units,
			counted_at,
			note,
			created_by,
			updated_by
		) VALUES (
			1,
			70,
			TIMESTAMP WITH TIME ZONE '2026-04-21 12:00:00+00',
			'Conteo reciente',
			'auditor@ponti.test',
			'auditor@ponti.test'
		)
	`)

	mustExec(t, db, `
		INSERT INTO public.supply_movements (
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
		(100, 'Remito oficial', TIMESTAMP '2026-04-15 09:00:00', 'REM-1', TRUE, 1, 1, 1, 1, 1),
		(-10, 'Movimiento interno', TIMESTAMP '2026-04-17 08:00:00', 'INT-OUT-1', FALSE, 1, 1, 1, 1, 1),
		(3, 'Movimiento interno entrada', TIMESTAMP '2026-04-18 07:30:00', 'INT-IN-1', TRUE, 1, 1, 1, 1, 1),
		(-20, 'Devolución', TIMESTAMP '2026-04-19 11:00:00', 'DEV-1', FALSE, 1, 1, 1, 1, 1)
	`)

	mustExec(t, db, `INSERT INTO public.lease_types (id, name) VALUES (1, 'Propio')`)
	mustExec(t, db, `INSERT INTO public.crops (id, name) VALUES (1, 'Maíz'), (2, 'Soja')`)
	mustExec(t, db, `INSERT INTO public.labor_types (id, name) VALUES (1, 'Aplicación')`)
	mustExec(t, db, `INSERT INTO public.labor_categories (id, name, type_id) VALUES (1, 'Pulverización', 1)`)
	mustExec(t, db, `
		INSERT INTO public.labors (
			id,
			project_id,
			name,
			category_id,
			price,
			contractor_name
		) VALUES (
			1,
			1,
			'Labor test',
			1,
			1,
			'Contratista'
		)
	`)
	mustExec(t, db, `
		INSERT INTO public.fields (
			id,
			name,
			project_id,
			lease_type_id
		) VALUES (
			1,
			'Campo 1',
			1,
			1
		)
	`)
	mustExec(t, db, `
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
	mustExec(t, db, `
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
	mustExec(t, db, `
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
