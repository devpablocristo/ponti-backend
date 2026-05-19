package crop

import (
	"context"
	"testing"

	contextkeys "github.com/devpablocristo/core/security/go/contextkeys"
	domain "github.com/devpablocristo/ponti-backend/internal/crop/usecases/domain"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type cropTenantGormEngine struct{ client *gorm.DB }

func (e cropTenantGormEngine) Client() *gorm.DB { return e.client }

func cropTenantContext(tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "tenant-user@example.com")
	ctx = context.WithValue(ctx, contextkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, contextkeys.Role, "tenant_admin")
	ctx = context.WithValue(ctx, contextkeys.Scopes, []string{"crops.read", "crops.write"})
	return ctx
}

func setupCropTenantDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE crops (
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

func TestCropRepositoryTenantIsolation(t *testing.T) {
	db := setupCropTenantDB(t)
	repo := NewRepository(cropTenantGormEngine{client: db})
	tenantA := uuid.New()
	tenantB := uuid.New()
	if err := db.Exec(`INSERT INTO crops (id, tenant_id, name) VALUES (1, ?, 'Soja'), (2, ?, 'Maiz')`, tenantA.String(), tenantB.String()).Error; err != nil {
		t.Fatalf("seed crops: %v", err)
	}

	ctxA := cropTenantContext(tenantA)
	list, total, err := repo.ListCrops(ctxA, 1, 50)
	if err != nil || total != 1 || len(list) != 1 || list[0].ID != 1 {
		t.Fatalf("expected tenant A crop only, total=%d list=%#v err=%v", total, list, err)
	}
	if _, err := repo.GetCrop(ctxA, 2); err == nil {
		t.Fatalf("expected cross-tenant get to fail")
	}
	if err := repo.UpdateCrop(ctxA, &domain.Crop{ID: 2, Name: "cross"}); err == nil {
		t.Fatalf("expected cross-tenant update to fail")
	}
	if err := repo.DeleteCrop(ctxA, 2); err == nil {
		t.Fatalf("expected cross-tenant delete to fail")
	}
}

func TestCropRepositoryRequiresTenantInStrictMode(t *testing.T) {
	t.Setenv("TENANT_STRICT_MODE", "true")
	repo := NewRepository(cropTenantGormEngine{client: setupCropTenantDB(t)})
	if _, _, err := repo.ListCrops(context.Background(), 1, 50); err == nil {
		t.Fatalf("expected strict list without tenant to fail")
	}
	if _, err := repo.CreateCrop(context.Background(), &domain.Crop{Name: "Soja"}); err == nil {
		t.Fatalf("expected strict create without tenant to fail")
	}
}
