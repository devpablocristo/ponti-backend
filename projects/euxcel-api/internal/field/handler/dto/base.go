package dto

import (
	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/field/usecases/domain"
)

// Field is the DTO for a specific Field.
type Field struct {
	ID              int64   `json:"id"`
	Name            string  `json:"name"`
	ProjectID       int64   `json:"project_id"`
	LeasePercentage float64 `json:"lease_percentage"`
	LeaseType       string  `json:"lease_type"`
}

// ToDomain converts the DTO Field to the domain entity.
func (f Field) ToDomain() *domain.Field {
	return &domain.Field{
		ID:              f.ID,
		Name:            f.Name,
		ProjectID:       f.ProjectID,
		LeasePercentage: f.LeasePercentage,
		LeaseType:       f.LeaseType,
	}
}

// FromDomain converts a domain Field to the DTO.
func FromDomain(d domain.Field) *Field {
	return &Field{
		ID:              d.ID,
		Name:            d.Name,
		ProjectID:       d.ProjectID,
		LeasePercentage: d.LeasePercentage,
		LeaseType:       d.LeaseType,
	}
}
