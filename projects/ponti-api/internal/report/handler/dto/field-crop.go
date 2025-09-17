// Package dto holds the Data Transfer Objects for reports.
package dto

import (
	"github.com/shopspring/decimal"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
)

/* =========================
   REQUEST DTOs
========================= */

// ReportFilterRequest represents the request filter for reports.
type ReportFilterRequest struct {
	CustomerID *int64 `json:"customer_id" binding:"omitempty"`
	ProjectID  *int64 `json:"project_id" binding:"omitempty"`
	CampaignID *int64 `json:"campaign_id" binding:"omitempty"`
	FieldID    *int64 `json:"field_id" binding:"omitempty"`
}

/* =========================
   RESPONSE DTOs — Field/Crop
   (Columns = Fields; Rows = Indicators, with first row = Crop)
========================= */

// FieldColumnResponse represents a report column (a field of the project).
type FieldColumnResponse struct {
	FieldID   int64  `json:"field_id"`
	FieldName string `json:"field_name"`
}

// TableCellValue supports both textual and numeric values per (row, field).
// Exactly one of Text or Number should be set.
type TableCellValue struct {
	Text   *string          `json:"text,omitempty"`
	Number *decimal.Decimal `json:"number,omitempty"`
}

// ReportRowValueType indicates the type of values stored in a row.
type ReportRowValueType string

const (
	ReportRowValueTypeText   ReportRowValueType = "text"   // e.g., "crop"
	ReportRowValueTypeNumber ReportRowValueType = "number" // e.g., surface, yield, prices
)

// TableRowResponse represents a row (indicator) in the pivot table.
// - key: machine-friendly identifier (e.g., "crop", "surface", "yield")
// - unit: optional (e.g., "ha", "tn", "usd", "usd/tn", "tn/ha")
// - value_type: "text" | "number"
// - values: map[field_id] -> TableCellValue
type TableRowResponse struct {
	Key       string                   `json:"key"`
	Unit      *string                  `json:"unit,omitempty"`
	ValueType ReportRowValueType       `json:"value_type"`
	Values    map[int64]TableCellValue `json:"values"`
}

// FieldCropReportResponse represents the field/crop report (pivot style).
// Columns are project fields; rows are indicators (first row = Crop).
type FieldCropReportResponse struct {
	ProjectID    int64  `json:"project_id"`
	ProjectName  string `json:"project_name"`
	CustomerID   int64  `json:"customer_id"`
	CustomerName string `json:"customer_name"`
	CampaignID   int64  `json:"campaign_id"`
	CampaignName string `json:"campaign_name"`

	// Columns: project fields
	Columns []FieldColumnResponse `json:"columns"`

	// Rows: indicators (first is "Crop" with text values, the rest numeric)
	Rows []TableRowResponse `json:"rows"`

	// Desglose de Labores por cultivo
	Labors map[string][]LaborMetricResponse `json:"labors,omitempty"` // key: crop_name, value: labor metrics

	// Desglose de Insumos por cultivo
	Supplies map[string][]SupplyMetricResponse `json:"supplies,omitempty"` // key: crop_name, value: supply metrics
}

/* =========================
   RESPONSE DTOs — Detailed metric per Field+Crop (kept for other views/APIs)
========================= */

