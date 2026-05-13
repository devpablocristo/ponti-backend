// Package customer implementa el repositorio de clientes.
package customer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/devpablocristo/core/errors/go/domainerr"
	actorsync "github.com/devpablocristo/ponti-backend/internal/actor"
	models "github.com/devpablocristo/ponti-backend/internal/customer/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/customer/usecases/domain"
	"github.com/devpablocristo/ponti-backend/internal/shared/authz"
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
	var id int64
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result, err := actorsync.EnsureCustomerFromActor(tx, actorsync.EnsureCustomerInput{
			ActorID:   c.ActorID,
			Name:      c.Name,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			CreatedBy: c.CreatedBy,
			UpdatedBy: c.UpdatedBy,
		})
		if err != nil {
			return err
		}
		if result != nil {
			id = result.CustomerID
			c.ActorID = &result.ActorID
			return nil
		}

		model := models.FromDomain(c)
		model.Base = sharedmodels.Base{
			CreatedBy: c.CreatedBy,
			UpdatedBy: c.UpdatedBy,
		}
		if tenantID, ok, err := authz.OptionalTenantOrStrict(ctx); err != nil {
			return err
		} else if ok {
			model.TenantID = tenantID
		}
		if err := tx.Create(model).Error; err != nil {
			return domainerr.Internal("failed to create customer")
		}
		id = model.ID
		return nil
	})
	return id, err
}

type listedCustomerRow struct {
	ID      int64
	Name    string
	ActorID *int64
}

func (r *Repository) ListCustomers(ctx context.Context, page, perPage int) ([]domain.ListedCustomer, int64, error) {
	var rows []listedCustomerRow
	var total int64

	db0 := r.db.Client().WithContext(ctx).
		Table("customers c").
		Where("c.deleted_at IS NULL")
	db0 = authz.MaybeTenantScope(ctx, db0, "c")

	// Conteo total
	if err := db0.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count customers")
	}

	// Consulta ligera: sólo id y name
	listDB := db0.
		Select("c.id, c.name, COALESCE(c.actor_id, m.actor_id) AS actor_id").
		Joins("LEFT JOIN legacy_actor_map m ON m.source_table = 'customers' AND m.source_id = c.id AND m.tenant_id = c.tenant_id")
	if err := listDB.
		Limit(perPage).
		Offset((page - 1) * perPage).
		Scan(&rows).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to list customers")
	}

	// Mapear a dominio ligero
	customers := make([]domain.ListedCustomer, len(rows))
	for i, m := range rows {
		customers[i] = domain.ListedCustomer{
			ID:      m.ID,
			Name:    m.Name,
			ActorID: m.ActorID,
		}
	}

	return customers, total, nil
}

func (r *Repository) ListArchivedCustomers(ctx context.Context, page, perPage int) ([]domain.ListedCustomer, int64, error) {
	var rows []listedCustomerRow
	var total int64

	db0 := r.db.Client().WithContext(ctx).
		Unscoped().
		Table("customers c").
		Where("c.deleted_at IS NOT NULL")
	db0 = authz.MaybeTenantScope(ctx, db0, "c")

	if err := db0.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count archived customers")
	}

	if err := db0.
		Select("c.id, c.name, COALESCE(c.actor_id, m.actor_id) AS actor_id").
		Joins("LEFT JOIN legacy_actor_map m ON m.source_table = 'customers' AND m.source_id = c.id AND m.tenant_id = c.tenant_id").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Scan(&rows).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to list archived customers")
	}

	customers := make([]domain.ListedCustomer, len(rows))
	for i, m := range rows {
		customers[i] = domain.ListedCustomer{
			ID:      m.ID,
			Name:    m.Name,
			ActorID: m.ActorID,
		}
	}

	return customers, total, nil
}

func (r *Repository) GetCustomer(ctx context.Context, id int64) (*domain.Customer, error) {
	var model models.Customer
	db0 := authz.MaybeTenantScope(ctx, r.db.Client().WithContext(ctx), "customers")
	err := db0.
		Where("id = ? AND deleted_at IS NULL", id).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerr.New(domainerr.KindNotFound, fmt.Sprintf("customer with id %d not found", id))
		}
		return nil, domainerr.Internal("failed to get customer")
	}
	out := model.ToDomain()
	if out.ActorID == nil {
		actorID, err := actorsync.ActorIDForLegacy(r.db.Client().WithContext(ctx), actorsync.LegacyCustomers, id)
		if err != nil {
			return nil, err
		}
		if actorID > 0 {
			out.ActorID = &actorID
		}
	}
	if out.ActorID != nil {
		actorID := *out.ActorID
		out.ActorID = &actorID
	}
	return out, nil
}

