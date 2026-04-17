package models

import (
	domain "github.com/alphacodinggroup/ponti-backend/internal/investor/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/internal/shared/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/internal/shared/models"
)

type Investor struct {
	ID   int64  `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"type:varchar(255);not null;unique"`
	sharedmodels.Base
}

func (i Investor) ToDomain() *domain.Investor {
	inv := &domain.Investor{
		ID:   i.ID,
		Name: i.Name,
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
		ID:   d.ID,
		Name: d.Name,
		Base: sharedmodels.Base{
			CreatedAt: d.CreatedAt,
			UpdatedAt: d.UpdatedAt,
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}
