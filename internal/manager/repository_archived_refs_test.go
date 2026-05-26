package manager

import (
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	domain "github.com/devpablocristo/ponti-backend/internal/manager/usecases/domain"
)

func newArchivedRefsTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`CREATE TABLE actors (id INTEGER PRIMARY KEY, deleted_at DATETIME);`).Error; err != nil {
		t.Fatalf("create actors: %v", err)
	}
	if err := db.Exec(`INSERT INTO actors (id, deleted_at) VALUES (1, NULL);`).Error; err != nil {
		t.Fatalf("seed active actor: %v", err)
	}
	if err := db.Exec(`INSERT INTO actors (id, deleted_at) VALUES (99, ?);`, time.Now()).Error; err != nil {
		t.Fatalf("seed archived actor: %v", err)
	}
	return db
}

func activeManager() *domain.Manager {
	actorID := int64(1)
	return &domain.Manager{ID: 0, Name: "Pepe Argento", ActorID: &actorID}
}

func TestAssertManagerReferencesActive_AllActive(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	if err := assertManagerReferencesActive(db, activeManager()); err != nil {
		t.Fatalf("expected nil error for all-active references, got %v", err)
	}
}

func TestAssertManagerReferencesActive_RejectsArchivedActor(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	archived := int64(99)
	m := activeManager()
	m.ActorID = &archived

	err := assertManagerReferencesActive(db, m)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !domainerr.IsConflict(err) {
		t.Fatalf("expected Conflict kind, got %v", err)
	}
	if !strings.Contains(err.Error(), "actor is archived") {
		t.Fatalf("expected message to contain %q, got %q", "actor is archived", err.Error())
	}
}

func TestAssertManagerReferencesActive_IgnoresNilActor(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	m := &domain.Manager{ID: 0, Name: "Sin actor", ActorID: nil}
	if err := assertManagerReferencesActive(db, m); err != nil {
		t.Fatalf("expected nil error for nil actor_id, got %v", err)
	}
}

func TestAssertManagerReferencesActive_NilSafe(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	if err := assertManagerReferencesActive(db, nil); err != nil {
		t.Fatalf("expected nil error for nil manager, got %v", err)
	}
}
