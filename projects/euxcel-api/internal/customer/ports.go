package customer

import (
	"context"

	domain "github.com/devpablocristo/monorepo/projects/qh/internal/customer/usecases/domain"
)

type UseCases interface {
	GetAllCustomers(context.Context) ([]domain.Customer, error)
	GetCustomerByID(context.Context, int64) (*domain.Customer, error)
	CreateCustomer(context.Context, *domain.Customer) error
	// GetCustomerByEmail(context.Context, string) (*domain.Customer, error)
	// UpdateCustomer(context.Context, *domain.Customer) error
	// DeleteCustomer(context.Context, int64) error
	// GetKPI(context.Context) (*domain.KPI, error)
}

type Repository interface {
	GetAllCustomers(context.Context) ([]domain.Customer, error)
	GetCustomerByID(context.Context, int64) (*domain.Customer, error)
	CreateCustomer(context.Context, *domain.Customer) error
	// Update(context.Context, *domain.Customer) error
	// Delete(context.Context, int64) error
	// GetByEmail(context.Context, string) (*domain.Customer, error)
}
