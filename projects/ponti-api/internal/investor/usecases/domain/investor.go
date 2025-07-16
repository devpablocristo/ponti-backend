package domain

import (
	"time"

	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
)

type Investor struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	Contributions    float64   // Total contribution amount (DECIMAL(10,2))
	ContributionDate time.Time // Date when the contribution was made
	Percentage       int       // Percentage of the investment in the project
	shareddomain.Base
}

type ListedInvestor struct {
	ID   int64  // Primary key
	Name string // Investor name
}
