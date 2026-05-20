package classtype

import (
	"context"
	"testing"

	contextkeys "github.com/devpablocristo/core/security/go/contextkeys"
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
	if err := db.Exec(`
		CREATE TABLE types (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT,
			name TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);
	`).Error; err != nil {
		t.Fatalf("create schema: %v", err)
	}
	return db
}

func TestClassTypeRepositoryTenantIsolation(t *testing.T) {
	db := setupClassTypeTenantDB(t)
	repo := NewRepository(classTypeTenantGormEngine{client: db})
	tenantA := uuid.New()
	tenantB := uuid.New()
	if err := db.Exec(`INSERT INTO types (id, tenant_id, name) VALUES (1, ?, 'Agroquimico'), (2, ?, 'Semilla')`, tenantA.String(), tenantB.String()).Error; err != nil {
		t.Fatalf("seed class types: %v", err)
	}

	ctxA := classTypeTenantContext(tenantA)
	list, total, err := repo.ListClassTypes(ctxA, 1, 50)
	if err != nil || total != 1 || len(list) != 1 || list[0].ID != 1 {
		t.Fatalf("expected tenant A type only, total=%d list=%#v err=%v", total, list, err)
	}
	if _, err := repo.GetClassType(ctxA, 2); err == nil {
		t.Fatalf("expected cross-tenant get to fail")
	}
	if err := repo.UpdateClassType(ctxA, &domain.ClassType{ID: 2, Name: "cross"}); err == nil {
		t.Fatalf("expected cross-tenant update to fail")
	}
	if err := repo.HardDeleteClassType(ctxA, 2); err == nil {
		t.Fatalf("expected cross-tenant hard delete to fail")
	}
}

func TestClassTypeRepositoryRequiresTenantInStrictMode(t *testing.T) {
	t.Setenv("TENANT_STRICT_MODE", "true")
	repo := NewRepository(classTypeTenantGormEngine{client: setupClassTypeTenantDB(t)})
	if _, _, err := repo.ListClassTypes(context.Background(), 1, 50); err == nil {
		t.Fatalf("expected strict list without tenant to fail")
	}
	if _, err := repo.CreateClassType(context.Background(), &domain.ClassType{Name: "Agroquimico"}); err == nil {
		t.Fatalf("expected strict create without tenant to fail")
	}
}
