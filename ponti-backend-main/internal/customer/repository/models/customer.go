package models

import (
	"github.com/alphacodinggroup/ponti-backend/internal/customer/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/internal/shared/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/internal/shared/models"
)

type Customer struct {
	ID   int64  `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"type:varchar(255);not null;unique"`

	sharedmodels.Base
}

func (c Customer) ToDomain() *domain.Customer {
	d := &domain.Customer{
		ID:   c.ID,
		Name: c.Name,
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
		ID:   d.ID,
		Name: d.Name,
		Base: sharedmodels.Base{
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}
