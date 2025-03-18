package supplier

import (
	"context"

	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/supplier/usecases/domain"
)

// UseCases defines business operations for Supplier.
type UseCases interface {
	CreateSupplier(ctx context.Context, s *domain.Supplier) (int64, error)
	ListSuppliers(ctx context.Context) ([]domain.Supplier, error)
	GetSupplier(ctx context.Context, id int64) (*domain.Supplier, error)
	UpdateSupplier(ctx context.Context, s *domain.Supplier) error
	DeleteSupplier(ctx context.Context, id int64) error
}

// Repository defines persistence operations for Supplier.
type Repository interface {
	CreateSupplier(ctx context.Context, s *domain.Supplier) (int64, error)
	ListSuppliers(ctx context.Context) ([]domain.Supplier, error)
	GetSupplier(ctx context.Context, id int64) (*domain.Supplier, error)
	UpdateSupplier(ctx context.Context, s *domain.Supplier) error
	DeleteSupplier(ctx context.Context, id int64) error
}
