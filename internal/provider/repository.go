package provider

import (
	"context"

	models "github.com/devpablocristo/ponti-backend/internal/provider/repository/models"
	"github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
	types "github.com/devpablocristo/ponti-backend/pkg/types"
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
	if err := r.db.Client().WithContext(ctx).Find(&providers).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list providers", err)
	}
	res := make([]domain.Provider, len(providers))
	for i := range providers {
		res[i] = *providers[i].ToDomain()
	}
	return res, nil
}
