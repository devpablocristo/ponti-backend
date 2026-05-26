package project

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	campdom "github.com/devpablocristo/ponti-backend/internal/campaign/usecases/domain"
	custdom "github.com/devpablocristo/ponti-backend/internal/customer/usecases/domain"
	invdom "github.com/devpablocristo/ponti-backend/internal/investor/usecases/domain"
	mgrdom "github.com/devpablocristo/ponti-backend/internal/manager/usecases/domain"
	domain "github.com/devpablocristo/ponti-backend/internal/project/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
)

// TestUpdateProjectPropagatesRenameToLegacyTables verifies that catalog
// entities referenced from a project (customer, manager, investor) are NOT
// renamed when the FE sends a project payload with a different name but the
// same ID. Renaming any of these belongs to its own dedicated editor — doing
// it inline from the project edit would silently rename a row that other
// projects also reference.
func TestUpdateProjectPropagatesRenameToLegacyTables(t *testing.T) {
	db := setupProjectTenantDB(t)
	repo := NewRepository(projectTenantGormEngine{client: db})

	tenantID := uuid.New()
	now := time.Now().UTC()

	if err := db.Exec(`
		INSERT INTO customers (id, tenant_id, name, deleted_at) VALUES
			(100, ?, 'AGRO LAJITAS 25-26', NULL);
		INSERT INTO campaigns (id, tenant_id, name) VALUES
			(100, ?, '2025-2026');
		INSERT INTO projects (id, tenant_id, name, customer_id, campaign_id, admin_cost, planned_cost, created_at, updated_at, deleted_at) VALUES
			(100, ?, 'LA CONCORDIA', 100, 100, 0, 0, ?, ?, NULL);
		INSERT INTO managers (id, tenant_id, name, created_at, updated_at, deleted_at) VALUES
			(100, ?, 'GERO', ?, ?, NULL);
		INSERT INTO project_managers (tenant_id, project_id, manager_id) VALUES
			(?, 100, 100);
		INSERT INTO investors (id, tenant_id, name) VALUES
			(100, ?, 'OLEGA SA');
		INSERT INTO project_investors (tenant_id, project_id, investor_id, percentage) VALUES
			(?, 100, 100, 100);
	`,
		tenantID.String(),
		tenantID.String(),
		tenantID.String(), now, now,
		tenantID.String(), now, now,
		tenantID.String(),
		tenantID.String(),
		tenantID.String(),
	).Error; err != nil {
		t.Fatalf("seed project graph: %v", err)
	}

	updated := &domain.Project{
		ID:          100,
		Name:        "LA CONCORDIA",
		AdminCost:   decimal.Zero,
		PlannedCost: decimal.Zero,
		Customer: custdom.Customer{
			ID:   100,
			Name: "AGRO LAJITAS",
		},
		Campaign: campdom.Campaign{
			ID:   100,
			Name: "2025-2026",
		},
		Managers: []mgrdom.Manager{
			{ID: 100, Name: "PEPITO"},
		},
		Investors: []invdom.Investor{
			{ID: 100, Name: "OLEGA SA RENAMED", Percentage: 100},
		},
		Base: shareddomain.Base{UpdatedAt: now},
	}

	if err := repo.UpdateProject(projectTenantContext(tenantID), updated); err != nil {
		t.Fatalf("update project: %v", err)
	}

	// Catalog rows referenced by id are NOT renamed from the project edit.
	// The seed values stay untouched even though the payload sent different
	// names. To rename any of these the caller uses the dedicated editor.
	var customerName string
	if err := db.Raw(`SELECT name FROM customers WHERE id = 100`).Scan(&customerName).Error; err != nil {
		t.Fatalf("read customer 100: %v", err)
	}
	if customerName != "AGRO LAJITAS 25-26" {
		t.Fatalf("expected customer 100 untouched (AGRO LAJITAS 25-26), got %q", customerName)
	}

	var managerName string
	if err := db.Raw(`SELECT name FROM managers WHERE id = 100`).Scan(&managerName).Error; err != nil {
		t.Fatalf("read manager 100: %v", err)
	}
	if managerName != "GERO" {
		t.Fatalf("expected manager 100 untouched (GERO), got %q", managerName)
	}

	var investorName string
	if err := db.Raw(`SELECT name FROM investors WHERE id = 100`).Scan(&investorName).Error; err != nil {
		t.Fatalf("read investor 100: %v", err)
	}
	if investorName != "OLEGA SA" {
		t.Fatalf("expected investor 100 untouched (OLEGA SA), got %q", investorName)
	}
}

