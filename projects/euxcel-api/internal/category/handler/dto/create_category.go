package dto

// CreateCategory is the DTO for the create request of a category.
// It embeds the base Category DTO.
type CreateCategory struct {
	Category
}

type CreateCategoryResponse struct {
	Message    string `json:"message"`
	CategoryID int64  `json:"item_id"`
}
