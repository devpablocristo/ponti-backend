package dto

import "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dollar/usecases/domain"

type DollarAverageItem struct {
	Month        string  `json:"month" binding:"required"`
	StartValue   float64 `json:"start_value" binding:"required"`
	EndValue     float64 `json:"end_value" binding:"required"`
	AverageValue float64 `json:"average_value" binding:"required"`
}

type BulkDollarAverageRequest struct {
	Year   int64               `json:"year" binding:"required"`
	Values []DollarAverageItem `json:"values" binding:"required,dive"`
}

func (b *BulkDollarAverageRequest) ToDomainSlice(projectID int64) []domain.DollarAverage {
	out := make([]domain.DollarAverage, len(b.Values))
	for i, item := range b.Values {
		out[i] = domain.DollarAverage{
			ProjectID:  projectID,
			Year:       b.Year,
			Month:      item.Month,
			StartValue: item.StartValue,
			EndValue:   item.EndValue,
			AvgValue:   item.AverageValue,
		}
	}
	return out
}