// FieldCropMetricResponse represents a detailed metric per field and crop.
// (Kept for APIs that require non-pivot, per (field,crop) detail.)
type FieldCropMetricResponse struct {
	ProjectID int64  `json:"project_id"`
	FieldID   int64  `json:"field_id"`
	FieldName string `json:"field_name"`
	CropID    int64  `json:"crop_id"`
	CropName  string `json:"crop_name"`

	// General info
	SurfaceHa       decimal.Decimal `json:"surface_ha"`
	ProductionTn    decimal.Decimal `json:"production_tn"`
	SownAreaHa      decimal.Decimal `json:"sown_area_ha"`
	HarvestedAreaHa decimal.Decimal `json:"harvested_area_ha"`

	// Yield
	YieldTnHa decimal.Decimal `json:"yield_tn_ha"`

	// Prices & commercialization
	GrossPriceUsdTn     decimal.Decimal `json:"gross_price_usd_tn"`
	FreightCostUsdTn    decimal.Decimal `json:"freight_cost_usd_tn"`
	CommercialCostUsdTn decimal.Decimal `json:"commercial_cost_usd_tn"`
	NetPriceUsdTn       decimal.Decimal `json:"net_price_usd_tn"`

	// Net income
	NetIncomeUsd   decimal.Decimal `json:"net_income_usd"`
	NetIncomeUsdHa decimal.Decimal `json:"net_income_usd_ha"`

	// Direct costs
	LaborsCostUsd       decimal.Decimal `json:"labors_cost_usd"`
	SuppliesCostUsd     decimal.Decimal `json:"supplies_cost_usd"`
	TotalDirectCostsUsd decimal.Decimal `json:"total_direct_costs_usd"`
	DirectCostsUsdHa    decimal.Decimal `json:"direct_costs_usd_ha"`

	// Gross margin
	GrossMarginUsd   decimal.Decimal `json:"gross_margin_usd"`
	GrossMarginUsdHa decimal.Decimal `json:"gross_margin_usd_ha"`

	// Rent (lease)
	LeaseUsd   decimal.Decimal `json:"lease_usd"`
	LeaseUsdHa decimal.Decimal `json:"lease_usd_ha"`

	// Administration
	AdminUsd   decimal.Decimal `json:"admin_usd"`
	AdminUsdHa decimal.Decimal `json:"admin_usd_ha"`

	// Operating result
	OperatingResultUsd   decimal.Decimal `json:"operating_result_usd"`
	OperatingResultUsdHa decimal.Decimal `json:"operating_result_usd_ha"`

	// Total invested
	TotalInvestedUsd   decimal.Decimal `json:"total_invested_usd"`
	TotalInvestedUsdHa decimal.Decimal `json:"total_invested_usd_ha"`

	// Calculated metrics
	ReturnPct              decimal.Decimal `json:"return_pct"`
	IndifferenceYieldTnHa  decimal.Decimal `json:"indifference_yield_tn_ha"`  // rinde de indiferencia (tn/ha)
	IndifferencePriceUsdTn decimal.Decimal `json:"indifference_price_usd_tn"` // precio de indiferencia (usd/tn)
}

/* =========================
   RESPONSE DTOs — Labors
========================= */

type LaborMetricResponse struct {
	LaborID        int64           `json:"labor_id"`
	LaborName      string          `json:"labor_name"`
	CategoryID     int64           `json:"category_id"`
	CategoryName   string          `json:"category_name"`
	SurfaceHa      decimal.Decimal `json:"surface_ha"`
	CostUsd        decimal.Decimal `json:"cost_usd"`
	CostPerHa      decimal.Decimal `json:"cost_per_ha"`
	WorkOrderCount int64           `json:"workorder_count"`
}

/* =========================
   RESPONSE DTOs — Supplies
========================= */

type SupplyMetricResponse struct {
	SupplyID       int64           `json:"supply_id"`
	SupplyName     string          `json:"supply_name"`
	CategoryID     int64           `json:"category_id"`
	CategoryName   string          `json:"category_name"`
	SurfaceHa      decimal.Decimal `json:"surface_ha"`
	QuantityUsed   decimal.Decimal `json:"quantity_used"`
	CostUsd        decimal.Decimal `json:"cost_usd"`
	CostPerHa      decimal.Decimal `json:"cost_per_ha"`
	WorkOrderCount int64           `json:"workorder_count"`
}

/* =========================
   RESPONSE DTOs — Table Format
========================= */

// ReportTableResponse represents the field/crop report in table format.
type ReportTableResponse struct {
	ProjectID    int64               `json:"project_id"`
	ProjectName  string              `json:"project_name"`
	CustomerID   *int64              `json:"customer_id,omitempty"`
	CustomerName *string             `json:"customer_name,omitempty"`
	CampaignID   *int64              `json:"campaign_id,omitempty"`
	CampaignName *string             `json:"campaign_name,omitempty"`
	Columns      []ReportTableColumn `json:"columns"`
	Rows         []ReportTableRow    `json:"rows"`
}

// ReportTableColumn represents a column in the report table.
// Each column represents a specific field+crop combination.
type ReportTableColumn struct {
	ID        string `json:"id"` // "fieldId-cropId"
	FieldID   int64  `json:"field_id"`
	FieldName string `json:"field_name"`
	CropID    int64  `json:"crop_id"`
	CropName  string `json:"crop_name"`
}

// NumberValue represents a numeric value in the table.
type NumberValue struct {
	Number string `json:"number"`
}

// ReportTableRow represents a row in the report table.
// Each row represents a specific metric with values for each column.
type ReportTableRow struct {
	Key       string                 `json:"key"`
	Unit      string                 `json:"unit"`
	ValueType string                 `json:"value_type"`
	Values    map[string]NumberValue `json:"values"`
}

