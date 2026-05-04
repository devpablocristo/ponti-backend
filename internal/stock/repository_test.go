package stock

import (
	"context"
	"testing"
	"time"

	projectdomain "github.com/devpablocristo/ponti-backend/internal/project/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	"github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type stockTestGormEngine struct {
	client *gorm.DB
}

func (e *stockTestGormEngine) Client() *gorm.DB { return e.client }

func newStockTestRepo(t *testing.T) (*Repository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	stmt := `CREATE TABLE stocks (
		id INTEGER PRIMARY KEY,
		project_id INTEGER NOT NULL,
		supply_id INTEGER NULL,
		investor_id INTEGER NULL,
		close_date DATETIME NULL,
		real_stock_units TEXT NOT NULL,
		initial_units TEXT NULL,
		year_period INTEGER NULL,
		month_period INTEGER NULL,
		units_entered TEXT NULL,
		units_consumed TEXT NULL,
		has_real_stock_count BOOLEAN NOT NULL,
		created_at DATETIME NULL,
		updated_at DATETIME NULL,
		created_by TEXT NULL,
		updated_by TEXT NULL,
		deleted_at DATETIME NULL,
		deleted_by TEXT NULL
	);`
	if err := db.Exec(stmt).Error; err != nil {
		t.Fatalf("exec schema: %v", err)
	}

	stmts := []string{
		`ATTACH DATABASE ':memory:' AS v4_report;`,
		`CREATE TABLE v4_report.stock_consumed_by_supply (
			project_id INTEGER NOT NULL,
			supply_id INTEGER NOT NULL,
			consumed TEXT NOT NULL
		);`,
		`CREATE TABLE projects (
			id INTEGER PRIMARY KEY,
			name TEXT,
			customer_id INTEGER,
			campaign_id INTEGER,
			admin_cost TEXT,
			planned_cost TEXT,
			created_at DATETIME NULL,
			updated_at DATETIME NULL,
			deleted_at DATETIME NULL,
			created_by TEXT NULL,
			updated_by TEXT NULL,
			deleted_by TEXT NULL
		);`,
		`CREATE TABLE types (
			id INTEGER PRIMARY KEY,
			name TEXT,
			created_at DATETIME NULL,
			updated_at DATETIME NULL,
			deleted_at DATETIME NULL,
			created_by TEXT NULL,
			updated_by TEXT NULL,
			deleted_by TEXT NULL
		);`,
		`CREATE TABLE categories (
			id INTEGER PRIMARY KEY,
			name TEXT,
			type_id INTEGER,
			created_at DATETIME NULL,
			updated_at DATETIME NULL,
			deleted_at DATETIME NULL,
			created_by TEXT NULL,
			updated_by TEXT NULL,
			deleted_by TEXT NULL
		);`,
		`CREATE TABLE supplies (
			id INTEGER PRIMARY KEY,
			project_id INTEGER,
			name TEXT,
			price TEXT,
			is_partial_price BOOLEAN,
			is_pending BOOLEAN,
			unit_id INTEGER,
			category_id INTEGER,
			type_id INTEGER,
			created_at DATETIME NULL,
			updated_at DATETIME NULL,
			deleted_at DATETIME NULL,
			created_by TEXT NULL,
			updated_by TEXT NULL,
			deleted_by TEXT NULL
		);`,
		`CREATE TABLE investors (
			id INTEGER PRIMARY KEY,
			name TEXT,
			created_at DATETIME NULL,
			updated_at DATETIME NULL,
			deleted_at DATETIME NULL,
			created_by TEXT NULL,
			updated_by TEXT NULL,
			deleted_by TEXT NULL
		);`,
		`CREATE TABLE supply_movements (
				id INTEGER PRIMARY KEY,
				stock_id INTEGER,
				quantity TEXT,
			movement_type TEXT,
			movement_date DATETIME NULL,
			reference_number TEXT,
			project_id INTEGER,
			project_destination_id INTEGER,
			supply_id INTEGER,
			investor_id INTEGER,
			provider_id INTEGER,
			is_entry BOOLEAN,
			created_at DATETIME NULL,
			updated_at DATETIME NULL,
			deleted_at DATETIME NULL,
			created_by TEXT NULL,
				updated_by TEXT NULL,
				deleted_by TEXT NULL
			);`,
		`CREATE TABLE workorders (
				id INTEGER PRIMARY KEY,
				project_id INTEGER,
				investor_id INTEGER,
				date DATETIME NULL,
				deleted_at DATETIME NULL
			);`,
		`CREATE TABLE workorder_items (
				id INTEGER PRIMARY KEY,
				workorder_id INTEGER,
				supply_id INTEGER,
				total_used TEXT,
				deleted_at DATETIME NULL
			);`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("exec schema: %v", err)
		}
	}

	return NewRepository(&stockTestGormEngine{client: db}), db
}

