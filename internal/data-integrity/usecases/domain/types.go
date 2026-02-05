// Package domain contiene los tipos de dominio para el módulo de integridad de datos
//
// ⚠️  ADVERTENCIA CRÍTICA - NO MODIFICAR SIN AUTORIZACIÓN EXPLÍCITA ⚠️
//
// ESTOS TIPOS SON CRÍTICOS PARA LA INTEGRIDAD DE DATOS.
// NUNCA ALTERAR SIN AUTORIZACIÓN EXPLÍCITA DEL USUARIO.
package domain

import "github.com/shopspring/decimal"

// IntegrityReport contiene el resultado de todas las validaciones de coherencia (14 controles)
type IntegrityReport struct {
	Checks []IntegrityCheck `json:"checks"`
}

// IntegrityCheck representa un control individual de coherencia de datos
// Cada control compara DOS cálculos INDEPENDIENTES: LEFT (origen/correcto) vs RIGHT (destino/a validar)
type IntegrityCheck struct {
	ControlNumber int    `json:"control_number"` // 1-14
	SourceModule  string `json:"source_module"`  // Pantalla origen: "Órdenes de trabajo", "Labores + Insumos", etc.
	DataToVerify  string `json:"data_to_verify"` // Dato a verificar: "Costos directos ejecutados", "Invertidos", etc.
	TargetModule  string `json:"target_module"`  // Pantalla destino: "Dashboard", "Lotes", "Informe de Aportes", etc.
	ControlRule   string `json:"control_rule"`   // Regla de control del CSV
	Description   string `json:"description"`    // Descripción breve del control

	// LEFT SIDE (ORIGEN - Fuente de verdad)
	LeftCalculation string          `json:"left_calculation"` // Descripción del cálculo LEFT
	LeftValue       decimal.Decimal `json:"left_value"`       // Valor calculado desde origen
	LeftSource      string          `json:"left_source"`      // Query/Vista/Tabla usada para LEFT
	LeftMeaning     string          `json:"left_interpretation"` // Interpretación simple de LEFT

	// RIGHT SIDE (DESTINO - A validar)
	RightCalculation string          `json:"right_calculation"` // Descripción del cálculo RIGHT
	RightValue       decimal.Decimal `json:"right_value"`       // Valor calculado desde destino
	RightSource      string          `json:"right_source"`      // Query/Vista/Tabla usada para RIGHT
	RightMeaning     string          `json:"right_interpretation"` // Interpretación simple de RIGHT

	CalculationMeaning string `json:"calculation_interpretation"` // Interpretación breve del cálculo

	// RESULTADO
	Difference decimal.Decimal `json:"difference"` // LeftValue - RightValue
	Status     string          `json:"status"`     // OK, ERROR
	Tolerance  decimal.Decimal `json:"tolerance"`  // Tolerancia permitida
}

// CostsCheckFilter contiene los filtros para la validación de costos
type CostsCheckFilter struct {
	ProjectID *int64
}
