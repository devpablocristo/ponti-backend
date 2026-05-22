// Package domain contiene las entidades de dominio para los reportes.
package domain

import (
	"github.com/shopspring/decimal"
)

// ===== ENTIDADES DE DOMINIO =====

// SummaryResults representa el resumen de resultados por cultivo
type SummaryResults struct {
	ProjectID int64
	CropID    int64
	CropName  string

	// Métricas por cultivo
	SurfaceHa          decimal.Decimal
	NetIncomeUsd       decimal.Decimal
	DirectCostsUsd     decimal.Decimal
	RentUsd            decimal.Decimal
	StructureUsd       decimal.Decimal
	TotalInvestedUsd   decimal.Decimal
	OperatingResultUsd decimal.Decimal
	CropReturnPct      decimal.Decimal

	// Totales del proyecto (para comparación)
	TotalSurfaceHa          decimal.Decimal
	TotalNetIncomeUsd       decimal.Decimal
	TotalDirectCostsUsd     decimal.Decimal
	TotalRentUsd            decimal.Decimal
	TotalStructureUsd       decimal.Decimal
	TotalInvestedProjectUsd decimal.Decimal
	TotalOperatingResultUsd decimal.Decimal
	ProjectReturnPct        decimal.Decimal
}

// ProjectTotals representa los totales del proyecto (GRAL CAMPOS).
type ProjectTotals struct {
	TotalSurfaceHa          decimal.Decimal
	TotalNetIncomeUsd       decimal.Decimal
	TotalDirectCostsUsd     decimal.Decimal
	TotalRentUsd            decimal.Decimal
	TotalStructureUsd       decimal.Decimal
	TotalInvestedProjectUsd decimal.Decimal
	TotalOperatingResultUsd decimal.Decimal
	ProjectReturnPct        decimal.Decimal
}

// GeneralCrops representa el resumen general de cultivos (GRAL CULTIVOS).
type GeneralCrops struct {
	TotalSurfaceHa          decimal.Decimal
	TotalNetIncomeUsd       decimal.Decimal
	TotalDirectCostsUsd     decimal.Decimal
	TotalRentUsd            decimal.Decimal
	TotalStructureUsd       decimal.Decimal
	TotalInvestedProjectUsd decimal.Decimal
	TotalOperatingResultUsd decimal.Decimal
	ProjectReturnPct        decimal.Decimal
}

// SummaryResultsResponse representa la respuesta completa del reporte
type SummaryResultsResponse struct {
	ProjectID    int64
	ProjectName  string
	CustomerID   int64
	CustomerName string
	CampaignID   int64
	CampaignName string

	// Resultados por cultivo
	Crops []SummaryResults

	// Totales del proyecto (GRAL CAMPOS)
	Totals ProjectTotals

	// Resumen general de cultivos (GRAL CULTIVOS)
	GeneralCrops GeneralCrops
}

// ===== FILTROS =====

// SummaryResultsFilter representa los filtros para el reporte de resumen de resultados
type SummaryResultsFilter struct {
	ProjectID  *int64
	CustomerID *int64
	CampaignID *int64
	FieldID    *int64
}