func TestRepository_GetStocks_UsesCurrentSupplyInvestorRollup(t *testing.T) {
	repo, db := newStockTestRepo(t)
	ctx := context.Background()

	seedStockLookupReferences(t, db)
	seedClosedAndActiveStockRollup(t, db)

	stocks, err := repo.GetStocks(ctx, 1, time.Time{})

	assert.NoError(t, err)
	if assert.Len(t, stocks, 1) {
		stock := stocks[0]
		assert.Equal(t, int64(102), stock.ID)
		assert.True(t, stock.GetEntryStock().Equal(decimal.NewFromInt(220)))
		assert.True(t, stock.Consumed.Equal(decimal.NewFromInt(220)))
		assert.True(t, stock.GetStockUnits().Equal(decimal.Zero))
	}
}

func TestRepository_GetStocks_ActiveUsesLatestStockAsReference(t *testing.T) {
	repo, db := newStockTestRepo(t)
	ctx := context.Background()

	seedStockLookupReferences(t, db)
	err := db.Exec(
		`INSERT INTO stocks (
			id, project_id, supply_id, investor_id, close_date, real_stock_units,
			initial_units, year_period, month_period, units_entered, units_consumed,
			has_real_stock_count
		) VALUES
			(101, 1, 7, 11, '2026-04-15', '8', '0', 2026, 3, '0', '0', true)`,
	).Error
	assert.NoError(t, err)

	stocks, err := repo.GetStocks(ctx, 1, time.Time{})

	assert.NoError(t, err)
	if assert.Len(t, stocks, 1) {
		assert.Equal(t, int64(101), stocks[0].ID)
		assert.Nil(t, stocks[0].CloseDate)
		assert.True(t, stocks[0].RealStockUnits.Equal(decimal.NewFromInt(8)))
	}
}

func TestRepository_GetStocks_ActiveBuildsInvestorRowFromMovementsAndConsumption(t *testing.T) {
	repo, db := newStockTestRepo(t)
	ctx := context.Background()

	seedStockLookupReferences(t, db)
	err := db.Exec(
		`INSERT INTO stocks (
			id, project_id, supply_id, investor_id, close_date, real_stock_units,
			initial_units, year_period, month_period, units_entered, units_consumed,
			has_real_stock_count
		) VALUES
			(101, 1, 7, 11, '2026-04-15', '0', '0', 2026, 4, '0', '0', false)`,
	).Error
	assert.NoError(t, err)
	err = db.Exec(
		`INSERT INTO supply_movements (
			id, stock_id, quantity, movement_type, movement_date, reference_number,
			project_id, project_destination_id, supply_id, investor_id, provider_id, is_entry
		) VALUES
			(501, 101, '500', 'Remito oficial', '2026-04-01', 'ING-1', 1, 0, 7, 11, 1, true)`,
	).Error
	assert.NoError(t, err)
	err = db.Exec(
		`INSERT INTO workorders (id, project_id, investor_id, date) VALUES
			(601, 1, 11, '2026-04-02')`,
	).Error
	assert.NoError(t, err)
	err = db.Exec(
		`INSERT INTO workorder_items (id, workorder_id, supply_id, total_used) VALUES
			(701, 601, 7, '455')`,
	).Error
	assert.NoError(t, err)

	stocks, err := repo.GetStocks(ctx, 1, time.Time{})

	assert.NoError(t, err)
	if assert.Len(t, stocks, 1) {
		stock := stocks[0]
		assert.Equal(t, int64(101), stock.ID)
		assert.Equal(t, "Investor B", stock.Investor.Name)
		assert.True(t, stock.GetEntryStock().Equal(decimal.NewFromInt(500)), "entry stock")
		assert.True(t, stock.Consumed.Equal(decimal.NewFromInt(455)), "consumed")
		assert.True(t, stock.GetStockUnits().Equal(decimal.NewFromInt(45)), "system stock")
	}
}

