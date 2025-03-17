package dto

type CreateSupplier struct {
	Supplier
}

// CreateSupplierResponse is the DTO for the response after creating a supplier.
type CreateSupplierResponse struct {
	Message    string `json:"message"`
	SupplierID int64  `json:"supplier_id"`
}
