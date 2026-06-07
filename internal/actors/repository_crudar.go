package actors

import (
	"context"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"gorm.io/gorm"

	domain "github.com/devpablocristo/ponti-backend/internal/actors/usecases/domain"
	identity "github.com/devpablocristo/ponti-backend/internal/identity"
	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"
)

// List devuelve actores del tenant paginados. status: active (default) | archived | all.
func (r *Repository) List(ctx context.Context, status string, page, perPage int) ([]domain.Actor, int64, error) {
	db := r.db.Client().WithContext(ctx)
	tenantID, err := identity.TenantFor(ctx, db)
	if err != nil {
		return nil, 0, err
	}
	where := "tenant_id = ?"
	switch status {
	case "archived":
		where += " AND deleted_at IS NOT NULL"
	case "all":
		// sin filtro de estado
	default:
		where += " AND deleted_at IS NULL"
	}

	var total int64
	if err := db.Raw("SELECT count(*) FROM actors WHERE "+where, tenantID).Scan(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count actors")
	}
	var ids []int64
	if err := db.Raw("SELECT id FROM actors WHERE "+where+" ORDER BY display_name LIMIT ? OFFSET ?",
		tenantID, perPage, (page-1)*perPage).Scan(&ids).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to list actors")
	}
	out := make([]domain.Actor, 0, len(ids))
	for _, id := range ids {
		if a, e := r.loadActor(db, id); e == nil && a != nil {
			out = append(out, *a)
		}
	}
	return out, total, nil
}

// Get carga un actor del tenant (404 si no es del tenant).
func (r *Repository) Get(ctx context.Context, id int64) (*domain.Actor, error) {
	db := r.db.Client().WithContext(ctx)
	tenantID, err := identity.TenantFor(ctx, db)
	if err != nil {
		return nil, err
	}
	var owner int64
	if err := db.Raw("SELECT id FROM actors WHERE id = ? AND tenant_id = ?", id, tenantID).Scan(&owner).Error; err != nil {
		return nil, domainerr.Internal("failed to get actor")
	}
	if owner == 0 {
		return nil, domainerr.NotFound("actor not found")
	}
	return r.loadActor(db, id)
}

// Update edita el display_name y party_type (las claves las maneja el resolver).
func (r *Repository) Update(ctx context.Context, a *domain.Actor) error {
	db := r.db.Client().WithContext(ctx)
	tenantID, err := identity.TenantFor(ctx, db)
	if err != nil {
		return err
	}
	res := db.Exec(
		"UPDATE actors SET display_name = ?, party_type = ?, updated_at = now() WHERE id = ? AND tenant_id = ? AND deleted_at IS NULL",
		a.DisplayName, a.PartyType, a.ID, tenantID)
	if res.Error != nil {
		return domainerr.Internal("failed to update actor")
	}
	if res.RowsAffected == 0 {
		return domainerr.NotFound("actor not found")
	}
	return nil
}

// Archive soft-borra el actor y DESACTIVA sus claves (lo saca del pool de dedup, así un
// alta nueva con el mismo CUIT/nombre crea una identidad fresca y no reusa una archivada).
func (r *Repository) Archive(ctx context.Context, id int64) error {
	db := r.db.Client().WithContext(ctx)
	tenantID, err := identity.TenantFor(ctx, db)
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		res := tx.Exec("UPDATE actors SET deleted_at = now() WHERE id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID)
		if res.Error != nil {
			return domainerr.Internal("failed to archive actor")
		}
		if res.RowsAffected == 0 {
			return domainerr.NotFound("actor not found")
		}
		return tx.Exec("UPDATE actor_keys SET active = false WHERE actor_id = ?", id).Error
	})
}

// Restore reactiva el actor y sus claves. Si otra identidad activa ya tomó una de esas
// claves, el índice único lo rechaza → 409.
func (r *Repository) Restore(ctx context.Context, id int64) error {
	db := r.db.Client().WithContext(ctx)
	tenantID, err := identity.TenantFor(ctx, db)
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		var owner int64
		if err := tx.Raw("SELECT id FROM actors WHERE id = ? AND tenant_id = ? AND deleted_at IS NOT NULL", id, tenantID).Scan(&owner).Error; err != nil {
			return domainerr.Internal("failed to restore actor")
		}
		if owner == 0 {
			return domainerr.NotFound("archived actor not found")
		}
		if err := tx.Exec("UPDATE actor_keys SET active = true WHERE actor_id = ?", id).Error; err != nil {
			if sharedrepo.IsUniqueViolation(err) {
				return domainerr.Conflict("cannot restore: an active identity now uses one of its keys")
			}
			return domainerr.Internal("failed to restore actor keys")
		}
		return tx.Exec("UPDATE actors SET deleted_at = NULL WHERE id = ?", id).Error
	})
}

// Delete hard-borra el actor (cascade de actor_keys/actor_roles; los portadores *_actor_id
// quedan en NULL por ON DELETE SET NULL).
func (r *Repository) Delete(ctx context.Context, id int64) error {
	db := r.db.Client().WithContext(ctx)
	tenantID, err := identity.TenantFor(ctx, db)
	if err != nil {
		return err
	}
	res := db.Exec("DELETE FROM actors WHERE id = ? AND tenant_id = ?", id, tenantID)
	if res.Error != nil {
		return domainerr.Internal("failed to delete actor")
	}
	if res.RowsAffected == 0 {
		return domainerr.NotFound("actor not found")
	}
	return nil
}
