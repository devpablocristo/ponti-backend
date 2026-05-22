package supply

import (
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	investordomain "github.com/devpablocristo/ponti-backend/internal/investor/usecases/domain"
	providerdomain "github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
	"github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
)

func newMovementArchivedRefsTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	tables := []string{"projects", "supplies", "investors", "providers", "actors"}
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

func activeSupplyMovement() *domain.SupplyMovement {
	actorID := int64(1)
	return &domain.SupplyMovement{
		ID:                   0,
		ProjectId:            1,
		ProjectDestinationId: 1,
		Supply:               &domain.Supply{ID: 1},
		Investor:             &investordomain.Investor{ID: 1, ActorID: &actorID},
		Provider:             &providerdomain.Provider{ID: 1},
	}
}

func TestAssertSupplyMovementReferencesActive_AllActive(t *testing.T) {
	t.Parallel()
	db := newMovementArchivedRefsTestDB(t)

	if err := assertSupplyMovementReferencesActive(db, activeSupplyMovement()); err != nil {
		t.Fatalf("expected nil error for all-active references, got %v", err)
	}
}

func TestAssertSupplyMovementReferencesActive_RejectsArchivedByEntity(t *testing.T) {
	t.Parallel()

	archivedActor := int64(99)
	cases := []struct {
		name      string
		mutate    func(*domain.SupplyMovement)
		wantLabel string
	}{
		{"archived project", func(m *domain.SupplyMovement) { m.ProjectId = 99 }, "project"},
		{"archived destination project", func(m *domain.SupplyMovement) { m.ProjectDestinationId = 99 }, "destination project"},
		{"archived supply", func(m *domain.SupplyMovement) { m.Supply.ID = 99 }, "supply"},
		{"archived investor", func(m *domain.SupplyMovement) { m.Investor.ID = 99 }, "investor"},
		{"archived investor actor", func(m *domain.SupplyMovement) { m.Investor.ActorID = &archivedActor }, "actor"},
		{"archived provider", func(m *domain.SupplyMovement) { m.Provider.ID = 99 }, "provider"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			db := newMovementArchivedRefsTestDB(t)

			m := activeSupplyMovement()
			tc.mutate(m)

			err := assertSupplyMovementReferencesActive(db, m)
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

func TestAssertSupplyMovementReferencesActive_IgnoresOptionalNil(t *testing.T) {
	t.Parallel()
	db := newMovementArchivedRefsTestDB(t)

	// Movement con sólo project + supply (sin investor/provider/destination).
	m := &domain.SupplyMovement{
		ID:        0,
		ProjectId: 1,
		Supply:    &domain.Supply{ID: 1},
	}
	if err := assertSupplyMovementReferencesActive(db, m); err != nil {
		t.Fatalf("expected nil error when optional refs are nil, got %v", err)
	}
}

func TestAssertSupplyMovementReferencesActive_NilSafe(t *testing.T) {
	t.Parallel()
	db := newMovementArchivedRefsTestDB(t)

	if err := assertSupplyMovementReferencesActive(db, nil); err != nil {
		t.Fatalf("expected nil error for nil movement, got %v", err)
	}
}
