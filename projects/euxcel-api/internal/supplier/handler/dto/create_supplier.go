package dto

// CreateSupplier is the DTO for the create request of a supplier.
// It embeds the base Supplier DTO.
type CreateSupplier struct {
	Supplier
}

// CreateSupplierResponse is the DTO for the response after creating a supplier.
type CreateSupplierResponse struct {
	Message    string `json:"message"`
	SupplierID int64  `json:"supplier_id"`
}
