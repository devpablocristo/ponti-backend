package models

import (
	"github.com/google/uuid"

	"github.com/devpablocristo/ponti-backend/internal/customer/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
)

type Customer struct {
	ID       int64     `gorm:"primaryKey;autoIncrement"`
	TenantID uuid.UUID `gorm:"column:tenant_id;type:uuid;index"`
	Name     string    `gorm:"type:varchar(255);not null"`
	ActorID  *int64    `gorm:"column:actor_id"`

	sharedmodels.Base
}

func (c Customer) ToDomain() *domain.Customer {
	d := &domain.Customer{
		ID:      c.ID,
		Name:    c.Name,
		ActorID: c.ActorID,
		Base: shareddomain.Base{
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			CreatedBy: c.CreatedBy,
			UpdatedBy: c.UpdatedBy,
		},
	}
	if c.DeletedAt.Valid {
		t := c.DeletedAt.Time
		d.ArchivedAt = &t
	}
	return d
}

func FromDomain(d *domain.Customer) *Customer {
	return &Customer{
		ID:      d.ID,
		Name:    d.Name,
		ActorID: d.ActorID,
		Base: sharedmodels.Base{
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}
