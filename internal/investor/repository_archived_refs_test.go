package investor

import (
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	domain "github.com/devpablocristo/ponti-backend/internal/investor/usecases/domain"
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

func activeInvestor() *domain.Investor {
	actorID := int64(1)
	return &domain.Investor{ID: 0, Name: "Inversor Test", ActorID: &actorID}
}

func TestAssertInvestorReferencesActive_AllActive(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	if err := assertInvestorReferencesActive(db, activeInvestor()); err != nil {
		t.Fatalf("expected nil error for all-active references, got %v", err)
	}
}

func TestAssertInvestorReferencesActive_RejectsArchivedActor(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	archived := int64(99)
	inv := activeInvestor()
	inv.ActorID = &archived

	err := assertInvestorReferencesActive(db, inv)
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

func TestAssertInvestorReferencesActive_IgnoresNilActor(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	inv := &domain.Investor{ID: 0, Name: "Sin actor", ActorID: nil}
	if err := assertInvestorReferencesActive(db, inv); err != nil {
		t.Fatalf("expected nil error for nil actor_id, got %v", err)
	}
}

func TestAssertInvestorReferencesActive_NilSafe(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	if err := assertInvestorReferencesActive(db, nil); err != nil {
		t.Fatalf("expected nil error for nil investor, got %v", err)
	}
}
