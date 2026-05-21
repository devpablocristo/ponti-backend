package supply

import (
	"context"
	"testing"
	"time"

	contextkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	classdomain "github.com/devpablocristo/ponti-backend/internal/class-type/usecases/domain"
	investordomain "github.com/devpablocristo/ponti-backend/internal/investor/usecases/domain"
	providerdomain "github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	domain "github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
)

type supplyTenantGormEngine struct {
	client *gorm.DB
}

func (e supplyTenantGormEngine) Client() *gorm.DB {
	return e.client
}

func supplyTenantContext(tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "tenant-user@example.com")
	ctx = context.WithValue(ctx, contextkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, contextkeys.Role, "tenant_admin")
	ctx = context.WithValue(ctx, contextkeys.Scopes, []string{"supplies.read", "supplies.write", "supplies.archive"})
	return ctx
}

func setupSupplyTenantDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.Exec(`
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
		CREATE TABLE projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			deleted_at DATETIME
		);
		CREATE TABLE providers (
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
		CREATE TABLE supply_movements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			stock_id INTEGER NOT NULL DEFAULT 0,
			quantity NUMERIC NOT NULL DEFAULT 0,
			movement_type TEXT NOT NULL,
			movement_date DATETIME,
			reference_number TEXT NOT NULL,
			project_id INTEGER NOT NULL,
			project_destination_id INTEGER NOT NULL DEFAULT 0,
			supply_id INTEGER NOT NULL,
			investor_id INTEGER NOT NULL,
			provider_id INTEGER NOT NULL,
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
			deleted_at DATETIME
		);
		CREATE TABLE workorder_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			workorder_id INTEGER NOT NULL,
			supply_id INTEGER NOT NULL,
			deleted_at DATETIME
		);
		CREATE TABLE stocks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			project_id INTEGER NOT NULL DEFAULT 0,
			supply_id INTEGER NOT NULL,
			investor_id INTEGER NOT NULL DEFAULT 0,
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
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT,
			deleted_at DATETIME
		);
	`).Error; err != nil {
		t.Fatalf("create schema: %v", err)
	}

	return db
}

