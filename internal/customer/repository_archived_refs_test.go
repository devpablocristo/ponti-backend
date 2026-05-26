package customer

import (
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	domain "github.com/devpablocristo/ponti-backend/internal/customer/usecases/domain"
)

// newArchivedRefsTestDB seeds an in-memory sqlite with the tables that
// `assertCustomerReferencesActive` queries. ID=1 is active, ID=99 is archived.
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

func activeCustomer() *domain.Customer {
	activeActor := int64(1)
	return &domain.Customer{
		ID:      0,
		Name:    "ACME SA",
		ActorID: &activeActor,
	}
}

func TestAssertCustomerReferencesActive_AllActive(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	if err := assertCustomerReferencesActive(db, activeCustomer()); err != nil {
		t.Fatalf("expected nil error for all-active references, got %v", err)
	}
}

func TestAssertCustomerReferencesActive_RejectsArchivedActor(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	archived := int64(99)
	c := activeCustomer()
	c.ActorID = &archived

	err := assertCustomerReferencesActive(db, c)
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

func TestAssertCustomerReferencesActive_IgnoresNilActor(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	c := &domain.Customer{ID: 0, Name: "Sin actor", ActorID: nil}
	if err := assertCustomerReferencesActive(db, c); err != nil {
		t.Fatalf("expected nil error for nil actor_id, got %v", err)
	}
}

func TestAssertCustomerReferencesActive_IgnoresZeroActor(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	zero := int64(0)
	c := &domain.Customer{ID: 0, Name: "Zero actor", ActorID: &zero}
	if err := assertCustomerReferencesActive(db, c); err != nil {
		t.Fatalf("expected nil error for zero actor_id, got %v", err)
	}
}

func TestAssertCustomerReferencesActive_NilSafe(t *testing.T) {
	t.Parallel()
	db := newArchivedRefsTestDB(t)

	if err := assertCustomerReferencesActive(db, nil); err != nil {
		t.Fatalf("expected nil error for nil customer, got %v", err)
	}
}
