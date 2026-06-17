package lot

import (
	"context"
	"testing"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// stubEngine implementa GormEnginePort sobre una conexión de test.
type stubEngine struct{ db *gorm.DB }

func (s stubEngine) Client() *gorm.DB { return s.db }

func newArchiveDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`CREATE TABLE lots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		field_id INTEGER,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME,
		created_by TEXT,
		updated_by TEXT,
		deleted_by TEXT);`).Error; err != nil {
		t.Fatalf("schema: %v", err)
	}
	return db
}

// Antes del fix, archivar un lote inexistente devolvía 204 (silencioso) porque no se
// chequeaba RowsAffected y el guard recibía el lot id contra la tabla fields.
func TestArchiveLot_MissingReturnsNotFound(t *testing.T) {
	repo := NewRepository(stubEngine{db: newArchiveDB(t)})
	err := repo.ArchiveLot(context.Background(), 999)
	if !domainerr.IsNotFound(err) {
		t.Fatalf("want NotFound for missing lot, got %v", err)
	}
}

func TestRestoreLot_MissingReturnsNotFound(t *testing.T) {
	repo := NewRepository(stubEngine{db: newArchiveDB(t)})
	err := repo.RestoreLot(context.Background(), 999)
	if !domainerr.IsNotFound(err) {
		t.Fatalf("want NotFound for missing lot, got %v", err)
	}
}

// Archivar dos veces el mismo lote debe seguir siendo idempotente (no 404 en el 2º).
func TestArchiveLot_Idempotent(t *testing.T) {
	db := newArchiveDB(t)
	if err := db.Exec(`INSERT INTO lots (id, name, field_id) VALUES (1, 'L1', 1)`).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
	repo := NewRepository(stubEngine{db: db})
	ctx := context.Background()
	if err := repo.ArchiveLot(ctx, 1); err != nil {
		t.Fatalf("first archive: %v", err)
	}
	if err := repo.ArchiveLot(ctx, 1); err != nil {
		t.Fatalf("second archive should be idempotent, got %v", err)
	}
}

// Restaurar un lote archivado lo deja activo de nuevo.
func TestRestoreLot_ReactivatesArchived(t *testing.T) {
	db := newArchiveDB(t)
	if err := db.Exec(`INSERT INTO lots (id, name, field_id, deleted_at) VALUES (1, 'L1', 1, CURRENT_TIMESTAMP)`).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
	repo := NewRepository(stubEngine{db: db})
	if err := repo.RestoreLot(context.Background(), 1); err != nil {
		t.Fatalf("restore: %v", err)
	}
	var deletedAt *string
	if err := db.Raw(`SELECT deleted_at FROM lots WHERE id = 1`).Scan(&deletedAt).Error; err != nil {
		t.Fatalf("load: %v", err)
	}
	if deletedAt != nil {
		t.Fatalf("expected deleted_at NULL after restore, got %v", *deletedAt)
	}
}
