package dto

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dollar/usecases/domain"
)

type MonthResponse struct {
	Month        string  `json:"month"`
	StartValue   float64 `json:"start_value"`
	EndValue     float64 `json:"end_value"`
	AverageValue float64 `json:"average_value"`
}

func FromDomainMonth(d *domain.DollarAverage) MonthResponse {
	return MonthResponse{
		Month:        d.Month,
		StartValue:   d.StartValue,
		EndValue:     d.EndValue,
		AverageValue: d.AvgValue,
	}
}
