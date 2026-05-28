package commercialization

import (
	"context"
	"testing"

	contextkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	domain "github.com/devpablocristo/ponti-backend/internal/commercialization/usecases/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type commercializationTenantGormEngine struct {
	client *gorm.DB
}

func (e commercializationTenantGormEngine) Client() *gorm.DB {
	return e.client
}

func commercializationTenantContext(tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "tenant-user@example.com")
	ctx = context.WithValue(ctx, contextkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, contextkeys.Role, "tenant_admin")
	ctx = context.WithValue(ctx, contextkeys.Scopes, []string{"projects.read", "projects.write"})
	return ctx
}

func setupCommercializationTenantDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE projects (
			id INTEGER PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			deleted_at DATETIME
		);
		CREATE TABLE crop_commercializations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT,
			project_id INTEGER NOT NULL,
			crop_id INTEGER NOT NULL,
			board_price NUMERIC NOT NULL,
			freight_cost NUMERIC NOT NULL,
			commercial_cost NUMERIC NOT NULL,
			net_price NUMERIC NOT NULL,
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

func TestCommercializationRepositoryProjectTenantIsolation(t *testing.T) {
	db := setupCommercializationTenantDB(t)
	repo := NewRepository(commercializationTenantGormEngine{client: db})

	tenantA := uuid.New()
	tenantB := uuid.New()
	if err := db.Exec(`
		INSERT INTO projects (id, tenant_id) VALUES (10, ?), (20, ?);
		INSERT INTO crop_commercializations
			(id, tenant_id, project_id, crop_id, board_price, freight_cost, commercial_cost, net_price)
		VALUES
			(1, ?, 10, 1, 100, 10, 5, 85),
			(2, ?, 20, 1, 200, 20, 5, 170);
	`, tenantA.String(), tenantB.String(), tenantA.String(), tenantB.String()).Error; err != nil {
		t.Fatalf("seed commercializations: %v", err)
	}

	ctxA := commercializationTenantContext(tenantA)
	if values, err := repo.ListByProject(ctxA, 10); err != nil || len(values) != 1 || values[0].ID != 1 {
		t.Fatalf("expected own project commercializations, values=%#v err=%v", values, err)
	}
	if _, err := repo.ListByProject(ctxA, 20); err == nil {
		t.Fatalf("expected cross-tenant list to fail")
	}
	if err := repo.CreateBulk(ctxA, []domain.CropCommercialization{{
		ProjectID:      20,
		CropID:         1,
		BoardPrice:     decimal.NewFromInt(100),
		FreightCost:    decimal.NewFromInt(10),
		CommercialCost: decimal.NewFromInt(5),
		NetPrice:       decimal.NewFromInt(85),
	}}); err == nil {
		t.Fatalf("expected cross-tenant create to fail")
	}
	if err := repo.Update(ctxA, &domain.CropCommercialization{
		ID:             2,
		ProjectID:      20,
		CropID:         1,
		BoardPrice:     decimal.NewFromInt(999),
		FreightCost:    decimal.NewFromInt(1),
		CommercialCost: decimal.NewFromInt(1),
		NetPrice:       decimal.NewFromInt(988),
	}); err == nil {
		t.Fatalf("expected cross-tenant update to fail")
	}

	var net string
	if err := db.Raw(`SELECT net_price FROM crop_commercializations WHERE id = 2`).Scan(&net).Error; err != nil {
		t.Fatalf("read cross-tenant row: %v", err)
	}
	if net != "170" && net != "170.0" && net != "170.00" {
		t.Fatalf("cross-tenant update changed net_price to %q", net)
	}
}

func TestCommercializationRepositoryRequiresTenantInStrictMode(t *testing.T) {
	t.Setenv("TENANT_STRICT_MODE", "true")

	db := setupCommercializationTenantDB(t)
	repo := NewRepository(commercializationTenantGormEngine{client: db})

	if _, err := repo.ListByProject(context.Background(), 10); err == nil {
		t.Fatalf("expected strict list without tenant to fail")
	}
}
