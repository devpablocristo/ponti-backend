package sharedrepo

import (
	"context"
	"testing"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type archRow struct {
	ID        int64 `gorm:"primaryKey"`
	Name      string
	TenantID  string
	DeletedAt gorm.DeletedAt
}

func (archRow) TableName() string { return "arch_rows" }

func newArchDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&archRow{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestSoftArchive_MissingReturnsNotFound(t *testing.T) {
	db := newArchDB(t)
	err := db.Transaction(func(tx *gorm.DB) error {
		return SoftArchive(context.Background(), tx, &archRow{}, 999, "arch_row", ArchiveOptions{})
	})
	if !domainerr.IsNotFound(err) {
		t.Fatalf("want NotFound, got %v", err)
	}
}

func TestSoftArchive_ArchivesThenIdempotent(t *testing.T) {
	db := newArchDB(t)
	db.Create(&archRow{ID: 1, Name: "A"})

	for i := 0; i < 2; i++ { // segundo archive debe ser no-op (nil), no 409/404
		err := db.Transaction(func(tx *gorm.DB) error {
			return SoftArchive(context.Background(), tx, &archRow{}, 1, "arch_row", ArchiveOptions{})
		})
		if err != nil {
			t.Fatalf("archive #%d: %v", i+1, err)
		}
	}
	var cnt int64
	db.Unscoped().Model(&archRow{}).Where("id = 1 AND deleted_at IS NOT NULL").Count(&cnt)
	if cnt != 1 {
		t.Fatalf("expected row 1 archived, got cnt=%d", cnt)
	}
}

func TestSoftRestore_RestoresThenIdempotent(t *testing.T) {
	db := newArchDB(t)
	db.Create(&archRow{ID: 1, Name: "A"})
	db.Delete(&archRow{}, 1) // archivar

	for i := 0; i < 2; i++ { // segundo restore debe ser no-op (nil)
		err := db.Transaction(func(tx *gorm.DB) error {
			return SoftRestore(context.Background(), tx, &archRow{}, 1, "arch_row", ArchiveOptions{})
		})
		if err != nil {
			t.Fatalf("restore #%d: %v", i+1, err)
		}
	}
	var cnt int64
	db.Model(&archRow{}).Where("id = 1").Count(&cnt) // scoped (no archivadas)
	if cnt != 1 {
		t.Fatalf("expected row 1 active after restore, got cnt=%d", cnt)
	}
}

func TestSoftRestore_MissingReturnsNotFound(t *testing.T) {
	db := newArchDB(t)
	err := db.Transaction(func(tx *gorm.DB) error {
		return SoftRestore(context.Background(), tx, &archRow{}, 999, "arch_row", ArchiveOptions{})
	})
	if !domainerr.IsNotFound(err) {
		t.Fatalf("want NotFound, got %v", err)
	}
}

func TestSoftArchive_BeforeArchiveConflict(t *testing.T) {
	db := newArchDB(t)
	db.Create(&archRow{ID: 1, Name: "A"})

	err := db.Transaction(func(tx *gorm.DB) error {
		return SoftArchive(context.Background(), tx, &archRow{}, 1, "arch_row", ArchiveOptions{
			BeforeArchive: func(tx *gorm.DB) error {
				return domainerr.Conflict("en uso")
			},
		})
	})
	if !domainerr.IsConflict(err) {
		t.Fatalf("want Conflict from BeforeArchive, got %v", err)
	}
	var cnt int64
	db.Unscoped().Model(&archRow{}).Where("id = 1 AND deleted_at IS NOT NULL").Count(&cnt)
	if cnt != 0 {
		t.Fatalf("BeforeArchive conflict debió abortar el archive, pero quedó archivado")
	}
}

func TestSoftArchive_ScopeCrossTenantReturnsNotFound(t *testing.T) {
	db := newArchDB(t)
	db.Create(&archRow{ID: 1, Name: "A", TenantID: "tenant-A"})

	// Scope que exige tenant-B: la fila (tenant-A) no matchea → 404.
	err := db.Transaction(func(tx *gorm.DB) error {
		return SoftArchive(context.Background(), tx, &archRow{}, 1, "arch_row", ArchiveOptions{
			Scope: func(q *gorm.DB) *gorm.DB { return q.Where("tenant_id = ?", "tenant-B") },
		})
	})
	if !domainerr.IsNotFound(err) {
		t.Fatalf("want NotFound for cross-tenant, got %v", err)
	}
}
