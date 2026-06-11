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

// Update edita display_name y party_type. Si el nombre canónico cambió, ROTA la clave de nombre
// (LEGAL_NAME/PERSON_NAME) en actor_keys para que la unicidad del nombre se mantenga también al
// editar (mismo patrón que SetTaxID con el CUIT). Si el nombre nuevo choca con otra identidad
// activa → 409. Cambiar solo party_type (mismo canónico) no toca claves.
func (r *Repository) Update(ctx context.Context, a *domain.Actor) error {
	db := r.db.Client().WithContext(ctx)
	tenantID, err := identity.TenantFor(ctx, db)
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		var row struct {
			ID          int64
			DisplayName string
		}
		if err := tx.Raw("SELECT id, display_name FROM actors WHERE id = ? AND tenant_id = ? AND deleted_at IS NULL", a.ID, tenantID).Scan(&row).Error; err != nil {
			return domainerr.Internal("failed to load actor")
		}
		if row.ID == 0 {
			return domainerr.NotFound("actor not found")
		}

		newParsed := identity.ParseLegalName(a.DisplayName)
		oldParsed := identity.ParseLegalName(row.DisplayName)
		if newParsed.KeyValue() != oldParsed.KeyValue() || newParsed.KeyType() != oldParsed.KeyType() {
			nv := newParsed.KeyValue()
			if nv == "" {
				return domainerr.Validation("nombre inválido")
			}
			// Desactivar las claves de nombre activas del actor distintas a la nueva.
			if err := tx.Exec(
				"UPDATE actor_keys SET active = false WHERE actor_id = ? AND key_type IN ('LEGAL_NAME','PERSON_NAME') AND active AND NOT (key_type = ? AND key_value = ?)",
				a.ID, newParsed.KeyType(), nv,
			).Error; err != nil {
				return domainerr.Internal("failed to rotate name key")
			}
			// Reactivar si el actor ya tuvo esta clave; si no, insertarla. Cualquiera puede
			// chocar con la clave de nombre activa de otra identidad → 409.
			res := tx.Exec("UPDATE actor_keys SET active = true WHERE actor_id = ? AND key_type = ? AND key_value = ?", a.ID, newParsed.KeyType(), nv)
			if res.Error != nil {
				if sharedrepo.IsUniqueViolation(res.Error) {
					return domainerr.Conflict("ese nombre ya lo usa otra identidad")
				}
				return domainerr.Internal("failed to set name key")
			}
			if res.RowsAffected == 0 {
				if err := tx.Exec(
					"INSERT INTO actor_keys (actor_id, tenant_id, key_type, key_value, active, source) VALUES (?, ?, ?, ?, true, 'direct')",
					a.ID, tenantID, newParsed.KeyType(), nv,
				).Error; err != nil {
					if sharedrepo.IsUniqueViolation(err) {
						return domainerr.Conflict("ese nombre ya lo usa otra identidad")
					}
					return domainerr.Internal("failed to set name key")
				}
			}
		}

		res := tx.Exec(
			"UPDATE actors SET display_name = ?, party_type = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND tenant_id = ? AND deleted_at IS NULL",
			a.DisplayName, a.PartyType, a.ID, tenantID)
		if res.Error != nil {
			return domainerr.Internal("failed to update actor")
		}
		if res.RowsAffected == 0 {
			return domainerr.NotFound("actor not found")
		}
		hasCustomer, err := r.actorHasCustomerRole(tx, a.ID)
		if err != nil {
			return err
		}
		if hasCustomer {
			if err := r.ensureLegacyCustomerForActor(ctx, tx, a.ID); err != nil {
				return err
			}
		}
		return nil
	})
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
		hasCustomer, err := r.actorHasCustomerRole(tx, id)
		if err != nil {
			return err
		}
		if hasCustomer {
			if err := r.archiveLegacyCustomerForActor(tx, id); err != nil {
				return err
			}
		}
		res := tx.Exec("UPDATE actors SET deleted_at = CURRENT_TIMESTAMP WHERE id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID)
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
		if err := tx.Exec("UPDATE actors SET deleted_at = NULL WHERE id = ?", id).Error; err != nil {
			return domainerr.Internal("failed to restore actor")
		}
		hasCustomer, err := r.actorHasCustomerRole(tx, id)
		if err != nil {
			return err
		}
		if hasCustomer {
			return r.ensureLegacyCustomerForActor(ctx, tx, id)
		}
		return nil
	})
}

