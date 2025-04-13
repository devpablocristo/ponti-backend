package models

import (
	"time"

	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/investor/usecases/domain"
)

// Investor represents the GORM model for an investor.
type Investor struct {
	ID               int64     `gorm:"primaryKey"`
	Name             string    `gorm:"type:varchar(255);not null"`
	FieldID          int64     `gorm:"not null"`
	Contributions    float64   `gorm:"type:decimal(10,2);not null"`
	ContributionDate time.Time `gorm:"type:date;not null"`
}

// ToDomain converts the Investor model to the domain entity.
func (i Investor) ToDomain() *domain.Investor {
	return &domain.Investor{
		ID:               i.ID,
		Name:             i.Name,
		FieldID:          i.FieldID,
		Contributions:    i.Contributions,
		ContributionDate: i.ContributionDate,
	}
}

// FromDomainInvestor converts a domain Investor to the GORM model.
func FromDomainInvestor(d *domain.Investor) *Investor {
	return &Investor{
		ID:               d.ID,
		Name:             d.Name,
		FieldID:          d.FieldID,
		Contributions:    d.Contributions,
		ContributionDate: d.ContributionDate,
	}
}
