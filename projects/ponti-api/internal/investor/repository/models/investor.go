package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
)

type Investor struct {
	ID   int64  `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"type:varchar(255);not null;unique"`
	sharedmodels.Base
}

func (i Investor) ToDomain() *domain.Investor {
	return &domain.Investor{
		ID:   i.ID,
		Name: i.Name,
	}
}

func FromDomain(d *domain.Investor) *Investor {
	return &Investor{
		ID:   d.ID,
		Name: d.Name,
	}
}