func TestRepository_GetStocks_ActiveCompensatesSameProjectSupplyAcrossInvestors(t *testing.T) {
	repo, db := newStockTestRepo(t)
	ctx := context.Background()

	seedStockLookupReferences(t, db)
	err := db.Exec(
		`INSERT INTO stocks (
			id, project_id, supply_id, investor_id, close_date, real_stock_units,
			initial_units, year_period, month_period, units_entered, units_consumed,
			has_real_stock_count
		) VALUES
			(101, 1, 7, 11, '2026-04-15', '0', '0', 2026, 4, '0', '0', false)`,
	).Error
	assert.NoError(t, err)
	err = db.Exec(
		`INSERT INTO supply_movements (
			id, stock_id, quantity, movement_type, movement_date, reference_number,
			project_id, project_destination_id, supply_id, investor_id, provider_id, is_entry
		) VALUES
			(501, 101, '500', 'Remito oficial', '2026-04-01', 'ING-1', 1, 0, 7, 11, 1, true)`,
	).Error
	assert.NoError(t, err)
	err = db.Exec(
		`INSERT INTO workorders (id, project_id, investor_id, date) VALUES
			(601, 1, 10, '2026-04-02')`,
	).Error
	assert.NoError(t, err)
	err = db.Exec(
		`INSERT INTO workorder_items (id, workorder_id, supply_id, total_used) VALUES
			(701, 601, 7, '455')`,
	).Error
	assert.NoError(t, err)

	stocks, err := repo.GetStocks(ctx, 1, time.Time{})

	assert.NoError(t, err)
	if assert.Len(t, stocks, 1) {
		stock := stocks[0]
		assert.Equal(t, "Investor B", stock.Investor.Name)
		assert.True(t, stock.GetEntryStock().Equal(decimal.NewFromInt(500)))
		assert.True(t, stock.Consumed.Equal(decimal.NewFromInt(455)))
		assert.True(t, stock.GetStockUnits().Equal(decimal.NewFromInt(45)))
	}
}

