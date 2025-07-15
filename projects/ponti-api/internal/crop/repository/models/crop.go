package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"

	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
)

// Crop represents a type of crop.
type Crop struct {
	ID   int64  `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name string `gorm:"uniqueIndex:idx_crops_name;size:50;not null"`
	sharedmodels.Base
}

func (Crop) TableName() string {
	return "crops"
}

// ToDomain converts the Crop model to the domain entity.
func (c Crop) ToDomain() *domain.Crop {
	return &domain.Crop{
		ID:   c.ID,
		Name: c.Name,
	}
}

// FromDomainCrop converts a domain Crop entity to the GORM model.
func FromDomainCrop(d *domain.Crop) *Crop {
	return &Crop{
		ID:   d.ID,
		Name: d.Name,
	}
}
