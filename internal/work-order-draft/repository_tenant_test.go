package workorderdraft

import (
	"context"
	"testing"
	"time"

	contextkeys "github.com/devpablocristo/core/security/go/contextkeys"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	sharedtypes "github.com/devpablocristo/ponti-backend/internal/shared/types"
	"github.com/devpablocristo/ponti-backend/internal/work-order-draft/usecases/domain"
)

type draftTenantGormEngine struct {
	client *gorm.DB
}

func (e draftTenantGormEngine) Client() *gorm.DB {
	return e.client
}

func draftTenantContext(tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "tenant-user@example.com")
	ctx = context.WithValue(ctx, contextkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, contextkeys.Role, "tenant_admin")
	ctx = context.WithValue(ctx, contextkeys.Scopes, []string{"workorders.read", "workorders.write", "workorders.archive"})
	return ctx
}

func setupDraftTenantDB(t *testing.T) *gorm.DB {
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
			deleted_at DATETIME
		);
		CREATE TABLE fields (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			project_id INTEGER NOT NULL,
			deleted_at DATETIME
		);
		CREATE TABLE work_order_drafts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			number TEXT NOT NULL,
			date DATETIME NOT NULL,
			customer_id INTEGER NOT NULL,
			project_id INTEGER NOT NULL,
			campaign_id INTEGER,
			field_id INTEGER NOT NULL,
			lot_id INTEGER NOT NULL,
			crop_id INTEGER NOT NULL,
			labor_id INTEGER NOT NULL,
			contractor TEXT NOT NULL,
			effective_area NUMERIC NOT NULL,
			observations TEXT,
			investor_id INTEGER NOT NULL,
			is_digital BOOLEAN DEFAULT false,
			status TEXT NOT NULL,
			reviewed_by INTEGER,
			published_work_order_id INTEGER,
			review_notes TEXT,
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
		CREATE TABLE work_order_draft_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			draft_id INTEGER NOT NULL,
			supply_id INTEGER NOT NULL,
			supply_name TEXT NOT NULL,
			total_used NUMERIC NOT NULL,
			final_dose NUMERIC NOT NULL,
			deleted_at DATETIME,
			deleted_by TEXT,
			archive_batch_id INTEGER,
			archive_origin_entity TEXT,
			archive_origin_id INTEGER,
			archive_reason TEXT
		);
		CREATE TABLE work_order_draft_investor_splits (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			draft_id INTEGER NOT NULL,
			investor_id INTEGER NOT NULL,
			percentage NUMERIC NOT NULL,
			deleted_at DATETIME,
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
	`).Error; err != nil {
		t.Fatalf("create schema: %v", err)
	}

	return db
}

func TestWorkOrderDraftRepositoryTenantIsolation(t *testing.T) {
	db := setupDraftTenantDB(t)
	repo := NewRepository(draftTenantGormEngine{client: db})

	tenantA := uuid.New()
	tenantB := uuid.New()
	now := time.Now().UTC()

	if err := db.Exec(`
		INSERT INTO projects (id, tenant_id, name) VALUES
			(10, ?, 'Project A'),
			(20, ?, 'Project B');
		INSERT INTO fields (id, tenant_id, name, project_id) VALUES
			(100, ?, 'Field A', 10),
			(200, ?, 'Field B', 20);
		INSERT INTO work_order_drafts (
			id, tenant_id, number, date, customer_id, project_id, field_id, lot_id,
			crop_id, labor_id, contractor, effective_area, investor_id, is_digital,
			status, created_at, updated_at
		) VALUES
			(1, ?, 'A-1', ?, 1, 10, 100, 1000, 1, 1, 'Contractor A', 10, 1, true, 'draft', ?, ?),
			(2, ?, 'B-1', ?, 2, 20, 200, 2000, 1, 1, 'Contractor B', 20, 2, true, 'draft', ?, ?)
	`, tenantA.String(), tenantB.String(), tenantA.String(), tenantB.String(),
		tenantA.String(), now, now, now, tenantB.String(), now, now, now).Error; err != nil {
		t.Fatalf("seed drafts: %v", err)
	}

	ctxA := draftTenantContext(tenantA)

	list, page, err := repo.ListWorkOrderDrafts(ctxA, "", "", nil, sharedtypes.Input{Page: 1, PageSize: 50})
	if err != nil {
		t.Fatalf("list drafts: %v", err)
	}
	if page.Total != 1 || len(list) != 1 || list[0].ID != 1 {
		t.Fatalf("expected only tenant A draft, page=%#v list=%#v", page, list)
	}

	if _, err := repo.GetWorkOrderDraftByID(ctxA, 2); err == nil {
		t.Fatalf("expected get cross-tenant draft to fail")
	}

	if err := repo.UpdateWorkOrderDraftByID(ctxA, &domain.WorkOrderDraft{
		ID:            2,
		Number:        "cross tenant update",
		Date:          now,
		CustomerID:    2,
		ProjectID:     20,
		FieldID:       200,
		LotID:         2000,
		CropID:        1,
		LaborID:       1,
		Contractor:    "Contractor B",
		EffectiveArea: decimal.NewFromInt(20),
		InvestorID:    2,
		Status:        domain.StatusDraft,
	}); err == nil {
		t.Fatalf("expected update cross-tenant draft to fail")
	}

	var number string
	if err := db.Raw(`SELECT number FROM work_order_drafts WHERE id = 2`).Scan(&number).Error; err != nil {
		t.Fatalf("read draft 2: %v", err)
	}
	if number != "B-1" {
		t.Fatalf("cross-tenant update changed draft 2 number to %q", number)
	}

	if err := repo.MarkWorkOrderDraftAsPublished(ctxA, 2, 99); err == nil {
		t.Fatalf("expected mark-published cross-tenant draft to fail")
	}
	var status string
	if err := db.Raw(`SELECT status FROM work_order_drafts WHERE id = 2`).Scan(&status).Error; err != nil {
		t.Fatalf("read draft status: %v", err)
	}
	if status != "draft" {
		t.Fatalf("cross-tenant mark published changed draft 2 status to %q", status)
	}

	if err := repo.HardDeleteWorkOrderDraftByID(ctxA, 2); err == nil {
		t.Fatalf("expected hard-delete cross-tenant draft to fail")
	}
	var exists int64
	if err := db.Raw(`SELECT COUNT(*) FROM work_order_drafts WHERE id = 2`).Scan(&exists).Error; err != nil {
		t.Fatalf("check hard delete side effect: %v", err)
	}
	if exists != 1 {
		t.Fatalf("cross-tenant hard delete removed draft 2")
	}
}

func TestWorkOrderDraftArchiveRestoreAndHardDelete(t *testing.T) {
	db := setupDraftTenantDB(t)
	repo := NewRepository(draftTenantGormEngine{client: db})

	tenantID := uuid.New()
	now := time.Now().UTC()
	if err := db.Exec(`
		INSERT INTO projects (id, tenant_id, name, deleted_at) VALUES
			(10, ?, 'Project A', NULL);
		INSERT INTO fields (id, tenant_id, name, project_id, deleted_at) VALUES
			(100, ?, 'Field A', 10, NULL);
		INSERT INTO work_order_drafts (
			id, tenant_id, number, date, customer_id, project_id, field_id, lot_id,
			crop_id, labor_id, contractor, effective_area, investor_id, is_digital,
			status, created_at, updated_at, deleted_at
		) VALUES
			(10, ?, 'D-10', ?, 1, 10, 100, 1000, 1, 1, 'Contractor A', 10, 1, true, 'draft', ?, ?, NULL);
		INSERT INTO work_order_draft_items (
			id, draft_id, supply_id, supply_name, total_used, final_dose, deleted_at
		) VALUES
			(10, 10, 1, 'Supply A', 1, 1, NULL),
			(11, 10, 2, 'Manual Archived Supply', 1, 1, ?);
		INSERT INTO work_order_draft_investor_splits (
			id, draft_id, investor_id, percentage, deleted_at
		) VALUES
			(10, 10, 1, 100, NULL);
	`, tenantID.String(),
		tenantID.String(),
		tenantID.String(), now, now, now, now,
		now,
	).Error; err != nil {
		t.Fatalf("seed draft: %v", err)
	}

	ctx := draftTenantContext(tenantID)
	if err := repo.ArchiveWorkOrderDraftByID(ctx, 10); err != nil {
		t.Fatalf("archive draft: %v", err)
	}

	var archivedDraft, archivedItem, manualItem int64
	if err := db.Raw(`SELECT COUNT(*) FROM work_order_drafts WHERE id = 10 AND deleted_at IS NOT NULL AND archive_batch_id IS NOT NULL AND archive_origin_entity = 'work_order_drafts' AND archive_origin_id = 10`).Scan(&archivedDraft).Error; err != nil {
		t.Fatalf("count archived draft: %v", err)
	}
	if err := db.Raw(`SELECT COUNT(*) FROM work_order_draft_items WHERE id = 10 AND deleted_at IS NOT NULL AND archive_batch_id IS NOT NULL AND archive_origin_entity = 'work_order_drafts' AND archive_origin_id = 10`).Scan(&archivedItem).Error; err != nil {
		t.Fatalf("count archived draft item: %v", err)
	}
	if err := db.Raw(`SELECT COUNT(*) FROM work_order_draft_items WHERE id = 11 AND deleted_at IS NOT NULL AND archive_batch_id IS NULL`).Scan(&manualItem).Error; err != nil {
		t.Fatalf("count manual draft item: %v", err)
	}
	if archivedDraft != 1 || archivedItem != 1 || manualItem != 1 {
		t.Fatalf("unexpected archive state draft=%d item=%d manual=%d", archivedDraft, archivedItem, manualItem)
	}

	if err := repo.RestoreWorkOrderDraftByID(ctx, 10); err != nil {
		t.Fatalf("restore draft: %v", err)
	}

	var restoredDraft, restoredItem, manualStillArchived int64
	if err := db.Raw(`SELECT COUNT(*) FROM work_order_drafts WHERE id = 10 AND deleted_at IS NULL AND archive_batch_id IS NULL`).Scan(&restoredDraft).Error; err != nil {
		t.Fatalf("count restored draft: %v", err)
	}
	if err := db.Raw(`SELECT COUNT(*) FROM work_order_draft_items WHERE id = 10 AND deleted_at IS NULL AND archive_batch_id IS NULL`).Scan(&restoredItem).Error; err != nil {
		t.Fatalf("count restored draft item: %v", err)
	}
	if err := db.Raw(`SELECT COUNT(*) FROM work_order_draft_items WHERE id = 11 AND deleted_at IS NOT NULL`).Scan(&manualStillArchived).Error; err != nil {
		t.Fatalf("count manual still archived: %v", err)
	}
	if restoredDraft != 1 || restoredItem != 1 || manualStillArchived != 1 {
		t.Fatalf("unexpected restore state draft=%d item=%d manual=%d", restoredDraft, restoredItem, manualStillArchived)
	}

	if err := repo.HardDeleteWorkOrderDraftByID(ctx, 10); err == nil {
		t.Fatalf("expected hard delete active draft to fail")
	}
	if err := repo.ArchiveWorkOrderDraftByID(ctx, 10); err != nil {
		t.Fatalf("archive draft before hard delete: %v", err)
	}
	if err := repo.HardDeleteWorkOrderDraftByID(ctx, 10); err != nil {
		t.Fatalf("hard delete archived draft: %v", err)
	}
	var exists int64
	if err := db.Raw(`SELECT COUNT(*) FROM work_order_drafts WHERE id = 10`).Scan(&exists).Error; err != nil {
		t.Fatalf("count hard deleted draft: %v", err)
	}
	if exists != 0 {
		t.Fatalf("expected hard deleted draft removed, got %d", exists)
	}
}

func TestWorkOrderDraftRepositoryRequiresTenantInStrictModeForNumberLookups(t *testing.T) {
	t.Setenv("TENANT_STRICT_MODE", "true")

	db := setupDraftTenantDB(t)
	repo := NewRepository(draftTenantGormEngine{client: db})

	ctx := context.Background()
	if _, err := repo.ListOccupiedWorkOrderNumbersByProject(ctx, 10); err == nil {
		t.Fatalf("expected strict occupied-number lookup without tenant to fail")
	}
	if _, err := repo.ListOccupiedWorkOrderNumbersByProjectExcludingDraft(ctx, 10, 1); err == nil {
		t.Fatalf("expected strict occupied-number excluding draft lookup without tenant to fail")
	}
	if _, err := repo.ListPublishedWorkOrderNumbersByProject(ctx, 10); err == nil {
		t.Fatalf("expected strict published-number lookup without tenant to fail")
	}
}
