// Package customer implementa el repositorio de clientes.
package customer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	models "github.com/devpablocristo/ponti-backend/internal/customer/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/customer/usecases/domain"
	identity "github.com/devpablocristo/ponti-backend/internal/identity"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"
	"gorm.io/gorm"
)

type GormEnginePort interface {
	Client() *gorm.DB
}

type Repository struct {
	db GormEnginePort
}

func NewRepository(db GormEnginePort) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateCustomer(ctx context.Context, c *domain.Customer) (int64, error) {
	if err := sharedrepo.ValidateEntity(c, "customer"); err != nil {
		return 0, err
	}
	model := models.FromDomain(c)
	model.Base = sharedmodels.Base{
		CreatedBy: c.CreatedBy,
		UpdatedBy: c.UpdatedBy,
	}

	// Inserta el customer + dual-write de tenant_id. Reutilizado por ambos caminos.
	create := func(db *gorm.DB) error {
		if err := db.WithContext(ctx).Create(model).Error; err != nil {
			if sharedrepo.IsUniqueViolation(err) {
				return domainerr.Conflict("a customer with that name already exists")
			}
			return domainerr.Internal("failed to create customer")
		}
		// T1.e: dual-write de tenant_id (flag-gated).
		if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
			if err := db.WithContext(ctx).Exec("UPDATE customers SET tenant_id = ? WHERE id = ?", orgID, model.ID).Error; err != nil {
				return domainerr.Internal("failed to set customer tenant")
			}
		}
		return nil
	}

	// Flag off → comportamiento idéntico al actual.
	if !sharedmodels.IdentityGateEnabled() {
		if err := create(r.db.Client()); err != nil {
			return 0, err
		}
		return model.ID, nil
	}

	// Identity Gate on: customer + resolución de identidad + stamp de actor_id en UNA tx.
	// Un unique-violation de la clave de identidad revierte el alta (→ 409, reintento reusa).
	if err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := create(tx); err != nil {
			return err
		}
		res, err := identity.ResolveOrCreateIdentity(ctx, tx, identity.RoleCustomer, identity.ResolveInput{RawName: c.Name})
		if err != nil {
			if sharedrepo.IsUniqueViolation(err) {
				return domainerr.Conflict("an entity with that identity already exists")
			}
			return domainerr.Internal("failed to resolve customer identity")
		}
		return tx.Exec("UPDATE customers SET actor_id = ? WHERE id = ?", res.ActorID, model.ID).Error
	}); err != nil {
		return 0, err
	}
	return model.ID, nil
}

func (r *Repository) ListCustomers(ctx context.Context, page, perPage int) ([]domain.ListedCustomer, int64, error) {
	var list []models.Customer
	var total int64

	db0 := r.db.Client().WithContext(ctx).
		Model(&models.Customer{}).
		Where("deleted_at IS NULL")

	// T1.e: acotar al tenant activo (flag-gated).
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		db0 = db0.Where("tenant_id = ?", orgID)
	}

	// Conteo total
	if err := db0.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count customers")
	}

	// Consulta ligera: sólo id y name
	if err := db0.
		Select("id, name").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Find(&list).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to list customers")
	}

	// Mapear a dominio ligero
	customers := make([]domain.ListedCustomer, len(list))
	for i, m := range list {
		customers[i] = domain.ListedCustomer{
			ID:   m.ID,
			Name: m.Name,
		}
	}

	return customers, total, nil
}

func (r *Repository) ListArchivedCustomers(ctx context.Context, page, perPage int) ([]domain.ListedCustomer, int64, error) {
	var list []models.Customer
	var total int64

	db0 := r.db.Client().WithContext(ctx).
		Unscoped().
		Model(&models.Customer{}).
		Where("deleted_at IS NOT NULL")

	// T1.e: acotar archivados al tenant activo (flag-gated) — antes era global.
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		db0 = db0.Where("tenant_id = ?", orgID)
	}

	if err := db0.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count archived customers")
	}

	if err := db0.
		Select("id, name").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Find(&list).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to list archived customers")
	}

	customers := make([]domain.ListedCustomer, len(list))
	for i, m := range list {
		customers[i] = domain.ListedCustomer{
			ID:   m.ID,
			Name: m.Name,
		}
	}

	return customers, total, nil
}

