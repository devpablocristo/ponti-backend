package models

import (
	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/field/usecases/domain"
)

// Field represents the GORM model for a field.
type Field struct {
	ID              int64   `gorm:"primaryKey"`
	Name            string  `gorm:"type:varchar(100);not null"`
	ProjectID       int64   `gorm:"not null"`
	LeasePercentage float64 `gorm:"not null"`
	LeaseType       string  `gorm:"type:varchar(100);not null"`
}

// ToDomain converts the Field model to the domain entity.
func (f Field) ToDomain() *domain.Field {
	return &domain.Field{
		ID:              f.ID,
		Name:            f.Name,
		ProjectID:       f.ProjectID,
		LeasePercentage: f.LeasePercentage,
		LeaseType:       f.LeaseType,
	}
}

// FromDomainField converts a domain Field to the GORM model.
func FromDomainField(d *domain.Field) *Field {
	return &Field{
		ID:              d.ID,
		Name:            d.Name,
		ProjectID:       d.ProjectID,
		LeasePercentage: d.LeasePercentage,
		LeaseType:       d.LeaseType,
	}
}
