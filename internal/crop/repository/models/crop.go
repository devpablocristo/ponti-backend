package models

import (
	"github.com/devpablocristo/ponti-backend/internal/crop/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
	"github.com/google/uuid"
)

type Crop struct {
	ID       int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	TenantID uuid.UUID `gorm:"column:tenant_id;type:uuid;index"`
	Name     string    `gorm:"size:50;not null"`

	sharedmodels.Base
}

func (Crop) TableName() string {
	return "crops"
}

func (c Crop) ToDomain() *domain.Crop {
	return &domain.Crop{
		ID:   c.ID,
		Name: c.Name,
		Base: shareddomain.Base{
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			CreatedBy: c.CreatedBy,
			UpdatedBy: c.UpdatedBy,
		},
	}
}

func FromDomainCrop(d *domain.Crop) *Crop {
	return &Crop{
		ID:   d.ID,
		Name: d.Name,
		Base: sharedmodels.Base{
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}
