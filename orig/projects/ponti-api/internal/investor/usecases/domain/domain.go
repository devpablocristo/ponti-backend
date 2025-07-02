package domain

import "time"

type Investor struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	Contributions    float64   // Total contribution amount (DECIMAL(10,2))
	ContributionDate time.Time // Date when the contribution was made
	Percentage       int       // Percentage of the investment in the project
}

type ListedInvestor struct {
	ID   int64  // Primary key
	Name string // Investor name
}
