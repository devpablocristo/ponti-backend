package provider

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	models "github.com/devpablocristo/ponti-backend/internal/provider/repository/models"
	"github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
	"github.com/devpablocristo/platform/persistence/gorm/go/tenancy"

	"github.com/devpablocristo/ponti-backend/internal/shared/authz"
	"github.com/devpablocristo/ponti-backend/internal/shared/lifecycle"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"
	"gorm.io/gorm"
)

// GormEnginePort expone el cliente de base de datos requerido.
type GormEnginePort interface {
	Client() *gorm.DB
}

type Repository struct {
	db GormEnginePort
}

func NewRepository(db GormEnginePort) *Repository {
	return &Repository{db: db}
}

type providerRow struct {
	ID      int64
	Name    string
	ActorID *int64
}

func (r *Repository) GetProviders(ctx context.Context) ([]domain.Provider, error) {
	if r.db.Client().Name() == "sqlite" {
		var providers []models.Provider
		if err := tenancy.Scope(ctx, r.db.Client().WithContext(ctx), "providers").Find(&providers).Error; err != nil {
			return nil, domainerr.Internal("failed to list providers")
		}
		res := make([]domain.Provider, len(providers))
		for i := range providers {
			res[i] = *providers[i].ToDomain()
		}
		return res, nil
	}

	var providers []providerRow
	base := r.db.Client().WithContext(ctx).
		Table("providers p").
		Where("p.deleted_at IS NULL")
	base = tenancy.Scope(ctx, base, "p")
	if err := base.
		Select("p.id, p.name, lm.actor_id").
		Joins("LEFT JOIN legacy_actor_map lm ON lm.source_table = 'providers' AND lm.source_id = p.id AND lm.tenant_id = p.tenant_id").
		Order("p.name ASC").
		Scan(&providers).Error; err != nil {
		return nil, domainerr.Internal("failed to list providers")
	}
	res := make([]domain.Provider, len(providers))
	for i := range providers {
		res[i] = domain.Provider{
			ID:      providers[i].ID,
			Name:    providers[i].Name,
			ActorID: providers[i].ActorID,
		}
	}
	return res, nil
}

func (r *Repository) ListArchivedProviders(ctx context.Context) ([]domain.Provider, error) {
	if r.db.Client().Name() == "sqlite" {
		var providers []models.Provider
		if err := tenancy.Scope(ctx, r.db.Client().WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL"), "providers").Find(&providers).Error; err != nil {
			return nil, domainerr.Internal("failed to list archived providers")
		}
		res := make([]domain.Provider, len(providers))
		for i := range providers {
			res[i] = *providers[i].ToDomain()
		}
		return res, nil
	}

	var providers []providerRow
	base := r.db.Client().WithContext(ctx).
		Unscoped().
		Table("providers p").
		Where("p.deleted_at IS NOT NULL")
	base = tenancy.Scope(ctx, base, "p")
	if err := base.
		Select("p.id, p.name, lm.actor_id").
		Joins("LEFT JOIN legacy_actor_map lm ON lm.source_table = 'providers' AND lm.source_id = p.id AND lm.tenant_id = p.tenant_id").
		Order("p.deleted_at DESC").
		Scan(&providers).Error; err != nil {
		return nil, domainerr.Internal("failed to list archived providers")
	}
	res := make([]domain.Provider, len(providers))
	for i := range providers {
		res[i] = domain.Provider{
			ID:      providers[i].ID,
			Name:    providers[i].Name,
			ActorID: providers[i].ActorID,
		}
	}
	return res, nil
}

func (r *Repository) GetProvider(ctx context.Context, id int64) (*domain.Provider, error) {
	if err := sharedrepo.ValidateID(id, "provider"); err != nil {
		return nil, err
	}
	var provider models.Provider
	if err := tenancy.Scope(ctx, r.db.Client().WithContext(ctx), "providers").Where("id = ?", id).First(&provider).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerr.NotFound("provider not found")
		}
		return nil, domainerr.Internal("failed to get provider")
	}
	return provider.ToDomain(), nil
}

