package lot

import (
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	cropdom "github.com/devpablocristo/ponti-backend/internal/crop/usecases/domain"
	domain "github.com/devpablocristo/ponti-backend/internal/lot/usecases/domain"
)

func newArchivedRefsTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	tables := []string{"fields", "crops"}
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

func activeLot() *domain.Lot {
	return &domain.Lot{
		ID:           0,
		FieldID:      1,
		PreviousCrop: cropdom.Crop{ID: 1},
		CurrentCrop:  cropdom.Crop{ID: 1},
	}
}

func TestAssertLotReferencesActive_AllActive(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	if err := assertLotReferencesActive(db, activeLot()); err != nil {
		t.Fatalf("expected nil error for all-active references, got %v", err)
	}
}

func TestAssertLotReferencesActive_RejectsArchivedByEntity(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		mutate    func(*domain.Lot)
		wantLabel string
	}{
		{"archived field", func(l *domain.Lot) { l.FieldID = 99 }, "field"},
		{"archived previous crop", func(l *domain.Lot) { l.PreviousCrop.ID = 99 }, "crop"},
		{"archived current crop", func(l *domain.Lot) { l.CurrentCrop.ID = 99 }, "crop"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			db := newArchivedRefsTestDB(t)

			l := activeLot()
			tc.mutate(l)

			err := assertLotReferencesActive(db, l)
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

func TestAssertLotReferencesActive_IgnoresZeroIDs(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	if err := assertLotReferencesActive(db, &domain.Lot{}); err != nil {
		t.Fatalf("expected nil error for zero IDs, got %v", err)
	}
}

func TestAssertLotReferencesActive_NilSafe(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	if err := assertLotReferencesActive(db, nil); err != nil {
		t.Fatalf("expected nil error for nil lot, got %v", err)
	}
}