// TestUpdateProjectCanonicalizesNames verifies that catalog entities
// referenced from a project (customer, manager) are NOT renamed by the
// project edit path even when the payload sends a differently-cased value.
// The rename test (in their own editor) handles canonicalization.
func TestUpdateProjectCanonicalizesNames(t *testing.T) {
	db := setupProjectTenantDB(t)
	repo := NewRepository(projectTenantGormEngine{client: db})

	tenantID := uuid.New()
	now := time.Now().UTC()

	if err := db.Exec(`
		INSERT INTO customers (id, tenant_id, name, deleted_at) VALUES
			(800, ?, 'agro lajitas s r l', NULL);
		INSERT INTO campaigns (id, tenant_id, name) VALUES
			(800, ?, '2025-2026');
		INSERT INTO projects (id, tenant_id, name, customer_id, campaign_id, admin_cost, planned_cost, created_at, updated_at, deleted_at) VALUES
			(800, ?, 'la concordia', 800, 800, 0, 0, ?, ?, NULL);
		INSERT INTO managers (id, tenant_id, name, created_at, updated_at, deleted_at) VALUES
			(800, ?, 'gero', ?, ?, NULL);
		INSERT INTO project_managers (tenant_id, project_id, manager_id) VALUES
			(?, 800, 800);
	`,
		tenantID.String(),
		tenantID.String(),
		tenantID.String(), now, now,
		tenantID.String(), now, now,
		tenantID.String(),
	).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	updated := &domain.Project{
		ID:          800,
		Name:        "la concordia",
		AdminCost:   decimal.Zero,
		PlannedCost: decimal.Zero,
		Customer:    custdom.Customer{ID: 800, Name: "  AGRO LAJITAS S.R.L.  "},
		Campaign:    campdom.Campaign{ID: 800, Name: "2025-2026"},
		Managers: []mgrdom.Manager{
			{ID: 800, Name: "María Ángeles"},
		},
		Base: shareddomain.Base{UpdatedAt: now},
	}
	if err := repo.UpdateProject(projectTenantContext(tenantID), updated); err != nil {
		t.Fatalf("update project: %v", err)
	}

	var customerName string
	if err := db.Raw(`SELECT name FROM customers WHERE id = 800`).Scan(&customerName).Error; err != nil {
		t.Fatalf("read customer 800: %v", err)
	}
	if customerName != "agro lajitas s r l" {
		t.Fatalf("expected canonical customer name, got %q", customerName)
	}

	var managerName string
	if err := db.Raw(`SELECT name FROM managers WHERE id = 800`).Scan(&managerName).Error; err != nil {
		t.Fatalf("read manager 800: %v", err)
	}
	if managerName != "gero" {
		t.Fatalf("expected manager 800 untouched (gero), got %q", managerName)
	}
}

func int64Ptr(v int64) *int64 { return &v }

