package supply

import (
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	classdomain "github.com/devpablocristo/ponti-backend/internal/class-type/usecases/domain"
	"github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
)

func newSupplyArchivedRefsTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	tables := []string{"projects", "categories", "types"}
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

func activeSupply() *domain.Supply {
	return &domain.Supply{
		ID:         0,
		ProjectID:  1,
		CategoryID: 1,
		Type:       classdomain.ClassType{ID: 1},
	}
}

func TestAssertSupplyReferencesActive_AllActive(t *testing.T) {
	t.Parallel()
	db := newSupplyArchivedRefsTestDB(t)

	if err := assertSupplyReferencesActive(db, activeSupply()); err != nil {
		t.Fatalf("expected nil error for all-active references, got %v", err)
	}
}

func TestAssertSupplyReferencesActive_RejectsArchivedByEntity(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		mutate    func(*domain.Supply)
		wantLabel string
	}{
		{"archived project", func(s *domain.Supply) { s.ProjectID = 99 }, "project"},
		{"archived category", func(s *domain.Supply) { s.CategoryID = 99 }, "category"},
		{"archived type", func(s *domain.Supply) { s.Type.ID = 99 }, "type"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			db := newSupplyArchivedRefsTestDB(t)

			s := activeSupply()
			tc.mutate(s)

			err := assertSupplyReferencesActive(db, s)
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

func TestAssertSupplyReferencesActive_IgnoresZeroIDs(t *testing.T) {
	t.Parallel()
	db := newSupplyArchivedRefsTestDB(t)

	if err := assertSupplyReferencesActive(db, &domain.Supply{}); err != nil {
		t.Fatalf("expected nil error for zero IDs, got %v", err)
	}
}

func TestAssertSupplyReferencesActive_NilSafe(t *testing.T) {
	t.Parallel()
	db := newSupplyArchivedRefsTestDB(t)

	if err := assertSupplyReferencesActive(db, nil); err != nil {
		t.Fatalf("expected nil error for nil supply, got %v", err)
	}
}
