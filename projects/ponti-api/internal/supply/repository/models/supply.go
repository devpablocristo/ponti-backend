package models

import (
	catmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/category/repository/models"
	classtype "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/classtype/repository/models"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/usecases/domain"
)

// Tablas auxiliares normalizadas

type SupplyUnit struct {
	ID   uint   `gorm:"primaryKey;autoIncrement;column:id"`
	Name string `gorm:"type:varchar(20);unique;not null"`
}

// Modelo principal de Supply
type Supply struct {
	ID        int64   `gorm:"primaryKey;autoIncrement;column:id"`
	ProjectID int64   `gorm:"not null;index"`
	Name      string  `gorm:"type:varchar(100);not null"`
	Price     float64 `gorm:"not null"`

	UnitID uint
	//Unit   SupplyUnit `gorm:"foreignKey:UnitID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	CategoryID uint
	Category   catmod.Category `gorm:"foreignKey:CategoryID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	TypeID uint
	Type   classtype.ClassType `gorm:"foreignKey:TypeID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	sharedmodels.Base // Campos de auditoría (CreatedAt, UpdatedAt, etc)
}

// De persistencia (models.Supply) → dominio (domain.Supply)
func (m *Supply) ToDomain() *domain.Supply {
	return &domain.Supply{
		ID:         m.ID,
		ProjectID:  m.ProjectID,
		Name:       m.Name,
		UnitID:     int64(m.UnitID),
		Price:      m.Price,
		CategoryID: int64(m.CategoryID),
		TypeID:     int64(m.TypeID),
		Base: shareddomain.Base{
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
		},
	}
}

func FromDomain(d *domain.Supply) *Supply {
	return &Supply{
		ID:         d.ID,
		ProjectID:  d.ProjectID,
		Name:       d.Name,
		Price:      d.Price,
		UnitID:     uint(d.UnitID),
		CategoryID: uint(d.CategoryID),
		TypeID:     uint(d.TypeID),
		Base: sharedmodels.Base{
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}
