package invoice

import (
	"context"
	"testing"
	"time"

	contextkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	domain "github.com/devpablocristo/ponti-backend/internal/invoice/usecases/domain"
)

type invoiceTenantGormEngine struct {
	client *gorm.DB
}

func (e invoiceTenantGormEngine) Client() *gorm.DB {
	return e.client
}

func invoiceTenantContext(tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "tenant-user@example.com")
	ctx = context.WithValue(ctx, contextkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, contextkeys.Role, "tenant_admin")
	ctx = context.WithValue(ctx, contextkeys.Scopes, []string{"invoices.read", "invoices.write"})
	return ctx
}

func setupInvoiceTenantDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE workorders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			project_id INTEGER NOT NULL,
			investor_id INTEGER NOT NULL,
			deleted_at DATETIME
		);
		CREATE TABLE workorder_investor_splits (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			workorder_id INTEGER NOT NULL,
			investor_id INTEGER NOT NULL,
			deleted_at DATETIME
		);
		CREATE TABLE invoices (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			work_order_id INTEGER NOT NULL,
			investor_id INTEGER NOT NULL,
			number TEXT NOT NULL,
			company TEXT NOT NULL,
			date DATETIME NOT NULL,
			status TEXT NOT NULL,
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

func TestInvoiceRepositoryTenantIsolation(t *testing.T) {
	db := setupInvoiceTenantDB(t)
	repo := NewRepository(invoiceTenantGormEngine{client: db})

	tenantA := uuid.New()
	tenantB := uuid.New()
	now := time.Now().UTC()

	if err := db.Exec(`
		INSERT INTO workorders (id, tenant_id, project_id, investor_id) VALUES
			(10, ?, 100, 1),
			(20, ?, 200, 2);
		INSERT INTO invoices (id, tenant_id, work_order_id, investor_id, number, company, date, status, created_at, updated_at) VALUES
			(1, ?, 10, 1, 'A-INV', 'Company A', ?, 'draft', ?, ?),
			(2, ?, 20, 2, 'B-INV', 'Company B', ?, 'draft', ?, ?);
	`, tenantA.String(), tenantB.String(), tenantA.String(), now, now, now, tenantB.String(), now, now, now).Error; err != nil {
		t.Fatalf("seed invoices: %v", err)
	}

	ctxA := invoiceTenantContext(tenantA)

	list, total, err := repo.ListByProjectID(ctxA, 100, 1, 50)
	if err != nil {
		t.Fatalf("list tenant invoices: %v", err)
	}
	if total != 1 || len(list) != 1 || list[0].ID != 1 {
		t.Fatalf("expected tenant A invoice, total=%d list=%#v", total, list)
	}
	list, total, err = repo.ListByProjectID(ctxA, 200, 1, 50)
	if err != nil {
		t.Fatalf("list cross-tenant project invoices: %v", err)
	}
	if total != 0 || len(list) != 0 {
		t.Fatalf("expected no cross-tenant invoices, total=%d list=%#v", total, list)
	}

	if _, err := repo.GetByWorkOrderAndInvestor(ctxA, 20, 2); err == nil {
		t.Fatalf("expected get cross-tenant invoice to fail")
	}

	if err := repo.Update(ctxA, &domain.Invoice{
		WorkOrderID: 20,
		InvestorID:  2,
		Number:      "cross tenant update",
		Date:        now,
		Status:      "paid",
	}); err == nil {
		t.Fatalf("expected update cross-tenant invoice to fail")
	}

	var number string
	if err := db.Raw(`SELECT number FROM invoices WHERE id = 2`).Scan(&number).Error; err != nil {
		t.Fatalf("read invoice 2: %v", err)
	}
	if number != "B-INV" {
		t.Fatalf("cross-tenant update changed invoice 2 number to %q", number)
	}

	ok, err := repo.InvestorBelongsToWorkOrder(ctxA, 20, 2)
	if err != nil {
		t.Fatalf("validate cross-tenant investor/workorder: %v", err)
	}
	if ok {
		t.Fatalf("expected investor/workorder validation to fail closed across tenants")
	}

	if err := repo.Delete(ctxA, 20, 2); err == nil {
		t.Fatalf("expected delete cross-tenant invoice to fail")
	}
	var exists int64
	if err := db.Raw(`SELECT COUNT(*) FROM invoices WHERE id = 2`).Scan(&exists).Error; err != nil {
		t.Fatalf("check delete side effect: %v", err)
	}
	if exists != 1 {
		t.Fatalf("cross-tenant delete removed invoice 2")
	}
}

func TestInvoiceRepositoryRequiresTenantInStrictModeForRawValidation(t *testing.T) {
	t.Setenv("TENANT_STRICT_MODE", "true")

	db := setupInvoiceTenantDB(t)
	repo := NewRepository(invoiceTenantGormEngine{client: db})

	if _, err := repo.InvestorBelongsToWorkOrder(context.Background(), 10, 1); err == nil {
		t.Fatalf("expected InvestorBelongsToWorkOrder without tenant to fail in strict mode")
	}
}
