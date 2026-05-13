package labor

import (
	"context"
	"testing"
	"time"

	contextkeys "github.com/devpablocristo/core/security/go/contextkeys"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	domain "github.com/devpablocristo/ponti-backend/internal/labor/usecases/domain"
)

type laborTenantGormEngine struct {
	client *gorm.DB
}

func (e laborTenantGormEngine) Client() *gorm.DB {
	return e.client
}

func laborTenantContext(tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "tenant-user@example.com")
	ctx = context.WithValue(ctx, contextkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, contextkeys.Role, "tenant_admin")
	ctx = context.WithValue(ctx, contextkeys.Scopes, []string{"labors.read", "labors.write", "labors.archive"})
	return ctx
}

func setupLaborTenantDB(t *testing.T) *gorm.DB {
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
		CREATE TABLE labors (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			contractor_name TEXT NOT NULL,
			price NUMERIC NOT NULL DEFAULT 0,
			is_partial_price BOOLEAN NOT NULL DEFAULT false,
			project_id INTEGER NOT NULL,
			category_id INTEGER NOT NULL,
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
			labor_id INTEGER NOT NULL,
			deleted_at DATETIME
		);
	`).Error; err != nil {
		t.Fatalf("create schema: %v", err)
	}

	return db
}

func TestLaborRepositoryTenantIsolation(t *testing.T) {
	db := setupLaborTenantDB(t)
	repo := NewRepository(laborTenantGormEngine{client: db})

	tenantA := uuid.New()
	tenantB := uuid.New()
	now := time.Now().UTC()

	if err := db.Exec(`
		INSERT INTO categories (id, name, type_id, created_at, updated_at, deleted_at) VALUES
			(1, 'Siembra', 1, ?, ?, NULL);
		INSERT INTO labors (
			id, tenant_id, name, contractor_name, price, is_partial_price, project_id,
			category_id, created_at, updated_at, deleted_at
		) VALUES
			(1, ?, 'Labor A', 'Contractor A', 10, false, 10, 1, ?, ?, NULL),
			(2, ?, 'Labor B', 'Contractor B', 20, false, 20, 1, ?, ?, NULL),
			(3, ?, 'Labor B archived', 'Contractor B', 30, false, 20, 1, ?, ?, ?),
			(4, ?, 'Labor A archived', 'Contractor A', 40, false, 10, 1, ?, ?, ?);
		INSERT INTO workorders (id, tenant_id, labor_id, deleted_at) VALUES
			(1, ?, 2, NULL);
	`, now, now,
		tenantA.String(), now, now,
		tenantB.String(), now, now,
		tenantB.String(), now, now, now,
		tenantA.String(), now, now, now,
		tenantB.String(),
	).Error; err != nil {
		t.Fatalf("seed labors: %v", err)
	}

	ctxA := laborTenantContext(tenantA)

	list, total, err := repo.ListLabor(ctxA, 1, 50, 10)
	if err != nil {
		t.Fatalf("list labors: %v", err)
	}
	if total != 1 || len(list) != 1 || list[0].ID != 1 {
		t.Fatalf("expected only tenant A labor, total=%d list=%#v", total, list)
	}

	list, total, err = repo.ListLabor(ctxA, 1, 50, 20)
	if err != nil {
		t.Fatalf("list cross-tenant project labors: %v", err)
	}
	if total != 0 || len(list) != 0 {
		t.Fatalf("expected no cross-tenant project labors, total=%d list=%#v", total, list)
	}

	archived, archivedTotal, err := repo.ListArchivedLabors(ctxA, 1, 50, 10)
	if err != nil {
		t.Fatalf("list archived labors: %v", err)
	}
	if archivedTotal != 1 || len(archived) != 1 || archived[0].ID != 4 {
		t.Fatalf("expected only tenant A archived labor, total=%d archived=%#v", archivedTotal, archived)
	}

	if _, err := repo.GetLabor(ctxA, 2); err == nil {
		t.Fatalf("expected get cross-tenant labor to fail")
	}

	exists, err := repo.ExistsLaborByProjectAndName(ctxA, 20, "Labor B")
	if err != nil {
		t.Fatalf("exists cross-tenant labor: %v", err)
	}
	if exists {
		t.Fatalf("expected duplicate check to ignore cross-tenant labor")
	}

	otherExists, err := repo.ExistsOtherLaborByProjectAndName(ctxA, 20, "Labor B", 1)
	if err != nil {
		t.Fatalf("exists other cross-tenant labor: %v", err)
	}
	if otherExists {
		t.Fatalf("expected other duplicate check to ignore cross-tenant labor")
	}

	count, err := repo.GetWorkOrdersByLaborID(ctxA, 2)
	if err != nil {
		t.Fatalf("count cross-tenant work orders by labor: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no cross-tenant work order count, got %d", count)
	}

	if err := repo.UpdateLabor(ctxA, &domain.Labor{
		ID:             2,
		Name:           "cross tenant update",
		ContractorName: "Contractor B",
		Price:          decimal.NewFromInt(99),
		ProjectId:      20,
		CategoryId:     1,
	}); err == nil {
		t.Fatalf("expected update cross-tenant labor to fail")
	}

	var name string
	if err := db.Raw(`SELECT name FROM labors WHERE id = 2`).Scan(&name).Error; err != nil {
		t.Fatalf("read labor 2: %v", err)
	}
	if name != "Labor B" {
		t.Fatalf("cross-tenant update changed labor 2 name to %q", name)
	}

	if err := repo.DeleteLabor(ctxA, 2); err == nil {
		t.Fatalf("expected delete cross-tenant labor to fail")
	}
	var deletedCount int64
	if err := db.Raw(`SELECT COUNT(*) FROM labors WHERE id = 2 AND deleted_at IS NOT NULL`).Scan(&deletedCount).Error; err != nil {
		t.Fatalf("check delete side effect: %v", err)
	}
	if deletedCount != 0 {
		t.Fatalf("cross-tenant delete modified labor 2")
	}

	if err := repo.ArchiveLabor(ctxA, 2); err == nil {
		t.Fatalf("expected archive cross-tenant labor to fail")
	}
	if err := db.Raw(`SELECT COUNT(*) FROM labors WHERE id = 2 AND deleted_at IS NOT NULL`).Scan(&deletedCount).Error; err != nil {
		t.Fatalf("check archive side effect: %v", err)
	}
	if deletedCount != 0 {
		t.Fatalf("cross-tenant archive modified labor 2")
	}

	if err := repo.RestoreLabor(ctxA, 3); err == nil {
		t.Fatalf("expected restore cross-tenant labor to fail")
	}
	if err := db.Raw(`SELECT COUNT(*) FROM labors WHERE id = 3 AND deleted_at IS NOT NULL`).Scan(&deletedCount).Error; err != nil {
		t.Fatalf("check restore side effect: %v", err)
	}
	if deletedCount != 1 {
		t.Fatalf("cross-tenant restore modified labor 3")
	}

	if err := repo.HardDeleteLabor(ctxA, 2); err == nil {
		t.Fatalf("expected hard-delete cross-tenant labor to fail")
	}
	var existsCount int64
	if err := db.Raw(`SELECT COUNT(*) FROM labors WHERE id = 2`).Scan(&existsCount).Error; err != nil {
		t.Fatalf("check hard delete side effect: %v", err)
	}
	if existsCount != 1 {
		t.Fatalf("cross-tenant hard delete removed labor 2")
	}
}
