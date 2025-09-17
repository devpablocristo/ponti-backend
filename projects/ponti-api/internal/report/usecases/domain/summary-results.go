// Package domain contiene las entidades de dominio para los reportes
package domain

import (
	"github.com/shopspring/decimal"
)

// ===== ENTIDADES DE DOMINIO =====

// SummaryResults representa el resumen de resultados por cultivo
type SummaryResults struct {
	ProjectID int64  `json:"project_id"`
	CropID    int64  `json:"crop_id"`
	CropName  string `json:"crop_name"`

	// Métricas por cultivo
	SuperficieHa          decimal.Decimal `json:"superficie_ha"`
	IngresoNetoUsd        decimal.Decimal `json:"ingreso_neto_usd"`
	CostosDirectosUsd     decimal.Decimal `json:"costos_directos_usd"`
	ArriendoUsd           decimal.Decimal `json:"arriendo_usd"`
	EstructuraUsd         decimal.Decimal `json:"estructura_usd"`
	TotalInvertidoUsd     decimal.Decimal `json:"total_invertido_usd"`
	ResultadoOperativoUsd decimal.Decimal `json:"resultado_operativo_usd"`
	RentaCultivoPct       decimal.Decimal `json:"renta_cultivo_pct"`

	// Totales del proyecto (para comparación)
	TotalSuperficieHa          decimal.Decimal `json:"total_superficie_ha"`
	TotalIngresoNetoUsd        decimal.Decimal `json:"total_ingreso_neto_usd"`
	TotalCostosDirectosUsd     decimal.Decimal `json:"total_costos_directos_usd"`
	TotalArriendoUsd           decimal.Decimal `json:"total_arriendo_usd"`
	TotalEstructuraUsd         decimal.Decimal `json:"total_estructura_usd"`
	TotalInvertidoProyectoUsd  decimal.Decimal `json:"total_invertido_proyecto_usd"`
	TotalResultadoOperativoUsd decimal.Decimal `json:"total_resultado_operativo_usd"`
	RentaProyectoPct           decimal.Decimal `json:"renta_proyecto_pct"`
}

// ProjectTotals representa los totales del proyecto
type ProjectTotals struct {
	TotalSuperficieHa          decimal.Decimal `json:"total_superficie_ha"`
	TotalIngresoNetoUsd        decimal.Decimal `json:"total_ingreso_neto_usd"`
	TotalCostosDirectosUsd     decimal.Decimal `json:"total_costos_directos_usd"`
	TotalArriendoUsd           decimal.Decimal `json:"total_arriendo_usd"`
	TotalEstructuraUsd         decimal.Decimal `json:"total_estructura_usd"`
	TotalInvertidoProyectoUsd  decimal.Decimal `json:"total_invertido_usd"`
	TotalResultadoOperativoUsd decimal.Decimal `json:"total_resultado_operativo_usd"`
	RentaProyectoPct           decimal.Decimal `json:"renta_proyecto_pct"`
}

// SummaryResultsResponse representa la respuesta completa del reporte
type SummaryResultsResponse struct {
	ProjectID    int64  `json:"project_id"`
	ProjectName  string `json:"project_name"`
	CustomerID   int64  `json:"customer_id"`
	CustomerName string `json:"customer_name"`
	CampaignID   int64  `json:"campaign_id"`
	CampaignName string `json:"campaign_name"`

	// Resultados por cultivo
	Cultivos []SummaryResults `json:"cultivos"`

	// Totales del proyecto
	Totales ProjectTotals `json:"totales"`
}

// ===== FILTROS =====

// SummaryResultsFilter representa los filtros para el reporte de resumen de resultados
type SummaryResultsFilter struct {
	ProjectID  *int64 `json:"project_id"`
	CustomerID *int64 `json:"customer_id"`
	CampaignID *int64 `json:"campaign_id"`
	FieldID    *int64 `json:"field_id"`
}
