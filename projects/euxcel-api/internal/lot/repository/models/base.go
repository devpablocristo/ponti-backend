package models

import (
	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/lot/usecases/domain"
)

// Field, Project, Customer, Crop e Input deben estar definidos en sus respectivos archivos de modelos.
// Se muestran a continuación ejemplos mínimos para que las relaciones funcionen:

type Customer struct {
	ID   int64  `gorm:"primaryKey" json:"id"`
	Name string `gorm:"size:255;not null" json:"name"`
}

type Project struct {
	ID         int64    `gorm:"primaryKey" json:"id"`
	Name       string   `gorm:"size:100;not null" json:"name"`
	CustomerID int64    `gorm:"not null;index" json:"customer_id"`
	Customer   Customer `gorm:"foreignKey:CustomerID" json:"customer"`
	// Otros campos...
}

type Field struct {
	ID         int64    `gorm:"primaryKey" json:"id"`
	Name       string   `gorm:"size:100;not null" json:"name"`
	Location   string   `gorm:"size:100" json:"location"`
	CustomerID int64    `gorm:"not null;index" json:"customer_id"`
	Customer   Customer `gorm:"foreignKey:CustomerID" json:"customer"`
	// Otros campos...
}

type Crop struct {
	ID   int64  `gorm:"primaryKey" json:"id"`
	Name string `gorm:"size:50;not null" json:"name"`
}

type Input struct {
	ID       int64  `gorm:"primaryKey" json:"id"`
	Name     string `gorm:"size:100;not null" json:"name"`
	Category string `gorm:"size:50" json:"category"`
	Unit     string `gorm:"size:20" json:"unit"`
}

// Lot representa el modelo GORM para un lote (subdivisión de un campo).
type Lot struct {
	ID             int64   `gorm:"primaryKey" json:"id"`
	Name           string  `gorm:"size:50;not null" json:"identifier"`
	FieldID        int64   `gorm:"not null;index" json:"field_id"`
	Field          Field   `gorm:"foreignKey:FieldID" json:"field"`
	ProjectID      int64   `gorm:"not null;index" json:"project_id"`
	Project        Project `gorm:"foreignKey:ProjectID" json:"project"`
	CurrentCropID  *int64  `gorm:"index" json:"current_crop_id"`
	CurrentCrop    *Crop   `gorm:"foreignKey:CurrentCropID" json:"current_crop"`
	PreviousCropID *int64  `gorm:"index" json:"previous_crop_id"`
	PreviousCrop   *Crop   `gorm:"foreignKey:PreviousCropID" json:"previous_crop"`
	Variety        string  `gorm:"size:100" json:"variety"`
	Area           float64 `gorm:"type:numeric(10,2)" json:"area"`
}

// ToDomain convierte el modelo Lot a la entidad del dominio.
func (l Lot) ToDomain() *domain.Lot {
	return &domain.Lot{
		ID:             l.ID,
		Name:           l.Name,
		FieldID:        l.FieldID,
		ProjectID:      l.ProjectID,
		CurrentCropID:  l.CurrentCropID,
		PreviousCropID: l.PreviousCropID,
		Variety:        l.Variety,
		Area:           l.Area,
	}
}

// FromDomainLot convierte una entidad del dominio a un modelo Lot.
func FromDomainLot(d *domain.Lot) *Lot {
	return &Lot{
		ID:             d.ID,
		Name:           d.Name,
		FieldID:        d.FieldID,
		ProjectID:      d.ProjectID,
		CurrentCropID:  d.CurrentCropID,
		PreviousCropID: d.PreviousCropID,
		Variety:        d.Variety,
		Area:           d.Area,
	}
}
