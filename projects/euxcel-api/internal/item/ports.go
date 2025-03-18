package item

import (
	"context"

	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/item/usecases/domain"
)

// UseCases define las operaciones de negocio para items.
type UseCases interface {
	CreateItem(ctx context.Context, item *domain.Item) (int64, error)
	ListItems(ctx context.Context) ([]domain.Item, error)
	GetItem(ctx context.Context, itemID int64) (*domain.Item, error)
	DeleteItem(ctx context.Context, itemID int64) error
	UpdateItem(ctx context.Context, updateItem *domain.Item) error
}

// Repository define las operaciones que el adaptador GORM debe implementar.
type Repository interface {
	CreateItem(ctx context.Context, item *domain.Item) (int64, error)
	ListItems(ctx context.Context) ([]domain.Item, error)
	GetItem(ctx context.Context, id int64) (*domain.Item, error)
	UpdateItem(ctx context.Context, item *domain.Item) error
	DeleteItem(ctx context.Context, id int64) error
}
