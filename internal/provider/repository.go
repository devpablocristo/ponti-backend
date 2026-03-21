package provider

import (
	"context"

	"github.com/devpablocristo/core/backend/go/domainerr"
	models "github.com/devpablocristo/ponti-backend/internal/provider/repository/models"
	"github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
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
		return nil, domainerr.Internal("failed to list providers")
	}
	res := make([]domain.Provider, len(providers))
	for i := range providers {
		res[i] = *providers[i].ToDomain()
	}
	return res, nil
}
