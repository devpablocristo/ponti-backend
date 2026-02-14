package dto

// CreateCustomer es el DTO para la creación de un customer.
// Embebe el DTO base Customer.
type CreateCustomer struct {
	Customer
}

type CreateCustomerResponse struct {
	Message string `json:"message"`
	ID      int64  `json:"id"`
}
