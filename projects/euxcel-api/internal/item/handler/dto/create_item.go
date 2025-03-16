package dto

// -----------------------------
// DTO para creación de Item
// -----------------------------

// CreateItem es el DTO para la solicitud de creación de un ítem.
// Se embebe el DTO Item, por lo que hereda su método ToDomain.
type CreateItem struct {
	Item
}

// CreateItemResponse es el DTO para la respuesta luego de crear un ítem.
type CreateItemResponse struct {
	Message string `json:"message"`
	ItemID  int64  `json:"item_id"`
}
