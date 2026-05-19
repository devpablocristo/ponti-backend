package bparams

import (
	"context"
	"testing"

	contextkeys "github.com/devpablocristo/core/security/go/contextkeys"
	domain "github.com/devpablocristo/ponti-backend/internal/business-parameters/usecases/domain"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type businessParameterTenantGormEngine struct{ client *gorm.DB }

func (e businessParameterTenantGormEngine) Client() *gorm.DB { return e.client }

func businessParameterTenantContext(tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "tenant-user@example.com")
	ctx = context.WithValue(ctx, contextkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, contextkeys.Role, "tenant_admin")
	ctx = context.WithValue(ctx, contextkeys.Scopes, []string{"admin.tenants"})
	return ctx
}

func setupBusinessParameterTenantDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE business_parameters (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT,
			key TEXT NOT NULL,
			value TEXT NOT NULL,
			type TEXT NOT NULL,
			category TEXT NOT NULL,
			description TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);
	`).Error; err != nil {
		t.Fatalf("create schema: %v", err)
	}
	return db
}

func TestBusinessParameterRepositoryTenantIsolation(t *testing.T) {
	db := setupBusinessParameterTenantDB(t)
	repo := NewRepository(businessParameterTenantGormEngine{client: db})
	tenantA := uuid.New()
	tenantB := uuid.New()
	if err := db.Exec(`INSERT INTO business_parameters (id, tenant_id, key, value, type, category, description) VALUES (1, ?, 'kg', '1', 'decimal', 'units', ''), (2, ?, 'tn', '1000', 'decimal', 'units', '')`, tenantA.String(), tenantB.String()).Error; err != nil {
		t.Fatalf("seed business parameters: %v", err)
	}

	ctxA := businessParameterTenantContext(tenantA)
	all, err := repo.ListAll(ctxA)
	if err != nil || len(all) != 1 || all[0].ID != 1 {
		t.Fatalf("expected tenant A parameter only, all=%#v err=%v", all, err)
	}
	if _, err := repo.GetByKey(ctxA, "tn"); err == nil {
		t.Fatalf("expected cross-tenant get by key to fail")
	}
	if err := repo.Update(ctxA, &domain.BusinessParameter{ID: 2, Key: "tn", Value: "999", Type: "decimal", Category: "units"}); err == nil {
		t.Fatalf("expected cross-tenant update to fail")
	}
	if err := repo.Delete(ctxA, 2); err == nil {
		t.Fatalf("expected cross-tenant delete to fail")
	}
}

func TestBusinessParameterRepositoryRequiresTenantInStrictMode(t *testing.T) {
	t.Setenv("TENANT_STRICT_MODE", "true")
	repo := NewRepository(businessParameterTenantGormEngine{client: setupBusinessParameterTenantDB(t)})
	if _, err := repo.ListAll(context.Background()); err == nil {
		t.Fatalf("expected strict list without tenant to fail")
	}
	if _, err := repo.Create(context.Background(), &domain.BusinessParameter{Key: "kg", Value: "1", Type: "decimal", Category: "units"}); err == nil {
		t.Fatalf("expected strict create without tenant to fail")
	}
}
