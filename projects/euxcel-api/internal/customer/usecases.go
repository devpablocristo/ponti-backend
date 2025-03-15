package customer

import (
	"context"

	types "github.com/devpablocristo/monorepo/pkg/types"

	domain "github.com/devpablocristo/monorepo/projects/qh/internal/customer/usecases/domain"
)

type useCases struct {
	repo Repository
}

func NewUseCases(r Repository) UseCases {
	return &useCases{
		repo: r,
	}
}

func (uc *useCases) GetAllCustomers(ctx context.Context) ([]domain.Customer, error) {
	customers, err := uc.repo.GetAllCustomers(ctx)
	if err != nil {
		// return nil, types.NewError(
		// 	types.ErrOperationFailed,
		// 	"failed to get customers",
		// 	err,
		// )
		return nil, err
	}
	return customers, nil
}

func (uc *useCases) GetCustomerByID(ctx context.Context, ID int64) (*domain.Customer, error) {
	customer, err := uc.repo.GetCustomerByID(ctx, ID)
	if err != nil {
		return nil, err
	}
	return customer, nil
}

func (uc *useCases) CreateCustomer(ctx context.Context, customer *domain.Customer) error {
	if customer == nil {
		return types.NewError(
			types.ErrInvalidInput,
			"customer cannot be nil",
			nil,
		)
	}

	if err := uc.repo.CreateCustomer(ctx, customer); err != nil {
		return err
	}
	return nil
}

// func (uc *useCases) GetCustomerByEmail(ctx context.Context, email string) (*domain.Customer, error) {
// 	customer, err := uc.repo.GetByEmail(ctx, email)
// 	if err != nil {
// 		if types.IsNotFound(err) {
// 			return nil, types.NewError(
// 				types.ErrNotFound,
// 				"customer not found",
// 				err,
// 			)
// 		}
// 		return nil, types.NewError(
// 			types.ErrOperationFailed,
// 			"failed to get customer by email",
// 			err,
// 		)
// 	}
// 	return customer, nil
// }

// func (uc *useCases) UpdateCustomer(ctx context.Context, customer *domain.Customer) error {
// 	if err := uc.repo.Update(ctx, customer); err != nil {
// 		if types.IsNotFound(err) {
// 			return err
// 		}
// 		if types.IsConflict(err) {
// 			return err
// 		}
// 		return types.NewError(
// 			types.ErrOperationFailed,
// 			"failed to update customer",
// 			err,
// 		)
// 	}
// 	return nil
// }

// func (uc *useCases) DeleteCustomer(ctx context.Context, ID int64) error {
// 	if err := uc.repo.Delete(ctx, ID); err != nil {
// 		if types.IsNotFound(err) {
// 			return err
// 		}
// 		return types.NewError(
// 			types.ErrOperationFailed,
// 			"failed to delete customer",
// 			err,
// 		)
// 	}
// 	return nil
// }

// func (uc *useCases) GetKPI(ctx context.Context) (*domain.KPI, error) {
// 	customers, err := uc.repo.GetAll(ctx)
// 	if err != nil {
// 		return nil, types.NewError(
// 			types.ErrOperationFailed,
// 			"failed to calculate KPI",
// 			err,
// 		)
// 	}

// 	return support.CalculateKPI(customers), nil
// }
