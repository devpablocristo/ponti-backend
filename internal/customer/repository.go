// Package customer implementa el repositorio de clientes.
package customer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/devpablocristo/core/errors/go/domainerr"
	models "github.com/devpablocristo/ponti-backend/internal/customer/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/customer/usecases/domain"
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
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, domainerr.Internal("failed to create customer")
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
	err := r.db.Client().WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&model).Error
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
	actor, err := sharedmodels.ActorFromContext(ctx)
	if err != nil {
		return err
	}
	deletedBy := &actor

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var customer models.Customer
		if err := tx.Unscoped().Where("id = ?", id).First(&customer).Error; err != nil {
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

		if err := tx.Model(&models.Customer{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"deleted_at": time.Now(),
				"deleted_by": deletedBy,
			}).Error; err != nil {
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
		if err := tx.Unscoped().Where("id = ?", id).First(&customer).Error; err != nil {
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

// HardDeleteCustomer elimina definitivamente un cliente.
// Bloquea con 409 si tiene proyectos (activos o archivados).
func (r *Repository) HardDeleteCustomer(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "customer"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Unscoped().Model(&models.Customer{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check customer existence")
		}
		if count == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("customer with id %d does not exist", id))
		}

		var activeCount, archivedCount int64
		if err := tx.Table("projects").
			Where("customer_id = ? AND deleted_at IS NULL", id).
			Count(&activeCount).Error; err != nil {
			return domainerr.Internal("failed to count active projects")
		}
		if err := tx.Unscoped().Table("projects").
			Where("customer_id = ? AND deleted_at IS NOT NULL", id).
			Count(&archivedCount).Error; err != nil {
			return domainerr.Internal("failed to count archived projects")
		}
		if activeCount+archivedCount > 0 {
			return domainerr.Conflict(fmt.Sprintf("customer has %d project(s) (%d active, %d archived); archive or hard-delete them first", activeCount+archivedCount, activeCount, archivedCount))
		}

		if err := tx.Unscoped().Delete(&models.Customer{}, "id = ?", id).Error; err != nil {
			return domainerr.Internal("failed to hard delete customer")
		}
		return nil
	})
}

// DeleteCustomer queda como alias hacia HardDeleteCustomer.
// Deprecated: usar HardDeleteCustomer.
func (r *Repository) DeleteCustomer(ctx context.Context, id int64) error {
	return r.HardDeleteCustomer(ctx, id)
}

