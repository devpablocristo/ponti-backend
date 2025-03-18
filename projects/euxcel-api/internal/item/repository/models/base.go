package models

import (
	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/item/usecases/domain"
)

// Item representa el modelo en la base de datos para un artículo o ítem.
type Item struct {
	ID         int64   `gorm:"primaryKey"`
	Name       string  `gorm:"type:varchar(150);not null"`
	PriceUSD   float64 `gorm:"not null"`
	CategoryID int64   `gorm:"not null"`
	SupplierID int64   `gorm:"not null"`
}

// ToDomain convierte el modelo Item a la entidad de dominio.
func (i Item) ToDomain() *domain.Item {
	return &domain.Item{
		ID:         i.ID,
		Name:       i.Name,
		PriceUSD:   i.PriceUSD,
		CategoryID: i.CategoryID,
		SupplierID: i.SupplierID,
	}
}

// FromDomainItem convierte una entidad de dominio a su modelo GORM.
func FromDomainItem(d *domain.Item) *Item {
	return &Item{
		ID:         d.ID,
		Name:       d.Name,
		PriceUSD:   d.PriceUSD,
		CategoryID: d.CategoryID,
		SupplierID: d.SupplierID,
	}
}
