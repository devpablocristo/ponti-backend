package customer

import (
	"context"
	"testing"
	"time"

	contextkeys "github.com/devpablocristo/core/security/go/contextkeys"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	domain "github.com/devpablocristo/ponti-backend/internal/customer/usecases/domain"
)

type customerTenantGormEngine struct {
	client *gorm.DB
}

func (e customerTenantGormEngine) Client() *gorm.DB {
	return e.client
}

func customerTenantContext(tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "tenant-user@example.com")
	ctx = context.WithValue(ctx, contextkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, contextkeys.Role, "tenant_admin")
	ctx = context.WithValue(ctx, contextkeys.Scopes, []string{"customers.read", "customers.write", "customers.archive"})
	return ctx
}

func setupCustomerTenantDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE customers (
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

func TestCustomerRepositoryTenantIsolation(t *testing.T) {
	db := setupCustomerTenantDB(t)
	repo := NewRepository(customerTenantGormEngine{client: db})

	tenantA := uuid.New()
	tenantB := uuid.New()
	now := time.Now().UTC()

	if err := db.Exec(`
		INSERT INTO customers (id, tenant_id, name, created_at, updated_at, deleted_at) VALUES
			(1, ?, 'Customer A', ?, ?, NULL),
			(2, ?, 'Customer B', ?, ?, NULL),
			(3, ?, 'Customer B archived', ?, ?, ?)
	`, tenantA.String(), now, now, tenantB.String(), now, now, tenantB.String(), now, now, now).Error; err != nil {
		t.Fatalf("seed customers: %v", err)
	}

	ctxA := customerTenantContext(tenantA)

	list, total, err := repo.ListCustomers(ctxA, 1, 50)
	if err != nil {
		t.Fatalf("list customers: %v", err)
	}
	if total != 1 || len(list) != 1 || list[0].ID != 1 {
		t.Fatalf("expected only tenant A customer, total=%d list=%#v", total, list)
	}

	if _, err := repo.GetCustomer(ctxA, 2); err == nil {
		t.Fatalf("expected get cross-tenant customer to fail")
	}

	if err := repo.UpdateCustomer(ctxA, &domain.Customer{ID: 2, Name: "cross tenant update"}); err == nil {
		t.Fatalf("expected update cross-tenant customer to fail")
	}

	var name string
	if err := db.Raw(`SELECT name FROM customers WHERE id = 2`).Scan(&name).Error; err != nil {
		t.Fatalf("read customer 2: %v", err)
	}
	if name != "Customer B" {
		t.Fatalf("cross-tenant update changed customer 2 name to %q", name)
	}

	if err := repo.ArchiveCustomer(ctxA, 2); err == nil {
		t.Fatalf("expected archive cross-tenant customer to fail")
	}
	var deletedCount int64
	if err := db.Raw(`SELECT COUNT(*) FROM customers WHERE id = 2 AND deleted_at IS NOT NULL`).Scan(&deletedCount).Error; err != nil {
		t.Fatalf("check archive side effect: %v", err)
	}
	if deletedCount != 0 {
		t.Fatalf("cross-tenant archive modified customer 2")
	}

	if err := repo.RestoreCustomer(ctxA, 3); err == nil {
		t.Fatalf("expected restore cross-tenant customer to fail")
	}
	if err := db.Raw(`SELECT COUNT(*) FROM customers WHERE id = 3 AND deleted_at IS NOT NULL`).Scan(&deletedCount).Error; err != nil {
		t.Fatalf("check restore side effect: %v", err)
	}
	if deletedCount != 1 {
		t.Fatalf("cross-tenant restore modified customer 3")
	}

	if err := repo.HardDeleteCustomer(ctxA, 2); err == nil {
		t.Fatalf("expected hard-delete cross-tenant customer to fail")
	}
	var exists int64
	if err := db.Raw(`SELECT COUNT(*) FROM customers WHERE id = 2`).Scan(&exists).Error; err != nil {
		t.Fatalf("check hard delete side effect: %v", err)
	}
	if exists != 1 {
		t.Fatalf("cross-tenant hard delete removed customer 2")
	}
}

func TestCustomerRepositoryRequiresTenantInStrictMode(t *testing.T) {
	t.Setenv("TENANT_STRICT_MODE", "true")

	db := setupCustomerTenantDB(t)
	repo := NewRepository(customerTenantGormEngine{client: db})

	if _, _, err := repo.ListCustomers(context.Background(), 1, 50); err == nil {
		t.Fatalf("expected strict mode to reject customer list without tenant context")
	}
}
