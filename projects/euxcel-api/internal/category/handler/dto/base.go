package dto

import (
	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/category/usecases/domain"
)

// Category is the DTO for a specific category.
type Category struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	MacroCategoryID int64  `json:"macro_category_id"`
}

// ToDomain converts the DTO Category to the domain entity.
func (c Category) ToDomain() *domain.Category {
	return &domain.Category{
		ID:              c.ID,
		Name:            c.Name,
		MacroCategoryID: c.MacroCategoryID,
	}
}

// FromDomain converts a domain Category to the DTO.
func FromDomain(d domain.Category) *Category {
	return &Category{
		ID:              d.ID,
		Name:            d.Name,
		MacroCategoryID: d.MacroCategoryID,
	}
}
