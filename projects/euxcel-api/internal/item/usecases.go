package item

import (
	"context"

	authe "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/authe"
	config "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/config"
	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/item/usecases/domain"
)

// useCases implementa la interfaz UseCases.
type useCases struct {
	repository Repository
	config     config.Loader
	autheUc    authe.UseCases
}

// NewUseCases crea una instancia de useCases con las dependencias adecuadas.
func NewUseCases(
	repo Repository,
	cfg config.Loader,
	au authe.UseCases,
) UseCases {
	return &useCases{
		repository: repo,
		config:     cfg,
		autheUc:    au,
	}
}

// CreateItem crea un nuevo item y lo guarda.
func (u *useCases) CreateItem(ctx context.Context, item *domain.Item) (int64, error) {
	return u.repository.CreateItem(ctx, item)
}

// ListItems obtiene la lista de todos los items.
func (u *useCases) ListItems(ctx context.Context) ([]domain.Item, error) {
	return u.repository.ListItems(ctx)
}

// GetItem obtiene un item por su ID.
func (u *useCases) GetItem(ctx context.Context, itemID int64) (*domain.Item, error) {
	return u.repository.GetItem(ctx, itemID)
}

// DeleteItem elimina un item por su ID.
func (u *useCases) DeleteItem(ctx context.Context, itemID int64) error {
	return u.repository.DeleteItem(ctx, itemID)
}

// UpdateItem actualiza un item existente.
func (u *useCases) UpdateItem(ctx context.Context, updateItem *domain.Item) error {
	return u.repository.UpdateItem(ctx, updateItem)
}
