package workorder

import (
	"context"
	"testing"
	"time"

	contextkeys "github.com/devpablocristo/core/security/go/contextkeys"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	domain "github.com/devpablocristo/ponti-backend/internal/work-order/usecases/domain"
)

type workOrderTenantGormEngine struct {
	client *gorm.DB
}

func (e workOrderTenantGormEngine) Client() *gorm.DB {
	return e.client
}

func workOrderTenantContext(tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "tenant-user@example.com")
	ctx = context.WithValue(ctx, contextkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, contextkeys.Role, "tenant_admin")
	ctx = context.WithValue(ctx, contextkeys.Scopes, []string{"workorders.read", "workorders.write", "workorders.archive"})
	return ctx
}

func setupWorkOrderTenantDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE workorders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			number TEXT NOT NULL,
			project_id INTEGER NOT NULL,
			field_id INTEGER NOT NULL,
			lot_id INTEGER NOT NULL,
			crop_id INTEGER NOT NULL,
			labor_id INTEGER NOT NULL,
			contractor TEXT,
			observations TEXT,
			date DATETIME NOT NULL,
			sequence_day INTEGER,
			investor_id INTEGER NOT NULL,
			effective_area NUMERIC NOT NULL DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		);
		CREATE TABLE workorder_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			workorder_id INTEGER NOT NULL,
			supply_id INTEGER NOT NULL,
			supply_name TEXT NOT NULL,
			total_used NUMERIC NOT NULL DEFAULT 0,
			final_dose NUMERIC NOT NULL DEFAULT 0
		);
		CREATE TABLE workorder_investor_splits (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			workorder_id INTEGER NOT NULL,
			investor_id INTEGER NOT NULL,
			percentage NUMERIC NOT NULL DEFAULT 0,
			payment_status TEXT NOT NULL DEFAULT 'Pendiente',
			deleted_at DATETIME
		);
		CREATE TABLE invoices (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			work_order_id INTEGER NOT NULL,
			deleted_at DATETIME
		);
		CREATE TABLE labors (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			work_order_id INTEGER,
			deleted_at DATETIME
		);
	`).Error; err != nil {
		t.Fatalf("create schema: %v", err)
	}

	return db
}

func TestWorkOrderRepositoryTenantIsolation(t *testing.T) {
	db := setupWorkOrderTenantDB(t)
	repo := NewRepository(workOrderTenantGormEngine{client: db})

	tenantA := uuid.New()
	tenantB := uuid.New()
	now := time.Now().UTC()

	if err := db.Exec(`
		INSERT INTO workorders (
			id, tenant_id, number, project_id, field_id, lot_id, crop_id, labor_id,
			contractor, observations, date, sequence_day, investor_id, effective_area,
			created_at, updated_at, deleted_at
		) VALUES
			(1, ?, '1000', 10, 100, 1000, 1, 1, 'A', '', ?, 1, 11, 10, ?, ?, NULL),
			(2, ?, '2000', 20, 200, 2000, 1, 1, 'B', '', ?, 1, 22, 20, ?, ?, NULL),
			(3, ?, '3000', 20, 200, 2000, 1, 1, 'B archived', '', ?, 1, 22, 30, ?, ?, ?),
			(4, ?, '4000', 10, 100, 1000, 1, 1, 'A archived', '', ?, 1, 11, 40, ?, ?, ?);
		INSERT INTO workorder_investor_splits (id, tenant_id, workorder_id, investor_id, percentage, payment_status, deleted_at) VALUES
			(1, ?, 2, 22, 100, 'Pendiente', NULL);
	`, tenantA.String(), now, now, now,
		tenantB.String(), now, now, now,
		tenantB.String(), now, now, now, now,
		tenantA.String(), now, now, now, now,
		tenantB.String(),
	).Error; err != nil {
		t.Fatalf("seed work orders: %v", err)
	}

	ctxA := workOrderTenantContext(tenantA)

	archived, total, err := repo.ListArchivedWorkOrders(ctxA, 1, 50, domain.ArchivedWorkOrderFilter{})
	if err != nil {
		t.Fatalf("list archived work orders: %v", err)
	}
	if total != 1 || len(archived) != 1 || archived[0].ID != 4 {
		t.Fatalf("expected only tenant A archived work order, total=%d archived=%#v", total, archived)
	}

	if _, err := repo.GetWorkOrderByID(ctxA, 2); err == nil {
		t.Fatalf("expected get cross-tenant work order to fail")
	}

	existing, err := repo.GetWorkOrderByNumberAndProjectID(ctxA, "2000", 20)
	if err != nil {
		t.Fatalf("get by number/project cross-tenant: %v", err)
	}
	if existing != nil {
		t.Fatalf("expected no cross-tenant work order by number/project, got %#v", existing)
	}

	if err := repo.UpdateWorkOrderByID(ctxA, &domain.WorkOrder{
		ID:            2,
		Number:        "cross tenant update",
		ProjectID:     20,
		FieldID:       200,
		LotID:         2000,
		CropID:        1,
		LaborID:       1,
		Date:          now,
		InvestorID:    22,
		EffectiveArea: decimal.NewFromInt(99),
		Base:          shareddomain.Base{UpdatedAt: now},
	}); err == nil {
		t.Fatalf("expected update cross-tenant work order to fail")
	}

	var number string
	if err := db.Raw(`SELECT number FROM workorders WHERE id = 2`).Scan(&number).Error; err != nil {
		t.Fatalf("read work order 2: %v", err)
	}
	if number != "2000" {
		t.Fatalf("cross-tenant update changed work order 2 number to %q", number)
	}

	if err := repo.UpdateInvestorPaymentStatus(ctxA, 2, 22, domain.InvestorPaymentStatusPaid); err == nil {
		t.Fatalf("expected payment status update on cross-tenant work order to fail")
	}
	var paymentStatus string
	if err := db.Raw(`SELECT payment_status FROM workorder_investor_splits WHERE id = 1`).Scan(&paymentStatus).Error; err != nil {
		t.Fatalf("read split 1: %v", err)
	}
	if paymentStatus != domain.InvestorPaymentStatusPending {
		t.Fatalf("cross-tenant payment status update changed split to %q", paymentStatus)
	}

	if err := repo.ArchiveWorkOrder(ctxA, 2); err == nil {
		t.Fatalf("expected archive cross-tenant work order to fail")
	}
	var deletedCount int64
	if err := db.Raw(`SELECT COUNT(*) FROM workorders WHERE id = 2 AND deleted_at IS NOT NULL`).Scan(&deletedCount).Error; err != nil {
		t.Fatalf("check archive side effect: %v", err)
	}
	if deletedCount != 0 {
		t.Fatalf("cross-tenant archive modified work order 2")
	}

	if err := repo.RestoreWorkOrder(ctxA, 3); err == nil {
		t.Fatalf("expected restore cross-tenant work order to fail")
	}
	if err := db.Raw(`SELECT COUNT(*) FROM workorders WHERE id = 3 AND deleted_at IS NOT NULL`).Scan(&deletedCount).Error; err != nil {
		t.Fatalf("check restore side effect: %v", err)
	}
	if deletedCount != 1 {
		t.Fatalf("cross-tenant restore modified work order 3")
	}

	if err := repo.HardDeleteWorkOrder(ctxA, 2); err == nil {
		t.Fatalf("expected hard-delete cross-tenant work order to fail")
	}
	var exists int64
	if err := db.Raw(`SELECT COUNT(*) FROM workorders WHERE id = 2`).Scan(&exists).Error; err != nil {
		t.Fatalf("check hard delete side effect: %v", err)
	}
	if exists != 1 {
		t.Fatalf("cross-tenant hard delete removed work order 2")
	}
}

func TestWorkOrderRepositoryRequiresTenantInStrictModeForMetrics(t *testing.T) {
	t.Setenv("TENANT_STRICT_MODE", "true")

	db := setupWorkOrderTenantDB(t)
	repo := NewRepository(workOrderTenantGormEngine{client: db})

	if _, err := repo.GetMetrics(context.Background(), domain.WorkOrderFilter{}); err == nil {
		t.Fatalf("expected strict mode to reject metrics without tenant context")
	}

	if _, err := repo.GetRawDirectCost(context.Background(), 0); err == nil {
		t.Fatalf("expected strict mode to reject raw direct cost without tenant context")
	}

	if _, _, _, err := repo.GetHarvestAreaSnapshot(context.Background(), 1, 1, 0); err == nil {
		t.Fatalf("expected strict mode to reject harvest area snapshot without tenant context")
	}
}
