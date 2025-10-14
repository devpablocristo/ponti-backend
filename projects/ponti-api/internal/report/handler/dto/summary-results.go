// Package dto holds the Data Transfer Objects for reports.
package dto

import (
	"encoding/json"

	"github.com/shopspring/decimal"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
)

// Decimal3 es un wrapper de decimal.Decimal que serializa con 3 decimales
type Decimal3 struct {
	decimal.Decimal
}

// MarshalJSON implementa json.Marshaler para formatear con 3 decimales
func (d Decimal3) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Decimal.StringFixed(3))
}

// NewDecimal3 crea un Decimal3 desde un decimal.Decimal
func NewDecimal3(d decimal.Decimal) Decimal3 {
	return Decimal3{Decimal: d}
}

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

	// Métricas por cultivo (formato 3 decimales)
	SurfaceHa          Decimal3 `json:"surface_ha"`
	NetIncomeUsd       Decimal3 `json:"net_income_usd"`
	DirectCostsUsd     Decimal3 `json:"direct_costs_usd"`
	RentUsd            Decimal3 `json:"rent_usd"`
	StructureUsd       Decimal3 `json:"structure_usd"`
	TotalInvestedUsd   Decimal3 `json:"total_invested_usd"`
	OperatingResultUsd Decimal3 `json:"operating_result_usd"`
	CropReturnPct      Decimal3 `json:"crop_return_pct"`
}

// ProjectTotalsResponse represents the project totals in the report (GRAL CAMPOS).
type ProjectTotalsResponse struct {
	TotalSurfaceHa          Decimal3 `json:"total_surface_ha"`
	TotalNetIncomeUsd       Decimal3 `json:"total_net_income_usd"`
	TotalDirectCostsUsd     Decimal3 `json:"total_direct_costs_usd"`
	TotalRentUsd            Decimal3 `json:"total_rent_usd"`
	TotalStructureUsd       Decimal3 `json:"total_structure_usd"`
	TotalInvestedProjectUsd Decimal3 `json:"total_invested_project_usd"`
	TotalOperatingResultUsd Decimal3 `json:"total_operating_result_usd"`
	ProjectReturnPct        Decimal3 `json:"project_return_pct"`
}

// GeneralCropsResponse represents the general crops summary (GRAL CULTIVOS).
type GeneralCropsResponse struct {
	TotalSurfaceHa          Decimal3 `json:"total_surface_ha"`
	TotalNetIncomeUsd       Decimal3 `json:"total_net_income_usd"`
	TotalDirectCostsUsd     Decimal3 `json:"total_direct_costs_usd"`
	TotalRentUsd            Decimal3 `json:"total_rent_usd"`
	TotalStructureUsd       Decimal3 `json:"total_structure_usd"`
	TotalInvestedProjectUsd Decimal3 `json:"total_invested_project_usd"`
	TotalOperatingResultUsd Decimal3 `json:"total_operating_result_usd"`
	ProjectReturnPct        Decimal3 `json:"project_return_pct"`
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
	// Mapear cultivos (con formato de 3 decimales)
	crops := make([]CropSummaryResponse, len(d.Crops))
	for i, crop := range d.Crops {
		crops[i] = CropSummaryResponse{
			CropID:             crop.CropID,
			CropName:           crop.CropName,
			SurfaceHa:          NewDecimal3(crop.SurfaceHa),
			NetIncomeUsd:       NewDecimal3(crop.NetIncomeUsd),
			DirectCostsUsd:     NewDecimal3(crop.DirectCostsUsd),
			RentUsd:            NewDecimal3(crop.RentUsd),
			StructureUsd:       NewDecimal3(crop.StructureUsd),
			TotalInvestedUsd:   NewDecimal3(crop.TotalInvestedUsd),
			OperatingResultUsd: NewDecimal3(crop.OperatingResultUsd),
			CropReturnPct:      NewDecimal3(crop.CropReturnPct),
		}
	}

	// Mapear totales (con formato de 3 decimales)
	totals := ProjectTotalsResponse{
		TotalSurfaceHa:          NewDecimal3(d.Totals.TotalSurfaceHa),
		TotalNetIncomeUsd:       NewDecimal3(d.Totals.TotalNetIncomeUsd),
		TotalDirectCostsUsd:     NewDecimal3(d.Totals.TotalDirectCostsUsd),
		TotalRentUsd:            NewDecimal3(d.Totals.TotalRentUsd),
		TotalStructureUsd:       NewDecimal3(d.Totals.TotalStructureUsd),
		TotalInvestedProjectUsd: NewDecimal3(d.Totals.TotalInvestedProjectUsd),
		TotalOperatingResultUsd: NewDecimal3(d.Totals.TotalOperatingResultUsd),
		ProjectReturnPct:        NewDecimal3(d.Totals.ProjectReturnPct),
	}

	// Mapear cultivos generales (GRAL CULTIVOS) con formato de 3 decimales
	generalCrops := GeneralCropsResponse{
		TotalSurfaceHa:          NewDecimal3(d.GeneralCrops.TotalSurfaceHa),
		TotalNetIncomeUsd:       NewDecimal3(d.GeneralCrops.TotalNetIncomeUsd),
		TotalDirectCostsUsd:     NewDecimal3(d.GeneralCrops.TotalDirectCostsUsd),
		TotalRentUsd:            NewDecimal3(d.GeneralCrops.TotalRentUsd),
		TotalStructureUsd:       NewDecimal3(d.GeneralCrops.TotalStructureUsd),
		TotalInvestedProjectUsd: NewDecimal3(d.GeneralCrops.TotalInvestedProjectUsd),
		TotalOperatingResultUsd: NewDecimal3(d.GeneralCrops.TotalOperatingResultUsd),
		ProjectReturnPct:        NewDecimal3(d.GeneralCrops.ProjectReturnPct),
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
