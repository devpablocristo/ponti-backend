// Package customer implementa el repositorio de clientes.
package customer

import (
	"context"
	"errors"
	"fmt"
	"time"

	sharedrepo "github.com/alphacodinggroup/ponti-backend/internal/shared/repository"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	models "github.com/alphacodinggroup/ponti-backend/internal/customer/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/internal/customer/usecases/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/internal/shared/models"
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
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create customer", err)
	}
	return model.ID, nil
}

func (r *Repository) ListCustomers(ctx context.Context, page, perPage int) ([]domain.ListedCustomer, int64, error) {
	var list []models.Customer
	var total int64

	db0 := r.db.Client().WithContext(ctx).
		Model(&models.Customer{}).
		Where("deleted_at IS NULL")

	// Conteo total
	if err := db0.Count(&total).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to count customers", err)
	}

	// Consulta ligera: sólo id y name
	if err := db0.
		Select("id, name").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Find(&list).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to list customers", err)
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

	if err := db0.Count(&total).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to count archived customers", err)
	}

	if err := db0.
		Select("id, name").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Find(&list).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to list archived customers", err)
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
	err := r.db.Client().WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, fmt.Sprintf("customer with id %d not found", id), err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get customer", err)
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
	result := updateTx.Updates(models.FromDomain(c))
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to update customer", result.Error)
	}
	if result.RowsAffected == 0 {
		if !c.UpdatedAt.IsZero() {
			return types.NewError(types.ErrConflict, "customer not found or outdated", nil)
		}
		return types.NewError(types.ErrNotFound, fmt.Sprintf("customer with id %d does not exist", c.ID), nil)
	}
	return nil
}