func (r *Repository) GetCustomer(ctx context.Context, id int64) (*domain.Customer, error) {
	var model models.Customer
	q := r.db.Client().WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id)
	// T1.e: guard de ownership (flag-gated) — 404 si el customer no es del tenant.
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		q = q.Where("tenant_id = ?", orgID)
	}
	err := q.First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerr.New(domainerr.KindNotFound, fmt.Sprintf("customer with id %d not found", id))
		}
		return nil, domainerr.Internal("failed to get customer")
	}
	return model.ToDomain(), nil
}

func (r *Repository) UpdateCustomer(ctx context.Context, c *domain.Customer) error {
	if err := sharedrepo.ValidateEntity(c, "customer"); err != nil {
		return err
	}
	if err := sharedrepo.ValidateID(c.ID, "customer"); err != nil {
		return err
	}
	updateTx := r.db.Client().WithContext(ctx).
		Model(&models.Customer{}).
		Where("id = ?", c.ID)
	if !c.UpdatedAt.IsZero() {
		updateTx = updateTx.Where("updated_at = ?", c.UpdatedAt)
	}
	// T1.e: guard de ownership (flag-gated) — solo actualiza si es del tenant.
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		updateTx = updateTx.Where("tenant_id = ?", orgID)
	}
	result := updateTx.Updates(models.FromDomain(c))
	if result.Error != nil {
		return domainerr.Internal("failed to update customer")
	}
	if result.RowsAffected == 0 {
		if !c.UpdatedAt.IsZero() {
			return domainerr.Conflict("customer not found or outdated")
		}
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("customer with id %d does not exist", c.ID))
	}
	return nil
}

func (r *Repository) ArchiveCustomer(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "customer"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var customer models.Customer
		loadQ := tx.Unscoped().Where("id = ?", id)
		// T1.e: guard de ownership (flag-gated).
		if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
			loadQ = loadQ.Where("tenant_id = ?", orgID)
		}
		if err := loadQ.First(&customer).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("customer %d not found", id))
			}
			return domainerr.Internal("failed to get customer")
		}
		if customer.DeletedAt.Valid {
			return domainerr.Conflict("customer already archived")
		}

		var activeProjects int64
		if err := tx.Table("projects").
			Where("customer_id = ? AND deleted_at IS NULL", id).
			Count(&activeProjects).Error; err != nil {
			return domainerr.Internal("failed to check active projects")
		}
		if activeProjects > 0 {
			return domainerr.Conflict("customer has active projects")
		}

		updates := map[string]any{
			"deleted_at": time.Now(),
		}
		updates["deleted_by"] = gorm.Expr("NULL")

		if err := tx.Model(&models.Customer{}).
			Where("id = ?", id).
			Updates(updates).Error; err != nil {
			return domainerr.Internal("failed to archive customer")
		}
		return nil
	})
}

func (r *Repository) RestoreCustomer(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "customer"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var customer models.Customer
		loadQ := tx.Unscoped().Where("id = ?", id)
		// T1.e: guard de ownership (flag-gated).
		if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
			loadQ = loadQ.Where("tenant_id = ?", orgID)
		}
		if err := loadQ.First(&customer).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("customer %d not found", id))
			}
			return domainerr.Internal("failed to get customer")
		}
		if !customer.DeletedAt.Valid {
			return domainerr.Conflict("customer is not archived")
		}

		if err := tx.Unscoped().Model(&models.Customer{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"deleted_at": nil,
				"deleted_by": nil,
				"updated_at": time.Now(),
			}).Error; err != nil {
			return domainerr.Internal("failed to restore customer")
		}
		return nil
	})
}

func (r *Repository) DeleteCustomer(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "customer"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Verificar que el customer existe
		var count int64
		existsQ := tx.Unscoped().Model(&models.Customer{}).Where("id = ?", id)
		// T1.e: guard de ownership (flag-gated).
		if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
			existsQ = existsQ.Where("tenant_id = ?", orgID)
		}
		if err := existsQ.Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check customer existence")
		}
		if count == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("customer with id %d does not exist", id))
		}

		// Obtener IDs de proyectos del customer
		var projectIDs []int64
		if err := tx.Unscoped().Table("projects").Where("customer_id = ?", id).Pluck("id", &projectIDs).Error; err != nil {
			return domainerr.Internal("failed to get projects for customer")
		}

		// Eliminar en cascada cada proyecto (mismo orden que HardDeleteProject)
		for _, projectID := range projectIDs {
			if err := hardDeleteProjectCascade(tx, projectID); err != nil {
				return err
			}
		}

		// Finalmente eliminar el customer
		if err := tx.Unscoped().Delete(&models.Customer{}, "id = ?", id).Error; err != nil {
			return domainerr.Internal("failed to hard delete customer")
		}
		return nil
	})
}

