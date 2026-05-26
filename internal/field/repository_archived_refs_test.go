package field

import (
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	domain "github.com/devpablocristo/ponti-backend/internal/field/usecases/domain"
	leasetypedom "github.com/devpablocristo/ponti-backend/internal/lease-type/usecases/domain"
	lotdom "github.com/devpablocristo/ponti-backend/internal/lot/usecases/domain"
	cropdom "github.com/devpablocristo/ponti-backend/internal/crop/usecases/domain"
)

func newArchivedRefsTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	tables := []string{"projects", "lease_types", "crops"}
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

func activeField() *domain.Field {
	return &domain.Field{
		ID:        0,
		ProjectID: 1,
		Name:      "Campo Test",
		LeaseType: &leasetypedom.LeaseType{ID: 1},
		Lots: []lotdom.Lot{
			{
				ID:           0,
				CurrentCrop:  cropdom.Crop{ID: 1},
				PreviousCrop: cropdom.Crop{ID: 1},
			},
		},
	}
}

func TestAssertFieldReferencesActive_AllActive(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	if err := assertFieldReferencesActive(db, activeField()); err != nil {
		t.Fatalf("expected nil error for all-active references, got %v", err)
	}
}

func TestAssertFieldReferencesActive_RejectsArchivedByEntity(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		mutate    func(*domain.Field)
		wantLabel string
	}{
		{"archived project", func(f *domain.Field) { f.ProjectID = 99 }, "project"},
		{"archived lease_type", func(f *domain.Field) { f.LeaseType.ID = 99 }, "lease type"},
		{"archived current crop", func(f *domain.Field) { f.Lots[0].CurrentCrop.ID = 99 }, "crop"},
		{"archived previous crop", func(f *domain.Field) { f.Lots[0].PreviousCrop.ID = 99 }, "crop"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			db := newArchivedRefsTestDB(t)

			f := activeField()
			tc.mutate(f)

			err := assertFieldReferencesActive(db, f)
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

func TestAssertFieldReferencesActive_IgnoresZeroIDs(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	if err := assertFieldReferencesActive(db, &domain.Field{}); err != nil {
		t.Fatalf("expected nil error for zero IDs, got %v", err)
	}
}

func TestAssertFieldReferencesActive_NilSafe(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	if err := assertFieldReferencesActive(db, nil); err != nil {
		t.Fatalf("expected nil error for nil field, got %v", err)
	}
}
