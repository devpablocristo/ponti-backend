package models

import (
	"time"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
)

type Investor struct {
	ID               int64     `gorm:"primaryKey;autoIncrement"`
	Name             string    `gorm:"type:varchar(255);not null"`
	Contributions    float64   `gorm:"type:decimal(10,2);not null"`
	ContributionDate time.Time `gorm:"not null"`
	Percentage       int       `gorm:"not null"`
}

func (i Investor) ToDomain() *domain.Investor {
	return &domain.Investor{
		ID:               i.ID,
		Name:             i.Name,
		Contributions:    i.Contributions,
		ContributionDate: i.ContributionDate,
		Percentage:       i.Percentage,
	}
}

func FromDomain(d *domain.Investor) *Investor {
	return &Investor{
		ID:               d.ID,
		Name:             d.Name,
		Contributions:    d.Contributions,
		ContributionDate: d.ContributionDate,
		Percentage:       d.Percentage,
	}
}