func TestRepository_GetStocks_CutoffDateTimeline(t *testing.T) {
	repo, db := newStockTestRepo(t)
	ctx := context.Background()
	date := func(month time.Month, day int) time.Time {
		return time.Date(2026, month, day, 0, 0, 0, 0, time.UTC)
	}

	seedStockLookupReferences(t, db)
	err := db.Exec(
		`INSERT INTO stocks (
			id, project_id, supply_id, investor_id, close_date, real_stock_units,
			initial_units, year_period, month_period, units_entered, units_consumed,
			has_real_stock_count
		) VALUES
			(201, 1, 7, 11, ?, '0', '0', 2026, 6, '0', '0', false),
			(202, 1, 7, 11, ?, '0', '0', 2026, 6, '0', '0', false),
			(203, 1, 7, 11, ?, '0', '0', 2026, 6, '0', '0', false),
			(204, 1, 7, 11, ?, '0', '0', 2026, 6, '0', '0', false)`,
		date(time.June, 2),
		date(time.June, 3),
		date(time.June, 4),
		date(time.June, 5),
	).Error
	assert.NoError(t, err)
	err = db.Exec(
		`INSERT INTO supply_movements (
			id, stock_id, quantity, movement_type, movement_date, reference_number,
			project_id, project_destination_id, supply_id, investor_id, provider_id, is_entry
		) VALUES
			(501, 201, '200', 'Remito oficial', ?, 'ING-1', 1, 0, 7, 11, 1, true)`,
		date(time.June, 1),
	).Error
	assert.NoError(t, err)
	err = db.Exec(
		`INSERT INTO workorders (id, project_id, investor_id, date) VALUES
			(601, 1, 11, ?),
			(602, 1, 11, ?),
			(603, 1, 11, ?)`,
		date(time.June, 3),
		date(time.June, 4),
		date(time.June, 5),
	).Error
	assert.NoError(t, err)
	err = db.Exec(
		`INSERT INTO workorder_items (id, workorder_id, supply_id, total_used) VALUES
			(701, 601, 7, '180'),
			(702, 602, 7, '10'),
			(703, 603, 7, '10')`,
	).Error
	assert.NoError(t, err)

	cases := []struct {
		name     string
		cutoff   time.Time
		entry    decimal.Decimal
		consumed decimal.Decimal
		system   decimal.Decimal
	}{
		{name: "25 May", cutoff: date(time.May, 25), entry: decimal.Zero, consumed: decimal.Zero, system: decimal.Zero},
		{name: "2 June", cutoff: date(time.June, 2), entry: decimal.NewFromInt(200), consumed: decimal.Zero, system: decimal.NewFromInt(200)},
		{name: "3 June", cutoff: date(time.June, 3), entry: decimal.NewFromInt(200), consumed: decimal.NewFromInt(180), system: decimal.NewFromInt(20)},
		{name: "4 June", cutoff: date(time.June, 4), entry: decimal.NewFromInt(200), consumed: decimal.NewFromInt(190), system: decimal.NewFromInt(10)},
		{name: "5 June", cutoff: date(time.June, 5), entry: decimal.NewFromInt(200), consumed: decimal.NewFromInt(200), system: decimal.Zero},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stocks, err := repo.GetStocks(ctx, 1, tc.cutoff)

			assert.NoError(t, err)
			if assert.Len(t, stocks, 1) {
				assert.True(t, stocks[0].GetEntryStock().Equal(tc.entry), "entry stock")
				assert.True(t, stocks[0].Consumed.Equal(tc.consumed), "consumed")
				assert.True(t, stocks[0].GetStockUnits().Equal(tc.system), "system stock")
			}
		})
	}
}

func TestRepository_GetStocks_ClosedStockMatchesPontiDocumentExample(t *testing.T) {
	repo, db := newStockTestRepo(t)
	ctx := context.Background()
	date := func(month time.Month, day int) time.Time {
		return time.Date(2026, month, day, 0, 0, 0, 0, time.UTC)
	}
	dateTime := func(month time.Month, day int, hour int) time.Time {
		return time.Date(2026, month, day, hour, 30, 0, 0, time.UTC)
	}
	closeDate := date(time.April, 15)

	seedStockLookupReferences(t, db)
	err := db.Exec(
		`INSERT INTO stocks (
			id, project_id, supply_id, investor_id, close_date, real_stock_units,
			initial_units, year_period, month_period, units_entered, units_consumed,
			has_real_stock_count
		) VALUES
			(301, 1, 7, 11, ?, '40', '0', 2026, 4, '0', '0', true)`,
		closeDate,
	).Error
	assert.NoError(t, err)
	err = db.Exec(
		`INSERT INTO supply_movements (
			id, stock_id, quantity, movement_type, movement_date, reference_number,
			project_id, project_destination_id, supply_id, investor_id, provider_id, is_entry
		) VALUES
			(801, 301, '20', 'Remito oficial', ?, '001-7398', 1, 0, 7, 11, 1, true),
			(802, 301, '200', 'Remito oficial', ?, 'REM MOCH', 1, 0, 7, 11, 1, true),
			(803, 301, '500', 'Remito oficial', ?, 'AFTER-CLOSE', 1, 0, 7, 11, 1, true)`,
		date(time.March, 16),
		date(time.April, 13),
		date(time.April, 16),
	).Error
	assert.NoError(t, err)
	err = db.Exec(
		`INSERT INTO workorders (id, project_id, investor_id, date) VALUES
			(901, 1, 11, ?),
			(902, 1, 11, ?),
			(903, 1, 11, ?),
			(904, 1, 11, ?)`,
		date(time.April, 13),
		date(time.April, 14),
		dateTime(time.April, 15, 18),
		date(time.April, 16),
	).Error
	assert.NoError(t, err)
	err = db.Exec(
		`INSERT INTO workorder_items (id, workorder_id, supply_id, total_used) VALUES
			(1001, 901, 7, '20'),
			(1002, 902, 7, '34'),
			(1003, 903, 7, '46'),
			(1004, 904, 7, '900')`,
	).Error
	assert.NoError(t, err)

	stocks, err := repo.GetStocks(ctx, 1, closeDate)

	assert.NoError(t, err)
	if assert.Len(t, stocks, 1) {
		stock := stocks[0]
		assert.True(t, stock.GetEntryStock().Equal(decimal.NewFromInt(220)), "entry stock")
		assert.True(t, stock.Consumed.Equal(decimal.NewFromInt(100)), "consumed")
		assert.True(t, stock.GetStockUnits().Equal(decimal.NewFromInt(120)), "system stock")
		assert.True(t, stock.RealStockUnits.Equal(decimal.NewFromInt(40)), "field stock")
		assert.True(t, stock.GetStockDifference().Equal(decimal.NewFromInt(-80)), "difference")
	}
}