func TestSupplyRepositoryTenantIsolation(t *testing.T) {
	db := setupSupplyTenantDB(t)
	repo := NewRepository(supplyTenantGormEngine{client: db})

	tenantA := uuid.New()
	tenantB := uuid.New()
	now := time.Now().UTC()

	if err := db.Exec(`
		INSERT INTO categories (id, name, type_id, created_at, updated_at, deleted_at) VALUES
			(1, 'Agroquimicos', 1, ?, ?, NULL);
		INSERT INTO types (id, name, created_at, updated_at, deleted_at) VALUES
			(1, 'Herbicida', ?, ?, NULL);
		INSERT INTO projects (id, tenant_id, name, deleted_at) VALUES
			(10, ?, 'Project A', NULL),
			(20, ?, 'Project B', NULL);
		INSERT INTO supplies (
			id, tenant_id, project_id, name, price, is_partial_price, is_pending,
			unit_id, category_id, type_id, created_at, updated_at, deleted_at
		) VALUES
			(1, ?, 10, 'Supply A', 10, false, false, 1, 1, 1, ?, ?, NULL),
			(2, ?, 20, 'Supply B', 20, false, false, 1, 1, 1, ?, ?, NULL),
			(3, ?, 20, 'Supply B archived', 30, false, false, 1, 1, 1, ?, ?, ?),
			(4, ?, 10, 'Supply A archived', 40, false, false, 1, 1, 1, ?, ?, ?);
		INSERT INTO supply_movements (
			id, tenant_id, stock_id, quantity, movement_type, movement_date,
			reference_number, project_id, project_destination_id, supply_id,
			investor_id, provider_id, is_entry, created_at, updated_at, deleted_at
		) VALUES
			(1, ?, 1, 10, 'Remito oficial', ?, 'REM-B', 20, 0, 2, 22, 1, true, ?, ?, NULL);
		INSERT INTO workorders (id, tenant_id, deleted_at) VALUES
			(1, ?, NULL);
		INSERT INTO workorder_items (id, tenant_id, workorder_id, supply_id, deleted_at) VALUES
			(1, ?, 1, 2, NULL);
	`, now, now,
		now, now,
		tenantA.String(), tenantB.String(),
		tenantA.String(), now, now,
		tenantB.String(), now, now,
		tenantB.String(), now, now, now,
		tenantA.String(), now, now, now,
		tenantB.String(), now, now, now,
		tenantB.String(),
		tenantB.String(),
	).Error; err != nil {
		t.Fatalf("seed supplies: %v", err)
	}

	ctxA := supplyTenantContext(tenantA)

	archived, total, err := repo.ListArchivedSupplies(ctxA, 1, 50)
	if err != nil {
		t.Fatalf("list archived supplies: %v", err)
	}
	if total != 1 || len(archived) != 1 || archived[0].ID != 4 {
		t.Fatalf("expected only tenant A archived supply, total=%d archived=%#v", total, archived)
	}

	if _, err := repo.GetSupply(ctxA, 2); err == nil {
		t.Fatalf("expected get cross-tenant supply to fail")
	}

	supplies, err := repo.GetSuppliesByIDs(ctxA, []int64{1, 2})
	if err != nil {
		t.Fatalf("get supplies by ids: %v", err)
	}
	if len(supplies) != 1 || supplies[0].ID != 1 {
		t.Fatalf("expected only tenant A supply by ids, got %#v", supplies)
	}

	if _, err := repo.GetSupplyByProjectAndName(ctxA, 20, "Supply B"); err == nil {
		t.Fatalf("expected get supply by cross-tenant project/name to fail")
	}

	projectExists, err := repo.ProjectExists(ctxA, 20)
	if err != nil {
		t.Fatalf("project exists cross-tenant: %v", err)
	}
	if projectExists {
		t.Fatalf("expected cross-tenant project to be invisible")
	}

	duplicate, err := repo.ExistsSupplyMovementByProjectReferenceAndSupply(ctxA, 20, "REM-B", 2)
	if err != nil {
		t.Fatalf("duplicate movement by project/reference/supply: %v", err)
	}
	if duplicate {
		t.Fatalf("expected duplicate movement check to ignore cross-tenant movement")
	}

	duplicateByType, err := repo.ExistsSupplyMovementByProjectReferenceAndType(ctxA, 20, "REM-B", "Remito oficial")
	if err != nil {
		t.Fatalf("duplicate movement by project/reference/type: %v", err)
	}
	if duplicateByType {
		t.Fatalf("expected duplicate movement by type check to ignore cross-tenant movement")
	}

	duplicateBySupplyAndType, err := repo.ExistsSupplyMovementByProjectReferenceSupplyAndType(ctxA, 20, "REM-B", 2, "Remito oficial")
	if err != nil {
		t.Fatalf("duplicate movement by project/reference/supply/type: %v", err)
	}
	if duplicateBySupplyAndType {
		t.Fatalf("expected duplicate movement by supply/type check to ignore cross-tenant movement")
	}

	woCount, err := repo.GetWorkOrdersBySupplyID(ctxA, 2)
	if err != nil {
		t.Fatalf("work orders by cross-tenant supply: %v", err)
	}
	if woCount != 0 {
		t.Fatalf("expected no cross-tenant work order count, got %d", woCount)
	}

	if err := repo.UpdateSupply(ctxA, &domain.Supply{
		ID:             2,
		ProjectID:      20,
		Name:           "cross tenant update",
		UnitID:         1,
		Price:          decimal.NewFromInt(99),
		CategoryID:     1,
		Type:           classdomain.ClassType{ID: 1},
		IsPartialPrice: false,
		Base:           shareddomain.Base{UpdatedAt: now},
	}); err == nil {
		t.Fatalf("expected update cross-tenant supply to fail")
	}

	var name string
	if err := db.Raw(`SELECT name FROM supplies WHERE id = 2`).Scan(&name).Error; err != nil {
		t.Fatalf("read supply 2: %v", err)
	}
	if name != "Supply B" {
		t.Fatalf("cross-tenant update changed supply 2 name to %q", name)
	}

	if err := repo.ArchiveSupply(ctxA, 2); err == nil {
		t.Fatalf("expected archive cross-tenant supply to fail")
	}
	var deletedCount int64
	if err := db.Raw(`SELECT COUNT(*) FROM supplies WHERE id = 2 AND deleted_at IS NOT NULL`).Scan(&deletedCount).Error; err != nil {
		t.Fatalf("check archive side effect: %v", err)
	}
	if deletedCount != 0 {
		t.Fatalf("cross-tenant archive modified supply 2")
	}

	if err := repo.RestoreSupply(ctxA, 3); err == nil {
		t.Fatalf("expected restore cross-tenant supply to fail")
	}
	if err := db.Raw(`SELECT COUNT(*) FROM supplies WHERE id = 3 AND deleted_at IS NOT NULL`).Scan(&deletedCount).Error; err != nil {
		t.Fatalf("check restore side effect: %v", err)
	}
	if deletedCount != 1 {
		t.Fatalf("cross-tenant restore modified supply 3")
	}

	if err := repo.HardDeleteSupply(ctxA, 2); err == nil {
		t.Fatalf("expected hard-delete cross-tenant supply to fail")
	}
	var exists int64
	if err := db.Raw(`SELECT COUNT(*) FROM supplies WHERE id = 2`).Scan(&exists).Error; err != nil {
		t.Fatalf("check hard delete side effect: %v", err)
	}
	if exists != 1 {
		t.Fatalf("cross-tenant delete removed supply 2")
	}
}

