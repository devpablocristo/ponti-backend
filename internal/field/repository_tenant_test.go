package field

import (
	"context"
	"testing"
	"time"

	contextkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	domain "github.com/devpablocristo/ponti-backend/internal/field/usecases/domain"
	leasetypedom "github.com/devpablocristo/ponti-backend/internal/lease-type/usecases/domain"
)

type fieldGormEngine struct {
	client *gorm.DB
}

func (e fieldGormEngine) Client() *gorm.DB {
	return e.client
}

func fieldTenantContext(tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "tenant-user@example.com")
	ctx = context.WithValue(ctx, contextkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, contextkeys.Role, "tenant_admin")
	ctx = context.WithValue(ctx, contextkeys.Scopes, []string{"fields.read", "fields.write", "fields.archive"})
	return ctx
}

func setupFieldTenantDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE fields (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			project_id INTEGER NOT NULL,
			lease_type_id INTEGER NOT NULL,
			lease_type_percent NUMERIC,
			lease_type_value NUMERIC,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		);
		CREATE TABLE lots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			field_id INTEGER NOT NULL,
			hectares NUMERIC DEFAULT 0,
			previous_crop_id INTEGER,
			current_crop_id INTEGER,
			season TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		);
	`).Error; err != nil {
		t.Fatalf("create schema: %v", err)
	}

	return db
}

func TestFieldRepositoryTenantIsolation(t *testing.T) {
	db := setupFieldTenantDB(t)
	repo := NewRepository(fieldGormEngine{client: db})

	tenantA := uuid.New()
	tenantB := uuid.New()
	now := time.Now().UTC()

	if err := db.Exec(`
		INSERT INTO fields (id, tenant_id, name, project_id, lease_type_id, created_at, updated_at, deleted_at) VALUES
			(1, ?, 'Field A', 10, 1, ?, ?, NULL),
			(2, ?, 'Field B', 20, 1, ?, ?, NULL),
			(3, ?, 'Field B archived', 20, 1, ?, ?, ?)
	`, tenantA.String(), now, now, tenantB.String(), now, now, tenantB.String(), now, now, now).Error; err != nil {
		t.Fatalf("seed fields: %v", err)
	}

	ctxA := fieldTenantContext(tenantA)

	list, total, err := repo.ListFields(ctxA, 1, 50)
	if err != nil {
		t.Fatalf("list fields: %v", err)
	}
	if total != 1 || len(list) != 1 || list[0].ID != 1 {
		t.Fatalf("expected only tenant A field, total=%d list=%#v", total, list)
	}

	if _, err := repo.GetField(ctxA, 2); err == nil {
		t.Fatalf("expected get cross-tenant field to fail")
	}

	if err := repo.UpdateField(ctxA, &domain.Field{
		ID:        2,
		Name:      "cross tenant update",
		LeaseType: &leasetypedom.LeaseType{ID: 2},
	}); err == nil {
		t.Fatalf("expected update cross-tenant field to fail")
	}

	var name string
	if err := db.Raw(`SELECT name FROM fields WHERE id = 2`).Scan(&name).Error; err != nil {
		t.Fatalf("read field 2: %v", err)
	}
	if name != "Field B" {
		t.Fatalf("cross-tenant update changed field 2 name to %q", name)
	}

	if err := repo.ArchiveField(ctxA, 2); err == nil {
		t.Fatalf("expected archive cross-tenant field to fail")
	}
	var deletedCount int64
	if err := db.Raw(`SELECT COUNT(*) FROM fields WHERE id = 2 AND deleted_at IS NOT NULL`).Scan(&deletedCount).Error; err != nil {
		t.Fatalf("check archive side effect: %v", err)
	}
	if deletedCount != 0 {
		t.Fatalf("cross-tenant archive modified field 2")
	}

	if err := repo.RestoreField(ctxA, 3); err == nil {
		t.Fatalf("expected restore cross-tenant field to fail")
	}
	if err := db.Raw(`SELECT COUNT(*) FROM fields WHERE id = 3 AND deleted_at IS NOT NULL`).Scan(&deletedCount).Error; err != nil {
		t.Fatalf("check restore side effect: %v", err)
	}
	if deletedCount != 1 {
		t.Fatalf("cross-tenant restore modified field 3")
	}

	if err := repo.HardDeleteField(ctxA, 2); err == nil {
		t.Fatalf("expected hard-delete cross-tenant field to fail")
	}
	var exists int64
	if err := db.Raw(`SELECT COUNT(*) FROM fields WHERE id = 2`).Scan(&exists).Error; err != nil {
		t.Fatalf("check hard delete side effect: %v", err)
	}
	if exists != 1 {
		t.Fatalf("cross-tenant hard delete removed field 2")
	}
}
