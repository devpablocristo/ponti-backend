package dto

import (
	"time"

	"github.com/devpablocristo/monorepo/projects/qh/internal/customer/usecases/domain"
)

// -----------------------------------------------------------------------------
// Definición de struct y tipo nombrado para listas
// -----------------------------------------------------------------------------

type CustomerJson struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name" binding:"required"`
	LastName  string    `json:"last_name" binding:"required"`
	Email     string    `json:"email" binding:"required"`
	Phone     string    `json:"phone"`
	Age       int       `json:"age" binding:"required"`
	BirthDate time.Time `json:"birth_date" binding:"required"`
}

// Definimos un tipo nombrado para poder asociarle métodos a la lista.
type CustomerJsonList []CustomerJson

// -----------------------------------------------------------------------------
// Conversiones: DTO -> Dominio
// -----------------------------------------------------------------------------

// ToDomain convierte un objeto CustomerJson a *domain.Customer.
func (cj *CustomerJson) ToDomain() *domain.Customer {
	if cj == nil {
		return nil
	}
	return &domain.Customer{
		ID:        cj.ID,
		Name:      cj.Name,
		LastName:  cj.LastName,
		Email:     cj.Email,
		Phone:     cj.Phone,
		Age:       cj.Age,
		BirthDate: cj.BirthDate,
	}
}

// ToDomainList convierte un slice de CustomerJson a un slice de domain.Customer.
func (cjList CustomerJsonList) ToDomainList() []domain.Customer {
	if len(cjList) == 0 {
		return []domain.Customer{}
	}

	domainCustomers := make([]domain.Customer, len(cjList))
	for i, c := range cjList {
		domainCustomers[i] = *c.ToDomain()
	}
	return domainCustomers
}

// -----------------------------------------------------------------------------
// Conversiones: Dominio -> DTO
// -----------------------------------------------------------------------------

// FromDomain crea un nuevo CustomerJson a partir de un *domain.Customer.
func FromCustomerDomain(customer *domain.Customer) *CustomerJson {
	if customer == nil {
		return nil
	}
	return &CustomerJson{
		ID:        customer.ID,
		Name:      customer.Name,
		LastName:  customer.LastName,
		Email:     customer.Email,
		Phone:     customer.Phone,
		Age:       customer.Age,
		BirthDate: customer.BirthDate,
	}
}

// FromDomainList crea un slice de CustomerJson a partir de un slice de domain.Customer.
func FromCustomerDomainList(customers []domain.Customer) CustomerJsonList {
	if len(customers) == 0 {
		return CustomerJsonList{}
	}

	jsonCustomers := make(CustomerJsonList, len(customers))
	for i, c := range customers {
		jsonCustomers[i] = *FromCustomerDomain(&c)
	}
	return jsonCustomers
}
