package customer

import (
	"context"
	"testing"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	contextkeys "github.com/devpablocristo/platform/security/go/contextkeys"
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
				actor_id INTEGER,
				created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT,
			archive_batch_id INTEGER,
			archive_origin_entity TEXT,
			archive_origin_id INTEGER,
			archive_reason TEXT
		);
		CREATE TABLE projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			customer_id INTEGER NOT NULL,
			campaign_id INTEGER NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT,
			archive_batch_id INTEGER,
			archive_origin_entity TEXT,
			archive_origin_id INTEGER,
			archive_reason TEXT
		);
		CREATE TABLE archive_batches (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT,
			root_entity TEXT NOT NULL,
			root_id INTEGER NOT NULL,
			action TEXT NOT NULL DEFAULT 'archive',
			reason TEXT,
			created_by TEXT,
			created_at DATETIME
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

func TestUpdateCustomerDuplicateNameReturnsConflict(t *testing.T) {
	db := setupCustomerTenantDB(t)
	repo := NewRepository(customerTenantGormEngine{client: db})

	tenantID := uuid.New()
	if err := db.Exec(`
		CREATE UNIQUE INDEX ux_customers_tenant_name_active
			ON customers (tenant_id, name)
			WHERE deleted_at IS NULL;
		INSERT INTO customers (id, tenant_id, name, deleted_at) VALUES
			(10, ?, 'EL SUEÑO', NULL),
			(11, ?, 'SOALEN SRL', NULL);
	`, tenantID.String(), tenantID.String()).Error; err != nil {
		t.Fatalf("seed customers: %v", err)
	}

	err := repo.UpdateCustomer(customerTenantContext(tenantID), &domain.Customer{
		ID:   11,
		Name: "EL SUEÑO",
	})
	if err == nil {
		t.Fatalf("expected duplicate customer name to fail")
	}
	if !domainerr.IsKind(err, domainerr.KindConflict) {
		t.Fatalf("expected conflict, got %T %v", err, err)
	}
	if err.Error() != "CONFLICT: customer already exists" {
		t.Fatalf("expected domain message, got %q", err.Error())
	}
}

func TestArchiveCustomerArchivesActiveProjects(t *testing.T) {
	db := setupCustomerTenantDB(t)
	repo := NewRepository(customerTenantGormEngine{client: db})

	tenantID := uuid.New()
	now := time.Now().UTC()

	if err := db.Exec(`
		INSERT INTO customers (id, tenant_id, name, created_at, updated_at, deleted_at) VALUES
			(10, ?, 'Customer With Projects', ?, ?, NULL);
		INSERT INTO projects (id, tenant_id, name, customer_id, campaign_id, created_at, updated_at, deleted_at) VALUES
			(20, ?, 'Project One', 10, 1, ?, ?, NULL),
			(21, ?, 'Project Two', 10, 1, ?, ?, NULL);
	`, tenantID.String(), now, now,
		tenantID.String(), now, now,
		tenantID.String(), now, now,
	).Error; err != nil {
		t.Fatalf("seed customer projects: %v", err)
	}

	if err := repo.ArchiveCustomer(customerTenantContext(tenantID), 10); err != nil {
		t.Fatalf("archive customer: %v", err)
	}

	var activeProjects int64
	if err := db.Raw(`SELECT COUNT(*) FROM projects WHERE customer_id = 10 AND deleted_at IS NULL`).Scan(&activeProjects).Error; err != nil {
		t.Fatalf("count active projects: %v", err)
	}
	if activeProjects != 0 {
		t.Fatalf("expected all customer projects archived, got %d active", activeProjects)
	}

	var archivedCustomer int64
	if err := db.Raw(`SELECT COUNT(*) FROM customers WHERE id = 10 AND deleted_at IS NOT NULL`).Scan(&archivedCustomer).Error; err != nil {
		t.Fatalf("count archived customer: %v", err)
	}
	if archivedCustomer != 1 {
		t.Fatalf("expected customer archived, got %d", archivedCustomer)
	}

	archived, total, err := repo.ListArchivedCustomers(customerTenantContext(tenantID), 1, 100)
	if err != nil {
		t.Fatalf("list archived customers: %v", err)
	}
	if total != 1 || len(archived) != 1 || archived[0].ID != 10 {
		t.Fatalf("expected archived customer 10, total=%d archived=%#v", total, archived)
	}

	if err := repo.RestoreCustomer(customerTenantContext(tenantID), 10); err != nil {
		t.Fatalf("restore customer: %v", err)
	}

	var restoredProjects int64
	if err := db.Raw(`SELECT COUNT(*) FROM projects WHERE customer_id = 10 AND deleted_at IS NULL`).Scan(&restoredProjects).Error; err != nil {
		t.Fatalf("count restored projects: %v", err)
	}
	if restoredProjects != 2 {
		t.Fatalf("expected both customer projects restored, got %d", restoredProjects)
	}
}

func TestRestoreCustomerDoesNotRestoreManuallyArchivedProject(t *testing.T) {
	db := setupCustomerTenantDB(t)
	repo := NewRepository(customerTenantGormEngine{client: db})

	tenantID := uuid.New()
	now := time.Now().UTC()

	if err := db.Exec(`
		INSERT INTO customers (id, tenant_id, name, created_at, updated_at, deleted_at) VALUES
			(30, ?, 'Customer With Mixed Projects', ?, ?, NULL);
		INSERT INTO projects (id, tenant_id, name, customer_id, campaign_id, created_at, updated_at, deleted_at) VALUES
			(40, ?, 'Active Project', 30, 1, ?, ?, NULL),
			(41, ?, 'Manually Archived Project', 30, 1, ?, ?, ?);
	`, tenantID.String(), now, now,
		tenantID.String(), now, now,
		tenantID.String(), now, now, now,
	).Error; err != nil {
		t.Fatalf("seed customer projects: %v", err)
	}

	if err := repo.ArchiveCustomer(customerTenantContext(tenantID), 30); err != nil {
		t.Fatalf("archive customer: %v", err)
	}
	if err := repo.RestoreCustomer(customerTenantContext(tenantID), 30); err != nil {
		t.Fatalf("restore customer: %v", err)
	}

	var activeProjectRestored int64
	if err := db.Raw(`SELECT COUNT(*) FROM projects WHERE id = 40 AND deleted_at IS NULL`).Scan(&activeProjectRestored).Error; err != nil {
		t.Fatalf("count restored project: %v", err)
	}
	if activeProjectRestored != 1 {
		t.Fatalf("expected project archived by customer restored")
	}

	var manuallyArchivedStillArchived int64
	if err := db.Raw(`SELECT COUNT(*) FROM projects WHERE id = 41 AND deleted_at IS NOT NULL`).Scan(&manuallyArchivedStillArchived).Error; err != nil {
		t.Fatalf("count manual archived project: %v", err)
	}
	if manuallyArchivedStillArchived != 1 {
		t.Fatalf("expected manually archived project to remain archived")
	}
}

func TestRestoreCustomerRestoresProjectGraphOnlyByCause(t *testing.T) {
	db := setupCustomerTenantDB(t)
	repo := NewRepository(customerTenantGormEngine{client: db})

	if err := db.Exec(`
		CREATE TABLE fields (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			project_id INTEGER NOT NULL,
			deleted_at DATETIME,
			archive_batch_id INTEGER,
			archive_origin_entity TEXT,
			archive_origin_id INTEGER
		);
		CREATE TABLE lots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			field_id INTEGER NOT NULL,
			deleted_at DATETIME,
			archive_batch_id INTEGER,
			archive_origin_entity TEXT,
			archive_origin_id INTEGER
		);
		CREATE TABLE workorders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			project_id INTEGER NOT NULL,
			deleted_at DATETIME,
			archive_batch_id INTEGER,
			archive_origin_entity TEXT,
			archive_origin_id INTEGER
		);
		CREATE TABLE workorder_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			workorder_id INTEGER NOT NULL,
			deleted_at DATETIME,
			archive_batch_id INTEGER,
			archive_origin_entity TEXT,
			archive_origin_id INTEGER
		);
		CREATE TABLE work_order_drafts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			project_id INTEGER NOT NULL,
			deleted_at DATETIME,
			archive_batch_id INTEGER,
			archive_origin_entity TEXT,
			archive_origin_id INTEGER
		);
		CREATE TABLE work_order_draft_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			draft_id INTEGER NOT NULL,
			deleted_at DATETIME,
			archive_batch_id INTEGER,
			archive_origin_entity TEXT,
			archive_origin_id INTEGER
		);
	`).Error; err != nil {
		t.Fatalf("create graph tables: %v", err)
	}

	tenantID := uuid.New()
	now := time.Now().UTC()
	if err := db.Exec(`
		INSERT INTO customers (id, tenant_id, name, created_at, updated_at, deleted_at) VALUES
			(80, ?, 'Customer Graph', ?, ?, NULL);
		INSERT INTO projects (id, tenant_id, name, customer_id, campaign_id, created_at, updated_at, deleted_at) VALUES
			(80, ?, 'Project Graph', 80, 1, ?, ?, NULL);
		INSERT INTO fields (id, tenant_id, name, project_id, deleted_at) VALUES
			(80, ?, 'Field Graph', 80, NULL);
		INSERT INTO lots (id, tenant_id, name, field_id, deleted_at) VALUES
			(80, ?, 'Lot Graph', 80, NULL),
			(81, ?, 'Manual Lot', 80, ?);
		INSERT INTO workorders (id, tenant_id, project_id, deleted_at) VALUES
			(80, ?, 80, NULL);
		INSERT INTO workorder_items (id, tenant_id, workorder_id, deleted_at) VALUES
			(80, ?, 80, NULL);
		INSERT INTO work_order_drafts (id, tenant_id, project_id, deleted_at) VALUES
			(80, ?, 80, NULL);
		INSERT INTO work_order_draft_items (id, draft_id, deleted_at) VALUES
			(80, 80, NULL);
	`, tenantID.String(), now, now,
		tenantID.String(), now, now,
		tenantID.String(),
		tenantID.String(),
		tenantID.String(), now,
		tenantID.String(),
		tenantID.String(),
		tenantID.String(),
	).Error; err != nil {
		t.Fatalf("seed graph: %v", err)
	}

	ctx := customerTenantContext(tenantID)
	if err := repo.ArchiveCustomer(ctx, 80); err != nil {
		t.Fatalf("archive customer graph: %v", err)
	}
	if err := repo.RestoreCustomer(ctx, 80); err != nil {
		t.Fatalf("restore customer graph: %v", err)
	}

	checks := map[string]string{
		"project":    `SELECT COUNT(*) FROM projects WHERE id = 80 AND deleted_at IS NULL AND archive_batch_id IS NULL`,
		"field":      `SELECT COUNT(*) FROM fields WHERE id = 80 AND deleted_at IS NULL AND archive_batch_id IS NULL`,
		"lot":        `SELECT COUNT(*) FROM lots WHERE id = 80 AND deleted_at IS NULL AND archive_batch_id IS NULL`,
		"workorder":  `SELECT COUNT(*) FROM workorders WHERE id = 80 AND deleted_at IS NULL AND archive_batch_id IS NULL`,
		"wo_item":    `SELECT COUNT(*) FROM workorder_items WHERE id = 80 AND deleted_at IS NULL AND archive_batch_id IS NULL`,
		"draft":      `SELECT COUNT(*) FROM work_order_drafts WHERE id = 80 AND deleted_at IS NULL AND archive_batch_id IS NULL`,
		"draft_item": `SELECT COUNT(*) FROM work_order_draft_items WHERE id = 80 AND deleted_at IS NULL AND archive_batch_id IS NULL`,
		"manual_lot": `SELECT COUNT(*) FROM lots WHERE id = 81 AND deleted_at IS NOT NULL`,
	}
	for name, query := range checks {
		var count int64
		if err := db.Raw(query).Scan(&count).Error; err != nil {
			t.Fatalf("check %s: %v", name, err)
		}
		if count != 1 {
			t.Fatalf("expected %s count 1, got %d", name, count)
		}
	}
}

func TestRestoreActiveCustomerRepairsChildrenArchivedBySameOrigin(t *testing.T) {
	db := setupCustomerTenantDB(t)
	repo := NewRepository(customerTenantGormEngine{client: db})

	if err := db.Exec(`
		CREATE TABLE fields (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			project_id INTEGER NOT NULL,
			deleted_at DATETIME,
			archive_batch_id INTEGER,
			archive_origin_entity TEXT,
			archive_origin_id INTEGER
		);
	`).Error; err != nil {
		t.Fatalf("create repair tables: %v", err)
	}

	tenantID := uuid.New()
	now := time.Now().UTC()
	if err := db.Exec(`
		INSERT INTO customers (id, tenant_id, name, created_at, updated_at, deleted_at, archive_batch_id, archive_origin_entity, archive_origin_id) VALUES
			(90, ?, 'Active Customer', ?, ?, NULL, NULL, NULL, NULL);
		INSERT INTO projects (id, tenant_id, name, customer_id, campaign_id, created_at, updated_at, deleted_at, archive_batch_id, archive_origin_entity, archive_origin_id) VALUES
			(90, ?, 'Active Project', 90, 1, ?, ?, NULL, NULL, NULL, NULL);
		INSERT INTO fields (id, tenant_id, name, project_id, deleted_at, archive_batch_id, archive_origin_entity, archive_origin_id) VALUES
			(90, ?, 'Broken Cascade Field', 90, ?, 77, 'customers', 90),
			(91, ?, 'Manual Field', 90, ?, 88, 'fields', 91);
	`, tenantID.String(), now, now,
		tenantID.String(), now, now,
		tenantID.String(), now,
		tenantID.String(), now,
	).Error; err != nil {
		t.Fatalf("seed repair graph: %v", err)
	}

	ctx := customerTenantContext(tenantID)
	if err := repo.RestoreCustomer(ctx, 90); err != nil {
		t.Fatalf("repair active customer graph: %v", err)
	}

	checks := map[string]string{
		"same_origin_restored": `SELECT COUNT(*) FROM fields WHERE id = 90 AND deleted_at IS NULL AND archive_batch_id IS NULL`,
		"manual_preserved":     `SELECT COUNT(*) FROM fields WHERE id = 91 AND deleted_at IS NOT NULL AND archive_batch_id = 88`,
	}
	for name, query := range checks {
		var count int64
		if err := db.Raw(query).Scan(&count).Error; err != nil {
			t.Fatalf("check %s: %v", name, err)
		}
		if count != 1 {
			t.Fatalf("expected %s count 1, got %d", name, count)
		}
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
