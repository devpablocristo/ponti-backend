package classtype

import (
	"context"
	"testing"

	contextkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	domain "github.com/devpablocristo/ponti-backend/internal/class-type/usecases/domain"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type classTypeTenantGormEngine struct{ client *gorm.DB }

func (e classTypeTenantGormEngine) Client() *gorm.DB { return e.client }

func classTypeTenantContext(tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "tenant-user@example.com")
	ctx = context.WithValue(ctx, contextkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, contextkeys.Role, "tenant_admin")
	ctx = context.WithValue(ctx, contextkeys.Scopes, []string{"supplies.read", "supplies.write"})
	return ctx
}

func setupClassTypeTenantDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	// `types` is a GLOBAL catalog — no `tenant_id` column. The schema
	// mirrors production (see migrations_v4 + the diagnosed 500 on
	// `/api/v1/types` when the Go model still pretended to be tenanted).
	if err := db.Exec(`
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
	`).Error; err != nil {
		t.Fatalf("create schema: %v", err)
	}
	return db
}

// ClassType is a global catalog: every tenant sees every type. The repo
// must list/get/update/delete without filtering by tenant. This test seeds
// rows that would have been "another tenant's" under the old (broken)
// shape and asserts they ARE visible to any caller.
func TestClassTypeRepositoryIsGlobalCatalog(t *testing.T) {
	db := setupClassTypeTenantDB(t)
	repo := NewRepository(classTypeTenantGormEngine{client: db})
	tenantA := uuid.New()

	if err := db.Exec(`INSERT INTO types (id, name) VALUES (1, 'Agroquimico'), (2, 'Semilla')`).Error; err != nil {
		t.Fatalf("seed class types: %v", err)
	}

	ctxA := classTypeTenantContext(tenantA)
	list, total, err := repo.ListClassTypes(ctxA, 1, 50)
	if err != nil || total != 2 || len(list) != 2 {
		t.Fatalf("expected both global types visible to tenant A, total=%d list=%#v err=%v", total, list, err)
	}

	// Read works for any id regardless of caller's tenant.
	got, err := repo.GetClassType(ctxA, 2)
	if err != nil || got == nil || got.ID != 2 {
		t.Fatalf("expected GetClassType(2) to succeed for tenant A, got=%v err=%v", got, err)
	}
}

// Strict-mode previously required a tenant for every list/create call.
// Now that the catalog is global, strict mode does NOT apply — list and
// create succeed without tenant context. This guards against regressing
// to the old shape (Go model + repo pretending multi-tenant on a DB table
// that has no tenant column).
func TestClassTypeRepositoryDoesNotRequireTenant(t *testing.T) {
	t.Setenv("TENANT_STRICT_MODE", "true")
	repo := NewRepository(classTypeTenantGormEngine{client: setupClassTypeTenantDB(t)})

	if _, _, err := repo.ListClassTypes(context.Background(), 1, 50); err != nil {
		t.Fatalf("expected list without tenant to succeed for global catalog, got %v", err)
	}
	if _, err := repo.CreateClassType(context.Background(), &domain.ClassType{Name: "Agroquimico"}); err != nil {
		t.Fatalf("expected create without tenant to succeed for global catalog, got %v", err)
	}
}
