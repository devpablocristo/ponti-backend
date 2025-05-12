package dto

import (
	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/field/usecases/domain"
)

// FieldDTO matches JSON payload for Field operations.
type Field struct {
	ID              int64   `json:"id,omitempty"`
	ProjectID       int64   `json:"project_id" binding:"required"`
	Name            string  `json:"name" binding:"required"`
	LeasePercentage float64 `json:"lease_percentage" binding:"required"`
	LeaseType       string  `json:"lease_type" binding:"required"`
}

// ToDomain converts DTO to domain entity.
func (f Field) ToDomain() *domain.Field {
	return &domain.Field{
		ID:              f.ID,
		ProjectID:       f.ProjectID,
		Name:            f.Name,
		LeasePercentage: f.LeasePercentage,
		LeaseType:       f.LeaseType,
	}
}

// FromDomain converts domain entity to DTO.
func FromDomain(d domain.Field) Field {
	return Field{
		ID:              d.ID,
		ProjectID:       d.ProjectID,
		Name:            d.Name,
		LeasePercentage: d.LeasePercentage,
		LeaseType:       d.LeaseType,
	}
}
