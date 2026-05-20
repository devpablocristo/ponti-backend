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

// TestUpdateProjectPropagatesRenameToLegacyTables verifies that when the FE
// sends a project payload where customer/manager/investor names changed but
// their IDs stayed the same, the BE updates the name columns in `customers`,
// `managers` and `investors`. (Actor table propagation is verified in
// integration tests against postgres; in sqlite, actorSyncDisabled() short-
// circuits the actor sync paths.)
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

	var customerName string
	if err := db.Raw(`SELECT name FROM customers WHERE id = 100`).Scan(&customerName).Error; err != nil {
		t.Fatalf("read customer 100: %v", err)
	}
	if customerName != "AGRO LAJITAS" {
		t.Fatalf("expected customer 100 renamed to AGRO LAJITAS, got %q", customerName)
	}

	var managerName string
	if err := db.Raw(`SELECT name FROM managers WHERE id = 100`).Scan(&managerName).Error; err != nil {
		t.Fatalf("read manager 100: %v", err)
	}
	if managerName != "PEPITO" {
		t.Fatalf("expected manager 100 renamed to PEPITO, got %q", managerName)
	}

	var investorName string
	if err := db.Raw(`SELECT name FROM investors WHERE id = 100`).Scan(&investorName).Error; err != nil {
		t.Fatalf("read investor 100: %v", err)
	}
	if investorName != "OLEGA SA RENAMED" {
		t.Fatalf("expected investor 100 renamed to OLEGA SA RENAMED, got %q", investorName)
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
