package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
)

type Investor struct {
	ID   int64  `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"type:varchar(255);not null"`
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
