package dto

import (
	"time"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dollar/usecases/domain"
)

type RecordResponse struct {
	ID           int64     `json:"id"`
	ProjectID    int64     `json:"project_id"`
	Year         int64     `json:"year"`
	Month        string    `json:"month"`
	StartValue   float64   `json:"start_value"`
	EndValue     float64   `json:"end_value"`
	AverageValue float64   `json:"average_value"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func FromDomain(d *domain.DollarAverage) RecordResponse {
	return RecordResponse{
		ID:           d.ID,
		ProjectID:    d.ProjectID,
		Year:         d.Year,
		Month:        d.Month,
		StartValue:   d.StartValue,
		EndValue:     d.EndValue,
		AverageValue: d.AvgValue,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}
