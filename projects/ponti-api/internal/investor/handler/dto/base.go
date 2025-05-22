package dto

import (
	"time"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
)

// Investor is the DTO for a specific investor.
type Investor struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	Contributions    float64   `json:"contributions"`
	ContributionDate time.Time `json:"contribution_date"`
	Percentage       int       `json:"percentage"`
}

// ToDomain converts the DTO Investor to the domain entity.
func (i Investor) ToDomain() *domain.Investor {
	return &domain.Investor{
		ID:               i.ID,
		Name:             i.Name,
		Contributions:    i.Contributions,
		ContributionDate: i.ContributionDate,
		Percentage:       i.Percentage,
	}
}

// FromDomain converts a domain Investor to the DTO.
func FromDomain(d domain.Investor) *Investor {
	return &Investor{
		ID:               d.ID,
		Name:             d.Name,
		Contributions:    d.Contributions,
		ContributionDate: d.ContributionDate,
		Percentage:       d.Percentage,
	}
}
