package stock

import (
	"context"
	"testing"
	"time"

	contextkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	investordomain "github.com/devpablocristo/ponti-backend/internal/investor/usecases/domain"
	projectdomain "github.com/devpablocristo/ponti-backend/internal/project/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	stockdomain "github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
	supplydomain "github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type stockTenantGormEngine struct {
	client *gorm.DB
}

func (e stockTenantGormEngine) Client() *gorm.DB {
	return e.client
}

func stockTenantContext(tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "tenant-user@example.com")
	ctx = context.WithValue(ctx, contextkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, contextkeys.Role, "tenant_admin")
	ctx = context.WithValue(ctx, contextkeys.Scopes, []string{"stock.read", "stock.write", "stock.archive"})
	return ctx
}

func setupStockTenantDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			customer_id INTEGER NOT NULL DEFAULT 0,
			campaign_id INTEGER NOT NULL DEFAULT 0,
			admin_cost NUMERIC NOT NULL DEFAULT 0,
			planned_cost NUMERIC NOT NULL DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		);
		CREATE TABLE categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			type_id INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		);
		CREATE TABLE types (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		);
		CREATE TABLE supplies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			project_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			price NUMERIC NOT NULL DEFAULT 0,
			is_partial_price BOOLEAN NOT NULL DEFAULT false,
			is_pending BOOLEAN NOT NULL DEFAULT false,
			unit_id INTEGER NOT NULL DEFAULT 0,
			category_id INTEGER NOT NULL DEFAULT 0,
			type_id INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		);
		CREATE TABLE investors (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		);
		CREATE TABLE stocks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			project_id INTEGER NOT NULL,
			supply_id INTEGER NOT NULL,
			investor_id INTEGER NOT NULL,
			close_date DATETIME,
			initial_units NUMERIC NOT NULL DEFAULT 0,
			year_period INTEGER NOT NULL DEFAULT 0,
			month_period INTEGER NOT NULL DEFAULT 0,
			units_entered NUMERIC NOT NULL DEFAULT 0,
			units_consumed NUMERIC NOT NULL DEFAULT 0,
			real_stock_units NUMERIC NOT NULL DEFAULT 0,
			has_real_stock_count BOOLEAN NOT NULL DEFAULT false,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		);
		CREATE TABLE supply_movements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			stock_id INTEGER NOT NULL,
			quantity NUMERIC NOT NULL DEFAULT 0,
			movement_type TEXT NOT NULL,
			movement_date DATETIME,
			reference_number TEXT NOT NULL,
			project_id INTEGER NOT NULL,
			project_destination_id INTEGER NOT NULL DEFAULT 0,
			supply_id INTEGER NOT NULL,
			investor_id INTEGER NOT NULL,
			provider_id INTEGER NOT NULL DEFAULT 0,
			is_entry BOOLEAN NOT NULL DEFAULT true,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		);
		CREATE TABLE workorders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			project_id INTEGER NOT NULL,
			investor_id INTEGER NOT NULL,
			date DATETIME,
			deleted_at DATETIME
		);
		CREATE TABLE workorder_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			workorder_id INTEGER NOT NULL,
			supply_id INTEGER NOT NULL,
			total_used NUMERIC NOT NULL DEFAULT 0,
			deleted_at DATETIME
		);
	`).Error; err != nil {
		t.Fatalf("create schema: %v", err)
	}

	return db
}

func TestStockRepositoryTenantIsolation(t *testing.T) {
	db := setupStockTenantDB(t)
	repo := NewRepository(stockTenantGormEngine{client: db})

	tenantA := uuid.New()
	tenantB := uuid.New()
	now := time.Now().UTC()

	mustExec := func(query string, args ...any) {
		t.Helper()
		if err := db.Exec(query, args...).Error; err != nil {
			t.Fatalf("seed stock: %v", err)
		}
	}

	mustExec(`INSERT INTO categories (id, name, type_id, created_at, updated_at, deleted_at) VALUES (1, 'Agroquimicos', 1, ?, ?, NULL)`, now, now)
	mustExec(`INSERT INTO types (id, name, created_at, updated_at, deleted_at) VALUES (1, 'Herbicida', ?, ?, NULL)`, now, now)
	mustExec(`INSERT INTO projects (id, tenant_id, name, created_at, updated_at, deleted_at) VALUES
		(10, ?, 'Project A', ?, ?, NULL),
		(20, ?, 'Project B', ?, ?, NULL)`, tenantA.String(), now, now, tenantB.String(), now, now)
	mustExec(`INSERT INTO investors (id, tenant_id, name, created_at, updated_at, deleted_at) VALUES
		(1, ?, 'Investor A', ?, ?, NULL),
		(2, ?, 'Investor B', ?, ?, NULL)`, tenantA.String(), now, now, tenantB.String(), now, now)
	mustExec(`INSERT INTO supplies (
		id, tenant_id, project_id, name, price, is_partial_price, is_pending,
		unit_id, category_id, type_id, created_at, updated_at, deleted_at
	) VALUES
		(1, ?, 10, 'Supply A', 10, false, false, 1, 1, 1, ?, ?, NULL),
		(2, ?, 20, 'Supply B', 20, false, false, 1, 1, 1, ?, ?, NULL)`, tenantA.String(), now, now, tenantB.String(), now, now)
	mustExec(`INSERT INTO stocks (
		id, tenant_id, project_id, supply_id, investor_id, close_date,
		initial_units, year_period, month_period, units_entered, units_consumed,
		real_stock_units, has_real_stock_count, created_at, updated_at, deleted_at
	) VALUES
		(1, ?, 10, 1, 1, NULL, 0, 2026, 5, 0, 5, 100, true, ?, ?, NULL),
		(2, ?, 20, 2, 2, NULL, 0, 2026, 5, 0, 7, 200, true, ?, ?, NULL),
		(3, ?, 10, 1, 1, ?, 0, 2026, 4, 0, 0, 80, true, ?, ?, NULL)`,
		tenantA.String(), now, now,
		tenantB.String(), now, now,
		tenantA.String(), now, now, now,
	)
	mustExec(`INSERT INTO supply_movements (
		id, tenant_id, stock_id, quantity, movement_type, movement_date,
		reference_number, project_id, project_destination_id, supply_id,
		investor_id, provider_id, is_entry, created_at, updated_at, deleted_at
	) VALUES
		(1, ?, 1, 10, 'Remito oficial', ?, 'REM-A', 10, 0, 1, 1, 1, true, ?, ?, NULL),
		(2, ?, 2, 20, 'Remito oficial', ?, 'REM-B', 20, 0, 2, 2, 2, true, ?, ?, NULL)`,
		tenantA.String(), now, now, now,
		tenantB.String(), now, now, now,
	)
	mustExec(`INSERT INTO workorders (id, tenant_id, project_id, investor_id, date, deleted_at) VALUES (1, ?, 20, 2, ?, NULL)`, tenantB.String(), now)
	mustExec(`INSERT INTO workorder_items (id, tenant_id, workorder_id, supply_id, total_used, deleted_at) VALUES (1, ?, 1, 2, 7, NULL)`, tenantB.String())

	ctxA := stockTenantContext(tenantA)

	stocks, err := repo.GetStocks(ctxA, 20, time.Time{})
	if err != nil {
		t.Fatalf("get cross-tenant stocks summary: %v", err)
	}
	if len(stocks) != 0 {
		t.Fatalf("expected no cross-tenant stock summary rows, got %#v", stocks)
	}

	allStocks, err := repo.ListAllStocks(ctxA)
	if err != nil {
		t.Fatalf("list all stocks: %v", err)
	}
	if len(allStocks) != 2 {
		t.Fatalf("expected only tenant A stocks, got %#v", allStocks)
	}

	if _, err := repo.GetStockByID(ctxA, 2); err == nil {
		t.Fatalf("expected get cross-tenant stock by id to fail")
	}

	active, err := repo.GetActiveStocksByProjectID(ctxA, 20)
	if err != nil {
		t.Fatalf("get cross-tenant active stocks: %v", err)
	}
	if len(active) != 0 {
		t.Fatalf("expected no active stocks for cross-tenant project, got %#v", active)
	}

	periods, err := repo.GetStocksPeriods(ctxA, 20)
	if err != nil {
		t.Fatalf("get cross-tenant stock periods: %v", err)
	}
	if len(periods) != 0 {
		t.Fatalf("expected no periods for cross-tenant project, got %#v", periods)
	}

	if stock, isFirst, err := repo.GetLastStockByProjectID(ctxA, 20, 2); err != nil || !isFirst || stock != nil {
		t.Fatalf("expected cross-tenant last stock by project/supply to be first nil, stock=%#v isFirst=%v err=%v", stock, isFirst, err)
	}
	if stock, isFirst, err := repo.GetLastStockByProjectInvestorID(ctxA, 20, 2, 2); err != nil || !isFirst || stock != nil {
		t.Fatalf("expected cross-tenant last stock by project/supply/investor to be first nil, stock=%#v isFirst=%v err=%v", stock, isFirst, err)
	}
	if _, err := repo.GetStockByPeriodAndProjectID(ctxA, 20); err == nil {
		t.Fatalf("expected get cross-tenant stock by period/project to fail")
	}

	actor := "tenant-user@example.com"
	if err := repo.UpdateCloseDateByProject(ctxA, 20, &stockdomain.Stock{
		CloseDate: &now,
		Base:      shareddomain.Base{UpdatedBy: &actor},
	}); err == nil {
		t.Fatalf("expected update close date on cross-tenant project to fail")
	}

	if err := repo.UpdateRealStockUnits(ctxA, 2, &stockdomain.Stock{
		Project:           &projectdomain.Project{ID: 20},
		RealStockUnits:    decimal.NewFromInt(999),
		HasRealStockCount: true,
	}); err == nil {
		t.Fatalf("expected update real stock on cross-tenant stock to fail")
	}

	if err := repo.UpdateUnitsConsumed(ctxA, stockdomain.Stock{
		ID:      2,
		Project: &projectdomain.Project{ID: 20},
	}, decimal.NewFromInt(5)); err == nil {
		t.Fatalf("expected update consumed on cross-tenant stock to fail")
	}

	var stock2 struct {
		RealStockUnits string
		UnitsConsumed  string
		CloseDate      *time.Time
	}
	if err := db.Raw(`SELECT real_stock_units, units_consumed, close_date FROM stocks WHERE id = 2`).Scan(&stock2).Error; err != nil {
		t.Fatalf("read stock 2: %v", err)
	}
	if stock2.RealStockUnits != "200" || stock2.UnitsConsumed != "7" || stock2.CloseDate != nil {
		t.Fatalf("cross-tenant mutation changed stock 2: %#v", stock2)
	}

	createdID, err := repo.CreateStock(ctxA, &stockdomain.Stock{
		Project: &projectdomain.Project{ID: 10},
		Supply:  &supplydomain.Supply{ID: 1},
		Investor: &investordomain.Investor{
			ID: 1,
		},
		RealStockUnits:    decimal.NewFromInt(3),
		InitialStock:      decimal.Zero,
		YearPeriod:        2026,
		MonthPeriod:       6,
		HasRealStockCount: true,
		Base:              shareddomain.Base{CreatedAt: now, UpdatedAt: now},
	})
	if err != nil {
		t.Fatalf("create stock with tenant: %v", err)
	}
	var createdTenant string
	if err := db.Raw(`SELECT tenant_id FROM stocks WHERE id = ?`, createdID).Scan(&createdTenant).Error; err != nil {
		t.Fatalf("read created stock tenant: %v", err)
	}
	if createdTenant != tenantA.String() {
		t.Fatalf("created stock tenant_id = %q, want %q", createdTenant, tenantA.String())
	}
}

func TestStockRepositoryRequiresTenantInStrictMode(t *testing.T) {
	t.Setenv("TENANT_STRICT_MODE", "true")

	db := setupStockTenantDB(t)
	repo := NewRepository(stockTenantGormEngine{client: db})

	if _, err := repo.GetStocks(context.Background(), 10, time.Time{}); err == nil {
		t.Fatalf("expected GetStocks without tenant to fail in strict mode")
	}
	if _, err := repo.ListAllStocks(context.Background()); err == nil {
		t.Fatalf("expected ListAllStocks without tenant to fail in strict mode")
	}
}
