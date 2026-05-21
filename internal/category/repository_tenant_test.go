package category

import (
	"context"
	"testing"

	contextkeys "github.com/devpablocristo/core/security/go/contextkeys"
	domain "github.com/devpablocristo/ponti-backend/internal/category/usecases/domain"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type categoryTenantGormEngine struct{ client *gorm.DB }

func (e categoryTenantGormEngine) Client() *gorm.DB { return e.client }

func categoryTenantContext(tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "tenant-user@example.com")
	ctx = context.WithValue(ctx, contextkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, contextkeys.Role, "tenant_admin")
	ctx = context.WithValue(ctx, contextkeys.Scopes, []string{"supplies.read", "supplies.write"})
	return ctx
}

func setupCategoryTenantDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT,
			name TEXT NOT NULL,
			type_id INTEGER NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);
	`).Error; err != nil {
		t.Fatalf("create schema: %v", err)
	}
	return db
}

func TestCategoryRepositoryTenantIsolation(t *testing.T) {
	db := setupCategoryTenantDB(t)
	repo := NewRepository(categoryTenantGormEngine{client: db})
	tenantA := uuid.New()
	tenantB := uuid.New()
	if err := db.Exec(`INSERT INTO categories (id, tenant_id, name, type_id) VALUES (1, ?, 'Herbicida', 1), (2, ?, 'Fertilizante', 1)`, tenantA.String(), tenantB.String()).Error; err != nil {
		t.Fatalf("seed categories: %v", err)
	}

	ctxA := categoryTenantContext(tenantA)
	list, total, err := repo.ListCategories(ctxA, domain.ListFilters{}, 1, 50)
	if err != nil || total != 1 || len(list) != 1 || list[0].ID != 1 {
		t.Fatalf("expected tenant A category only, total=%d list=%#v err=%v", total, list, err)
	}
	if _, err := repo.GetCategory(ctxA, 2); err == nil {
		t.Fatalf("expected cross-tenant get to fail")
	}
	if err := repo.UpdateCategory(ctxA, &domain.Category{ID: 2, Name: "cross", TypeID: 1}); err == nil {
		t.Fatalf("expected cross-tenant update to fail")
	}
	if err := repo.HardDeleteCategory(ctxA, 2); err == nil {
		t.Fatalf("expected cross-tenant hard delete to fail")
	}
}

func TestCategoryRepositoryRequiresTenantInStrictMode(t *testing.T) {
	t.Setenv("TENANT_STRICT_MODE", "true")
	repo := NewRepository(categoryTenantGormEngine{client: setupCategoryTenantDB(t)})
	if _, _, err := repo.ListCategories(context.Background(), domain.ListFilters{}, 1, 50); err == nil {
		t.Fatalf("expected strict list without tenant to fail")
	}
	if _, err := repo.CreateCategory(context.Background(), &domain.Category{Name: "Herbicida", TypeID: 1}); err == nil {
		t.Fatalf("expected strict create without tenant to fail")
	}
}
