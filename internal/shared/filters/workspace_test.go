package filters

import (
	"context"
	"testing"

	contextkeys "github.com/devpablocristo/core/security/go/contextkeys"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func tenantContext(tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "user-1")
	ctx = context.WithValue(ctx, contextkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, contextkeys.Role, "tenant_viewer")
	ctx = context.WithValue(ctx, contextkeys.Scopes, []string{"projects.read"})
	return ctx
}

func TestResolveProjectIDsScopesByTenant(t *testing.T) {
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

	tenantA := uuid.New()
	tenantB := uuid.New()
	if err := db.Exec(
		`INSERT INTO projects (id, tenant_id, customer_id, campaign_id) VALUES
			(1, ?, 10, 100),
			(2, ?, 20, 200)`,
		tenantA.String(), tenantB.String(),
	).Error; err != nil {
		t.Fatalf("seed projects: %v", err)
	}
	if err := db.Exec(
		`INSERT INTO fields (id, tenant_id, project_id) VALUES
			(101, ?, 1),
			(202, ?, 2)`,
		tenantA.String(), tenantB.String(),
	).Error; err != nil {
		t.Fatalf("seed fields: %v", err)
	}

	projectID := int64(2)
	ids, err := ResolveProjectIDs(tenantContext(tenantA), db, WorkspaceFilter{ProjectID: &projectID})
	if err == nil || len(ids) != 0 {
		t.Fatalf("expected cross-tenant project to fail closed, ids=%#v err=%v", ids, err)
	}

	fieldID := int64(202)
	ids, err = ResolveProjectIDs(tenantContext(tenantA), db, WorkspaceFilter{FieldID: &fieldID})
	if err != nil {
		t.Fatalf("resolve field from other tenant: %v", err)
	}
	if len(ids) != 1 || ids[0] != 0 {
		t.Fatalf("expected cross-tenant field to fail closed, ids=%#v err=%v", ids, err)
	}

	ids, err = ResolveProjectIDs(tenantContext(tenantA), db, WorkspaceFilter{})
	if err != nil {
		t.Fatalf("resolve tenant projects: %v", err)
	}
	if len(ids) != 1 || ids[0] != 1 {
		t.Fatalf("expected tenant A project only, got %#v", ids)
	}
}
