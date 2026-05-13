package campaign

import (
	"context"
	"testing"
	"time"

	contextkeys "github.com/devpablocristo/core/security/go/contextkeys"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	domain "github.com/devpablocristo/ponti-backend/internal/campaign/usecases/domain"
)

type campaignGormEngine struct {
	client *gorm.DB
}

func (e campaignGormEngine) Client() *gorm.DB {
	return e.client
}

func campaignTenantContext(tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "tenant-user@example.com")
	ctx = context.WithValue(ctx, contextkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, contextkeys.Role, "tenant_admin")
	ctx = context.WithValue(ctx, contextkeys.Scopes, []string{"campaigns.read", "campaigns.write", "campaigns.archive"})
	return ctx
}

func setupCampaignTenantDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE campaigns (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		);
		CREATE TABLE projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			customer_id INTEGER NOT NULL,
			campaign_id INTEGER NOT NULL,
			admin_cost NUMERIC DEFAULT 0,
			planned_cost NUMERIC DEFAULT 0,
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

func TestCampaignRepositoryTenantIsolation(t *testing.T) {
	db := setupCampaignTenantDB(t)
	repo := NewRepository(campaignGormEngine{client: db})

	tenantA := uuid.New()
	tenantB := uuid.New()
	now := time.Now().UTC()

	if err := db.Exec(`
		INSERT INTO campaigns (id, tenant_id, name, created_at, updated_at, deleted_at) VALUES
			(1, ?, 'A 2025', ?, ?, NULL),
			(2, ?, 'B 2025', ?, ?, NULL),
			(3, ?, 'B archived', ?, ?, ?)
	`, tenantA.String(), now, now, tenantB.String(), now, now, tenantB.String(), now, now, now).Error; err != nil {
		t.Fatalf("seed campaigns: %v", err)
	}
	if err := db.Exec(`
		INSERT INTO projects (id, tenant_id, name, customer_id, campaign_id, created_at, updated_at) VALUES
			(10, ?, 'Same project', 100, 1, ?, ?),
			(20, ?, 'Same project', 100, 2, ?, ?)
	`, tenantA.String(), now, now, tenantB.String(), now, now).Error; err != nil {
		t.Fatalf("seed projects: %v", err)
	}

	ctxA := campaignTenantContext(tenantA)

	list, err := repo.ListCampaigns(ctxA, 0, "")
	if err != nil {
		t.Fatalf("list campaigns: %v", err)
	}
	if len(list) != 1 || list[0].ID != 1 {
		t.Fatalf("expected only tenant A campaign, got %#v", list)
	}

	filtered, err := repo.ListCampaigns(ctxA, 100, "Same project")
	if err != nil {
		t.Fatalf("list filtered campaigns: %v", err)
	}
	if len(filtered) != 1 || filtered[0].ID != 1 || filtered[0].ProjectID != 10 {
		t.Fatalf("expected only tenant A filtered campaign/project, got %#v", filtered)
	}

	if _, err := repo.GetCampaign(ctxA, 2); err == nil {
		t.Fatalf("expected get cross-tenant campaign to fail")
	}

	if err := repo.UpdateCampaign(ctxA, &domain.Campaign{ID: 2, Name: "cross tenant update"}); err == nil {
		t.Fatalf("expected update cross-tenant campaign to fail")
	}

	var name string
	if err := db.Raw(`SELECT name FROM campaigns WHERE id = 2`).Scan(&name).Error; err != nil {
		t.Fatalf("read campaign 2: %v", err)
	}
	if name != "B 2025" {
		t.Fatalf("cross-tenant update changed campaign 2 name to %q", name)
	}

	if err := repo.ArchiveCampaign(ctxA, 2); err == nil {
		t.Fatalf("expected archive cross-tenant campaign to fail")
	}
	var deletedCount int64
	if err := db.Raw(`SELECT COUNT(*) FROM campaigns WHERE id = 2 AND deleted_at IS NOT NULL`).Scan(&deletedCount).Error; err != nil {
		t.Fatalf("check archive side effect: %v", err)
	}
	if deletedCount != 0 {
		t.Fatalf("cross-tenant archive modified campaign 2")
	}

	if err := repo.RestoreCampaign(ctxA, 3); err == nil {
		t.Fatalf("expected restore cross-tenant campaign to fail")
	}
	if err := db.Raw(`SELECT COUNT(*) FROM campaigns WHERE id = 3 AND deleted_at IS NOT NULL`).Scan(&deletedCount).Error; err != nil {
		t.Fatalf("check restore side effect: %v", err)
	}
	if deletedCount != 1 {
		t.Fatalf("cross-tenant restore modified campaign 3")
	}

	if err := repo.HardDeleteCampaign(ctxA, 2); err == nil {
		t.Fatalf("expected hard-delete cross-tenant campaign to fail")
	}
	var exists int64
	if err := db.Raw(`SELECT COUNT(*) FROM campaigns WHERE id = 2`).Scan(&exists).Error; err != nil {
		t.Fatalf("check hard delete side effect: %v", err)
	}
	if exists != 1 {
		t.Fatalf("cross-tenant hard delete removed campaign 2")
	}
}
