package domain

import (
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
)

type DollarAverage struct {
	ID                int64
	ProjectID         int64
	Year              int64
	Month             string
	StartValue        float64
	EndValue          float64
	AvgValue          float64
	
	shareddomain.Base // CreatedAt, UpdatedAt, CreatedBy, UpdatedBy, DeletedBy
}
