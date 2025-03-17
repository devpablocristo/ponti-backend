package dto

// CreateMacroCategory is the DTO for the create request of a macro category.
// It embeds the base MacroCategory DTO.
type CreateMacroCategory struct {
	MacroCategory
}

// CreateMacroCategoryResponse is the DTO for the response after creating a macro category.
type CreateMacroCategoryResponse struct {
	Message         string `json:"message"`
	MacroCategoryID int64  `json:"macro_category_id"`
}
