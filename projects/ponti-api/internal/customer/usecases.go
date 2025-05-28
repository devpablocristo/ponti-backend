package customer

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/usecases/domain"
)

type RepositoryPort interface {
	CreateCustomer(ctx context.Context, c *domain.Customer) (int64, error)
	ListCustomers(ctx context.Context, page, perPage int) ([]domain.ListedCustomer, int64, error)
	GetCustomer(ctx context.Context, id int64) (*domain.Customer, error)
	UpdateCustomer(ctx context.Context, c *domain.Customer) error
	DeleteCustomer(ctx context.Context, id int64) error
}

type UseCases struct {
	repo RepositoryPort
}

// NewUseCases crea una instancia de los casos de uso para Customer.
func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) CreateCustomer(ctx context.Context, c *domain.Customer) (int64, error) {
	return u.repo.CreateCustomer(ctx, c)
}

func (u *UseCases) ListCustomers(ctx context.Context, page, perPage int) ([]domain.ListedCustomer, int64, error) {
	return u.repo.ListCustomers(ctx, page, perPage)
}

func (u *UseCases) GetCustomer(ctx context.Context, id int64) (*domain.Customer, error) {
	return u.repo.GetCustomer(ctx, id)
}

func (u *UseCases) UpdateCustomer(ctx context.Context, c *domain.Customer) error {
	return u.repo.UpdateCustomer(ctx, c)
}

func (u *UseCases) DeleteCustomer(ctx context.Context, id int64) error {
	return u.repo.DeleteCustomer(ctx, id)
}
