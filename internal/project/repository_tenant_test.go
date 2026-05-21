package project

import (
	"context"
	"testing"
	"time"

	contextkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	cropmod "github.com/devpablocristo/ponti-backend/internal/crop/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/project/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
)

type projectTenantGormEngine struct {
	client *gorm.DB
}

func (e projectTenantGormEngine) Client() *gorm.DB {
	return e.client
}

func projectTenantContext(tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "tenant-user@example.com")
	ctx = context.WithValue(ctx, contextkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, contextkeys.Role, "tenant_admin")
	ctx = context.WithValue(ctx, contextkeys.Scopes, []string{"projects.read", "projects.write", "projects.archive"})
	return ctx
}

func setupProjectTenantDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			customer_id INTEGER NOT NULL,
			campaign_id INTEGER NOT NULL,
			admin_cost NUMERIC NOT NULL DEFAULT 0,
			planned_cost NUMERIC NOT NULL DEFAULT 0,
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
				deleted_by TEXT
			);
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
		CREATE TABLE fields (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			project_id INTEGER NOT NULL,
			lease_type_id INTEGER NOT NULL DEFAULT 0,
			lease_type_percent NUMERIC,
			lease_type_value NUMERIC,
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
		CREATE TABLE lots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			field_id INTEGER NOT NULL,
			hectares NUMERIC NOT NULL DEFAULT 0,
			previous_crop_id INTEGER NOT NULL DEFAULT 0,
			current_crop_id INTEGER NOT NULL DEFAULT 0,
			season TEXT NOT NULL DEFAULT '',
			variety TEXT,
			sowing_date DATETIME,
			tons NUMERIC DEFAULT 0,
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
		CREATE TABLE managers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
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
		CREATE TABLE project_managers (
			tenant_id TEXT NOT NULL,
			project_id INTEGER NOT NULL,
			manager_id INTEGER NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT,
			archive_batch_id INTEGER,
			archive_origin_entity TEXT,
			archive_origin_id INTEGER,
			archive_reason TEXT,
			UNIQUE(project_id, manager_id)
		);
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
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		);
		CREATE TABLE admin_cost_investors (
			tenant_id TEXT NOT NULL,
			project_id INTEGER NOT NULL,
			investor_id INTEGER NOT NULL,
			percentage INTEGER NOT NULL DEFAULT 0,
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
			tenant_id TEXT,
			source_table TEXT NOT NULL,
			source_id INTEGER NOT NULL,
			source_key TEXT NOT NULL,
			source_text TEXT,
			actor_id INTEGER NOT NULL,
			confidence NUMERIC,
			mapping_status TEXT,
			PRIMARY KEY (tenant_id, source_table, source_key)
		);
	`).Error; err != nil {
		t.Fatalf("create schema: %v", err)
	}

	return db
}

func TestProjectRepositoryTenantIsolation(t *testing.T) {
	db := setupProjectTenantDB(t)
	repo := NewRepository(projectTenantGormEngine{client: db})

	tenantA := uuid.New()
	tenantB := uuid.New()
	now := time.Now().UTC()

	if err := db.Exec(`
		INSERT INTO customers (id, tenant_id, name, deleted_at) VALUES
			(1, ?, 'Customer A', NULL),
			(2, ?, 'Customer B', NULL);
		INSERT INTO campaigns (id, tenant_id, name) VALUES
			(1, ?, '2025-2026'),
			(2, ?, '2025-2026');
		INSERT INTO projects (id, tenant_id, name, customer_id, campaign_id, admin_cost, planned_cost, created_at, updated_at, deleted_at) VALUES
			(1, ?, 'Project A', 1, 1, 10, 100, ?, ?, NULL),
			(2, ?, 'Project B', 2, 2, 20, 200, ?, ?, NULL),
			(3, ?, 'Project B archived', 2, 2, 30, 300, ?, ?, ?);
		INSERT INTO fields (id, tenant_id, name, project_id, lease_type_id, created_at, updated_at, deleted_at) VALUES
			(1, ?, 'Field A', 1, 0, ?, ?, NULL),
			(2, ?, 'Field B', 2, 0, ?, ?, NULL),
			(3, ?, 'Corrupt cross tenant field', 1, 0, ?, ?, NULL);
		INSERT INTO lots (id, tenant_id, name, field_id, hectares, previous_crop_id, current_crop_id, season, created_at, updated_at, deleted_at) VALUES
			(1, ?, 'Lot A', 1, 10, 0, 0, '2025-2026', ?, ?, NULL),
			(2, ?, 'Lot B', 2, 20, 0, 0, '2025-2026', ?, ?, NULL),
			(3, ?, 'Corrupt cross tenant lot', 3, 999, 0, 0, '2025-2026', ?, ?, NULL);
	`, tenantA.String(), tenantB.String(),
		tenantA.String(), tenantB.String(),
		tenantA.String(), now, now,
		tenantB.String(), now, now,
		tenantB.String(), now, now, now,
		tenantA.String(), now, now,
		tenantB.String(), now, now,
		tenantB.String(), now, now,
		tenantA.String(), now, now,
		tenantB.String(), now, now,
		tenantB.String(), now, now,
	).Error; err != nil {
		t.Fatalf("seed projects: %v", err)
	}

	ctxA := projectTenantContext(tenantA)

	list, total, err := repo.ListProjects(ctxA, 1, 50)
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}
	if total != 1 || len(list) != 1 || list[0].Name != "Project A" {
		t.Fatalf("expected only tenant A project, total=%d list=%#v", total, list)
	}

	projects, totalHectares, total, err := repo.GetProjects(ctxA, "", 0, 0, 1, 50)
	if err != nil {
		t.Fatalf("get projects: %v", err)
	}
	if total != 1 || len(projects) != 1 || projects[0].ID != 1 {
		t.Fatalf("expected only tenant A hydrated project, total=%d projects=%#v", total, projects)
	}
	if !totalHectares.Equal(decimal.NewFromInt(10)) {
		t.Fatalf("expected tenant-safe hectares 10, got %s", totalHectares.String())
	}

	fields, err := repo.GetFieldsByProjectID(ctxA, 2)
	if err != nil {
		t.Fatalf("list cross-tenant fields: %v", err)
	}
	if len(fields) != 0 {
		t.Fatalf("expected no cross-tenant fields, got %#v", fields)
	}

	if _, err := repo.GetProject(ctxA, 2); err == nil {
		t.Fatalf("expected get cross-tenant project to fail")
	}

	if err := repo.UpdateProject(ctxA, &domain.Project{
		ID:   2,
		Name: "cross tenant update",
		Base: shareddomain.Base{UpdatedAt: now},
	}); err == nil {
		t.Fatalf("expected update cross-tenant project to fail")
	}

	var name string
	if err := db.Raw(`SELECT name FROM projects WHERE id = 2`).Scan(&name).Error; err != nil {
		t.Fatalf("read project 2: %v", err)
	}
	if name != "Project B" {
		t.Fatalf("cross-tenant update changed project 2 name to %q", name)
	}

	if err := repo.ArchiveProject(ctxA, 2); err == nil {
		t.Fatalf("expected archive cross-tenant project to fail")
	}
	var deletedCount int64
	if err := db.Raw(`SELECT COUNT(*) FROM projects WHERE id = 2 AND deleted_at IS NOT NULL`).Scan(&deletedCount).Error; err != nil {
		t.Fatalf("check archive side effect: %v", err)
	}
	if deletedCount != 0 {
		t.Fatalf("cross-tenant archive modified project 2")
	}

	if err := repo.RestoreProject(ctxA, 3); err == nil {
		t.Fatalf("expected restore cross-tenant project to fail")
	}
	if err := db.Raw(`SELECT COUNT(*) FROM projects WHERE id = 3 AND deleted_at IS NOT NULL`).Scan(&deletedCount).Error; err != nil {
		t.Fatalf("check restore side effect: %v", err)
	}
	if deletedCount != 1 {
		t.Fatalf("cross-tenant restore modified project 3")
	}

	if err := repo.HardDeleteProject(ctxA, 2); err == nil {
		t.Fatalf("expected hard-delete cross-tenant project to fail")
	}
	var exists int64
	if err := db.Raw(`SELECT COUNT(*) FROM projects WHERE id = 2`).Scan(&exists).Error; err != nil {
		t.Fatalf("check hard delete side effect: %v", err)
	}
	if exists != 1 {
		t.Fatalf("cross-tenant hard delete removed project 2")
	}
}

func TestProjectRepositoryRequiresTenantInStrictModeForCreate(t *testing.T) {
	t.Setenv("TENANT_STRICT_MODE", "true")

	db := setupProjectTenantDB(t)
	repo := NewRepository(projectTenantGormEngine{client: db})

	ctx := context.WithValue(context.Background(), contextkeys.Actor, "tenant-user@example.com")
	if _, err := repo.CreateProject(ctx, &domain.Project{}); err == nil {
		t.Fatalf("expected CreateProject without tenant to fail in strict mode")
	}
}

func TestArchiveProjectUsesBatchAndDoesNotArchiveCustomer(t *testing.T) {
	db := setupProjectTenantDB(t)
	repo := NewRepository(projectTenantGormEngine{client: db})

	tenantID := uuid.New()
	now := time.Now().UTC()
	if err := db.Exec(`
		INSERT INTO customers (id, tenant_id, name, deleted_at) VALUES
			(10, ?, 'Customer A', NULL);
		INSERT INTO campaigns (id, tenant_id, name) VALUES
			(10, ?, '2025-2026');
		INSERT INTO projects (id, tenant_id, name, customer_id, campaign_id, admin_cost, planned_cost, created_at, updated_at, deleted_at) VALUES
			(10, ?, 'Project A', 10, 10, 0, 0, ?, ?, NULL);
		INSERT INTO fields (id, tenant_id, name, project_id, lease_type_id, created_at, updated_at, deleted_at) VALUES
			(10, ?, 'Field A', 10, 0, ?, ?, NULL);
		INSERT INTO lots (id, tenant_id, name, field_id, hectares, previous_crop_id, current_crop_id, season, created_at, updated_at, deleted_at) VALUES
			(10, ?, 'Lot A', 10, 1, 0, 0, '2025-2026', ?, ?, NULL);
	`, tenantID.String(),
		tenantID.String(),
		tenantID.String(), now, now,
		tenantID.String(), now, now,
		tenantID.String(), now, now,
	).Error; err != nil {
		t.Fatalf("seed project: %v", err)
	}

	if err := repo.ArchiveProject(projectTenantContext(tenantID), 10); err != nil {
		t.Fatalf("archive project: %v", err)
	}

	var archivedProject, archivedField, archivedLot, archivedCustomer int64
	if err := db.Raw(`SELECT COUNT(*) FROM projects WHERE id = 10 AND deleted_at IS NOT NULL AND archive_batch_id IS NOT NULL AND archive_origin_entity = 'projects' AND archive_origin_id = 10`).Scan(&archivedProject).Error; err != nil {
		t.Fatalf("count archived project: %v", err)
	}
	if err := db.Raw(`SELECT COUNT(*) FROM fields WHERE id = 10 AND deleted_at IS NOT NULL AND archive_batch_id IS NOT NULL AND archive_origin_entity = 'projects' AND archive_origin_id = 10`).Scan(&archivedField).Error; err != nil {
		t.Fatalf("count archived field: %v", err)
	}
	if err := db.Raw(`SELECT COUNT(*) FROM lots WHERE id = 10 AND deleted_at IS NOT NULL AND archive_batch_id IS NOT NULL AND archive_origin_entity = 'projects' AND archive_origin_id = 10`).Scan(&archivedLot).Error; err != nil {
		t.Fatalf("count archived lot: %v", err)
	}
	if err := db.Raw(`SELECT COUNT(*) FROM customers WHERE id = 10 AND deleted_at IS NOT NULL`).Scan(&archivedCustomer).Error; err != nil {
		t.Fatalf("count archived customer: %v", err)
	}
	if archivedProject != 1 || archivedField != 1 || archivedLot != 1 {
		t.Fatalf("expected project/field/lot archived with project cause, got project=%d field=%d lot=%d", archivedProject, archivedField, archivedLot)
	}
	if archivedCustomer != 0 {
		t.Fatalf("project archive must not archive customer")
	}

	if err := repo.RestoreProject(projectTenantContext(tenantID), 10); err != nil {
		t.Fatalf("restore project: %v", err)
	}

	var activeProject, activeField, activeLot int64
	if err := db.Raw(`SELECT COUNT(*) FROM projects WHERE id = 10 AND deleted_at IS NULL AND archive_batch_id IS NULL`).Scan(&activeProject).Error; err != nil {
		t.Fatalf("count active project: %v", err)
	}
	if err := db.Raw(`SELECT COUNT(*) FROM fields WHERE id = 10 AND deleted_at IS NULL AND archive_batch_id IS NULL`).Scan(&activeField).Error; err != nil {
		t.Fatalf("count active field: %v", err)
	}
	if err := db.Raw(`SELECT COUNT(*) FROM lots WHERE id = 10 AND deleted_at IS NULL AND archive_batch_id IS NULL`).Scan(&activeLot).Error; err != nil {
		t.Fatalf("count active lot: %v", err)
	}
	if activeProject != 1 || activeField != 1 || activeLot != 1 {
		t.Fatalf("expected project/field/lot restored, got project=%d field=%d lot=%d", activeProject, activeField, activeLot)
	}
}

func TestArchiveProjectDoesNotOverwriteManuallyArchivedChildren(t *testing.T) {
	db := setupProjectTenantDB(t)
	repo := NewRepository(projectTenantGormEngine{client: db})

	tenantID := uuid.New()
	now := time.Now().UTC()
	if err := db.Exec(`
		INSERT INTO customers (id, tenant_id, name, deleted_at) VALUES
			(50, ?, 'Customer A', NULL);
		INSERT INTO campaigns (id, tenant_id, name) VALUES
			(50, ?, '2025-2026');
		INSERT INTO archive_batches (id, tenant_id, root_entity, root_id, action, created_at) VALUES
			(950, ?, 'lots', 950, 'archive', ?);
		INSERT INTO projects (id, tenant_id, name, customer_id, campaign_id, admin_cost, planned_cost, created_at, updated_at, deleted_at) VALUES
			(50, ?, 'Project A', 50, 50, 0, 0, ?, ?, NULL);
		INSERT INTO fields (id, tenant_id, name, project_id, lease_type_id, created_at, updated_at, deleted_at) VALUES
			(50, ?, 'Field A', 50, 0, ?, ?, NULL);
		INSERT INTO lots (
			id, tenant_id, name, field_id, hectares, previous_crop_id, current_crop_id, season,
			created_at, updated_at, deleted_at, archive_batch_id, archive_origin_entity, archive_origin_id
		) VALUES
			(50, ?, 'Active Lot', 50, 1, 0, 0, '2025-2026', ?, ?, NULL, NULL, NULL, NULL),
			(51, ?, 'Manual Lot', 50, 1, 0, 0, '2025-2026', ?, ?, ?, 950, 'lots', 950);
	`, tenantID.String(),
		tenantID.String(),
		tenantID.String(), now,
		tenantID.String(), now, now,
		tenantID.String(), now, now,
		tenantID.String(), now, now,
		tenantID.String(), now, now, now,
	).Error; err != nil {
		t.Fatalf("seed project: %v", err)
	}

	if err := repo.ArchiveProject(projectTenantContext(tenantID), 50); err != nil {
		t.Fatalf("archive project: %v", err)
	}

	var activeLotArchived, manualLotPreserved int64
	if err := db.Raw(`SELECT COUNT(*) FROM lots WHERE id = 50 AND deleted_at IS NOT NULL AND archive_origin_entity = 'projects' AND archive_origin_id = 50`).Scan(&activeLotArchived).Error; err != nil {
		t.Fatalf("count active lot archive cause: %v", err)
	}
	if err := db.Raw(`SELECT COUNT(*) FROM lots WHERE id = 51 AND deleted_at IS NOT NULL AND archive_batch_id = 950 AND archive_origin_entity = 'lots' AND archive_origin_id = 950`).Scan(&manualLotPreserved).Error; err != nil {
		t.Fatalf("count manual lot cause: %v", err)
	}
	if activeLotArchived != 1 || manualLotPreserved != 1 {
		t.Fatalf("expected active lot archived by project and manual lot preserved, got active=%d manual=%d", activeLotArchived, manualLotPreserved)
	}

	if err := repo.RestoreProject(projectTenantContext(tenantID), 50); err != nil {
		t.Fatalf("restore project: %v", err)
	}

	var restoredActiveLot, manualStillArchived int64
	if err := db.Raw(`SELECT COUNT(*) FROM lots WHERE id = 50 AND deleted_at IS NULL AND archive_batch_id IS NULL`).Scan(&restoredActiveLot).Error; err != nil {
		t.Fatalf("count restored lot: %v", err)
	}
	if err := db.Raw(`SELECT COUNT(*) FROM lots WHERE id = 51 AND deleted_at IS NOT NULL AND archive_batch_id = 950`).Scan(&manualStillArchived).Error; err != nil {
		t.Fatalf("count manual still archived: %v", err)
	}
	if restoredActiveLot != 1 || manualStillArchived != 1 {
		t.Fatalf("expected project restore to skip manual lot, got restored=%d manual=%d", restoredActiveLot, manualStillArchived)
	}
}

func TestRestoreProjectRequiresActiveCustomer(t *testing.T) {
	db := setupProjectTenantDB(t)
	repo := NewRepository(projectTenantGormEngine{client: db})

	tenantID := uuid.New()
	now := time.Now().UTC()
	if err := db.Exec(`
		INSERT INTO customers (id, tenant_id, name, deleted_at) VALUES
			(20, ?, 'Archived Customer', ?);
		INSERT INTO campaigns (id, tenant_id, name) VALUES
			(20, ?, '2025-2026');
		INSERT INTO archive_batches (id, tenant_id, root_entity, root_id, action, created_at) VALUES
			(20, ?, 'projects', 20, 'archive', ?);
		INSERT INTO projects (
			id, tenant_id, name, customer_id, campaign_id, admin_cost, planned_cost,
			created_at, updated_at, deleted_at, archive_batch_id, archive_origin_entity, archive_origin_id
		) VALUES
			(20, ?, 'Project A', 20, 20, 0, 0, ?, ?, ?, 20, 'projects', 20);
	`, tenantID.String(), now,
		tenantID.String(),
		tenantID.String(), now,
		tenantID.String(), now, now, now,
	).Error; err != nil {
		t.Fatalf("seed archived project: %v", err)
	}

	if err := repo.RestoreProject(projectTenantContext(tenantID), 20); err == nil {
		t.Fatalf("expected restore project to fail when customer is archived")
	}
}

func TestHardDeleteProjectRequiresArchivedState(t *testing.T) {
	db := setupProjectTenantDB(t)
	repo := NewRepository(projectTenantGormEngine{client: db})

	tenantID := uuid.New()
	now := time.Now().UTC()
	if err := db.Exec(`
		INSERT INTO customers (id, tenant_id, name, deleted_at) VALUES
			(30, ?, 'Customer A', NULL);
		INSERT INTO campaigns (id, tenant_id, name) VALUES
			(30, ?, '2025-2026');
		INSERT INTO projects (id, tenant_id, name, customer_id, campaign_id, admin_cost, planned_cost, created_at, updated_at, deleted_at) VALUES
			(30, ?, 'Active Project', 30, 30, 0, 0, ?, ?, NULL),
			(31, ?, 'Archived Project', 30, 30, 0, 0, ?, ?, ?);
	`, tenantID.String(),
		tenantID.String(),
		tenantID.String(), now, now,
		tenantID.String(), now, now, now,
	).Error; err != nil {
		t.Fatalf("seed projects: %v", err)
	}

	if err := repo.HardDeleteProject(projectTenantContext(tenantID), 30); err == nil {
		t.Fatalf("expected hard delete active project to fail")
	}
	if err := repo.HardDeleteProject(projectTenantContext(tenantID), 31); err != nil {
		t.Fatalf("hard delete archived project: %v", err)
	}
	var exists int64
	if err := db.Raw(`SELECT COUNT(*) FROM projects WHERE id = 31`).Scan(&exists).Error; err != nil {
		t.Fatalf("count hard deleted project: %v", err)
	}
	if exists != 0 {
		t.Fatalf("expected archived project to be hard deleted")
	}
}

func TestProjectEnsureCropScopesByTenant(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE crops (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT,
			name TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		);
	`).Error; err != nil {
		t.Fatalf("create crops schema: %v", err)
	}

	tenantA := uuid.New()
	tenantB := uuid.New()
	if err := db.Exec(`INSERT INTO crops (id, tenant_id, name) VALUES (1, ?, 'Soja'), (2, ?, 'Soja')`, tenantA.String(), tenantB.String()).Error; err != nil {
		t.Fatalf("seed crops: %v", err)
	}

	err = db.WithContext(projectTenantContext(tenantA)).Transaction(func(tx *gorm.DB) error {
		id, err := ensureCrop(tx, &cropmod.Crop{Name: "Soja"})
		if err != nil {
			return err
		}
		if id != 1 {
			t.Fatalf("expected tenant A crop id 1, got %d", id)
		}

		id, err = ensureCrop(tx, &cropmod.Crop{ID: 2, Name: "Soja"})
		if err != nil {
			return err
		}
		if id == 2 {
			t.Fatalf("must not reuse cross-tenant crop by explicit id")
		}
		if id != 1 {
			t.Fatalf("expected fallback to tenant A crop id 1, got %d", id)
		}

		id, err = ensureCrop(tx, &cropmod.Crop{Name: "Maiz"})
		if err != nil {
			return err
		}
		if id == 0 {
			t.Fatalf("expected new tenant-scoped crop id")
		}
		var createdTenant string
		if err := tx.Raw(`SELECT tenant_id FROM crops WHERE id = ?`, id).Scan(&createdTenant).Error; err != nil {
			return err
		}
		if createdTenant != tenantA.String() {
			t.Fatalf("expected created crop tenant %s, got %s", tenantA, createdTenant)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("ensure crop tenant transaction: %v", err)
	}
}
