package report

import (
	"context"
	"testing"

	contextkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/devpablocristo/ponti-backend/internal/report/usecases/domain"
)

type reportTenantGormEngine struct {
	client *gorm.DB
}

func (e reportTenantGormEngine) Client() *gorm.DB {
	return e.client
}

func reportTenantContext(tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "tenant-user@example.com")
	ctx = context.WithValue(ctx, contextkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, contextkeys.Role, "tenant_viewer")
	ctx = context.WithValue(ctx, contextkeys.Scopes, []string{"reports.read"})
	return ctx
}

func setupReportTenantDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE projects (
			id INTEGER PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			customer_id INTEGER,
			campaign_id INTEGER,
			deleted_at DATETIME
		);
			CREATE TABLE customers (
				id INTEGER PRIMARY KEY,
				tenant_id TEXT NOT NULL,
				name TEXT NOT NULL,
				actor_id INTEGER,
				deleted_at DATETIME
			);
		CREATE TABLE campaigns (
			id INTEGER PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
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

func TestReportRepositoryProjectResolutionTenantIsolation(t *testing.T) {
	db := setupReportTenantDB(t)
	repo := NewReportRepository(reportTenantGormEngine{client: db})

	tenantA := uuid.New()
	tenantB := uuid.New()

	if err := db.Exec(`
		INSERT INTO customers (id, tenant_id, name, deleted_at) VALUES
			(1, ?, 'Customer A', NULL),
			(2, ?, 'Customer B', NULL);
		INSERT INTO campaigns (id, tenant_id, name, deleted_at) VALUES
			(1, ?, '2025-2026', NULL),
			(2, ?, '2025-2026', NULL);
		INSERT INTO projects (id, tenant_id, name, customer_id, campaign_id, deleted_at) VALUES
			(1, ?, 'Project A', 1, 1, NULL),
			(2, ?, 'Project B', 2, 2, NULL);
		INSERT INTO fields (id, tenant_id, project_id, deleted_at) VALUES
			(101, ?, 1, NULL),
			(202, ?, 2, NULL);
	`, tenantA.String(), tenantB.String(),
		tenantA.String(), tenantB.String(),
		tenantA.String(), tenantB.String(),
		tenantA.String(), tenantB.String(),
	).Error; err != nil {
		t.Fatalf("seed workspace: %v", err)
	}

	ctxA := reportTenantContext(tenantA)

	info, err := repo.GetProjectInfo(ctxA, 1)
	if err != nil {
		t.Fatalf("get tenant project info: %v", err)
	}
	if info.ProjectID != 1 || info.CustomerName != "Customer A" {
		t.Fatalf("expected tenant A project info, got %#v", info)
	}

	if info, err := repo.GetProjectInfo(ctxA, 2); err == nil {
		t.Fatalf("expected cross-tenant project info to fail closed, info=%#v", info)
	}

	projectIDs, err := repo.getRelatedProjectIDs(ctxA, emptyReportFilter())
	if err != nil {
		t.Fatalf("resolve all tenant report projects: %v", err)
	}
	if len(projectIDs) != 1 || projectIDs[0] != 1 {
		t.Fatalf("expected only tenant A project, got %#v", projectIDs)
	}
}

func emptyReportFilter() domain.ReportFilter {
	return domain.ReportFilter{}
}