func TestRepository_GetStocks_ClosedFilterIncludesBackdatedActiveMovement(t *testing.T) {
	repo, db := newStockTestRepo(t)
	ctx := context.Background()
	closeDate := time.Date(2026, time.April, 15, 0, 0, 0, 0, time.UTC)
	backdated := time.Date(2026, time.April, 15, 18, 30, 0, 0, time.UTC)

	seedStockLookupReferences(t, db)
	err := db.Exec(
		`INSERT INTO stocks (
			id, project_id, supply_id, investor_id, close_date, real_stock_units,
			initial_units, year_period, month_period, units_entered, units_consumed,
			has_real_stock_count
		) VALUES
			(401, 1, 7, 11, ?, '0', '0', 2026, 4, '0', '0', false),
			(402, 1, 7, 11, NULL, '0', '0', 2026, 5, '0', '0', false)`,
		closeDate,
	).Error
	assert.NoError(t, err)
	err = db.Exec(
		`INSERT INTO supply_movements (
			id, stock_id, quantity, movement_type, movement_date, reference_number,
			project_id, project_destination_id, supply_id, investor_id, provider_id, is_entry
		) VALUES
			(1101, 402, '30', 'Remito oficial', ?, 'BACKDATED-ACTIVE', 1, 0, 7, 11, 1, true)`,
		backdated,
	).Error
	assert.NoError(t, err)

	stocks, err := repo.GetStocks(ctx, 1, closeDate)

	assert.NoError(t, err)
	if assert.Len(t, stocks, 1) {
		assert.True(t, stocks[0].GetEntryStock().Equal(decimal.NewFromInt(30)))
		assert.True(t, stocks[0].GetStockUnits().Equal(decimal.NewFromInt(30)))
	}
}

func TestRepository_GetStocks_ClosedPeriodUsesProjectSupplyAndCutoffDate(t *testing.T) {
	repo, db := newStockTestRepo(t)
	ctx := context.Background()
	closeDate := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)

	seedStockLookupReferences(t, db)
	err := db.Exec(
		`INSERT INTO stocks (
			id, project_id, supply_id, investor_id, close_date, real_stock_units,
			initial_units, year_period, month_period, units_entered, units_consumed,
			has_real_stock_count
		) VALUES
			(101, 1, 7, 10, ?, '0', '0', 2026, 3, '0', '0', false),
			(102, 1, 7, 11, ?, '0', '0', 2026, 3, '0', '0', false)`,
		closeDate,
		closeDate,
	).Error
	assert.NoError(t, err)
	err = db.Exec(`INSERT INTO workorders (id, project_id, investor_id, date) VALUES
		(201, 1, 10, '2026-04-10'),
		(202, 1, 11, '2026-04-10'),
		(203, 1, 10, '2026-04-20')`).Error
	assert.NoError(t, err)
	err = db.Exec(`INSERT INTO workorder_items (id, workorder_id, supply_id, total_used) VALUES
		(301, 201, 7, '45'),
		(302, 202, 7, '5'),
		(303, 203, 7, '100')`).Error
	assert.NoError(t, err)
	err = db.Exec(
		`INSERT INTO supply_movements (
			id, stock_id, quantity, movement_type, movement_date, reference_number,
			project_id, project_destination_id, supply_id, investor_id, provider_id, is_entry
		) VALUES
			(401, 101, '50', 'Remito oficial', '2026-04-10', 'BEFORE', 1, 0, 7, 10, 1, true),
			(402, 102, '6', 'Remito oficial', '2026-04-10', 'BEFORE', 1, 0, 7, 11, 1, true),
			(403, 101, '100', 'Remito oficial', '2026-04-20', 'AFTER', 1, 0, 7, 10, 1, true)`,
	).Error
	assert.NoError(t, err)

	stocks, err := repo.GetStocks(ctx, 1, closeDate)

	assert.NoError(t, err)
	if assert.Len(t, stocks, 1) {
		assert.Equal(t, int64(0), stocks[0].ID)
		assert.True(t, stocks[0].GetEntryStock().Equal(decimal.NewFromInt(56)))
		assert.True(t, stocks[0].Consumed.Equal(decimal.NewFromInt(50)))
		assert.True(t, stocks[0].GetStockUnits().Equal(decimal.NewFromInt(6)))
	}
}

