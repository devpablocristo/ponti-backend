// Package dto holds the Data Transfer Objects for reports.
package dto

import (
	"github.com/shopspring/decimal"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
)

/* =========================
   REQUEST DTOs
========================= */

// SummaryResultsRequest represents the request for summary results report.
type SummaryResultsRequest struct {
	ProjectID  *int64 `form:"project_id" binding:"omitempty"`
	CustomerID *int64 `form:"customer_id" binding:"omitempty"`
	CampaignID *int64 `form:"campaign_id" binding:"omitempty"`
	FieldID    *int64 `form:"field_id" binding:"omitempty"`
}

/* =========================
   RESPONSE DTOs
========================= */

// SummaryResultsResponse represents the summary results report response.
type SummaryResultsResponse struct {
	ProjectID    int64  `json:"project_id"`
	ProjectName  string `json:"project_name"`
	CustomerID   int64  `json:"customer_id"`
	CustomerName string `json:"customer_name"`
	CampaignID   int64  `json:"campaign_id"`
	CampaignName string `json:"campaign_name"`

	// Resultados por cultivo
	Cultivos []CropSummaryResponse `json:"cultivos"`

	// Totales del proyecto
	Totales ProjectTotalsResponse `json:"totales"`
}

// CropSummaryResponse represents a crop summary in the report.
type CropSummaryResponse struct {
	CropID   int64  `json:"crop_id"`
	CropName string `json:"crop_name"`

	// Métricas por cultivo
	SuperficieHa          decimal.Decimal `json:"superficie_ha"`
	IngresoNetoUsd        decimal.Decimal `json:"ingreso_neto_usd"`
	CostosDirectosUsd     decimal.Decimal `json:"costos_directos_usd"`
	ArriendoUsd           decimal.Decimal `json:"arriendo_usd"`
	EstructuraUsd         decimal.Decimal `json:"estructura_usd"`
	TotalInvertidoUsd     decimal.Decimal `json:"total_invertido_usd"`
	ResultadoOperativoUsd decimal.Decimal `json:"resultado_operativo_usd"`
	RentaCultivoPct       decimal.Decimal `json:"renta_cultivo_pct"`
}

// ProjectTotalsResponse represents the project totals in the report.
type ProjectTotalsResponse struct {
	TotalSuperficieHa          decimal.Decimal `json:"total_superficie_ha"`
	TotalIngresoNetoUsd        decimal.Decimal `json:"total_ingreso_neto_usd"`
	TotalCostosDirectosUsd     decimal.Decimal `json:"total_costos_directos_usd"`
	TotalArriendoUsd           decimal.Decimal `json:"total_arriendo_usd"`
	TotalEstructuraUsd         decimal.Decimal `json:"total_estructura_usd"`
	TotalInvertidoProyectoUsd  decimal.Decimal `json:"total_invertido_proyecto_usd"`
	TotalResultadoOperativoUsd decimal.Decimal `json:"total_resultado_operativo_usd"`
	RentaProyectoPct           decimal.Decimal `json:"renta_proyecto_pct"`
}

/* =========================
   MAPPING FUNCTIONS
========================= */

// ToDomainSummaryResultsFilter maps DTO to domain filters.
func ToDomainSummaryResultsFilter(in SummaryResultsRequest) domain.SummaryResultsFilter {
	return domain.SummaryResultsFilter{
		ProjectID:  in.ProjectID,
		CustomerID: in.CustomerID,
		CampaignID: in.CampaignID,
		FieldID:    in.FieldID,
	}
}

// FromDomainSummaryResults maps domain to DTO response.
func FromDomainSummaryResults(d *domain.SummaryResultsResponse) *SummaryResultsResponse {
	// Mapear cultivos
	cultivos := make([]CropSummaryResponse, len(d.Cultivos))
	for i, cultivo := range d.Cultivos {
		cultivos[i] = CropSummaryResponse{
			CropID:                cultivo.CropID,
			CropName:              cultivo.CropName,
			SuperficieHa:          cultivo.SuperficieHa,
			IngresoNetoUsd:        cultivo.IngresoNetoUsd,
			CostosDirectosUsd:     cultivo.CostosDirectosUsd,
			ArriendoUsd:           cultivo.ArriendoUsd,
			EstructuraUsd:         cultivo.EstructuraUsd,
			TotalInvertidoUsd:     cultivo.TotalInvertidoUsd,
			ResultadoOperativoUsd: cultivo.ResultadoOperativoUsd,
			RentaCultivoPct:       cultivo.RentaCultivoPct,
		}
	}

	// Mapear totales
	totales := ProjectTotalsResponse{
		TotalSuperficieHa:          d.Totales.TotalSuperficieHa,
		TotalIngresoNetoUsd:        d.Totales.TotalIngresoNetoUsd,
		TotalCostosDirectosUsd:     d.Totales.TotalCostosDirectosUsd,
		TotalArriendoUsd:           d.Totales.TotalArriendoUsd,
		TotalEstructuraUsd:         d.Totales.TotalEstructuraUsd,
		TotalInvertidoProyectoUsd:  d.Totales.TotalInvertidoProyectoUsd,
		TotalResultadoOperativoUsd: d.Totales.TotalResultadoOperativoUsd,
		RentaProyectoPct:           d.Totales.RentaProyectoPct,
	}

	return &SummaryResultsResponse{
		ProjectID:    d.ProjectID,
		ProjectName:  d.ProjectName,
		CustomerID:   d.CustomerID,
		CustomerName: d.CustomerName,
		CampaignID:   d.CampaignID,
		CampaignName: d.CampaignName,
		Cultivos:     cultivos,
		Totales:      totales,
	}
}