// Case 1: create project with a brand-new customer (no id, no actor_id).
// Expected: the BE creates the customer in `customers` (and an actor in
// `actors` under postgres, here only legacy is asserted because sqlite
// disables the actor sync).
func TestCreateProjectCreatesCustomerWithoutId(t *testing.T) {
	db := setupProjectTenantDB(t)
	repo := NewRepository(projectTenantGormEngine{client: db})

	tenantID := uuid.New()
	now := time.Now().UTC()
	if err := db.Exec(`
		INSERT INTO campaigns (id, tenant_id, name) VALUES (300, ?, '2025-2026');
	`, tenantID.String()).Error; err != nil {
		t.Fatalf("seed campaign: %v", err)
	}

	pid, err := repo.CreateProject(projectTenantContext(tenantID), &domain.Project{
		Name:        "PROYECTO NUEVO",
		AdminCost:   decimal.Zero,
		PlannedCost: decimal.Zero,
		Customer:    custdom.Customer{Name: "CLIENTE NUEVO"},
		Campaign:    campdom.Campaign{ID: 300, Name: "2025-2026"},
		Base:        shareddomain.Base{CreatedAt: now, UpdatedAt: now},
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if pid <= 0 {
		t.Fatalf("expected project id > 0, got %d", pid)
	}

	var custID int64
	if err := db.Raw(`SELECT customer_id FROM projects WHERE id = ?`, pid).Scan(&custID).Error; err != nil {
		t.Fatalf("read project customer_id: %v", err)
	}
	if custID <= 0 {
		t.Fatalf("expected project linked to a customer, got customer_id=%d", custID)
	}
	var custName string
	if err := db.Raw(`SELECT name FROM customers WHERE id = ?`, custID).Scan(&custName).Error; err != nil {
		t.Fatalf("read customer name: %v", err)
	}
	if custName != "CLIENTE NUEVO" {
		t.Fatalf("expected customer name CLIENTE NUEVO, got %q", custName)
	}
}

// Case 4: swap the project's customer to a different one by sending a new
// customer.id in the PUT payload.
func TestUpdateProjectSwapsCustomer(t *testing.T) {
	db := setupProjectTenantDB(t)
	repo := NewRepository(projectTenantGormEngine{client: db})

	tenantID := uuid.New()
	now := time.Now().UTC()
	if err := db.Exec(`
		INSERT INTO customers (id, tenant_id, name, deleted_at) VALUES
			(400, ?, 'OLD CUSTOMER', NULL),
			(401, ?, 'NEW CUSTOMER', NULL);
		INSERT INTO campaigns (id, tenant_id, name) VALUES (400, ?, '2025-2026');
		INSERT INTO projects (id, tenant_id, name, customer_id, campaign_id, admin_cost, planned_cost, created_at, updated_at, deleted_at) VALUES
			(400, ?, 'SWAP TEST', 400, 400, 0, 0, ?, ?, NULL);
	`,
		tenantID.String(), tenantID.String(),
		tenantID.String(),
		tenantID.String(), now, now,
	).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	if err := repo.UpdateProject(projectTenantContext(tenantID), &domain.Project{
		ID:          400,
		Name:        "SWAP TEST",
		AdminCost:   decimal.Zero,
		PlannedCost: decimal.Zero,
		Customer:    custdom.Customer{ID: 401, Name: "NEW CUSTOMER"},
		Campaign:    campdom.Campaign{ID: 400, Name: "2025-2026"},
		Base:        shareddomain.Base{UpdatedAt: now},
	}); err != nil {
		t.Fatalf("update project: %v", err)
	}

	var custID int64
	if err := db.Raw(`SELECT customer_id FROM projects WHERE id = 400`).Scan(&custID).Error; err != nil {
		t.Fatalf("read project customer_id: %v", err)
	}
	if custID != 401 {
		t.Fatalf("expected project customer_id swapped to 401, got %d", custID)
	}
}

// Case 5: create project with a new manager that has no actor_id (legacy
// flow). Expected: a row in `managers` is created with the requested name.
func TestCreateProjectCreatesManagerWithoutActorID(t *testing.T) {
	db := setupProjectTenantDB(t)
	repo := NewRepository(projectTenantGormEngine{client: db})

	tenantID := uuid.New()
	now := time.Now().UTC()
	if err := db.Exec(`
		INSERT INTO customers (id, tenant_id, name, deleted_at) VALUES (500, ?, 'CUSTOMER', NULL);
		INSERT INTO campaigns (id, tenant_id, name) VALUES (500, ?, '2025-2026');
	`, tenantID.String(), tenantID.String()).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	pid, err := repo.CreateProject(projectTenantContext(tenantID), &domain.Project{
		Name:        "PROYECTO CON MANAGER NUEVO",
		AdminCost:   decimal.Zero,
		PlannedCost: decimal.Zero,
		Customer:    custdom.Customer{ID: 500, Name: "CUSTOMER"},
		Campaign:    campdom.Campaign{ID: 500, Name: "2025-2026"},
		Managers: []mgrdom.Manager{
			{Name: "MANAGER NUEVO"},
		},
		Base: shareddomain.Base{CreatedAt: now, UpdatedAt: now},
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	var managerName string
	if err := db.Raw(`
		SELECT m.name FROM managers m
		JOIN project_managers pm ON pm.manager_id = m.id
		WHERE pm.project_id = ?
	`, pid).Scan(&managerName).Error; err != nil {
		t.Fatalf("read manager via project: %v", err)
	}
	if managerName != "MANAGER NUEVO" {
		t.Fatalf("expected manager name MANAGER NUEVO, got %q", managerName)
	}
}

// Case 7: update project swapping the manager slot to a different one
// referenced only by actor_id. The FE clears manager.id and sets actor_id,
// and the BE resolves the legacy row via legacy_actor_map.
func TestUpdateProjectSwapsManagerByActorID(t *testing.T) {
	db := setupProjectTenantDB(t)
	repo := NewRepository(projectTenantGormEngine{client: db})

	tenantID := uuid.New()
	now := time.Now().UTC()
	if err := db.Exec(`
		INSERT INTO customers (id, tenant_id, name, deleted_at) VALUES (700, ?, 'CUSTOMER', NULL);
		INSERT INTO campaigns (id, tenant_id, name) VALUES (700, ?, '2025-2026');
		INSERT INTO projects (id, tenant_id, name, customer_id, campaign_id, admin_cost, planned_cost, created_at, updated_at, deleted_at) VALUES
			(700, ?, 'SWAP MGR', 700, 700, 0, 0, ?, ?, NULL);
		INSERT INTO managers (id, tenant_id, name, created_at, updated_at, deleted_at) VALUES
			(700, ?, 'OLD MGR', ?, ?, NULL),
			(701, ?, 'NEW MGR', ?, ?, NULL);
		INSERT INTO project_managers (tenant_id, project_id, manager_id) VALUES
			(?, 700, 700);
		INSERT INTO legacy_actor_map (tenant_id, source_table, source_id, source_key, source_text, actor_id, confidence, mapping_status) VALUES
			(?, 'managers', 700, '700', 'OLD MGR', 9001, 1.0, 'auto_matched'),
			(?, 'managers', 701, '701', 'NEW MGR', 9002, 1.0, 'auto_matched');
		INSERT INTO actors (id, tenant_id, deleted_at) VALUES
			(9001, ?, NULL),
			(9002, ?, NULL);
	`,
		tenantID.String(),
		tenantID.String(),
		tenantID.String(), now, now,
		tenantID.String(), now, now,
		tenantID.String(), now, now,
		tenantID.String(),
		tenantID.String(),
		tenantID.String(),
		tenantID.String(),
		tenantID.String(),
	).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	if err := repo.UpdateProject(projectTenantContext(tenantID), &domain.Project{
		ID:          700,
		Name:        "SWAP MGR",
		AdminCost:   decimal.Zero,
		PlannedCost: decimal.Zero,
		Customer:    custdom.Customer{ID: 700, Name: "CUSTOMER"},
		Campaign:    campdom.Campaign{ID: 700, Name: "2025-2026"},
		Managers: []mgrdom.Manager{
			{ID: 0, ActorID: int64Ptr(9002), Name: "NEW MGR"},
		},
		Base: shareddomain.Base{UpdatedAt: now},
	}); err != nil {
		t.Fatalf("update project: %v", err)
	}

	var linkedManagerID int64
	if err := db.Raw(`SELECT manager_id FROM project_managers WHERE project_id = 700`).Scan(&linkedManagerID).Error; err != nil {
		t.Fatalf("read linked manager: %v", err)
	}
	if linkedManagerID != 701 {
		t.Fatalf("expected project_managers.manager_id=701 after swap, got %d", linkedManagerID)
	}
}

// TestGetProjectHydratesActorIDFromLegacyMap verifies that when a manager
// has a row in legacy_actor_map but no direct actor_id column (which is the
// default for managers/investors before this hydration was added), GetProject
// fills the in-memory ActorID so the FE editor can render the slot with its
// real identity and the duplicate-name guard does not fire false positives.
func TestGetProjectHydratesActorIDFromLegacyMap(t *testing.T) {
	db := setupProjectTenantDB(t)
	repo := NewRepository(projectTenantGormEngine{client: db})

	tenantID := uuid.New()
	now := time.Now().UTC()

	if err := db.Exec(`
		INSERT INTO customers (id, tenant_id, name, actor_id, deleted_at) VALUES
			(200, ?, 'AGRO HYDRA', 700, NULL);
		INSERT INTO campaigns (id, tenant_id, name) VALUES
			(200, ?, '2025-2026');
		INSERT INTO projects (id, tenant_id, name, customer_id, campaign_id, admin_cost, planned_cost, created_at, updated_at, deleted_at) VALUES
			(200, ?, 'HYDRA', 200, 200, 0, 0, ?, ?, NULL);
		INSERT INTO managers (id, tenant_id, name, created_at, updated_at, deleted_at) VALUES
			(200, ?, 'GERO', ?, ?, NULL);
		INSERT INTO project_managers (tenant_id, project_id, manager_id) VALUES
			(?, 200, 200);
		INSERT INTO investors (id, tenant_id, name) VALUES
			(200, ?, 'OLEGA SA');
		INSERT INTO project_investors (tenant_id, project_id, investor_id, percentage) VALUES
			(?, 200, 200, 100);
		INSERT INTO legacy_actor_map (tenant_id, source_table, source_id, source_key, source_text, actor_id, confidence, mapping_status) VALUES
			(?, 'managers', 200, '200', 'GERO', 7001, 1.0, 'auto_matched'),
			(?, 'investors', 200, '200', 'OLEGA SA', 7002, 1.0, 'auto_matched');
	`,
		tenantID.String(),
		tenantID.String(),
		tenantID.String(), now, now,
		tenantID.String(), now, now,
		tenantID.String(),
		tenantID.String(),
		tenantID.String(),
		tenantID.String(),
		tenantID.String(),
	).Error; err != nil {
		t.Fatalf("seed project graph: %v", err)
	}

	p, err := repo.GetProject(projectTenantContext(tenantID), 200)
	if err != nil {
		t.Fatalf("get project: %v", err)
	}
	if len(p.Managers) != 1 {
		t.Fatalf("expected 1 manager, got %d", len(p.Managers))
	}
	if p.Managers[0].ActorID == nil || *p.Managers[0].ActorID != 7001 {
		t.Fatalf("expected manager ActorID 7001, got %v", p.Managers[0].ActorID)
	}
	if len(p.Investors) != 1 {
		t.Fatalf("expected 1 investor, got %d", len(p.Investors))
	}
	if p.Investors[0].ActorID == nil || *p.Investors[0].ActorID != 7002 {
		t.Fatalf("expected investor ActorID 7002, got %v", p.Investors[0].ActorID)
	}
}
