// Package domain define los tipos del módulo de integridad de datos.
package domain

import "github.com/shopspring/decimal"

// IntegrityReport contiene el resultado de todas las validaciones de coherencia (16 controles)
type IntegrityReport struct {
	Checks []IntegrityCheck `json:"checks"`
}

// IntegrityCheck representa un control individual de coherencia de datos.
// Cada control compara hasta 3 valores:
//   - SystemValue: el valor EXACTO que el sistema muestra al usuario
//   - RecalcA: recálculo independiente por camino A
//   - RecalcB: recálculo independiente por camino B (opcional)
type IntegrityCheck struct {
	ControlNumber int    `json:"control_number"` // 1-9
	DataToVerify  string `json:"data_to_verify"` // Dato a verificar
	Description   string `json:"description"`    // Descripción breve del control
	ControlRule   string `json:"control_rule"`   // Regla de control

	// SYSTEM VALUE (lo que el usuario ve en pantalla)
	SystemCalculation string          `json:"system_calculation"`
	SystemValue       decimal.Decimal `json:"system_value"`
	SystemSource      string          `json:"system_source"`
	SystemMeaning     string          `json:"system_meaning"` // Qué representa y cómo se calcula

	// RECALC A (primer recálculo independiente)
	RecalcACalculation string          `json:"recalc_a_calculation"`
	RecalcAValue       decimal.Decimal `json:"recalc_a_value"`
	RecalcASource      string          `json:"recalc_a_source"`
	RecalcAMeaning     string          `json:"recalc_a_meaning"` // Qué representa y cómo se calcula

	// RECALC B (segundo recálculo independiente, opcional)
	RecalcBCalculation *string          `json:"recalc_b_calculation,omitempty"`
	RecalcBValue       *decimal.Decimal `json:"recalc_b_value,omitempty"`
	RecalcBSource      *string          `json:"recalc_b_source,omitempty"`
	RecalcBMeaning     *string          `json:"recalc_b_meaning,omitempty"` // Qué representa y cómo se calcula

	// RESULTADO
	DifferenceA decimal.Decimal  `json:"difference_a"`           // SystemValue - RecalcAValue
	DifferenceB *decimal.Decimal `json:"difference_b,omitempty"` // SystemValue - RecalcBValue (nil si RecalcB no aplica)
	Status      string           `json:"status"`                 // OK, ERROR
	Tolerance   decimal.Decimal  `json:"tolerance"`
}

// CostsCheckFilter contiene los filtros para la validación de costos
type CostsCheckFilter struct {
	ProjectID *int64
}
