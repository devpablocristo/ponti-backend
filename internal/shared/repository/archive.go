package sharedrepo

import (
	"context"
	"database/sql"
	"fmt"

	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
)

// ArchiveOptions parametriza el soft-delete/restore unificado por entidad.
//
//   - Scope: aplica la acotación de tenant EXISTENTE de la entidad al query (no se cambia acá
//     el comportamiento de tenant). Ej.: func(q) { return sharedfilters.ScopeTenant(ctx, q) }
//     para catálogos; para hijas, q.Where(cond, args...) con TenantProjectScope/TenantFieldScope;
//     para actors, q.Where("tenant_id = ?", tenantID). Si es nil, no se acota.
//   - BeforeArchive: guard de conflicto previo al archivado (ej. customer con proyectos activos);
//     si devuelve error (típicamente domainerr.Conflict) se aborta. Solo aplica a SoftArchive.
//   - RestoreConflictMsg: mensaje del 409 cuando restaurar dispara un unique-violation (nombre/
//     CUIT/key ya activo). Si vacío, un unique-violation se trata como error interno.
type ArchiveOptions struct {
	Scope              func(*gorm.DB) *gorm.DB
	BeforeArchive      func(tx *gorm.DB) error
	RestoreConflictMsg string
}

// probeState devuelve si la fila existe (en el tenant acotado) y si está archivada, leyendo
// solo deleted_at. Universal: toda entidad con soft-delete tiene esa columna.
func probeState(tx *gorm.DB, model any, id int64, scope func(*gorm.DB) *gorm.DB) (exists, archived bool, err error) {
	q := tx.Unscoped().Model(model).Where("id = ?", id)
	if scope != nil {
		q = scope(q)
	}
	var deletedAts []sql.NullTime
	if err = q.Pluck("deleted_at", &deletedAts).Error; err != nil {
		return false, false, err
	}
	if len(deletedAts) == 0 {
		return false, false, nil
	}
	return true, deletedAts[0].Valid, nil
}

// SoftArchive aplica la semántica UNIFICADA de archivado (soft-delete) para cualquier entidad:
//   - 404 (NotFound) si no existe o no es del tenant del caller.
//   - no-op nil (200) si YA está archivada → idempotente.
//   - 409 (Conflict) si BeforeArchive lo determina (referencias activas, etc.).
//   - si no, marca deleted_at (soft-delete nativo de gorm).
//
// Corre dentro de la tx del caller. `model` es un puntero al modelo gorm (con gorm.DeletedAt).
func SoftArchive(ctx context.Context, tx *gorm.DB, model any, id int64, entity string, opts ArchiveOptions) error {
	exists, archived, err := probeState(tx, model, id, opts.Scope)
	if err != nil {
		return domainerr.Internal(fmt.Sprintf("failed to load %s", entity))
	}
	if !exists {
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("%s %d not found", entity, id))
	}
	if archived {
		return nil // idempotente: ya archivado
	}
	if opts.BeforeArchive != nil {
		if err := opts.BeforeArchive(tx); err != nil {
			return err
		}
	}
	q := tx.WithContext(ctx).Where("id = ?", id)
	if opts.Scope != nil {
		q = opts.Scope(q)
	}
	if err := q.Delete(model).Error; err != nil {
		return domainerr.Internal(fmt.Sprintf("failed to archive %s", entity))
	}
	return nil
}

// SoftRestore aplica la semántica UNIFICADA de restauración:
//   - 404 (NotFound) si no existe o no es del tenant del caller.
//   - no-op nil (200) si YA está activa → idempotente.
//   - 409 (Conflict) si reactivar choca con un duplicado activo (unique-violation) y hay
//     RestoreConflictMsg; si no, error interno.
//   - si no, limpia deleted_at.
func SoftRestore(ctx context.Context, tx *gorm.DB, model any, id int64, entity string, opts ArchiveOptions) error {
	exists, archived, err := probeState(tx, model, id, opts.Scope)
	if err != nil {
		return domainerr.Internal(fmt.Sprintf("failed to load %s", entity))
	}
	if !exists {
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("%s %d not found", entity, id))
	}
	if !archived {
		return nil // idempotente: ya activa
	}
	q := tx.WithContext(ctx).Unscoped().Model(model).Where("id = ?", id)
	if opts.Scope != nil {
		q = opts.Scope(q)
	}
	if err := q.Update("deleted_at", nil).Error; err != nil {
		if opts.RestoreConflictMsg != "" && IsUniqueViolation(err) {
			return domainerr.Conflict(opts.RestoreConflictMsg)
		}
		return domainerr.Internal(fmt.Sprintf("failed to restore %s", entity))
	}
	return nil
}