func TestSupplyMovementRepositoryTenantIsolation(t *testing.T) {
	db := setupSupplyTenantDB(t)
	repo := NewRepository(supplyTenantGormEngine{client: db})

	tenantA := uuid.New()
	tenantB := uuid.New()
	now := time.Now().UTC()

	if err := db.Exec(`
		INSERT INTO categories (id, name, type_id, created_at, updated_at, deleted_at) VALUES
			(1, 'Agroquimicos', 1, ?, ?, NULL);
		INSERT INTO types (id, name, created_at, updated_at, deleted_at) VALUES
			(1, 'Herbicida', ?, ?, NULL);
		INSERT INTO projects (id, tenant_id, name, deleted_at) VALUES
			(10, ?, 'Project A', NULL),
			(20, ?, 'Project B', NULL);
		INSERT INTO investors (id, tenant_id, name, created_at, updated_at, deleted_at) VALUES
			(1, ?, 'Investor A', ?, ?, NULL),
			(2, ?, 'Investor B', ?, ?, NULL);
		INSERT INTO providers (id, tenant_id, name, created_at, updated_at, deleted_at) VALUES
			(1, ?, 'Provider A', ?, ?, NULL),
			(2, ?, 'Provider B', ?, ?, NULL);
		INSERT INTO supplies (
			id, tenant_id, project_id, name, price, is_partial_price, is_pending,
			unit_id, category_id, type_id, created_at, updated_at, deleted_at
		) VALUES
			(1, ?, 10, 'Supply A', 10, false, false, 1, 1, 1, ?, ?, NULL),
			(2, ?, 20, 'Supply B', 20, false, false, 1, 1, 1, ?, ?, NULL);
		INSERT INTO stocks (
			id, tenant_id, project_id, supply_id, investor_id, real_stock_units,
			has_real_stock_count, created_at, updated_at, deleted_at
		) VALUES
			(1, ?, 10, 1, 1, 100, true, ?, ?, NULL),
			(2, ?, 20, 2, 2, 200, true, ?, ?, NULL);
		INSERT INTO supply_movements (
			id, tenant_id, stock_id, quantity, movement_type, movement_date,
			reference_number, project_id, project_destination_id, supply_id,
			investor_id, provider_id, is_entry, created_at, updated_at, deleted_at
		) VALUES
			(1, ?, 1, 10, 'Remito oficial', ?, 'REM-A', 10, 0, 1, 1, 1, true, ?, ?, NULL),
			(2, ?, 2, 20, 'Remito oficial', ?, 'REM-B', 20, 0, 2, 2, 2, true, ?, ?, NULL),
			(3, ?, 2, 30, 'Remito oficial', ?, 'REM-B-ARCH', 20, 0, 2, 2, 2, true, ?, ?, ?);
	`, now, now,
		now, now,
		tenantA.String(), tenantB.String(),
		tenantA.String(), now, now,
		tenantB.String(), now, now,
		tenantA.String(), now, now,
		tenantB.String(), now, now,
		tenantA.String(), now, now,
		tenantB.String(), now, now,
		tenantA.String(), now, now,
		tenantB.String(), now, now,
		tenantA.String(), now, now, now,
		tenantB.String(), now, now, now,
		tenantB.String(), now, now, now, now,
	).Error; err != nil {
		t.Fatalf("seed supply movements: %v", err)
	}

	ctxA := supplyTenantContext(tenantA)

	if _, err := repo.GetSupplyMovementByID(ctxA, 2); err == nil {
		t.Fatalf("expected get cross-tenant supply movement to fail")
	}

	archived, err := repo.ListArchivedSupplyMovements(ctxA, 20)
	if err != nil {
		t.Fatalf("list archived cross-tenant supply movements: %v", err)
	}
	if len(archived) != 0 {
		t.Fatalf("expected no archived movements for cross-tenant project, got %#v", archived)
	}

	if err := repo.UpdateSupplyMovement(ctxA, &domain.SupplyMovement{
		ID:                   2,
		StockId:              2,
		Quantity:             decimal.NewFromInt(999),
		MovementType:         "Remito oficial",
		MovementDate:         &now,
		ReferenceNumber:      "REM-B",
		ProjectId:            20,
		ProjectDestinationId: 0,
		Supply:               &domain.Supply{ID: 2},
		Investor:             &investordomain.Investor{ID: 2},
		Provider:             &providerdomain.Provider{ID: 2},
		IsEntry:              true,
		Base:                 shareddomain.Base{UpdatedAt: now},
	}); err == nil {
		t.Fatalf("expected update cross-tenant supply movement to fail")
	}

	var quantity string
	if err := db.Raw(`SELECT quantity FROM supply_movements WHERE id = 2`).Scan(&quantity).Error; err != nil {
		t.Fatalf("read movement 2 quantity: %v", err)
	}
	if quantity != "20" {
		t.Fatalf("cross-tenant update changed movement 2 quantity to %q", quantity)
	}

	if err := repo.DeleteSupplyMovement(ctxA, 20, 2); err == nil {
		t.Fatalf("expected delete cross-tenant supply movement to fail")
	}
	if err := repo.ArchiveSupplyMovement(ctxA, 20, 2); err == nil {
		t.Fatalf("expected archive cross-tenant supply movement to fail")
	}

	var archivedCount int64
	if err := db.Raw(`SELECT COUNT(*) FROM supply_movements WHERE id = 2 AND deleted_at IS NOT NULL`).Scan(&archivedCount).Error; err != nil {
		t.Fatalf("check archive side effect: %v", err)
	}
	if archivedCount != 0 {
		t.Fatalf("cross-tenant archive modified movement 2")
	}

	if err := repo.RestoreSupplyMovement(ctxA, 20, 3); err == nil {
		t.Fatalf("expected restore cross-tenant supply movement to fail")
	}
	if err := db.Raw(`SELECT COUNT(*) FROM supply_movements WHERE id = 3 AND deleted_at IS NOT NULL`).Scan(&archivedCount).Error; err != nil {
		t.Fatalf("check restore side effect: %v", err)
	}
	if archivedCount != 1 {
		t.Fatalf("cross-tenant restore modified movement 3")
	}

	if err := repo.HardDeleteSupplyMovement(ctxA, 20, 2); err == nil {
		t.Fatalf("expected hard-delete cross-tenant supply movement to fail")
	}
	var exists int64
	if err := db.Raw(`SELECT COUNT(*) FROM supply_movements WHERE id = 2`).Scan(&exists).Error; err != nil {
		t.Fatalf("check hard delete side effect: %v", err)
	}
	if exists != 1 {
		t.Fatalf("cross-tenant hard delete removed movement 2")
	}

	if err := repo.ResetFieldStockCounts(ctxA, 20, nil); err != nil {
		t.Fatalf("reset cross-tenant stock counts: %v", err)
	}
	var realStock string
	if err := db.Raw(`SELECT real_stock_units FROM stocks WHERE id = 2`).Scan(&realStock).Error; err != nil {
		t.Fatalf("read stock 2 real stock: %v", err)
	}
	if realStock != "200" {
		t.Fatalf("cross-tenant stock reset changed stock 2 to %q", realStock)
	}
}

func TestSupplyRepositoryRequiresTenantInStrictMode(t *testing.T) {
	t.Setenv("TENANT_STRICT_MODE", "true")

	db := setupSupplyTenantDB(t)
	repo := NewRepository(supplyTenantGormEngine{client: db})

	if _, _, err := repo.ListAllSupplies(context.Background(), domain.SupplyFilter{}); err == nil {
		t.Fatalf("expected ListAllSupplies without tenant to fail in strict mode")
	}
	if _, err := repo.GetEntriesSupplyMovementsByProjectID(context.Background(), 10); err == nil {
		t.Fatalf("expected GetEntriesSupplyMovementsByProjectID without tenant to fail in strict mode")
	}
	if _, err := repo.ListArchivedSupplyMovements(context.Background(), 10); err == nil {
		t.Fatalf("expected ListArchivedSupplyMovements without tenant to fail in strict mode")
	}
}
