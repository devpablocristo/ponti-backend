package dashboard

import (
	"context"
	"testing"

	contextkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	domain "github.com/devpablocristo/ponti-backend/internal/dashboard/usecases/domain"
)

type dashboardTenantGormEngine struct {
	client *gorm.DB
}

func (e dashboardTenantGormEngine) Client() *gorm.DB {
	return e.client
}

func dashboardTenantContext(tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "tenant-user@example.com")
	ctx = context.WithValue(ctx, contextkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, contextkeys.Role, "tenant_viewer")
	ctx = context.WithValue(ctx, contextkeys.Scopes, []string{"dashboard.read"})
	return ctx
}

func setupDashboardTenantDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE projects (
			id INTEGER PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			customer_id INTEGER,
			campaign_id INTEGER,
			deleted_at DATETIME
		);
		CREATE TABLE fields (
			id INTEGER PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			project_id INTEGER NOT NULL,
			deleted_at DATETIME
		);
	`).Error; err != nil {
		t.Fatalf("create schema: %v", err)
	}

	return db
}

func TestDashboardRepositoryResolveProjectIDsTenantIsolation(t *testing.T) {
	db := setupDashboardTenantDB(t)
	repo := NewRepository(dashboardTenantGormEngine{client: db})

	tenantA := uuid.New()
	tenantB := uuid.New()

	if err := db.Exec(`
		INSERT INTO projects (id, tenant_id, customer_id, campaign_id, deleted_at) VALUES
			(1, ?, 10, 100, NULL),
			(2, ?, 20, 200, NULL);
		INSERT INTO fields (id, tenant_id, project_id, deleted_at) VALUES
			(101, ?, 1, NULL),
			(202, ?, 2, NULL);
	`, tenantA.String(), tenantB.String(), tenantA.String(), tenantB.String()).Error; err != nil {
		t.Fatalf("seed workspace: %v", err)
	}

	ctxA := dashboardTenantContext(tenantA)

	ids, err := repo.resolveProjectIDs(ctxA, domain.DashboardFilter{})
	if err != nil {
		t.Fatalf("resolve all tenant projects: %v", err)
	}
	if len(ids) != 1 || ids[0] != 1 {
		t.Fatalf("expected only tenant A project, got %#v", ids)
	}

	crossProjectID := int64(2)
	if ids, err := repo.resolveProjectIDs(ctxA, domain.DashboardFilter{ProjectID: &crossProjectID}); err == nil || len(ids) != 0 {
		t.Fatalf("expected cross-tenant project to fail closed, ids=%#v err=%v", ids, err)
	}

	crossFieldID := int64(202)
	ids, err = repo.resolveProjectIDs(ctxA, domain.DashboardFilter{FieldID: &crossFieldID})
	if err != nil {
		t.Fatalf("resolve cross-tenant field: %v", err)
	}
	if len(ids) != 1 || ids[0] != 0 {
		t.Fatalf("expected cross-tenant field to resolve to empty sentinel, got %#v", ids)
	}
}
