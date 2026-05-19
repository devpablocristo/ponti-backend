package provider

import (
	"context"
	"testing"

	contextkeys "github.com/devpablocristo/core/security/go/contextkeys"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type providerTenantGormEngine struct {
	client *gorm.DB
}

func (e providerTenantGormEngine) Client() *gorm.DB {
	return e.client
}

func providerTenantContext(tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "tenant-user@example.com")
	ctx = context.WithValue(ctx, contextkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, contextkeys.Role, "tenant_viewer")
	ctx = context.WithValue(ctx, contextkeys.Scopes, []string{"providers.read"})
	return ctx
}

func TestProviderRepositoryTenantIsolation(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE providers (
			id INTEGER PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);
	`).Error; err != nil {
		t.Fatalf("create schema: %v", err)
	}

	tenantA := uuid.New()
	tenantB := uuid.New()
	if err := db.Exec(`
		INSERT INTO providers (id, tenant_id, name, deleted_at) VALUES
			(1, ?, 'Provider A', NULL),
			(2, ?, 'Provider B', NULL);
	`, tenantA.String(), tenantB.String()).Error; err != nil {
		t.Fatalf("seed providers: %v", err)
	}

	repo := NewRepository(providerTenantGormEngine{client: db})
	providers, err := repo.GetProviders(providerTenantContext(tenantA))
	if err != nil {
		t.Fatalf("list providers: %v", err)
	}
	if len(providers) != 1 || providers[0].ID != 1 {
		t.Fatalf("expected only tenant A provider, got %#v", providers)
	}
}
