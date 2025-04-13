package dto

// CreateLot is the DTO for the create request of a Lot.
// It embeds the base Lot DTO.
type CreateLot struct {
	Lot
}

type CreateLotResponse struct {
	Message string `json:"message"`
	ID      int64  `json:"lot_id"`
}
