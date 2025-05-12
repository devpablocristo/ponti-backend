package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/usecases/domain"
)

// Customer representa el modelo GORM para un customer.
type Customer struct {
	ID   int64  `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(100);not null"`
	Type string `gorm:"type:varchar(100);not null"`
}

// ToDomain convierte el modelo Customer a la entidad de dominio.
func (c Customer) ToDomain() *domain.Customer {
	return &domain.Customer{
		ID:   c.ID,
		Name: c.Name,
		Type: c.Type,
	}
}

// FromDomainCustomer convierte una entidad de dominio a su modelo GORM.
func FromDomainCustomer(d *domain.Customer) *Customer {
	return &Customer{
		ID:   d.ID,
		Name: d.Name,
		Type: d.Type,
	}
}