// SetRoles reemplaza el conjunto de roles del actor por los dados (valida cada rol,
// tenant-guarded). Idempotente: borra los que sobran e inserta los que faltan.
func (r *Repository) SetRoles(ctx context.Context, id int64, roles []string) error {
	db := r.db.Client().WithContext(ctx)
	tenantID, err := identity.TenantFor(ctx, db)
	if err != nil {
		return err
	}
	clean := make([]string, 0, len(roles))
	seen := make(map[string]bool)
	for _, raw := range roles {
		role, e := toRole(raw)
		if e != nil {
			return e
		}
		if !seen[string(role)] {
			seen[string(role)] = true
			clean = append(clean, string(role))
		}
	}
	if len(clean) == 0 {
		return domainerr.Validation("at least one role is required")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		var owner int64
		if err := tx.Raw("SELECT id FROM actors WHERE id = ? AND tenant_id = ?", id, tenantID).Scan(&owner).Error; err != nil {
			return err
		}
		if owner == 0 {
			return domainerr.NotFound("actor not found")
		}
		hadCustomer, err := r.actorHasCustomerRole(tx, id)
		if err != nil {
			return err
		}
		wantsCustomer := hasRoleValue(clean, identity.RoleCustomer)
		if hadCustomer && !wantsCustomer {
			if err := r.archiveLegacyCustomerForActor(tx, id); err != nil {
				return err
			}
		}
		if err := tx.Exec("DELETE FROM actor_roles WHERE actor_id = ? AND role NOT IN ?", id, clean).Error; err != nil {
			return err
		}
		for _, role := range clean {
			if err := tx.Exec("INSERT INTO actor_roles (actor_id, role) VALUES (?, ?) ON CONFLICT (actor_id, role) DO NOTHING", id, role).Error; err != nil {
				return err
			}
		}
		if wantsCustomer {
			if err := r.ensureLegacyCustomerForActor(ctx, tx, id); err != nil {
				return err
			}
		}
		return nil
	})
}

// SetTaxID corrige (o agrega) la clave fiscal CUIT/DNI del actor SIN tocar su identidad
// interna: el actor_id no se mueve, así que todo lo colgado (clientes, trabajos, FKs
// *_actor_id) sigue apuntando igual. Desactiva la TAX_ID activa anterior y activa/inserta la
// nueva normalizada, en una sola tx. Si OTRA identidad activa del tenant ya tiene ese CUIT, el
// índice único parcial lo rechaza → 409 (caso "merge", fuera de alcance). Idempotente.
func (r *Repository) SetTaxID(ctx context.Context, id int64, rawTaxID string) error {
	db := r.db.Client().WithContext(ctx)
	tenantID, err := identity.TenantFor(ctx, db)
	if err != nil {
		return err
	}
	if !identity.TaxIDIsNumeric(rawTaxID) {
		return domainerr.Validation("el id fiscal solo puede contener números")
	}
	val := identity.NormalizeTaxID(rawTaxID)
	if val == "" {
		return domainerr.Validation("tax_id is required")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		var owner int64
		if err := tx.Raw("SELECT id FROM actors WHERE id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID).Scan(&owner).Error; err != nil {
			return domainerr.Internal("failed to load actor")
		}
		if owner == 0 {
			return domainerr.NotFound("actor not found")
		}
		// Desactivar la TAX_ID activa actual del actor (si tiene otra distinta).
		if err := tx.Exec("UPDATE actor_keys SET active = false WHERE actor_id = ? AND key_type = 'TAX_ID' AND active AND key_value <> ?", id, val).Error; err != nil {
			return domainerr.Internal("failed to rotate tax_id")
		}
		// Reactivar si el actor ya tuvo esta clave; si no, insertarla. Cualquiera de las dos
		// puede chocar con la TAX_ID activa de otra identidad → unique-violation → 409.
		res := tx.Exec("UPDATE actor_keys SET active = true WHERE actor_id = ? AND key_type = 'TAX_ID' AND key_value = ?", id, val)
		if res.Error != nil {
			if sharedrepo.IsUniqueViolation(res.Error) {
				return domainerr.Conflict("another active identity already uses that CUIT/DNI")
			}
			return domainerr.Internal("failed to set tax_id")
		}
		if res.RowsAffected == 0 {
			if err := tx.Exec(
				"INSERT INTO actor_keys (actor_id, tenant_id, key_type, key_value, active, source) VALUES (?, ?, 'TAX_ID', ?, true, 'direct')",
				id, tenantID, val,
			).Error; err != nil {
				if sharedrepo.IsUniqueViolation(err) {
					return domainerr.Conflict("another active identity already uses that CUIT/DNI")
				}
				return domainerr.Internal("failed to set tax_id")
			}
		}
		return nil
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
