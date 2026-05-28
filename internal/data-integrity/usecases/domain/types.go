// Package domain define los tipos del módulo de integridad de datos.
package domain

import "github.com/shopspring/decimal"

// IntegrityReport contiene el resultado de todas las validaciones de coherencia (17 controles)
type IntegrityReport struct {
	Checks []IntegrityCheck
}

// IntegrityCheck representa un control individual de coherencia de datos.
// Cada control compara hasta 3 valores:
//   - SystemValue: el valor EXACTO que el sistema muestra al usuario
//   - RecalcA: recálculo independiente por camino A
//   - RecalcB: recálculo independiente por camino B (opcional)
type IntegrityCheck struct {
	ControlNumber  int    // 1-17
	DataToVerify   string // Dato a verificar
	Description    string // Descripción breve del control
	ControlRule    string // Regla de control
	CheckType      string // STRONG, WEAK, FORMULA_ALIGNMENT
	Severity       string // INFO, WARNING, ERROR
	Recommendation string // Qué hacer si el control requiere acción

	// SYSTEM VALUE (lo que el usuario ve en pantalla)
	SystemCalculation string
	SystemValue       decimal.Decimal
	SystemSource      string
	SystemMeaning     string // Qué representa y cómo se calcula

	// RECALC A (primer recálculo independiente)
	RecalcACalculation string
	RecalcAValue       decimal.Decimal
	RecalcASource      string
	RecalcAMeaning     string // Qué representa y cómo se calcula

	// RECALC B (segundo recálculo independiente, opcional)
	RecalcBCalculation *string
	RecalcBValue       *decimal.Decimal
	RecalcBSource      *string
	RecalcBMeaning     *string // Qué representa y cómo se calcula

	// RESULTADO
	DifferenceA decimal.Decimal  // SystemValue - RecalcAValue
	DifferenceB *decimal.Decimal // SystemValue - RecalcBValue (nil si RecalcB no aplica)
	Status      string           // OK, ERROR, WARNING, SKIPPED
	Tolerance   decimal.Decimal
}

// CostsCheckFilter contiene los filtros para la validación de costos
type CostsCheckFilter struct {
	ProjectID *int64
}

type TentativePricesFilter struct {
	CustomerID *int64
	ProjectID  *int64
	CampaignID *int64
	FieldID    *int64
}

type TentativePriceItem struct {
	SupplyID     int64
	Name         string
	CategoryName string
	Price        decimal.Decimal
}

type TentativePricesReport struct {
	Count int64
	Items []TentativePriceItem
}