func TestRepository_GetLastStockByProjectInvestorID_UsesCurrentSupplyInvestorRollup(t *testing.T) {
	repo, db := newStockTestRepo(t)
	ctx := context.Background()

	seedStockLookupReferences(t, db)
	seedClosedAndActiveStockRollup(t, db)

	stock, isFirst, err := repo.GetLastStockByProjectInvestorID(ctx, 1, 7, 11)

	assert.NoError(t, err)
	assert.False(t, isFirst)
	if assert.NotNil(t, stock) {
		assert.Equal(t, int64(102), stock.ID)
		assert.True(t, stock.GetEntryStock().Equal(decimal.NewFromInt(220)))
		assert.True(t, stock.Consumed.Equal(decimal.NewFromInt(220)))
		assert.True(t, stock.GetStockUnits().Equal(decimal.Zero))
	}
}

func TestRepository_UpdateRealStockUnits_ScopesByProject(t *testing.T) {
	repo, db := newStockTestRepo(t)
	ctx := context.Background()
	currentVersion := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)

	err := db.Exec(
		"INSERT INTO stocks (id, project_id, real_stock_units, has_real_stock_count, updated_at) VALUES (?, ?, ?, ?, ?)",
		10,
		2,
		"3",
		false,
		currentVersion,
	).Error
	assert.NoError(t, err)

	err = repo.UpdateRealStockUnits(ctx, 10, &domain.Stock{
		Project:           &projectdomain.Project{ID: 1},
		RealStockUnits:    decimal.NewFromInt(9),
		HasRealStockCount: true,
	})

	assert.ErrorContains(t, err, "no stock found to update")

	var realStockUnits string
	err = db.Raw("SELECT real_stock_units FROM stocks WHERE id = ?", 10).Scan(&realStockUnits).Error
	assert.NoError(t, err)
	assert.Equal(t, "3", realStockUnits)
}

func TestRepository_UpdateRealStockUnits_AllowsClosedStock(t *testing.T) {
	repo, db := newStockTestRepo(t)
	ctx := context.Background()
	closedAt := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
	currentVersion := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)

	err := db.Exec(
		"INSERT INTO stocks (id, project_id, close_date, real_stock_units, has_real_stock_count, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		10,
		1,
		closedAt,
		"3",
		true,
		currentVersion,
	).Error
	assert.NoError(t, err)

	err = repo.UpdateRealStockUnits(ctx, 10, &domain.Stock{
		Project:           &projectdomain.Project{ID: 1},
		RealStockUnits:    decimal.NewFromInt(9),
		HasRealStockCount: true,
		Base: shareddomain.Base{
			UpdatedAt: currentVersion,
		},
	})

	assert.NoError(t, err)

	var realStockUnits string
	err = db.Raw("SELECT real_stock_units FROM stocks WHERE id = ?", 10).Scan(&realStockUnits).Error
	assert.NoError(t, err)
	assert.Equal(t, "9", realStockUnits)
}

