package lot

import (
	"context"
	"testing"
	"time"

	contextkeys "github.com/devpablocristo/core/security/go/contextkeys"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	cropdom "github.com/devpablocristo/ponti-backend/internal/crop/usecases/domain"
	domain "github.com/devpablocristo/ponti-backend/internal/lot/usecases/domain"
)

type lotTenantGormEngine struct {
	client *gorm.DB
}

func (e lotTenantGormEngine) Client() *gorm.DB {
	return e.client
}

func lotTenantContext(tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "tenant-user@example.com")
	ctx = context.WithValue(ctx, contextkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, contextkeys.Role, "tenant_admin")
	ctx = context.WithValue(ctx, contextkeys.Scopes, []string{"lots.read", "lots.write", "lots.archive"})
	return ctx
}

func setupLotTenantDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE lots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			field_id INTEGER NOT NULL,
			hectares NUMERIC NOT NULL DEFAULT 0,
			previous_crop_id INTEGER NOT NULL DEFAULT 0,
			current_crop_id INTEGER NOT NULL DEFAULT 0,
			season TEXT NOT NULL DEFAULT '',
			variety TEXT,
			sowing_date DATETIME,
			tons NUMERIC DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		);
		CREATE TABLE lot_dates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			lot_id INTEGER NOT NULL,
			sowing_date DATETIME,
			harvest_date DATETIME,
			sequence INTEGER,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		);
		CREATE TABLE workorders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			lot_id INTEGER NOT NULL,
			crop_id INTEGER,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		);
	`).Error; err != nil {
		t.Fatalf("create schema: %v", err)
	}

	return db
}

func TestLotRepositoryTenantIsolation(t *testing.T) {
	db := setupLotTenantDB(t)
	repo := NewRepository(lotTenantGormEngine{client: db})

	tenantA := uuid.New()
	tenantB := uuid.New()
	now := time.Now().UTC()

	if err := db.Exec(`
		INSERT INTO lots (id, tenant_id, name, field_id, hectares, previous_crop_id, current_crop_id, season, created_at, updated_at, deleted_at) VALUES
			(1, ?, 'Lot A', 10, 10, 1, 2, '2025-2026', ?, ?, NULL),
			(2, ?, 'Lot B', 20, 20, 1, 2, '2025-2026', ?, ?, NULL),
			(3, ?, 'Lot B archived', 20, 30, 1, 2, '2025-2026', ?, ?, ?)
	`, tenantA.String(), now, now, tenantB.String(), now, now, tenantB.String(), now, now, now).Error; err != nil {
		t.Fatalf("seed lots: %v", err)
	}

	ctxA := lotTenantContext(tenantA)

	list, err := repo.ListLotsByField(ctxA, 10)
	if err != nil {
		t.Fatalf("list tenant lots: %v", err)
	}
	if len(list) != 1 || list[0].ID != 1 {
		t.Fatalf("expected tenant A lot, got %#v", list)
	}
	list, err = repo.ListLotsByField(ctxA, 20)
	if err != nil {
		t.Fatalf("list cross-tenant lots by field: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected no cross-tenant lots, got %#v", list)
	}

	if _, err := repo.GetLot(ctxA, 2); err == nil {
		t.Fatalf("expected get cross-tenant lot to fail")
	}

	if err := repo.UpdateLot(ctxA, &domain.Lot{
		ID:           2,
		Name:         "cross tenant update",
		FieldID:      20,
		PreviousCrop: cropdom.Crop{},
		CurrentCrop:  cropdom.Crop{},
	}); err == nil {
		t.Fatalf("expected update cross-tenant lot to fail")
	}

	var name string
	if err := db.Raw(`SELECT name FROM lots WHERE id = 2`).Scan(&name).Error; err != nil {
		t.Fatalf("read lot 2: %v", err)
	}
	if name != "Lot B" {
		t.Fatalf("cross-tenant update changed lot 2 name to %q", name)
	}

	if err := repo.UpdateLotTons(ctxA, 2, decimal.NewFromInt(99)); err == nil {
		t.Fatalf("expected update tons cross-tenant lot to fail")
	}

	if err := repo.ArchiveLot(ctxA, 2); err == nil {
		t.Fatalf("expected archive cross-tenant lot to fail")
	}
	var deletedCount int64
	if err := db.Raw(`SELECT COUNT(*) FROM lots WHERE id = 2 AND deleted_at IS NOT NULL`).Scan(&deletedCount).Error; err != nil {
		t.Fatalf("check archive side effect: %v", err)
	}
	if deletedCount != 0 {
		t.Fatalf("cross-tenant archive modified lot 2")
	}

	if err := repo.RestoreLot(ctxA, 3); err == nil {
		t.Fatalf("expected restore cross-tenant lot to fail")
	}
	if err := db.Raw(`SELECT COUNT(*) FROM lots WHERE id = 3 AND deleted_at IS NOT NULL`).Scan(&deletedCount).Error; err != nil {
		t.Fatalf("check restore side effect: %v", err)
	}
	if deletedCount != 1 {
		t.Fatalf("cross-tenant restore modified lot 3")
	}

	if err := repo.HardDeleteLot(ctxA, 2); err == nil {
		t.Fatalf("expected hard-delete cross-tenant lot to fail")
	}
	var exists int64
	if err := db.Raw(`SELECT COUNT(*) FROM lots WHERE id = 2`).Scan(&exists).Error; err != nil {
		t.Fatalf("check hard delete side effect: %v", err)
	}
	if exists != 1 {
		t.Fatalf("cross-tenant hard delete removed lot 2")
	}
}

func TestLotRepositoryRequiresTenantInStrictModeForReportViews(t *testing.T) {
	t.Setenv("TENANT_STRICT_MODE", "true")

	db := setupLotTenantDB(t)
	repo := NewRepository(lotTenantGormEngine{client: db})

	if _, err := repo.GetMetrics(context.Background(), 0, 0, 0); err == nil {
		t.Fatalf("expected strict mode to reject lot metrics without tenant context")
	}

	if _, _, _, _, err := repo.ListLots(context.Background(), domain.LotListFilter{}, 1, 50); err == nil {
		t.Fatalf("expected strict mode to reject lot list without tenant context")
	}
}
