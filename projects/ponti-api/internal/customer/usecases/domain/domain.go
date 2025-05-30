package domain

// Customer representa una entidad de customer.
type Customer struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}
