package models

import (
	"github.com/google/uuid"

	"github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
)

type Provider struct {
	ID       int64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID uuid.UUID `gorm:"column:tenant_id;type:uuid;index"`
	Name     string    `gorm:"type:varchar(50);not null"`
	sharedmodels.Base
}

func (p *Provider) ToDomain() *domain.Provider {
	return &domain.Provider{
		ID:   p.ID,
		Name: p.Name,
		Base: shareddomain.Base{
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
			CreatedBy: p.CreatedBy,
			UpdatedBy: p.UpdatedBy,
		},
	}
}

func FromDomain(d *domain.Provider) *Provider {
	return &Provider{
		ID:   d.ID,
		Name: d.Name,
		Base: sharedmodels.Base{
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}
