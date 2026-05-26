package project

import (
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	campdom "github.com/devpablocristo/ponti-backend/internal/campaign/usecases/domain"
	cropdom "github.com/devpablocristo/ponti-backend/internal/crop/usecases/domain"
	customerdom "github.com/devpablocristo/ponti-backend/internal/customer/usecases/domain"
	fielddom "github.com/devpablocristo/ponti-backend/internal/field/usecases/domain"
	invdom "github.com/devpablocristo/ponti-backend/internal/investor/usecases/domain"
	lotdom "github.com/devpablocristo/ponti-backend/internal/lot/usecases/domain"
	mandom "github.com/devpablocristo/ponti-backend/internal/manager/usecases/domain"
	domain "github.com/devpablocristo/ponti-backend/internal/project/usecases/domain"
)

// newArchivedRefsTestDB seeds an in-memory sqlite with every table that
// `assertProjectReferencesActive` touches. ID=1 is active; ID=99 is archived.
func newArchivedRefsTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	tables := []string{"customers", "campaigns", "managers", "investors", "fields", "lots", "crops", "actors"}
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

func activeProject() *domain.Project {
	actorID := int64(1)
	return &domain.Project{
		ID:       0,
		Name:     "Proyecto Test",
		Customer: customerdom.Customer{ID: 1, ActorID: &actorID},
		Campaign: campdom.Campaign{ID: 1},
		Managers: []mandom.Manager{
			{ID: 1, ActorID: &actorID},
		},
		Investors: []invdom.Investor{
			{ID: 1, ActorID: &actorID},
		},
		AdminCostInvestors: []invdom.Investor{
			{ID: 1, ActorID: &actorID},
		},
		Fields: []fielddom.Field{
			{
				ID: 1,
				Investors: []invdom.Investor{
					{ID: 1, ActorID: &actorID},
				},
				Lots: []lotdom.Lot{
					{
						ID:           1,
						CurrentCrop:  cropdom.Crop{ID: 1},
						PreviousCrop: cropdom.Crop{ID: 1},
					},
				},
			},
		},
	}
}

func TestAssertProjectReferencesActive_AllActive(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	if err := assertProjectReferencesActive(db, activeProject()); err != nil {
		t.Fatalf("expected nil error for all-active references, got %v", err)
	}
}

func TestAssertProjectReferencesActive_RejectsArchivedByEntity(t *testing.T) {
	t.Parallel()

	archivedActor := int64(99)
	cases := []struct {
		name      string
		mutate    func(*domain.Project)
		wantLabel string
	}{
		{
			name:      "archived customer",
			mutate:    func(p *domain.Project) { p.Customer.ID = 99 },
			wantLabel: "customer",
		},
		{
			name:      "archived customer actor",
			mutate:    func(p *domain.Project) { p.Customer.ActorID = &archivedActor },
			wantLabel: "actor",
		},
		{
			name:      "archived campaign",
			mutate:    func(p *domain.Project) { p.Campaign.ID = 99 },
			wantLabel: "campaign",
		},
		{
			name:      "archived manager",
			mutate:    func(p *domain.Project) { p.Managers[0].ID = 99 },
			wantLabel: "manager",
		},
		{
			name:      "archived manager actor",
			mutate:    func(p *domain.Project) { p.Managers[0].ActorID = &archivedActor },
			wantLabel: "actor",
		},
		{
			name:      "archived investor",
			mutate:    func(p *domain.Project) { p.Investors[0].ID = 99 },
			wantLabel: "investor",
		},
		{
			name:      "archived investor actor",
			mutate:    func(p *domain.Project) { p.Investors[0].ActorID = &archivedActor },
			wantLabel: "actor",
		},
		{
			name:      "archived admin-cost investor",
			mutate:    func(p *domain.Project) { p.AdminCostInvestors[0].ID = 99 },
			wantLabel: "investor",
		},
		{
			name:      "archived field",
			mutate:    func(p *domain.Project) { p.Fields[0].ID = 99 },
			wantLabel: "field",
		},
		{
			name:      "archived field investor",
			mutate:    func(p *domain.Project) { p.Fields[0].Investors[0].ID = 99 },
			wantLabel: "investor",
		},
		{
			name:      "archived lot",
			mutate:    func(p *domain.Project) { p.Fields[0].Lots[0].ID = 99 },
			wantLabel: "lot",
		},
		{
			name:      "archived current crop",
			mutate:    func(p *domain.Project) { p.Fields[0].Lots[0].CurrentCrop.ID = 99 },
			wantLabel: "crop",
		},
		{
			name:      "archived previous crop",
			mutate:    func(p *domain.Project) { p.Fields[0].Lots[0].PreviousCrop.ID = 99 },
			wantLabel: "crop",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			db := newArchivedRefsTestDB(t)

			p := activeProject()
			tc.mutate(p)

			err := assertProjectReferencesActive(db, p)
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

func TestAssertProjectReferencesActive_IgnoresZeroIDs(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	// All zeros — caller has not selected refs yet. Should not error; the
	// ensure*-helpers create new rows for ID == 0.
	if err := assertProjectReferencesActive(db, &domain.Project{}); err != nil {
		t.Fatalf("expected nil error for zero IDs, got %v", err)
	}
}

func TestAssertProjectReferencesActive_NilSafe(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	if err := assertProjectReferencesActive(db, nil); err != nil {
		t.Fatalf("expected nil error for nil project, got %v", err)
	}
}