func (r *Repository) CreateProvider(ctx context.Context, provider *domain.Provider) (int64, error) {
	if provider == nil || provider.Name == "" {
		return 0, domainerr.Validation("provider name is required")
	}
	model := models.FromDomain(provider)
	if tenantID, ok, err := authz.OptionalTenantOrStrict(ctx); err != nil {
		return 0, err
	} else if ok {
		model.TenantID = tenantID
	}
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, domainerr.Internal("failed to create provider")
	}
	return model.ID, nil
}

func (r *Repository) UpdateProvider(ctx context.Context, provider *domain.Provider) error {
	if provider == nil || provider.Name == "" {
		return domainerr.Validation("provider name is required")
	}
	if err := sharedrepo.ValidateID(provider.ID, "provider"); err != nil {
		return err
	}
	result := tenancy.Scope(ctx, r.db.Client().WithContext(ctx).Model(&models.Provider{}), "providers").
		Where("id = ?", provider.ID).
		Updates(map[string]any{"name": provider.Name})
	if result.Error != nil {
		return domainerr.Internal("failed to update provider")
	}
	if result.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("provider %d not found", provider.ID))
	}
	return nil
}

func (r *Repository) ArchiveProvider(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "provider"); err != nil {
		return err
	}
	actor, err := sharedmodels.ActorFromContext(ctx)
	if err != nil {
		return err
	}
	deletedBy := &actor
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var provider models.Provider
		if err := tenancy.Scope(ctx, tx.Unscoped(), "providers").Where("id = ?", id).First(&provider).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.NotFound("provider not found")
			}
			return domainerr.Internal("failed to get provider")
		}
		if provider.DeletedAt.Valid {
			return domainerr.Conflict("provider already archived")
		}
		archivedAt := time.Now()
		cause, err := lifecycle.RootCause(tx, provider.TenantID, "providers", id, nil, deletedBy)
		if err != nil {
			return err
		}
		result := tenancy.Scope(ctx, tx.Model(&models.Provider{}), "providers").
			Where("id = ?", id).
			Updates(lifecycle.ArchiveUpdates(tx, "providers", archivedAt, deletedBy, cause))
		if result.Error != nil {
			return domainerr.Internal("failed to archive provider")
		}
		if result.RowsAffected == 0 {
			return domainerr.NotFound("provider not found")
		}
		return nil
	})
}

func (r *Repository) RestoreProvider(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "provider"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var provider models.Provider
		if err := tenancy.Scope(ctx, tx.Unscoped(), "providers").Where("id = ?", id).First(&provider).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.NotFound("provider not found")
			}
			return domainerr.Internal("failed to get provider")
		}
		if !provider.DeletedAt.Valid {
			return domainerr.Conflict("provider is not archived")
		}
		result := tenancy.Scope(ctx, tx.Unscoped().Model(&models.Provider{}), "providers").
			Where("id = ?", id).
			Updates(lifecycle.RestoreUpdates(tx, "providers", time.Now()))
		if result.Error != nil {
			return domainerr.Internal("failed to restore provider")
		}
		if result.RowsAffected == 0 {
			return domainerr.NotFound("provider not found")
		}
		return nil
	})
}

func (r *Repository) HardDeleteProvider(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "provider"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		providerDB := tenancy.Scope(ctx, tx.Unscoped().Table("providers"), "providers")
		var count int64
		if err := providerDB.Where("id = ?", id).Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check provider existence")
		}
		if count == 0 {
			return domainerr.NotFound("provider not found")
		}
		if err := lifecycle.RequireArchived(providerDB, "providers", "provider", id); err != nil {
			return err
		}
		var movements int64
		if err := tenancy.Scope(ctx, tx.Unscoped().Table("supply_movements"), "supply_movements").
			Where("provider_id = ?", id).
			Count(&movements).Error; err != nil {
			return domainerr.Internal("failed to check provider usage")
		}
		if movements > 0 {
			return domainerr.Conflict(fmt.Sprintf("provider has %d supply movement reference(s); remove them first", movements))
		}
		result := tenancy.Scope(ctx, tx.Unscoped(), "providers").Delete(&models.Provider{}, id)
		if result.Error != nil {
			return domainerr.Internal("failed to hard delete provider")
		}
		if result.RowsAffected == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("provider %d not found", id))
		}
		return nil
	})
}
