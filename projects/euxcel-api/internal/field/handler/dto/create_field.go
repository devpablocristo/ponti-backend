package dto

// CreateField is the DTO for the create request of a Field.
// It embeds the base Field DTO.
type CreateField struct {
	Field
}

type CreateFieldResponse struct {
	Message string `json:"message"`
	FieldID int64  `json:"field_id"`
}
