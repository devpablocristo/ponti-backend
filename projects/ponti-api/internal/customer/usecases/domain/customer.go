package domain

// Customer representa una entidad de customer.
type Customer struct {
	ID   int64
	Name string
}

// Customer representa una entidad de customer.
type ListedCustomer struct {
	ID   int64  // Llave primaria (numérica)
	Name string // Nombre del customer
}
