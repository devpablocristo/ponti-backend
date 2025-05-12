package models

import (
	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/field/usecases/domain"
)

// Field represents the GORM model for a field.
type Field struct {
	ID         int64  `gorm:"primaryKey" json:"id"`
	Name       string `gorm:"size:100;not null" json:"name"`
	Location   string `gorm:"size:100" json:"location"`
	CustomerID int64  `gorm:"not null;index" json:"customer_id"`
}

// ToDomain converts the Field model to the domain entity.
func (f Field) ToDomain() *domain.Field {
	return &domain.Field{
		ID:         f.ID,
		Name:       f.Name,
		Location:   f.Location,
		CustomerID: f.CustomerID,
	}
}

// FromDomainField converts a domain Field entity to the GORM model.
func FromDomainField(d *domain.Field) *Field {
	return &Field{
		ID:         d.ID,
		Name:       d.Name,
		Location:   d.Location,
		CustomerID: d.CustomerID,
	}
}
