package provider

import (
	"context"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	models "github.com/devpablocristo/ponti-backend/internal/provider/repository/models"
	"github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
	sharedfilters "github.com/devpablocristo/ponti-backend/internal/shared/filters"
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

func (r *Repository) GetProviders(ctx context.Context) ([]domain.Provider, error) {
	var providers []models.Provider

	db0 := r.db.Client().WithContext(ctx).
		Model(&models.Provider{})

	// T1.e: acotar al tenant activo (flag-gated).
	db0 = sharedfilters.ScopeTenant(ctx, db0)

	if err := db0.Find(&providers).Error; err != nil {
		return nil, domainerr.Internal("failed to list providers")
	}
	res := make([]domain.Provider, len(providers))
	for i := range providers {
		res[i] = *providers[i].ToDomain()
	}
	return res, nil
}