func TestRepository_UpdateRealStockUnits_RejectsStaleVersion(t *testing.T) {
	repo, db := newStockTestRepo(t)
	ctx := context.Background()
	currentVersion := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
	staleVersion := currentVersion.Add(-time.Minute)

	err := db.Exec(
		"INSERT INTO stocks (id, project_id, real_stock_units, has_real_stock_count, updated_at) VALUES (?, ?, ?, ?, ?)",
		10,
		1,
		"3",
		false,
		currentVersion,
	).Error
	assert.NoError(t, err)

	err = repo.UpdateRealStockUnits(ctx, 10, &domain.Stock{
		Project:           &projectdomain.Project{ID: 1},
		RealStockUnits:    decimal.NewFromInt(9),
		HasRealStockCount: true,
		Base: shareddomain.Base{
			UpdatedAt: staleVersion,
		},
	})

	assert.ErrorContains(t, err, "stock not found or outdated")

	var realStockUnits string
	err = db.Raw("SELECT real_stock_units FROM stocks WHERE id = ?", 10).Scan(&realStockUnits).Error
	assert.NoError(t, err)
	assert.Equal(t, "3", realStockUnits)
}

func TestRepository_UpdateRealStockUnits_AllowsMatchingProjectAndVersion(t *testing.T) {
	repo, db := newStockTestRepo(t)
	ctx := context.Background()
	currentVersion := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
	actor := "user@example.com"

	err := db.Exec(
		"INSERT INTO stocks (id, project_id, real_stock_units, has_real_stock_count, updated_at) VALUES (?, ?, ?, ?, ?)",
		10,
		1,
		"3",
		false,
		currentVersion,
	).Error
	assert.NoError(t, err)

	err = repo.UpdateRealStockUnits(ctx, 10, &domain.Stock{
		Project:           &projectdomain.Project{ID: 1},
		RealStockUnits:    decimal.NewFromInt(9),
		HasRealStockCount: true,
		Base: shareddomain.Base{
			UpdatedAt: currentVersion,
			UpdatedBy: &actor,
		},
	})
	assert.NoError(t, err)

	var row struct {
		RealStockUnits    string
		HasRealStockCount bool
		UpdatedBy         string
	}
	err = db.Raw(
		"SELECT real_stock_units, has_real_stock_count, updated_by FROM stocks WHERE id = ?",
		10,
	).Scan(&row).Error
	assert.NoError(t, err)
	assert.Equal(t, "9", row.RealStockUnits)
	assert.True(t, row.HasRealStockCount)
	assert.Equal(t, actor, row.UpdatedBy)
}

func TestRepository_UpdateUnitsConsumed_AllowsClosedStock(t *testing.T) {
	repo, db := newStockTestRepo(t)
	ctx := context.Background()
	closedAt := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
	currentVersion := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)

	err := db.Exec(
		"INSERT INTO stocks (id, project_id, close_date, real_stock_units, units_consumed, has_real_stock_count, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		10,
		1,
		closedAt,
		"3",
		"2",
		true,
		currentVersion,
	).Error
	assert.NoError(t, err)

	err = repo.UpdateUnitsConsumed(ctx, domain.Stock{
		ID:      10,
		Project: &projectdomain.Project{ID: 1},
		Base: shareddomain.Base{
			UpdatedAt: currentVersion,
		},
	}, decimal.NewFromInt(5))

	assert.NoError(t, err)

	var unitsConsumed string
	err = db.Raw("SELECT units_consumed FROM stocks WHERE id = ?", 10).Scan(&unitsConsumed).Error
	assert.NoError(t, err)
	assert.Equal(t, "7", unitsConsumed)
}

