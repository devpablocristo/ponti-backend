package models

import (
	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/category/usecases/domain"
)

// Category representa el modelo en la base de datos para una categoría.
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
