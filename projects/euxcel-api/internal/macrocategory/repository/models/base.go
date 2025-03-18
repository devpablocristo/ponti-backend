package models

import (
	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/macrocategory/usecases/domain"
)

// MacroCategory represents the database model for a macro category.
type MacroCategory struct {
	ID   int64  `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(100);not null"`
}

// ToDomain converts the MacroCategory model to the domain entity.
func (m MacroCategory) ToDomain() *domain.MacroCategory {
	return &domain.MacroCategory{
		ID:   m.ID,
		Name: m.Name,
	}
}

// FromDomainMacroCategory converts a domain MacroCategory to the GORM model.
func FromDomainMacroCategory(d *domain.MacroCategory) *MacroCategory {
	return &MacroCategory{
		ID:   d.ID,
		Name: d.Name,
	}
}