func TestRepository_GetLastStockByProjectInvestorID_FiltersByInvestor(t *testing.T) {
	repo, db := newStockTestRepo(t)
	ctx := context.Background()

	seedStockLookupReferences(t, db)
	err := db.Exec(
		`INSERT INTO stocks (
			id, project_id, supply_id, investor_id, close_date, real_stock_units,
			initial_units, year_period, month_period, units_entered, units_consumed,
			has_real_stock_count
		) VALUES
			(101, 1, 7, 10, NULL, '1', '0', 2026, 4, '0', '0', false),
			(102, 1, 7, 11, NULL, '2', '0', 2026, 4, '0', '0', false)`,
	).Error
	assert.NoError(t, err)
	err = db.Exec("INSERT INTO workorders (id, project_id, investor_id, date) VALUES (201, 1, 11, '2026-04-01')").Error
	assert.NoError(t, err)
	err = db.Exec("INSERT INTO workorder_items (id, workorder_id, supply_id, total_used) VALUES (301, 201, 7, '5')").Error
	assert.NoError(t, err)

	stock, isFirst, err := repo.GetLastStockByProjectInvestorID(ctx, 1, 7, 11)

	assert.NoError(t, err)
	assert.False(t, isFirst)
	if assert.NotNil(t, stock) {
		assert.Equal(t, int64(102), stock.ID)
		assert.Equal(t, int64(1), stock.Project.ID)
		assert.Equal(t, int64(7), stock.Supply.ID)
		assert.Equal(t, int64(11), stock.Investor.ID)
		assert.True(t, stock.Consumed.Equal(decimal.NewFromInt(5)))
	}
}

func TestRepository_GetLastStockByProjectInvestorID_ReturnsFirstWhenInvestorHasNoActiveStock(t *testing.T) {
	repo, db := newStockTestRepo(t)
	ctx := context.Background()

	seedStockLookupReferences(t, db)
	err := db.Exec(
		`INSERT INTO stocks (
			id, project_id, supply_id, investor_id, close_date, real_stock_units,
			initial_units, year_period, month_period, units_entered, units_consumed,
			has_real_stock_count
		) VALUES
			(101, 1, 7, 10, NULL, '1', '0', 2026, 4, '0', '0', false)`,
	).Error
	assert.NoError(t, err)

	stock, isFirst, err := repo.GetLastStockByProjectInvestorID(ctx, 1, 7, 11)

	assert.NoError(t, err)
	assert.True(t, isFirst)
	assert.Nil(t, stock)
}

func seedClosedAndActiveStockRollup(t *testing.T, db *gorm.DB) {
	t.Helper()

	stmts := []string{
		`INSERT INTO stocks (
			id, project_id, supply_id, investor_id, close_date, real_stock_units,
			initial_units, year_period, month_period, units_entered, units_consumed,
			has_real_stock_count
		) VALUES
			(101, 1, 7, 11, '2026-04-15', '0', '0', 2026, 3, '0', '0', false),
			(102, 1, 7, 11, NULL, '0', '0', 2026, 4, '0', '0', true)`,
		`INSERT INTO supply_movements (
			id, stock_id, quantity, movement_type, movement_date, reference_number,
			project_id, project_destination_id, supply_id, investor_id, provider_id, is_entry
		) VALUES
			(201, 101, '20', 'Remito oficial', '2026-03-16', 'MAR', 1, 0, 7, 11, 1, true),
			(202, 102, '200', 'Remito oficial', '2026-04-13', 'APR', 1, 0, 7, 11, 1, true)`,
		`INSERT INTO workorders (id, project_id, investor_id, date) VALUES
			(301, 1, 11, '2026-03-16'),
			(302, 1, 11, '2026-04-13')`,
		`INSERT INTO workorder_items (id, workorder_id, supply_id, total_used) VALUES
			(401, 301, 7, '20'),
			(402, 302, 7, '200')`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("seed current rollup: %v", err)
		}
	}
}

func seedStockLookupReferences(t *testing.T, db *gorm.DB) {
	t.Helper()

	stmts := []string{
		"INSERT INTO projects (id, name, customer_id, campaign_id, admin_cost, planned_cost) VALUES (1, 'Project', 1, 1, '0', '0')",
		"INSERT INTO types (id, name) VALUES (1, 'Agroquimico')",
		"INSERT INTO categories (id, name, type_id) VALUES (1, 'Herbicida', 1)",
		"INSERT INTO supplies (id, project_id, name, price, is_partial_price, is_pending, unit_id, category_id, type_id) VALUES (7, 1, 'Sempra', '1', false, false, 1, 1, 1)",
		"INSERT INTO investors (id, name) VALUES (10, 'Investor A')",
		"INSERT INTO investors (id, name) VALUES (11, 'Investor B')",
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("seed lookup references: %v", err)
		}
	}
}
