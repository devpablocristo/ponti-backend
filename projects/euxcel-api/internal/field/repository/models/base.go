// File: internal/field/repository/models/base.go
package models

import (
	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/field/usecases/domain"
)

// Field is the GORM model for a field.
type Field struct {
	ID              int64   `gorm:"primaryKey" json:"id"`
	ProjectID       int64   `gorm:"not null;index" json:"project_id"`
	Name            string  `gorm:"size:100;not null" json:"name"`
	LeasePercentage float64 `json:"lease_percentage"`
	LeaseType       string  `gorm:"size:50;not null" json:"lease_type"`
}

// ToDomain converts the GORM model to the domain entity.
func (m Field) ToDomain() *domain.Field {
	return &domain.Field{
		ID:              m.ID,
		ProjectID:       m.ProjectID,
		Name:            m.Name,
		LeasePercentage: m.LeasePercentage,
		LeaseType:       m.LeaseType,
		Lots:            nil, // loaded separately
	}
}

// FromDomainField converts a domain entity to the GORM model.
func FromDomainField(d *domain.Field) *Field {
	return &Field{
		ID:              d.ID,
		ProjectID:       d.ProjectID,
		Name:            d.Name,
		LeasePercentage: d.LeasePercentage,
		LeaseType:       d.LeaseType,
	}
}
