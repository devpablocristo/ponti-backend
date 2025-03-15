package support

import (
	"database/sql"
	"errors"

	types "github.com/devpablocristo/monorepo/pkg/types"
	model "github.com/devpablocristo/monorepo/projects/qh/internal/customer/repository/models"
)

const (
	// Schema definition
	Schema = `
        CREATE TABLE IF NOT EXISTS customers (
            id          INTEGER PRIMARY KEY AUTOINCREMENT,
            name        TEXT NOT NULL,
            last_name   TEXT NOT NULL,
            email       TEXT NOT NULL UNIQUE,
            phone       TEXT NOT NULL,
            age         INTEGER NOT NULL,
            birth_date  DATETIME NOT NULL
        );
    `

	// Base select query
	SelectAllCustomersQuery = `
        SELECT  id, 
                name, 
                last_name, 
                email, 
                phone, 
                age, 
                birth_date
        FROM    customers
    `

	// Select queries
	SelectCustomerByIDQuery    = SelectAllCustomersQuery + ` WHERE id = ?`
	SelectCustomerByEmailQuery = SelectAllCustomersQuery + ` WHERE email = ?`

	// Insert query
	InsertCustomerQuery = `
        INSERT INTO customers (
            name,
            last_name,
            email,
            phone,
            age,
            birth_date
        ) VALUES (?, ?, ?, ?, ?, ?)
    `

	// Update query
	UpdateCustomerQuery = `
        UPDATE  customers 
        SET     name = ?, 
                last_name = ?, 
                email = ?, 
                phone = ?, 
                age = ?, 
                birth_date = ?
        WHERE   id = ?
    `

	// Delete query
	DeleteCustomerQuery = `DELETE FROM customers WHERE id = ?`
)

func ValidateRows(result sql.Result) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return types.NewError(
			types.ErrOperationFailed,
			"failed to get affected rows",
			err,
		)
	}

	if rows == 0 {
		return types.NewError(
			types.ErrNotFound,
			"No rows were affected",
			errors.New("no details available"),
		)
	}
	return nil
}

func ScanCustomer(row *sql.Row) (*model.Customer, error) {
	var model model.Customer
	err := row.Scan(
		&model.ID, &model.Name, &model.LastName, &model.Email,
		&model.Phone, &model.Age, &model.BirthDate,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, types.NewError(
				types.ErrNotFound,
				"customer not found",
				err,
			)
		}
		return nil, types.NewError(
			types.ErrOperationFailed,
			"failed to fetch customer",
			err,
		)
	}
	return &model, nil
}
