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
	Crops []CropSummaryResponse `json:"crops"`

	// Totales del proyecto (GRAL CAMPOS)
	Totals ProjectTotalsResponse `json:"totals"`

	// Resumen general de cultivos (GRAL CULTIVOS)
	GeneralCrops GeneralCropsResponse `json:"general_crops"`
}

// CropSummaryResponse represents a crop summary in the report.
type CropSummaryResponse struct {
	CropID   int64  `json:"crop_id"`
	CropName string `json:"crop_name"`

	// Métricas por cultivo
	SurfaceHa          decimal.Decimal `json:"surface_ha"`
	NetIncomeUsd       decimal.Decimal `json:"net_income_usd"`
	DirectCostsUsd     decimal.Decimal `json:"direct_costs_usd"`
	RentUsd            decimal.Decimal `json:"rent_usd"`
	StructureUsd       decimal.Decimal `json:"structure_usd"`
	TotalInvestedUsd   decimal.Decimal `json:"total_invested_usd"`
	OperatingResultUsd decimal.Decimal `json:"operating_result_usd"`
	CropReturnPct      decimal.Decimal `json:"crop_return_pct"`
}

// ProjectTotalsResponse represents the project totals in the report (GRAL CAMPOS).
type ProjectTotalsResponse struct {
	TotalSurfaceHa          decimal.Decimal `json:"total_surface_ha"`
	TotalNetIncomeUsd       decimal.Decimal `json:"total_net_income_usd"`
	TotalDirectCostsUsd     decimal.Decimal `json:"total_direct_costs_usd"`
	TotalRentUsd            decimal.Decimal `json:"total_rent_usd"`
	TotalStructureUsd       decimal.Decimal `json:"total_structure_usd"`
	TotalInvestedProjectUsd decimal.Decimal `json:"total_invested_project_usd"`
	TotalOperatingResultUsd decimal.Decimal `json:"total_operating_result_usd"`
	ProjectReturnPct        decimal.Decimal `json:"project_return_pct"`
}

// GeneralCropsResponse represents the general crops summary (GRAL CULTIVOS).
type GeneralCropsResponse struct {
	TotalSurfaceHa          decimal.Decimal `json:"total_surface_ha"`
	TotalNetIncomeUsd       decimal.Decimal `json:"total_net_income_usd"`
	TotalDirectCostsUsd     decimal.Decimal `json:"total_direct_costs_usd"`
	TotalRentUsd            decimal.Decimal `json:"total_rent_usd"`
	TotalStructureUsd       decimal.Decimal `json:"total_structure_usd"`
	TotalInvestedProjectUsd decimal.Decimal `json:"total_invested_project_usd"`
	TotalOperatingResultUsd decimal.Decimal `json:"total_operating_result_usd"`
	ProjectReturnPct        decimal.Decimal `json:"project_return_pct"`
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
	crops := make([]CropSummaryResponse, len(d.Crops))
	for i, crop := range d.Crops {
		crops[i] = CropSummaryResponse{
			CropID:             crop.CropID,
			CropName:           crop.CropName,
			SurfaceHa:          crop.SurfaceHa,
			NetIncomeUsd:       crop.NetIncomeUsd,
			DirectCostsUsd:     crop.DirectCostsUsd,
			RentUsd:            crop.RentUsd,
			StructureUsd:       crop.StructureUsd,
			TotalInvestedUsd:   crop.TotalInvestedUsd,
			OperatingResultUsd: crop.OperatingResultUsd,
			CropReturnPct:      crop.CropReturnPct,
		}
	}

	// Mapear totales
	totals := ProjectTotalsResponse{
		TotalSurfaceHa:          d.Totals.TotalSurfaceHa,
		TotalNetIncomeUsd:       d.Totals.TotalNetIncomeUsd,
		TotalDirectCostsUsd:     d.Totals.TotalDirectCostsUsd,
		TotalRentUsd:            d.Totals.TotalRentUsd,
		TotalStructureUsd:       d.Totals.TotalStructureUsd,
		TotalInvestedProjectUsd: d.Totals.TotalInvestedProjectUsd,
		TotalOperatingResultUsd: d.Totals.TotalOperatingResultUsd,
		ProjectReturnPct:        d.Totals.ProjectReturnPct,
	}

	// Mapear cultivos generales (GRAL CULTIVOS)
	generalCrops := GeneralCropsResponse{
		TotalSurfaceHa:          d.GeneralCrops.TotalSurfaceHa,
		TotalNetIncomeUsd:       d.GeneralCrops.TotalNetIncomeUsd,
		TotalDirectCostsUsd:     d.GeneralCrops.TotalDirectCostsUsd,
		TotalRentUsd:            d.GeneralCrops.TotalRentUsd,
		TotalStructureUsd:       d.GeneralCrops.TotalStructureUsd,
		TotalInvestedProjectUsd: d.GeneralCrops.TotalInvestedProjectUsd,
		TotalOperatingResultUsd: d.GeneralCrops.TotalOperatingResultUsd,
		ProjectReturnPct:        d.GeneralCrops.ProjectReturnPct,
	}

	return &SummaryResultsResponse{
		ProjectID:    d.ProjectID,
		ProjectName:  d.ProjectName,
		CustomerID:   d.CustomerID,
		CustomerName: d.CustomerName,
		CampaignID:   d.CampaignID,
		CampaignName: d.CampaignName,
		Crops:        crops,
		Totals:       totals,
		GeneralCrops: generalCrops,
	}
}
