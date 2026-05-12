package provider

import (
	"context"

	"github.com/devpablocristo/core/errors/go/domainerr"
	models "github.com/devpablocristo/ponti-backend/internal/provider/repository/models"
	"github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
	"github.com/devpablocristo/ponti-backend/internal/shared/authz"
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
		if err := r.db.Client().WithContext(ctx).Find(&providers).Error; err != nil {
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
	base = authz.MaybeTenantScope(ctx, base, "p")
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
