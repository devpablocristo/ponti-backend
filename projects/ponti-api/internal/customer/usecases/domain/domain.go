package domain

// Customer representa una entidad de customer.
type Customer struct {
	ID   int64  // Llave primaria (numérica)
	Name string // Nombre del customer
	Type string // Tipo del customer
}

// Customer representa una entidad de customer.
type ListedCustomer struct {
	ID   int64  // Llave primaria (numérica)
	Name string // Nombre del customer
}
