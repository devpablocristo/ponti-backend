package models

import (
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/unit/usecases/domain"
)

type Unit struct {
	ID   int64  `gorm:"primaryKey;autoIncrement;column:id"`
	Name string `gorm:"type:varchar(30);unique;not null"`

	sharedmodels.Base // Campos de auditoría (CreatedAt, UpdatedAt, DeletedAt, CreatedBy, UpdatedBy, DeletedBy)
}

func (m *Unit) ToDomain() *domain.Unit {
	return &domain.Unit{
		ID:   m.ID,
		Name: m.Name,
		Base: shareddomain.Base{
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
		},
	}
}

func FromDomain(d *domain.Unit) *Unit {
	return &Unit{
		ID:   d.ID,
		Name: d.Name,
		Base: sharedmodels.Base{
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}