// hardDeleteProjectCascade elimina un proyecto y sus entidades relacionadas.
// Orden según FKs: invoices → workorder_items → workorders → labors → supply_movements →
// stocks → crop_commercializations → project_dollar_values → field_investors → lots →
// fields → project_managers → project_investors → admin_cost_investors → projects.
func hardDeleteProjectCascade(tx *gorm.DB, projectID int64) error {
	var fieldIDs []int64
	if err := tx.Unscoped().Table("fields").Where("project_id = ?", projectID).Pluck("id", &fieldIDs).Error; err != nil {
		return domainerr.Internal("failed to get field ids")
	}

	// invoices (FK → workorders via work_order_id, CASCADE pero explícito)
	if err := tx.Exec(`
		DELETE FROM invoices 
		WHERE work_order_id IN (SELECT id FROM workorders WHERE project_id = ?)
	`, projectID).Error; err != nil {
		return domainerr.Internal("failed to hard delete invoices")
	}
	// workorder_items (FK → workorders)
	if err := tx.Exec(`
		DELETE FROM workorder_items 
		WHERE workorder_id IN (SELECT id FROM workorders WHERE project_id = ?)
	`, projectID).Error; err != nil {
		return domainerr.Internal("failed to hard delete workorder_items")
	}
	// workorders (FK → labors, projects)
	if err := tx.Exec("DELETE FROM workorders WHERE project_id = ?", projectID).Error; err != nil {
		return domainerr.Internal("failed to hard delete workorders")
	}
	// labors (FK → projects, RESTRICT)
	if err := tx.Exec("DELETE FROM labors WHERE project_id = ?", projectID).Error; err != nil {
		return domainerr.Internal("failed to hard delete labors")
	}
	// supply_movements (FK → projects, RESTRICT)
	if err := tx.Exec("DELETE FROM supply_movements WHERE project_id = ?", projectID).Error; err != nil {
		return domainerr.Internal("failed to hard delete supply_movements")
	}
	// stocks (FK → projects, RESTRICT)
	if err := tx.Exec("DELETE FROM stocks WHERE project_id = ?", projectID).Error; err != nil {
		return domainerr.Internal("failed to hard delete stocks")
	}
	// crop_commercializations
	if err := tx.Exec("DELETE FROM crop_commercializations WHERE project_id = ?", projectID).Error; err != nil {
		return domainerr.Internal("failed to hard delete commercializations")
	}
	// project_dollar_values (RESTRICT)
	if err := tx.Exec("DELETE FROM project_dollar_values WHERE project_id = ?", projectID).Error; err != nil {
		return domainerr.Internal("failed to hard delete dollar values")
	}
	if len(fieldIDs) > 0 {
		// field_investors
		if err := tx.Exec("DELETE FROM field_investors WHERE field_id IN ?", fieldIDs).Error; err != nil {
			return domainerr.Internal("failed to hard delete field_investors")
		}
		// lot_dates (si existe FK a lots)
		tx.Exec("DELETE FROM lot_dates WHERE lot_id IN (SELECT id FROM lots WHERE field_id IN ?)", fieldIDs)
		// lots
		if err := tx.Exec("DELETE FROM lots WHERE field_id IN ?", fieldIDs).Error; err != nil {
			return domainerr.Internal("failed to hard delete lots")
		}
	}
	// fields
	if err := tx.Exec("DELETE FROM fields WHERE project_id = ?", projectID).Error; err != nil {
		return domainerr.Internal("failed to hard delete fields")
	}
	// project_managers
	if err := tx.Exec("DELETE FROM project_managers WHERE project_id = ?", projectID).Error; err != nil {
		return domainerr.Internal("failed to hard delete project_managers")
	}
	// project_investors
	if err := tx.Exec("DELETE FROM project_investors WHERE project_id = ?", projectID).Error; err != nil {
		return domainerr.Internal("failed to hard delete project_investors")
	}
	// admin_cost_investors
	if err := tx.Exec("DELETE FROM admin_cost_investors WHERE project_id = ?", projectID).Error; err != nil {
		return domainerr.Internal("failed to hard delete admin_cost_investors")
	}
	// Finalmente el proyecto
	if err := tx.Unscoped().Exec("DELETE FROM projects WHERE id = ?", projectID).Error; err != nil {
		return domainerr.Internal("failed to hard delete project")
	}
	return nil
}
