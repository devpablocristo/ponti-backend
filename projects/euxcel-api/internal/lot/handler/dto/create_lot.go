package dto

// CreateLot es el DTO para la creación de un lote.
// Embebe el DTO base Lot.
type CreateLot struct {
	Lot
}

type CreateLotResponse struct {
	Message string `json:"message"`
	ID      int64  `json:"id"`
}
