package dto

import (
	"errors"
	"fmt"
	"time"

	domain "github.com/devpablocristo/monorepo/projects/qh/internal/customer/usecases/domain"
)

// -----------------------------------------------------------------------------
// Definición de DTO (uno) y lista de DTO
// -----------------------------------------------------------------------------

type Customer struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	LastName  string    `db:"last_name"`
	Email     string    `db:"email"`
	Phone     string    `db:"phone"`
	Age       int       `db:"age"`
	BirthDate time.Time `db:"birth_date"`
}

// Definimos un tipo nombrado para poder asociarle métodos a la lista.
type CustomerList []Customer

// -----------------------------------------------------------------------------
// Conversiones: DTO -> Dominio (similar al CustomerJson que ya tienes)
// -----------------------------------------------------------------------------

// ToDomain convierte un solo Customer en un *domain.Customer.
func (m *Customer) ToDomain() *domain.Customer {
	if m == nil {
		return nil
	}
	return &domain.Customer{
		ID:        m.ID,
		Name:      m.Name,
		LastName:  m.LastName,
		Email:     m.Email,
		Phone:     m.Phone,
		Age:       m.Age,
		BirthDate: m.BirthDate,
	}
}

// ToDomainList convierte un slice de Customer en un slice de domain.Customer.
// Observa que devolvemos un slice de valores (no de punteros), igual que en tu JSON.
func (ml CustomerList) ToDomainList() []domain.Customer {
	if len(ml) == 0 {
		return []domain.Customer{}
	}

	domainCustomers := make([]domain.Customer, len(ml))
	for i, model := range ml {
		// Cada model (Customer) se convierte a *domain.Customer y luego
		// asignamos el valor con '*'.
		domainCustomers[i] = *model.ToDomain()
	}
	return domainCustomers
}

// -----------------------------------------------------------------------------
// Conversiones: Dominio -> DTO
// -----------------------------------------------------------------------------

// FromDomain convierte un *domain.Customer en un *Customer.
func FromDomain(customer *domain.Customer) (*Customer, error) {
	if customer == nil {
		return nil, errors.New("candidate cannot be nil")
	}

	return &Customer{
		ID:        customer.ID,
		Name:      customer.Name,
		LastName:  customer.LastName,
		Email:     customer.Email,
		Phone:     customer.Phone,
		Age:       customer.Age,
		BirthDate: customer.BirthDate,
	}, nil
}

// FromDomainList convierte un slice de domain.Customer en un CustomerList.
func FromDomainList(customers []domain.Customer) (CustomerList, error) {
	if len(customers) == 0 {
		return CustomerList{}, fmt.Errorf("customers slice empty")
	}

	models := make(CustomerList, len(customers))
	for i, c := range customers {
		cu, err := FromDomain(&c)
		if err != nil {
			return nil, err
		}
		models[i] = *cu
	}
	return models, nil
}
