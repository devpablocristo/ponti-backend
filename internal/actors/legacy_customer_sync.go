package actors

import (
	"context"
	"strings"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"gorm.io/gorm"

	"github.com/devpablocristo/ponti-backend/internal/identity"
	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"
)

func hasRoleValue(roles []string, role identity.Role) bool {
	want := string(role)
	for _, raw := range roles {
		if strings.EqualFold(strings.TrimSpace(raw), want) {
			return true
		}
	}
	return false
}

func (r *Repository) actorHasCustomerRole(tx *gorm.DB, actorID int64) (bool, error) {
	var n int64
	if err := tx.Raw(
		"SELECT count(*) FROM actor_roles WHERE actor_id = ? AND role = ?",
		actorID,
		string(identity.RoleCustomer),
	).Scan(&n).Error; err != nil {
		return false, domainerr.Internal("failed to check actor roles")
	}
	return n > 0, nil
}

func nameMatchPredicate(tx *gorm.DB) string {
	if tx.Name() == "postgres" {
		return "normalize_name(name) = normalize_name(?)"
	}
	return "lower(name) = lower(?)"
}

func uniqueConflict(err error, message string) error {
	if sharedrepo.IsUniqueViolation(err) || strings.Contains(strings.ToLower(err.Error()), "unique") {
		return domainerr.Conflict(message)
	}
	return err
}

// ensureLegacyCustomerForActor materializa o vincula la fila legacy customers para
// actores con rol customer. Los selectores del sistema siguen leyendo esa tabla.
func (r *Repository) ensureLegacyCustomerForActor(ctx context.Context, tx *gorm.DB, actorID int64) error {
	tenantID, err := identity.TenantFor(ctx, tx)
	if err != nil {
		return err
	}

	var actor struct {
		ID          int64
		DisplayName string
		DeletedAt   gorm.DeletedAt
	}
	if err := tx.Raw(
		"SELECT id, display_name, deleted_at FROM actors WHERE id = ? AND tenant_id = ?",
		actorID,
		tenantID,
	).Scan(&actor).Error; err != nil {
		return domainerr.Internal("failed to load actor")
	}
	if actor.ID == 0 {
		return domainerr.NotFound("actor not found")
	}
	name := strings.TrimSpace(actor.DisplayName)
	if name == "" {
		return domainerr.Validation("customer name is required")
	}
	if actor.DeletedAt.Valid {
		return r.archiveLegacyCustomerForActor(tx, actorID)
	}

	var linked struct {
		ID int64
	}
	if err := tx.Raw(
		"SELECT id FROM customers WHERE actor_id = ? LIMIT 1",
		actorID,
	).Scan(&linked).Error; err != nil {
		return domainerr.Internal("failed to find linked customer")
	}
	if linked.ID != 0 {
		err := tx.Exec(
			"UPDATE customers SET name = ?, tenant_id = ?, actor_id = ?, deleted_at = NULL, deleted_by = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			name,
			tenantID,
			actorID,
			linked.ID,
		).Error
		if err != nil {
			return uniqueConflict(err, "a customer with that name already exists")
		}
		return nil
	}

	var byName struct {
		ID      int64
		ActorID *int64
	}
	predicate := nameMatchPredicate(tx)
	if err := tx.Raw(
		"SELECT id, actor_id FROM customers WHERE deleted_at IS NULL AND tenant_id = ? AND "+predicate+" LIMIT 1",
		tenantID,
		name,
	).Scan(&byName).Error; err != nil {
		return domainerr.Internal("failed to find customer by name")
	}
	if byName.ID != 0 {
		if byName.ActorID != nil && *byName.ActorID != actorID {
			return domainerr.Conflict("customer is already linked to another actor")
		}
		err := tx.Exec(
			"UPDATE customers SET actor_id = ?, name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			actorID,
			name,
			byName.ID,
		).Error
		if err != nil {
			return uniqueConflict(err, "a customer with that name already exists")
		}
		return nil
	}

	err = tx.Exec(
		"INSERT INTO customers (name, tenant_id, actor_id, created_at, updated_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		name,
		tenantID,
		actorID,
	).Error
	if err != nil {
		return uniqueConflict(err, "a customer with that name already exists")
	}
	return nil
}

func (r *Repository) archiveLegacyCustomerForActor(tx *gorm.DB, actorID int64) error {
	var customer struct {
		ID int64
	}
	if err := tx.Raw(
		"SELECT id FROM customers WHERE actor_id = ? LIMIT 1",
		actorID,
	).Scan(&customer).Error; err != nil {
		return domainerr.Internal("failed to find linked customer")
	}
	if customer.ID == 0 {
		return nil
	}

	var activeProjects int64
	if err := tx.Raw(
		"SELECT count(*) FROM projects WHERE customer_id = ? AND deleted_at IS NULL",
		customer.ID,
	).Scan(&activeProjects).Error; err != nil {
		return domainerr.Internal("failed to check active projects")
	}
	if activeProjects > 0 {
		return domainerr.Conflict("customer has active projects")
	}

	if err := tx.Exec(
		"UPDATE customers SET deleted_at = CURRENT_TIMESTAMP, deleted_by = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND deleted_at IS NULL",
		customer.ID,
	).Error; err != nil {
		return domainerr.Internal("failed to archive linked customer")
	}
	return nil
}
