package workorder

import (
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/devpablocristo/ponti-backend/internal/work-order/usecases/domain"
)

// newArchivedRefsTestDB sets up an in-memory sqlite with only the tables that
// `assertWorkOrderReferencesActive` looks at. Seeds active rows by default;
// callers archive specific IDs in their test setup.
func newArchivedRefsTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	tables := []string{"projects", "fields", "lots", "crops", "labors", "supplies", "investors"}
	for _, table := range tables {
		if err := db.Exec(`CREATE TABLE ` + table + ` (
			id INTEGER PRIMARY KEY,
			deleted_at DATETIME
		);`).Error; err != nil {
			t.Fatalf("create %s: %v", table, err)
		}
		// Seed one active row per table at ID=1 — the happy-path uses these.
		if err := db.Exec(`INSERT INTO ` + table + ` (id, deleted_at) VALUES (1, NULL);`).Error; err != nil {
			t.Fatalf("seed %s: %v", table, err)
		}
		// Seed one archived row per table at ID=99 — failure-path uses these.
		if err := db.Exec(`INSERT INTO `+table+` (id, deleted_at) VALUES (99, ?);`, time.Now()).Error; err != nil {
			t.Fatalf("seed archived %s: %v", table, err)
		}
	}
	return db
}

func activeWorkOrder() *domain.WorkOrder {
	return &domain.WorkOrder{
		ProjectID:  1,
		FieldID:    1,
		LotID:      1,
		CropID:     1,
		LaborID:    1,
		InvestorID: 1,
		Items: []domain.WorkOrderItem{
			{SupplyID: 1},
		},
	}
}

func TestAssertWorkOrderReferencesActive_AllActive(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	if err := assertWorkOrderReferencesActive(db, activeWorkOrder()); err != nil {
		t.Fatalf("expected nil error for all-active references, got %v", err)
	}
}

func TestAssertWorkOrderReferencesActive_RejectsArchivedByEntity(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		mutate    func(*domain.WorkOrder)
		wantLabel string
	}{
		{
			name:      "archived project",
			mutate:    func(o *domain.WorkOrder) { o.ProjectID = 99 },
			wantLabel: "project",
		},
		{
			name:      "archived field",
			mutate:    func(o *domain.WorkOrder) { o.FieldID = 99 },
			wantLabel: "field",
		},
		{
			name:      "archived lot",
			mutate:    func(o *domain.WorkOrder) { o.LotID = 99 },
			wantLabel: "lot",
		},
		{
			name:      "archived crop",
			mutate:    func(o *domain.WorkOrder) { o.CropID = 99 },
			wantLabel: "crop",
		},
		{
			name:      "archived labor",
			mutate:    func(o *domain.WorkOrder) { o.LaborID = 99 },
			wantLabel: "labor",
		},
		{
			name:      "archived investor (header)",
			mutate:    func(o *domain.WorkOrder) { o.InvestorID = 99 },
			wantLabel: "investor",
		},
		{
			name: "archived investor (split)",
			mutate: func(o *domain.WorkOrder) {
				o.InvestorSplits = []domain.WorkOrderInvestorSplit{{InvestorID: 99}}
			},
			wantLabel: "investor",
		},
		{
			name: "archived supply (item)",
			mutate: func(o *domain.WorkOrder) {
				o.Items = []domain.WorkOrderItem{{SupplyID: 99}}
			},
			wantLabel: "supply",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			db := newArchivedRefsTestDB(t)

			wo := activeWorkOrder()
			tc.mutate(wo)

			err := assertWorkOrderReferencesActive(db, wo)
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

func TestAssertWorkOrderReferencesActive_IgnoresZeroIDs(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	// All zeros — caller hasn't picked any references yet. Should not error;
	// upstream validators (validateItems, validateInvestorSplits) own the
	// "id must be > 0" rule.
	if err := assertWorkOrderReferencesActive(db, &domain.WorkOrder{}); err != nil {
		t.Fatalf("expected nil error for zero IDs, got %v", err)
	}
}

func TestAssertWorkOrderReferencesActive_NilSafe(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	if err := assertWorkOrderReferencesActive(db, nil); err != nil {
		t.Fatalf("expected nil error for nil work order, got %v", err)
	}
}
