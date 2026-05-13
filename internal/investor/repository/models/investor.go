package models

import (
	"github.com/google/uuid"

	domain "github.com/devpablocristo/ponti-backend/internal/investor/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
)

type Investor struct {
	ID       int64     `gorm:"primaryKey;autoIncrement"`
	TenantID uuid.UUID `gorm:"column:tenant_id;type:uuid;index"`
	Name     string    `gorm:"type:varchar(255);not null"`
	ActorID  *int64    `gorm:"-"`
	sharedmodels.Base
}

func (i Investor) ToDomain() *domain.Investor {
	inv := &domain.Investor{
		ID:      i.ID,
		Name:    i.Name,
		ActorID: i.ActorID,
		Base: shareddomain.Base{
			CreatedAt: i.CreatedAt,
			UpdatedAt: i.UpdatedAt,
			CreatedBy: i.CreatedBy,
			UpdatedBy: i.UpdatedBy,
		},
	}
	if i.DeletedAt.Valid {
		t := i.DeletedAt.Time
		inv.ArchivedAt = &t
	}
	return inv
}

func FromDomain(d *domain.Investor) *Investor {
	return &Investor{
		ID:      d.ID,
		Name:    d.Name,
		ActorID: d.ActorID,
		Base: sharedmodels.Base{
			CreatedAt: d.CreatedAt,
			UpdatedAt: d.UpdatedAt,
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}
