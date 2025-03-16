package models

import (
	"github.com/alphacodinggroup/euxcel-backend/internal/item/usecases/domain"
)

// MacroCategory representa la entidad en la base de datos para una macro categoría.
type MacroCategory struct {
	ID   int64  `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(100);not null"`
}

// ToDomain convierte el modelo MacroCategory a la entidad de dominio.
func (m MacroCategory) ToDomain() *domain.MacroCategory {
	return &domain.MacroCategory{
		ID:   m.ID,
		Name: m.Name,
	}
}

// FromDomainMacroCategory convierte una entidad de dominio a su modelo GORM.
func FromDomainMacroCategory(d *domain.MacroCategory) *MacroCategory {
	return &MacroCategory{
		ID:   d.ID,
		Name: d.Name,
	}
}

// Category representa el modelo en la base de datos para una categoría específica.
type Category struct {
	ID              int64  `gorm:"primaryKey"`
	Name            string `gorm:"type:varchar(100);not null"`
	MacroCategoryID int64  `gorm:"not null"`
}

// ToDomain convierte el modelo Category a la entidad de dominio.
func (c Category) ToDomain() *domain.Category {
	return &domain.Category{
		ID:              c.ID,
		Name:            c.Name,
		MacroCategoryID: c.MacroCategoryID,
	}
}

// FromDomainCategory convierte una entidad de dominio a su modelo GORM.
func FromDomainCategory(d *domain.Category) *Category {
	return &Category{
		ID:              d.ID,
		Name:            d.Name,
		MacroCategoryID: d.MacroCategoryID,
	}
}

// Supplier representa el modelo en la base de datos para un proveedor.
type Supplier struct {
	ID   int64  `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(100);not null"`
}

// ToDomain convierte el modelo Supplier a la entidad de dominio.
func (s Supplier) ToDomain() *domain.Supplier {
	return &domain.Supplier{
		ID:   s.ID,
		Name: s.Name,
	}
}

// FromDomainSupplier convierte una entidad de dominio a su modelo GORM.
func FromDomainSupplier(d *domain.Supplier) *Supplier {
	return &Supplier{
		ID:   d.ID,
		Name: d.Name,
	}
}

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
