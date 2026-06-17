package customer

import (
	"context"
	"testing"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type archStubEngine struct{ db *gorm.DB }

func (s archStubEngine) Client() *gorm.DB { return s.db }

func newCustomerArchiveDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	stmts := []string{
		`CREATE TABLE customers (
			id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, tenant_id TEXT,
			created_at DATETIME, updated_at DATETIME, deleted_at DATETIME,
			created_by TEXT, updated_by TEXT, deleted_by TEXT);`,
		`CREATE TABLE projects (id INTEGER PRIMARY KEY AUTOINCREMENT, customer_id INTEGER, deleted_at DATETIME);`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			t.Fatalf("schema: %v", err)
		}
	}
	return db
}

// El guard de negocio (BeforeArchive) debe bloquear archivar un customer con proyectos activos.
func TestArchiveCustomer_BlockedByActiveProjects(t *testing.T) {
	db := newCustomerArchiveDB(t)
	db.Exec(`INSERT INTO customers (id, name) VALUES (1, 'C1')`)
	db.Exec(`INSERT INTO projects (id, customer_id) VALUES (1, 1)`) // activo
	repo := NewRepository(archStubEngine{db: db})

	err := repo.ArchiveCustomer(context.Background(), 1)
	if !domainerr.IsConflict(err) {
		t.Fatalf("want Conflict (active projects), got %v", err)
	}
	var n int64
	db.Raw(`SELECT count(*) FROM customers WHERE id=1 AND deleted_at IS NOT NULL`).Scan(&n)
	if n != 0 {
		t.Fatalf("el guard debió abortar el archive, pero el customer quedó archivado")
	}
}

// Sin proyectos activos: archiva; re-archivar es no-op idempotente.
func TestArchiveCustomer_NoActiveProjectsIsIdempotent(t *testing.T) {
	db := newCustomerArchiveDB(t)
	db.Exec(`INSERT INTO customers (id, name) VALUES (1, 'C1')`)
	repo := NewRepository(archStubEngine{db: db})
	ctx := context.Background()
	if err := repo.ArchiveCustomer(ctx, 1); err != nil {
		t.Fatalf("archive: %v", err)
	}
	if err := repo.ArchiveCustomer(ctx, 1); err != nil {
		t.Fatalf("re-archive debe ser no-op, got %v", err)
	}
}

func TestArchiveCustomer_MissingReturnsNotFound(t *testing.T) {
	db := newCustomerArchiveDB(t)
	repo := NewRepository(archStubEngine{db: db})
	if err := repo.ArchiveCustomer(context.Background(), 999); !domainerr.IsNotFound(err) {
		t.Fatalf("want NotFound, got %v", err)
	}
}
