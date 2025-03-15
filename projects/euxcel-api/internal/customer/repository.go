package customer

import (
	"context"
	"database/sql"

	sqlite "github.com/devpablocristo/monorepo/pkg/databases/sql/sqlite"
	types "github.com/devpablocristo/monorepo/pkg/types"

	model "github.com/devpablocristo/monorepo/projects/qh/internal/customer/repository/models"
	support "github.com/devpablocristo/monorepo/projects/qh/internal/customer/repository/support"
	domain "github.com/devpablocristo/monorepo/projects/qh/internal/customer/usecases/domain"
)

type repository struct {
	db sqlite.Repository
}

func NewRepository() (Repository, error) {
	r, err := sqlite.Bootstrap()
	if err != nil {
		return nil, types.NewError(
			types.ErrOperationFailed,
			"failed to bootstrap sqlite",
			err,
		)
	}

	if err := initSchema(r.DB()); err != nil {
		return nil, types.NewError(
			types.ErrOperationFailed,
			"failed to initialize schema",
			err,
		)
	}

	return &repository{
		db: r,
	}, nil
}

func initSchema(db *sql.DB) error {
	_, err := db.Exec(support.Schema)
	if err != nil {
		return types.NewError(
			types.ErrOperationFailed,
			"failed to create schema",
			err,
		)
	}

	return nil
}

func (r *repository) GetAllCustomers(ctx context.Context) ([]domain.Customer, error) {
	var customers model.CustomerList
	err := r.db.SelectContext(ctx, &customers, support.SelectAllCustomersQuery)
	if err != nil {
		return nil, types.NewError(
			types.ErrOperationFailed,
			"failed to fetch customers",
			err,
		)
	}

	return customers.ToDomainList(), nil
}

func (r *repository) GetCustomerByID(ctx context.Context, id int64) (*domain.Customer, error) {
	model, err := support.ScanCustomer(r.db.QueryRowContext(ctx, support.SelectCustomerByIDQuery, id))
	if err != nil {
		if types.IsNotFound(err) {
			return nil, types.NewError(
				types.ErrNotFound,
				"customer not found",
				err,
			)
		}
		return nil, types.NewError(
			types.ErrOperationFailed,
			"failed to get customer",
			err,
		)
	}
	return model.ToDomain(), nil
}

// func (r *repository) GetByEmail(ctx context.Context, email string) (*domain.Customer, error) {
// 	model, err := scanCustomer(r.db.QueryRowContext(ctx, selectCustomerByEmailQuery, email))
// 	if err != nil {
// 		return nil, err
// 	}
// 	return model.CustomerToDomain(model), nil
// }

func (r *repository) CreateCustomer(ctx context.Context, customer *domain.Customer) error {
	// if err := r.validateEmailConflict(ctx, 0, customer.Email); err != nil {
	// 	return err
	// }

	model, err := model.FromDomain(customer)
	if err != nil {
		return err
	}
	_, err = r.db.DB().ExecContext(ctx, support.InsertCustomerQuery,
		model.Name, model.LastName, model.Email,
		model.Phone, model.Age, model.BirthDate,
	)
	if err != nil {
		if types.IsConflict(err) {
			return err
		}
		return types.NewError(
			types.ErrOperationFailed,
			"failed to create customer",
			err,
		)
	}
	return nil
}

// func (r *repository) Update(ctx context.Context, customer *domain.Customer) error {
// 	// Verificar que existe el customer primero
// 	var customers []model.Customer
// 	err := r.db.SelectContext(ctx, &customers, selectCustomerByIDQuery, customer.ID)
// 	if err != nil {
// 		return types.NewError(
// 			types.ErrOperationFailed,
// 			"failed to fetch customer",
// 			err,
// 		)
// 	}
// 	if len(customers) == 0 {
// 		return types.NewError(
// 			types.ErrNotFound,
// 			"customer not found",
// 			errors.New("no details available"),
// 		)
// 	}

// 	// Validar conflicto de email
// 	if err := r.validateEmailConflict(ctx, customer.ID, customer.Email); err != nil {
// 		return err
// 	}

// 	// Hacer el update - usando el mismo wrapper
// 	model := model.DomainToCustomer(customer)
// 	result, err := r.db.ExecContext(ctx, updateCustomerQuery, // <-- Aquí está el cambio
// 		model.Name, model.LastName, model.Email,
// 		model.Phone, model.Age, model.BirthDate, model.ID,
// 	)
// 	if err != nil {
// 		return types.NewError(
// 			types.ErrOperationFailed,
// 			"failed to update customer",
// 			err,
// 		)
// 	}

// 	return validateRows(result)
// }

// func (r *repository) Update(ctx context.Context, customer *domain.Customer) error {

// 	fmt.Println(customer)
// 	var existingCustomer model.Customer
// 	query := selectCustomerByIDQuery // usar la constante que ya tienes definida
// 	fmt.Printf("Query: %s\nID: %d\n", query, customer.ID)
// 	err := r.db.SelectContext(ctx, &existingCustomer, query, customer.ID)
// 	if err != nil {
// 		fmt.Printf("Error: %v\n", err)
// 	}
// 	fmt.Printf("Existing customer: %+v\n", existingCustomer)

// 	if err := r.validateEmailConflict(ctx, customer.ID, customer.Email); err != nil {
// 		return err
// 	}

// 	model := model.DomainToCustomer(customer)
// 	result, err := r.db.DB().ExecContext(ctx, updateCustomerQuery,
// 		model.Name, model.LastName, model.Email,
// 		model.Phone, model.Age, model.BirthDate, model.ID,
// 	)
// 	if err != nil {
// 		return types.NewError(
// 			types.ErrOperationFailed,
// 			"failed to update customer",
// 			err,
// 		)
// 	}

// 	return validateRows(result)
// }

// func (r *repository) Delete(ctx context.Context, id int64) error {
// 	result, err := r.db.DB().ExecContext(ctx, deleteCustomerQuery, id)
// 	if err != nil {
// 		return types.NewError(
// 			types.ErrOperationFailed,
// 			"failed to delete customer",
// 			err,
// 		)
// 	}

// 	return validateRows(result)
// }

// func (r *repository) validateEmailConflict(ctx context.Context, customerID int64, email string) error {
// 	existing, err := r.GetByEmail(ctx, email)
// 	if err != nil {
// 		if !types.IsNotFound(err) {
// 			return err
// 		}
// 		return nil
// 	}
// 	if existing != nil && existing.ID != customerID {
// 		return types.NewError(
// 			types.ErrConflict,
// 			fmt.Sprintf("email %s is already in use by another customer", email),
// 			errors.New("no details available"),
// 		)
// 	}
// 	return nil
// }