/* =========================
   MAPPING FUNCTIONS
========================= */

// FromDomainFieldCrop convierte el dominio a DTO simple
func FromDomainFieldCrop(table domain.FieldCrop) ReportTableResponse {
	// Convertir columnas
	columns := make([]ReportTableColumn, 0, len(table.Columns))
	for _, col := range table.Columns {
		columns = append(columns, ReportTableColumn{
			ID:        col.ID,
			FieldID:   col.FieldID,
			FieldName: col.FieldName,
			CropID:    col.CropID,
			CropName:  col.CropName,
		})
	}

	// Convertir filas
	rows := make([]ReportTableRow, 0, len(table.Rows))
	for _, row := range table.Rows {
		values := make(map[string]NumberValue)
		for fieldCropKey, value := range row.Values {
			values[fieldCropKey] = NumberValue{
				Number: value.Number.String(),
			}
		}

		rows = append(rows, ReportTableRow{
			Key:       row.Key,
			Unit:      row.Unit,
			ValueType: row.ValueType,
			Values:    values,
		})
	}

	return ReportTableResponse{
		ProjectID:    table.ProjectID,
		ProjectName:  table.ProjectName,
		CustomerID:   table.CustomerID,
		CustomerName: table.CustomerName,
		CampaignID:   table.CampaignID,
		CampaignName: table.CampaignName,
		Columns:      columns,
		Rows:         rows,
	}
}

/* =========================
   MAPPING FUNCTIONS ONLY
========================= */

/* =========================
   MAPPERS
========================= */

// FromDomainFieldCropMetric maps domain to DTO (detailed per field+crop).
func FromDomainFieldCropMetric(d *domain.FieldCropMetric) *FieldCropMetricResponse {
	return &FieldCropMetricResponse{
		ProjectID:              d.ProjectID,
		FieldID:                d.FieldID,
		FieldName:              d.FieldName,
		CropID:                 d.CropID,
		CropName:               d.CropName,
		SurfaceHa:              d.SuperficieHa, // domain already computed
		ProductionTn:           d.ProduccionTn,
		SownAreaHa:             d.AreaSembradaHa,
		HarvestedAreaHa:        d.AreaCosechadaHa,
		YieldTnHa:              d.RendimientoTnHa,
		GrossPriceUsdTn:        d.PrecioBrutoUsdTn,
		FreightCostUsdTn:       d.GastoFleteUsdTn,
		CommercialCostUsdTn:    d.GastoComercialUsdTn,
		NetPriceUsdTn:          d.PrecioNetoUsdTn,
		NetIncomeUsd:           d.IngresoNetoUsd,
		NetIncomeUsdHa:         d.IngresoNetoUsdHa,
		LaborsCostUsd:          d.CostosLaboresUsd,
		SuppliesCostUsd:        d.CostosInsumosUsd,
		TotalDirectCostsUsd:    d.TotalCostosDirectosUsd,
		DirectCostsUsdHa:       d.CostosDirectosUsdHa,
		GrossMarginUsd:         d.MargenBrutoUsd,
		GrossMarginUsdHa:       d.MargenBrutoUsdHa,
		LeaseUsd:               d.ArriendoUsd,
		LeaseUsdHa:             d.ArriendoUsdHa,
		AdminUsd:               d.AdministracionUsd,
		AdminUsdHa:             d.AdministracionUsdHa,
		OperatingResultUsd:     d.ResultadoOperativoUsd,
		OperatingResultUsdHa:   d.ResultadoOperativoUsdHa,
		TotalInvestedUsd:       d.TotalInvertidoUsd,
		TotalInvestedUsdHa:     d.TotalInvertidoUsdHa,
		ReturnPct:              d.RentaPct,
		IndifferenceYieldTnHa:  d.RindeIndiferenciaUsdTn, // NOTE: ensure domain naming matches semantics
		IndifferencePriceUsdTn: d.PrecioNetoUsdTn,        // TODO: replace with domain field if available
	}
}

// ToDomainReportFilter maps DTO to domain filters.
func ToDomainReportFilter(in ReportFilterRequest) domain.ReportFilter {
	return domain.ReportFilter{
		CustomerID: in.CustomerID,
		ProjectID:  in.ProjectID,
		CampaignID: in.CampaignID,
		FieldID:    in.FieldID,
	}
}