func (r *Repository) UpdateCustomer(ctx context.Context, c *domain.Customer) error {
	if err := sharedrepo.ValidateEntity(c, "customer"); err != nil {
		return err
	}
	if err := sharedrepo.ValidateID(c.ID, "customer"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updateTx := authz.MaybeTenantScope(ctx, tx.Model(&models.Customer{}), "customers").Where("id = ?", c.ID)
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
		if c.ActorID != nil && *c.ActorID > 0 {
			_, err := actorsync.LinkLegacyEntityToActor(tx, actorsync.LegacyActorSync{
				SourceTable: actorsync.LegacyCustomers,
				SourceID:    c.ID,
				Name:        c.Name,
				ActorKind:   actorsync.KindOrganization,
				Role:        actorsync.RoleCliente,
				UpdatedAt:   time.Now(),
				UpdatedBy:   c.UpdatedBy,
			}, *c.ActorID)
			return err
		}
		_, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
			SourceTable: actorsync.LegacyCustomers,
			SourceID:    c.ID,
			Name:        c.Name,
			ActorKind:   actorsync.KindOrganization,
			Role:        actorsync.RoleCliente,
			UpdatedAt:   time.Now(),
			UpdatedBy:   c.UpdatedBy,
		})
		return err
	})
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
		customerQuery := authz.MaybeTenantScope(ctx, tx.Unscoped(), "customers")
		if err := customerQuery.Where("id = ?", id).First(&customer).Error; err != nil {
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
			Where("tenant_id = ?", customer.TenantID).
			Count(&activeProjects).Error; err != nil {
			return domainerr.Internal("failed to check active projects")
		}
		if activeProjects > 0 {
			return domainerr.Conflict("customer has active projects")
		}

		archivedAt := time.Now()
		if err := authz.MaybeTenantScope(ctx, tx.Model(&models.Customer{}), "customers").
			Where("id = ?", id).
			Updates(map[string]any{
				"deleted_at": archivedAt,
				"deleted_by": deletedBy,
			}).Error; err != nil {
			return domainerr.Internal("failed to archive customer")
		}
		if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
			SourceTable: actorsync.LegacyCustomers,
			SourceID:    customer.ID,
			Name:        customer.Name,
			ActorKind:   actorsync.KindOrganization,
			Role:        actorsync.RoleCliente,
			ArchivedAt:  &archivedAt,
			UpdatedAt:   archivedAt,
			UpdatedBy:   customer.UpdatedBy,
			DeletedBy:   deletedBy,
		}); err != nil {
			return err
		}
		return nil
	})
}

func (r *Repository) RestoreCustomer(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "customer"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		restoredAt := time.Now()
		var customer models.Customer
		customerQuery := authz.MaybeTenantScope(ctx, tx.Unscoped(), "customers")
		if err := customerQuery.Where("id = ?", id).First(&customer).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("customer %d not found", id))
			}
			return domainerr.Internal("failed to get customer")
		}
		if !customer.DeletedAt.Valid {
			return domainerr.Conflict("customer is not archived")
		}

		if err := authz.MaybeTenantScope(ctx, tx.Unscoped().Model(&models.Customer{}), "customers").
			Where("id = ?", id).
			Updates(map[string]any{
				"deleted_at": nil,
				"deleted_by": nil,
				"updated_at": restoredAt,
			}).Error; err != nil {
			return domainerr.Internal("failed to restore customer")
		}
		if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
			SourceTable: actorsync.LegacyCustomers,
			SourceID:    customer.ID,
			Name:        customer.Name,
			ActorKind:   actorsync.KindOrganization,
			Role:        actorsync.RoleCliente,
			UpdatedAt:   restoredAt,
			UpdatedBy:   customer.UpdatedBy,
		}); err != nil {
			return err
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
		customerDB := authz.MaybeTenantScope(ctx, tx.Unscoped().Model(&models.Customer{}), "customers")
		if err := customerDB.Where("id = ?", id).Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check customer existence")
		}
		if count == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("customer with id %d does not exist", id))
		}

		var activeCount, archivedCount int64
		if err := tx.Table("projects").
			Where("customer_id = ? AND deleted_at IS NULL", id).
			Where("tenant_id IN (SELECT tenant_id FROM customers WHERE id = ?)", id).
			Count(&activeCount).Error; err != nil {
			return domainerr.Internal("failed to count active projects")
		}
		if err := tx.Unscoped().Table("projects").
			Where("customer_id = ? AND deleted_at IS NOT NULL", id).
			Where("tenant_id IN (SELECT tenant_id FROM customers WHERE id = ?)", id).
			Count(&archivedCount).Error; err != nil {
			return domainerr.Internal("failed to count archived projects")
		}
		if activeCount+archivedCount > 0 {
			return domainerr.Conflict(fmt.Sprintf("customer has %d project(s) (%d active, %d archived); archive or hard-delete them first", activeCount+archivedCount, activeCount, archivedCount))
		}

		var deletedBy *string
		if actor, err := sharedmodels.ActorFromContext(ctx); err == nil {
			deletedBy = &actor
		}
		if err := actorsync.DeleteLegacyActor(tx, actorsync.LegacyCustomers, id, actorsync.RoleCliente, deletedBy); err != nil {
			return err
		}
		if err := authz.MaybeTenantScope(ctx, tx.Unscoped(), "customers").Delete(&models.Customer{}, "id = ?", id).Error; err != nil {
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