func (r *Repository) ArchiveCustomer(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "customer"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var customer models.Customer
		if err := tx.Unscoped().Where("id = ?", id).First(&customer).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return types.NewError(types.ErrNotFound, fmt.Sprintf("customer %d not found", id), err)
			}
			return types.NewError(types.ErrInternal, "failed to get customer", err)
		}
		if customer.DeletedAt.Valid {
			return types.NewError(types.ErrConflict, "customer already archived", nil)
		}

		var activeProjects int64
		if err := tx.Table("projects").
			Where("customer_id = ? AND deleted_at IS NULL", id).
			Count(&activeProjects).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check active projects", err)
		}
		if activeProjects > 0 {
			return types.NewError(types.ErrConflict, "customer has active projects", nil)
		}

		updates := map[string]any{
			"deleted_at": time.Now(),
		}
		updates["deleted_by"] = gorm.Expr("NULL")

		if err := tx.Model(&models.Customer{}).
			Where("id = ?", id).
			Updates(updates).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to archive customer", err)
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
		if err := tx.Unscoped().Where("id = ?", id).First(&customer).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return types.NewError(types.ErrNotFound, fmt.Sprintf("customer %d not found", id), err)
			}
			return types.NewError(types.ErrInternal, "failed to get customer", err)
		}
		if !customer.DeletedAt.Valid {
			return types.NewError(types.ErrConflict, "customer is not archived", nil)
		}

		if err := tx.Unscoped().Model(&models.Customer{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"deleted_at": nil,
				"deleted_by": nil,
				"updated_at": time.Now(),
			}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to restore customer", err)
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
		if err := tx.Unscoped().Model(&models.Customer{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check customer existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("customer with id %d does not exist", id), nil)
		}

		// Obtener IDs de proyectos del customer
		var projectIDs []int64
		if err := tx.Unscoped().Table("projects").Where("customer_id = ?", id).Pluck("id", &projectIDs).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to get projects for customer", err)
		}

		// Eliminar en cascada cada proyecto (mismo orden que HardDeleteProject)
		for _, projectID := range projectIDs {
			if err := hardDeleteProjectCascade(tx, projectID); err != nil {
				return err
			}
		}

		// Finalmente eliminar el customer
		if err := tx.Unscoped().Delete(&models.Customer{}, "id = ?", id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to hard delete customer", err)
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
		return types.NewError(types.ErrInternal, "failed to get field ids", err)
	}

	// invoices (FK → workorders via work_order_id, CASCADE pero explícito)
	if err := tx.Exec(`
		DELETE FROM invoices 
		WHERE work_order_id IN (SELECT id FROM workorders WHERE project_id = ?)
	`, projectID).Error; err != nil {
		return types.NewError(types.ErrInternal, "failed to hard delete invoices", err)
	}
	// workorder_items (FK → workorders)
	if err := tx.Exec(`
		DELETE FROM workorder_items 
		WHERE workorder_id IN (SELECT id FROM workorders WHERE project_id = ?)
	`, projectID).Error; err != nil {
		return types.NewError(types.ErrInternal, "failed to hard delete workorder_items", err)
	}
	// workorders (FK → labors, projects)
	if err := tx.Exec("DELETE FROM workorders WHERE project_id = ?", projectID).Error; err != nil {
		return types.NewError(types.ErrInternal, "failed to hard delete workorders", err)
	}
	// labors (FK → projects, RESTRICT)
	if err := tx.Exec("DELETE FROM labors WHERE project_id = ?", projectID).Error; err != nil {
		return types.NewError(types.ErrInternal, "failed to hard delete labors", err)
	}
	// supply_movements (FK → projects, RESTRICT)
	if err := tx.Exec("DELETE FROM supply_movements WHERE project_id = ?", projectID).Error; err != nil {
		return types.NewError(types.ErrInternal, "failed to hard delete supply_movements", err)
	}
	// stocks (FK → projects, RESTRICT)
	if err := tx.Exec("DELETE FROM stocks WHERE project_id = ?", projectID).Error; err != nil {
		return types.NewError(types.ErrInternal, "failed to hard delete stocks", err)
	}
	// crop_commercializations
	if err := tx.Exec("DELETE FROM crop_commercializations WHERE project_id = ?", projectID).Error; err != nil {
		return types.NewError(types.ErrInternal, "failed to hard delete commercializations", err)
	}
	// project_dollar_values (RESTRICT)
	if err := tx.Exec("DELETE FROM project_dollar_values WHERE project_id = ?", projectID).Error; err != nil {
		return types.NewError(types.ErrInternal, "failed to hard delete dollar values", err)
	}
	if len(fieldIDs) > 0 {
		// field_investors
		if err := tx.Exec("DELETE FROM field_investors WHERE field_id IN ?", fieldIDs).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to hard delete field_investors", err)
		}
		// lot_dates (si existe FK a lots)
		tx.Exec("DELETE FROM lot_dates WHERE lot_id IN (SELECT id FROM lots WHERE field_id IN ?)", fieldIDs)
		// lots
		if err := tx.Exec("DELETE FROM lots WHERE field_id IN ?", fieldIDs).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to hard delete lots", err)
		}
	}
	// fields
	if err := tx.Exec("DELETE FROM fields WHERE project_id = ?", projectID).Error; err != nil {
		return types.NewError(types.ErrInternal, "failed to hard delete fields", err)
	}
	// project_managers
	if err := tx.Exec("DELETE FROM project_managers WHERE project_id = ?", projectID).Error; err != nil {
		return types.NewError(types.ErrInternal, "failed to hard delete project_managers", err)
	}
	// project_investors
	if err := tx.Exec("DELETE FROM project_investors WHERE project_id = ?", projectID).Error; err != nil {
		return types.NewError(types.ErrInternal, "failed to hard delete project_investors", err)
	}
	// admin_cost_investors
	if err := tx.Exec("DELETE FROM admin_cost_investors WHERE project_id = ?", projectID).Error; err != nil {
		return types.NewError(types.ErrInternal, "failed to hard delete admin_cost_investors", err)
	}
	// Finalmente el proyecto
	if err := tx.Unscoped().Exec("DELETE FROM projects WHERE id = ?", projectID).Error; err != nil {
		return types.NewError(types.ErrInternal, "failed to hard delete project", err)
	}
	return nil
}
