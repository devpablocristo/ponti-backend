package workorderdraft

import (
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/devpablocristo/ponti-backend/internal/work-order-draft/usecases/domain"
)

func newArchivedRefsTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	tables := []string{"projects", "fields", "lots", "crops", "labors", "supplies", "investors", "customers", "campaigns"}
	for _, table := range tables {
		if err := db.Exec(`CREATE TABLE ` + table + ` (id INTEGER PRIMARY KEY, deleted_at DATETIME);`).Error; err != nil {
			t.Fatalf("create %s: %v", table, err)
		}
		if err := db.Exec(`INSERT INTO ` + table + ` (id, deleted_at) VALUES (1, NULL);`).Error; err != nil {
			t.Fatalf("seed active %s: %v", table, err)
		}
		if err := db.Exec(`INSERT INTO `+table+` (id, deleted_at) VALUES (99, ?);`, time.Now()).Error; err != nil {
			t.Fatalf("seed archived %s: %v", table, err)
		}
	}
	return db
}

func activeDraft() *domain.WorkOrderDraft {
	campaignID := int64(1)
	return &domain.WorkOrderDraft{
		ProjectID:  1,
		FieldID:    1,
		LotID:      1,
		CropID:     1,
		LaborID:    1,
		CustomerID: 1,
		CampaignID: &campaignID,
		InvestorID: 1,
		Items: []domain.WorkOrderDraftItem{
			{SupplyID: 1},
		},
		InvestorSplits: []domain.WorkOrderDraftInvestorSplit{
			{InvestorID: 1},
		},
	}
}

func TestAssertWorkOrderDraftReferencesActive_AllActive(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	if err := assertWorkOrderDraftReferencesActive(db, activeDraft()); err != nil {
		t.Fatalf("expected nil error for all-active references, got %v", err)
	}
}

func TestAssertWorkOrderDraftReferencesActive_RejectsArchivedByEntity(t *testing.T) {
	t.Parallel()

	archivedCampaign := int64(99)
	cases := []struct {
		name      string
		mutate    func(*domain.WorkOrderDraft)
		wantLabel string
	}{
		{"archived project", func(d *domain.WorkOrderDraft) { d.ProjectID = 99 }, "project"},
		{"archived field", func(d *domain.WorkOrderDraft) { d.FieldID = 99 }, "field"},
		{"archived lot", func(d *domain.WorkOrderDraft) { d.LotID = 99 }, "lot"},
		{"archived crop", func(d *domain.WorkOrderDraft) { d.CropID = 99 }, "crop"},
		{"archived labor", func(d *domain.WorkOrderDraft) { d.LaborID = 99 }, "labor"},
		{"archived customer", func(d *domain.WorkOrderDraft) { d.CustomerID = 99 }, "customer"},
		{"archived campaign", func(d *domain.WorkOrderDraft) { d.CampaignID = &archivedCampaign }, "campaign"},
		{"archived investor (header)", func(d *domain.WorkOrderDraft) { d.InvestorID = 99 }, "investor"},
		{
			name: "archived investor (split)",
			mutate: func(d *domain.WorkOrderDraft) {
				d.InvestorSplits = []domain.WorkOrderDraftInvestorSplit{{InvestorID: 99}}
			},
			wantLabel: "investor",
		},
		{
			name: "archived supply (item)",
			mutate: func(d *domain.WorkOrderDraft) {
				d.Items = []domain.WorkOrderDraftItem{{SupplyID: 99}}
			},
			wantLabel: "supply",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			db := newArchivedRefsTestDB(t)

			d := activeDraft()
			tc.mutate(d)

			err := assertWorkOrderDraftReferencesActive(db, d)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !domainerr.IsConflict(err) {
				t.Fatalf("expected Conflict kind, got %v", err)
			}
			wantMsg := tc.wantLabel + " is archived"
			if !strings.Contains(err.Error(), wantMsg) {
				t.Fatalf("expected message to contain %q, got %q", wantMsg, err.Error())
			}
		})
	}
}

func TestAssertWorkOrderDraftReferencesActive_IgnoresZeroIDs(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	if err := assertWorkOrderDraftReferencesActive(db, &domain.WorkOrderDraft{}); err != nil {
		t.Fatalf("expected nil error for zero IDs, got %v", err)
	}
}

func TestAssertWorkOrderDraftReferencesActive_NilSafe(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	if err := assertWorkOrderDraftReferencesActive(db, nil); err != nil {
		t.Fatalf("expected nil error for nil draft, got %v", err)
	}
}
