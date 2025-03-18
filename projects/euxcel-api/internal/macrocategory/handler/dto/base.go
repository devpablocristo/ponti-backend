package dto

import (
	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/macrocategory/usecases/domain"
)

// MacroCategory is the DTO for a macro category.
type MacroCategory struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ToDomain converts the DTO to the domain entity.
func (m MacroCategory) ToDomain() *domain.MacroCategory {
	return &domain.MacroCategory{
		ID:   m.ID,
		Name: m.Name,
	}
}

// FromDomain converts a domain MacroCategory to the DTO.
func FromDomain(d domain.MacroCategory) *MacroCategory {
	return &MacroCategory{
		ID:   d.ID,
		Name: d.Name,
	}
}
