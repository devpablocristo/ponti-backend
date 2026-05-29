package models

import (
	domain "github.com/devpablocristo/ponti-backend/internal/business-parameters/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
)

type BusinessParameter struct {
	ID          int64  `gorm:"primaryKey;autoIncrement;column:id"`
	Key         string `gorm:"uniqueIndex;size:100;not null"`
	Value       string `gorm:"size:255;not null"`
	Type        string `gorm:"size:20;not null"` // decimal, integer, string, boolean
	Category    string `gorm:"size:50;not null"` // units, calculations, business_rules
	Description string `gorm:"type:text"`

	sharedmodels.Base
}

func (m *BusinessParameter) ToDomain() *domain.BusinessParameter {
	return &domain.BusinessParameter{
		ID:          m.ID,
		Key:         m.Key,
		Value:       m.Value,
		Type:        m.Type,
		Category:    m.Category,
		Description: m.Description,
		Base: shareddomain.Base{
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
		},
	}
}

func FromDomain(d *domain.BusinessParameter) *BusinessParameter {
	return &BusinessParameter{
		ID:          d.ID,
		Key:         d.Key,
		Value:       d.Value,
		Type:        d.Type,
		Category:    d.Category,
		Description: d.Description,
		Base: sharedmodels.Base{
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}
