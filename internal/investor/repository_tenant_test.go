package investor

import (
	"context"
	"testing"
	"time"

	contextkeys "github.com/devpablocristo/core/security/go/contextkeys"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	domain "github.com/devpablocristo/ponti-backend/internal/investor/usecases/domain"
)

type investorGormEngine struct {
	client *gorm.DB
}

func (e investorGormEngine) Client() *gorm.DB {
	return e.client
}

func investorTenantContext(tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "tenant-user@example.com")
	ctx = context.WithValue(ctx, contextkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, contextkeys.Role, "tenant_admin")
	ctx = context.WithValue(ctx, contextkeys.Scopes, []string{"investors.read", "investors.write", "investors.archive"})
	return ctx
}

func setupInvestorTenantDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE investors (
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
		CREATE TABLE project_investors (
			tenant_id TEXT NOT NULL,
			project_id INTEGER NOT NULL,
			investor_id INTEGER NOT NULL,
			percentage INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);
		CREATE TABLE field_investors (
			tenant_id TEXT NOT NULL,
			field_id INTEGER NOT NULL,
			investor_id INTEGER NOT NULL,
			percentage INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);
		CREATE TABLE admin_cost_investors (
			tenant_id TEXT NOT NULL,
			project_id INTEGER NOT NULL,
			investor_id INTEGER NOT NULL,
			percentage INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);
		CREATE TABLE legacy_actor_map (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			source_table TEXT NOT NULL,
			source_id INTEGER NOT NULL,
			actor_id INTEGER NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);
	`).Error; err != nil {
		t.Fatalf("create schema: %v", err)
	}

	return db
}

func TestInvestorRepositoryTenantIsolation(t *testing.T) {
	db := setupInvestorTenantDB(t)
	repo := NewRepository(investorGormEngine{client: db})

	tenantA := uuid.New()
	tenantB := uuid.New()
	now := time.Now().UTC()

	if err := db.Exec(`
		INSERT INTO investors (id, tenant_id, name, created_at, updated_at, deleted_at) VALUES
			(1, ?, 'Investor A', ?, ?, NULL),
			(2, ?, 'Investor B', ?, ?, NULL),
			(3, ?, 'Investor B archived', ?, ?, ?)
	`, tenantA.String(), now, now, tenantB.String(), now, now, tenantB.String(), now, now, now).Error; err != nil {
		t.Fatalf("seed investors: %v", err)
	}

	ctxA := investorTenantContext(tenantA)

	list, total, err := repo.ListInvestors(ctxA, 1, 50)
	if err != nil {
		t.Fatalf("list investors: %v", err)
	}
	if total != 1 || len(list) != 1 || list[0].ID != 1 {
		t.Fatalf("expected only tenant A investor, total=%d list=%#v", total, list)
	}

	if _, err := repo.GetInvestor(ctxA, 2); err == nil {
		t.Fatalf("expected get cross-tenant investor to fail")
	}

	if err := repo.UpdateInvestor(ctxA, &domain.Investor{ID: 2, Name: "cross tenant update"}); err == nil {
		t.Fatalf("expected update cross-tenant investor to fail")
	}

	var name string
	if err := db.Raw(`SELECT name FROM investors WHERE id = 2`).Scan(&name).Error; err != nil {
		t.Fatalf("read investor 2: %v", err)
	}
	if name != "Investor B" {
		t.Fatalf("cross-tenant update changed investor 2 name to %q", name)
	}

	if err := repo.ArchiveInvestor(ctxA, 2); err == nil {
		t.Fatalf("expected archive cross-tenant investor to fail")
	}
	var deletedCount int64
	if err := db.Raw(`SELECT COUNT(*) FROM investors WHERE id = 2 AND deleted_at IS NOT NULL`).Scan(&deletedCount).Error; err != nil {
		t.Fatalf("check archive side effect: %v", err)
	}
	if deletedCount != 0 {
		t.Fatalf("cross-tenant archive modified investor 2")
	}

	if err := repo.RestoreInvestor(ctxA, 3); err == nil {
		t.Fatalf("expected restore cross-tenant investor to fail")
	}
	if err := db.Raw(`SELECT COUNT(*) FROM investors WHERE id = 3 AND deleted_at IS NOT NULL`).Scan(&deletedCount).Error; err != nil {
		t.Fatalf("check restore side effect: %v", err)
	}
	if deletedCount != 1 {
		t.Fatalf("cross-tenant restore modified investor 3")
	}

	if err := repo.HardDeleteInvestor(ctxA, 2); err == nil {
		t.Fatalf("expected hard-delete cross-tenant investor to fail")
	}
	var exists int64
	if err := db.Raw(`SELECT COUNT(*) FROM investors WHERE id = 2`).Scan(&exists).Error; err != nil {
		t.Fatalf("check hard delete side effect: %v", err)
	}
	if exists != 1 {
		t.Fatalf("cross-tenant hard delete removed investor 2")
	}
}
