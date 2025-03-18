package models

import (
	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/supplier/usecases/domain"
)

// Supplier represents the database model for a supplier.
type Supplier struct {
	ID   int64  `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(100);not null"`
}

// ToDomain converts the Supplier model to the domain entity.
func (s Supplier) ToDomain() *domain.Supplier {
	return &domain.Supplier{
		ID:   s.ID,
		Name: s.Name,
	}
}

// FromDomainSupplier converts a domain Supplier to the GORM model.
func FromDomainSupplier(d *domain.Supplier) *Supplier {
	return &Supplier{
		ID:   d.ID,
		Name: d.Name,
	}
}
