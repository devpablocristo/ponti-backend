package manager

import (
	"context"
	"testing"
	"time"

	contextkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	domain "github.com/devpablocristo/ponti-backend/internal/manager/usecases/domain"
)

type managerGormEngine struct {
	client *gorm.DB
}

func (e managerGormEngine) Client() *gorm.DB {
	return e.client
}

func managerTenantContext(tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "tenant-user@example.com")
	ctx = context.WithValue(ctx, contextkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, contextkeys.Role, "tenant_admin")
	ctx = context.WithValue(ctx, contextkeys.Scopes, []string{"managers.read", "managers.write", "managers.archive"})
	return ctx
}

func setupManagerTenantDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE managers (
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
		CREATE TABLE project_managers (
			tenant_id TEXT NOT NULL,
			project_id INTEGER NOT NULL,
			manager_id INTEGER NOT NULL,
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

func TestManagerRepositoryTenantIsolation(t *testing.T) {
	db := setupManagerTenantDB(t)
	repo := NewRepository(managerGormEngine{client: db})

	tenantA := uuid.New()
	tenantB := uuid.New()
	now := time.Now().UTC()

	if err := db.Exec(`
		INSERT INTO managers (id, tenant_id, name, created_at, updated_at, deleted_at) VALUES
			(1, ?, 'Manager A', ?, ?, NULL),
			(2, ?, 'Manager B', ?, ?, NULL),
			(3, ?, 'Manager B archived', ?, ?, ?)
	`, tenantA.String(), now, now, tenantB.String(), now, now, tenantB.String(), now, now, now).Error; err != nil {
		t.Fatalf("seed managers: %v", err)
	}

	ctxA := managerTenantContext(tenantA)

	list, total, err := repo.ListManagers(ctxA, 1, 50)
	if err != nil {
		t.Fatalf("list managers: %v", err)
	}
	if total != 1 || len(list) != 1 || list[0].ID != 1 {
		t.Fatalf("expected only tenant A manager, total=%d list=%#v", total, list)
	}

	if _, err := repo.GetManager(ctxA, 2); err == nil {
		t.Fatalf("expected get cross-tenant manager to fail")
	}

	if err := repo.UpdateManager(ctxA, &domain.Manager{ID: 2, Name: "cross tenant update"}); err == nil {
		t.Fatalf("expected update cross-tenant manager to fail")
	}

	var name string
	if err := db.Raw(`SELECT name FROM managers WHERE id = 2`).Scan(&name).Error; err != nil {
		t.Fatalf("read manager 2: %v", err)
	}
	if name != "Manager B" {
		t.Fatalf("cross-tenant update changed manager 2 name to %q", name)
	}

	if err := repo.ArchiveManager(ctxA, 2); err == nil {
		t.Fatalf("expected archive cross-tenant manager to fail")
	}
	var deletedCount int64
	if err := db.Raw(`SELECT COUNT(*) FROM managers WHERE id = 2 AND deleted_at IS NOT NULL`).Scan(&deletedCount).Error; err != nil {
		t.Fatalf("check archive side effect: %v", err)
	}
	if deletedCount != 0 {
		t.Fatalf("cross-tenant archive modified manager 2")
	}

	if err := repo.RestoreManager(ctxA, 3); err == nil {
		t.Fatalf("expected restore cross-tenant manager to fail")
	}
	if err := db.Raw(`SELECT COUNT(*) FROM managers WHERE id = 3 AND deleted_at IS NOT NULL`).Scan(&deletedCount).Error; err != nil {
		t.Fatalf("check restore side effect: %v", err)
	}
	if deletedCount != 1 {
		t.Fatalf("cross-tenant restore modified manager 3")
	}

	if err := repo.HardDeleteManager(ctxA, 2); err == nil {
		t.Fatalf("expected hard-delete cross-tenant manager to fail")
	}
	var exists int64
	if err := db.Raw(`SELECT COUNT(*) FROM managers WHERE id = 2`).Scan(&exists).Error; err != nil {
		t.Fatalf("check hard delete side effect: %v", err)
	}
	if exists != 1 {
		t.Fatalf("cross-tenant hard delete removed manager 2")
	}
}
