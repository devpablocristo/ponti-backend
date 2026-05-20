package lot

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/devpablocristo/core/errors/go/domainerr"
	cropdom "github.com/devpablocristo/ponti-backend/internal/crop/usecases/domain"
	domain "github.com/devpablocristo/ponti-backend/internal/lot/usecases/domain"
)

// TestLotRepositoryCRUDARFlow comprueba empíricamente el flujo completo
// CRUDAR del repositorio de lot, incluyendo sus particularidades:
//   - Create / List / Get / Update funcionan en orden.
//   - Archive marca deleted_at sin borrar la fila.
//   - HardDelete requiere que el lot esté archivado primero (RequireArchived).
//   - HardDelete devuelve 409 con prefijo BLOCKED_BY_WORKORDERS:<count>|...
//     cuando el lot tiene workorders referenciándolo.
//   - Restore devuelve el lot a estado activo.
func TestLotRepositoryCRUDARFlow(t *testing.T) {
	db := setupLotTenantDB(t)
	// RestoreLot consulta fields y projects para chequear cascada inversa.
	if err := db.Exec(`
		CREATE TABLE fields (
			id INTEGER PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			project_id INTEGER NOT NULL,
			deleted_at DATETIME
		);
		CREATE TABLE projects (
			id INTEGER PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			deleted_at DATETIME
		);
	`).Error; err != nil {
		t.Fatalf("create fields/projects schema: %v", err)
	}

	repo := NewRepository(lotTenantGormEngine{client: db})

	tenantID := uuid.New()
	ctx := lotTenantContext(tenantID)

	// Field activo asociado al lot (field_id=10, project_id=100).
	if err := db.Exec(
		`INSERT INTO fields (id, tenant_id, project_id, deleted_at)
		 VALUES (10, ?, 100, NULL);
		 INSERT INTO projects (id, tenant_id, deleted_at)
		 VALUES (100, ?, NULL);`,
		tenantID.String(), tenantID.String(),
	).Error; err != nil {
		t.Fatalf("seed field/project: %v", err)
	}

	// ── Create ────────────────────────────────────────────────────────────
	lot := &domain.Lot{
		Name:         "LOTE 1",
		FieldID:      10,
		Hectares:     decimal.NewFromInt(50),
		PreviousCrop: cropdom.Crop{},
		CurrentCrop:  cropdom.Crop{},
		Season:       "2025-2026",
	}
	lotID, err := repo.CreateLot(ctx, lot)
	if err != nil {
		t.Fatalf("CreateLot: %v", err)
	}
	if lotID == 0 {
		t.Fatalf("CreateLot returned empty id")
	}

	// ── Read (List by field) ──────────────────────────────────────────────
	list, err := repo.ListLotsByField(ctx, 10)
	if err != nil {
		t.Fatalf("ListLotsByField: %v", err)
	}
	if len(list) != 1 || list[0].ID != lotID {
		t.Fatalf("expected 1 lot id=%d, got %#v", lotID, list)
	}

	// ── Read (Get by id) ──────────────────────────────────────────────────
	got, err := repo.GetLot(ctx, lotID)
	if err != nil {
		t.Fatalf("GetLot: %v", err)
	}
	if got.Name != "LOTE 1" {
		t.Fatalf("GetLot name = %q, want %q", got.Name, "LOTE 1")
	}

	// ── Update ────────────────────────────────────────────────────────────
	got.Name = "LOTE 1 (renombrado)"
	if err := repo.UpdateLot(ctx, got); err != nil {
		t.Fatalf("UpdateLot: %v", err)
	}
	got2, err := repo.GetLot(ctx, lotID)
	if err != nil {
		t.Fatalf("GetLot after update: %v", err)
	}
	if got2.Name != "LOTE 1 (renombrado)" {
		t.Fatalf("UpdateLot did not persist: %q", got2.Name)
	}

	// ── HardDelete sin archive primero → RequireArchived bloquea ─────────
	if err := repo.HardDeleteLot(ctx, lotID); err == nil {
		t.Fatalf("HardDeleteLot debió fallar antes de archive")
	} else if !domainerr.IsKind(err, domainerr.KindConflict) {
		t.Fatalf("HardDelete pre-archive: kind=%v want CONFLICT, msg=%q",
			err, err.Error())
	}

	// ── Archive ───────────────────────────────────────────────────────────
	if err := repo.ArchiveLot(ctx, lotID); err != nil {
		t.Fatalf("ArchiveLot: %v", err)
	}
	var deletedAtCount int64
	if err := db.Raw(
		`SELECT COUNT(*) FROM lots WHERE id = ? AND deleted_at IS NOT NULL`, lotID,
	).Scan(&deletedAtCount).Error; err != nil {
		t.Fatalf("check archive side effect: %v", err)
	}
	if deletedAtCount != 1 {
		t.Fatalf("ArchiveLot no marcó deleted_at")
	}

	// ── HardDelete con workorder referenciando → BLOCKED_BY_WORKORDERS ───
	if err := db.Exec(
		`INSERT INTO workorders (id, tenant_id, lot_id, created_at, updated_at)
		 VALUES (1, ?, ?, datetime('now'), datetime('now'))`,
		tenantID.String(), lotID,
	).Error; err != nil {
		t.Fatalf("seed workorder: %v", err)
	}

	err = repo.HardDeleteLot(ctx, lotID)
	if err == nil {
		t.Fatalf("HardDeleteLot debió fallar con workorder referenciando")
	}
	if !domainerr.IsKind(err, domainerr.KindConflict) {
		t.Fatalf("HardDelete con workorder: kind=%v want CONFLICT", err)
	}
	if !strings.Contains(err.Error(), "BLOCKED_BY_WORKORDERS:1|") {
		t.Fatalf("HardDelete con workorder: falta prefijo en mensaje: %q",
			err.Error())
	}

	// ── Restore (volver a activo) ─────────────────────────────────────────
	if err := repo.RestoreLot(ctx, lotID); err != nil {
		t.Fatalf("RestoreLot: %v", err)
	}
	if err := db.Raw(
		`SELECT COUNT(*) FROM lots WHERE id = ? AND deleted_at IS NULL`, lotID,
	).Scan(&deletedAtCount).Error; err != nil {
		t.Fatalf("check restore side effect: %v", err)
	}
	if deletedAtCount != 1 {
		t.Fatalf("RestoreLot no quitó deleted_at")
	}

	// ── HardDelete final (limpiar workorder + archive + delete) ──────────
	if err := db.Exec(`DELETE FROM workorders WHERE id = 1`).Error; err != nil {
		t.Fatalf("clear workorder: %v", err)
	}
	if err := repo.ArchiveLot(ctx, lotID); err != nil {
		t.Fatalf("re-archive: %v", err)
	}
	if err := repo.HardDeleteLot(ctx, lotID); err != nil {
		t.Fatalf("HardDeleteLot final: %v", err)
	}
	var exists int64
	if err := db.Raw(`SELECT COUNT(*) FROM lots WHERE id = ?`, lotID).
		Scan(&exists).Error; err != nil {
		t.Fatalf("check hard delete: %v", err)
	}
	if exists != 0 {
		t.Fatalf("HardDeleteLot no borró fila físicamente")
	}
}

