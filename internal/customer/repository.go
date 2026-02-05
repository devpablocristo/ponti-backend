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

func (r *Repository) DeleteCustomer(ctx context.Context, id int64) error {
	return r.ArchiveCustomer(ctx, id)
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

func (r *Repository) HardDeleteCustomer(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "customer"); err != nil {
		return err
	}
	result := r.db.Client().WithContext(ctx).
		Unscoped().
		Delete(&models.Customer{}, "id = ?", id)
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to hard delete customer", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("customer with id %d does not exist", id), nil)
	}
	return nil
}
