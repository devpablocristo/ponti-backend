package dto

import (
	"github.com/shopspring/decimal"

	"github.com/devpablocristo/ponti-backend/internal/data-integrity/usecases/domain"
)

// IntegrityReportResponse representa la respuesta del endpoint de integridad (9 controles)
type IntegrityReportResponse struct {
	Checks []IntegrityCheckDTO `json:"checks"`
}

// IntegrityCheckDTO representa un control individual de coherencia de datos
type IntegrityCheckDTO struct {
	ControlNumber int    `json:"control_number"`
	DataToVerify  string `json:"data_to_verify"`
	Description   string `json:"description"`
	ControlRule   string `json:"control_rule"`

	SystemCalculation string `json:"system_calculation"`
	SystemValue       string `json:"system_value"`
	SystemSource      string `json:"system_source"`
	SystemMeaning     string `json:"system_meaning"`

	RecalcACalculation string `json:"recalc_a_calculation"`
	RecalcAValue       string `json:"recalc_a_value"`
	RecalcASource      string `json:"recalc_a_source"`
	RecalcAMeaning     string `json:"recalc_a_meaning"`

	RecalcBCalculation *string `json:"recalc_b_calculation,omitempty"`
	RecalcBValue       *string `json:"recalc_b_value,omitempty"`
	RecalcBSource      *string `json:"recalc_b_source,omitempty"`
	RecalcBMeaning     *string `json:"recalc_b_meaning,omitempty"`

	DifferenceA string  `json:"difference_a"`
	DifferenceB *string `json:"difference_b,omitempty"`
	Status      string  `json:"status"`
	Tolerance   string  `json:"tolerance"`
}

// ToIntegrityReportResponse convierte domain.IntegrityReport a DTO
func ToIntegrityReportResponse(report *domain.IntegrityReport) *IntegrityReportResponse {
	checks := make([]IntegrityCheckDTO, len(report.Checks))

	for i, check := range report.Checks {
		dto := IntegrityCheckDTO{
			ControlNumber:      check.ControlNumber,
			DataToVerify:       check.DataToVerify,
			Description:        check.Description,
			ControlRule:        check.ControlRule,
			SystemCalculation:  check.SystemCalculation,
			SystemValue:        formatDecimal(check.SystemValue),
			SystemSource:       check.SystemSource,
			SystemMeaning:      check.SystemMeaning,
			RecalcACalculation: check.RecalcACalculation,
			RecalcAValue:       formatDecimal(check.RecalcAValue),
			RecalcASource:      check.RecalcASource,
			RecalcAMeaning:     check.RecalcAMeaning,
			DifferenceA:        formatDecimal(check.DifferenceA),
			Status:             check.Status,
			Tolerance:          formatDecimal(check.Tolerance),
		}

		if check.RecalcBCalculation != nil {
			dto.RecalcBCalculation = check.RecalcBCalculation
		}
		if check.RecalcBValue != nil {
			v := formatDecimal(*check.RecalcBValue)
			dto.RecalcBValue = &v
		}
		if check.RecalcBSource != nil {
			dto.RecalcBSource = check.RecalcBSource
		}
		if check.RecalcBMeaning != nil {
			dto.RecalcBMeaning = check.RecalcBMeaning
		}
		if check.DifferenceB != nil {
			v := formatDecimal(*check.DifferenceB)
			dto.DifferenceB = &v
		}

		checks[i] = dto
	}

	return &IntegrityReportResponse{
		Checks: checks,
	}
}

// formatDecimal formatea un decimal con 2 decimales
func formatDecimal(d decimal.Decimal) string {
	return d.StringFixed(2)
}
