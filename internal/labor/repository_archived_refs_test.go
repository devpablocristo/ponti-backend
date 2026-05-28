package labor

import (
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/devpablocristo/ponti-backend/internal/labor/usecases/domain"
)

func newArchivedRefsTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	tables := []string{"projects", "categories"}
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

func activeLabor() *domain.Labor {
	return &domain.Labor{
		ID:         0,
		Name:       "Labor Test",
		ProjectId:  1,
		CategoryId: 1,
	}
}

func TestAssertLaborReferencesActive_AllActive(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	if err := assertLaborReferencesActive(db, activeLabor()); err != nil {
		t.Fatalf("expected nil error for all-active references, got %v", err)
	}
}

func TestAssertLaborReferencesActive_RejectsArchivedByEntity(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		mutate    func(*domain.Labor)
		wantLabel string
	}{
		{"archived project", func(l *domain.Labor) { l.ProjectId = 99 }, "project"},
		{"archived category", func(l *domain.Labor) { l.CategoryId = 99 }, "category"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			db := newArchivedRefsTestDB(t)

			l := activeLabor()
			tc.mutate(l)

			err := assertLaborReferencesActive(db, l)
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

func TestAssertLaborReferencesActive_IgnoresZeroIDs(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	if err := assertLaborReferencesActive(db, &domain.Labor{}); err != nil {
		t.Fatalf("expected nil error for zero IDs, got %v", err)
	}
}

func TestAssertLaborReferencesActive_NilSafe(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	if err := assertLaborReferencesActive(db, nil); err != nil {
		t.Fatalf("expected nil error for nil labor, got %v", err)
	}
}
