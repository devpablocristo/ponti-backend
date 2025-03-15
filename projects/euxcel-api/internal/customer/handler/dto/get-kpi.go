package dto

import (
	"github.com/devpablocristo/monorepo/projects/qh/internal/customer/usecases/domain"
)

type GetKPIJson struct {
	AverageAge      float64 `json:"average_age"`
	AgeStdDeviation float64 `json:"age_std_deviation"`
}

func ToGetKPIJson(kpi *domain.KPI) *GetKPIJson {
	return &GetKPIJson{
		AverageAge:      kpi.AverageAge,
		AgeStdDeviation: kpi.AgeStdDeviation,
	}
}
