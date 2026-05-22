// Package domain contiene los modelos de dominio para los reportes.
package domain

import (
	"github.com/shopspring/decimal"
)

// ===== DOMAIN MODELS =====

// ReportFilter representa los filtros para los reportes
type ReportFilter struct {
	CustomerID *int64
	ProjectID  *int64
	CampaignID *int64
	FieldID    *int64
}

// ProjectInfo representa la información básica de un proyecto.
// El mapping a columnas SQL vive en internal/report/repository/models/project_info.go;
// el domain no conoce de persistencia.
type ProjectInfo struct {
	ProjectID    int64
	ProjectName  string
	CustomerID   int64
	CustomerName string
	CampaignID   int64
	CampaignName string
}

// FieldCropMetric representa una métrica por campo y cultivo
type FieldCropMetric struct {
	ProjectID int64
	FieldID   int64
	FieldName string
	CropID    int64
	CropName  string

	// Información general
	SurfaceHa       decimal.Decimal
	ProductionTn    decimal.Decimal
	SownAreaHa      decimal.Decimal
	HarvestedAreaHa decimal.Decimal

	// Rendimiento
	YieldTnHa decimal.Decimal

	// Precios y comercialización
	GrossPriceUsdTn     decimal.Decimal
	FreightCostUsdTn    decimal.Decimal
	CommercialCostUsdTn decimal.Decimal
	NetPriceUsdTn       decimal.Decimal

	// Ingreso neto
	NetIncomeUsd   decimal.Decimal
	NetIncomeUsdHa decimal.Decimal

	// Costos directos
	LaborCostsUsd       decimal.Decimal
	LaborCostsUsdHa     decimal.Decimal // TODO: Confirmar en vista v4
	SupplyCostsUsd      decimal.Decimal
	SupplyCostsUsdHa    decimal.Decimal // TODO: Confirmar en vista v4
	TotalDirectCostsUsd decimal.Decimal
	DirectCostsUsdHa    decimal.Decimal

	// Margen bruto
	GrossMarginUsd   decimal.Decimal
	GrossMarginUsdHa decimal.Decimal

	// Arriendo
	RentUsd   decimal.Decimal
	RentUsdHa decimal.Decimal

	// Costos administrativos
	AdministrationUsd   decimal.Decimal
	AdministrationUsdHa decimal.Decimal

	// Resultado operativo
	OperatingResultUsd   decimal.Decimal
	OperatingResultUsdHa decimal.Decimal

	// Total invertido
	TotalInvestedUsd   decimal.Decimal
	TotalInvestedUsdHa decimal.Decimal

	// Métricas calculadas
	ReturnPct              decimal.Decimal
	IndifferenceYieldUsdTn decimal.Decimal
}

// ===== TABLE DOMAIN MODELS =====

// FieldCrop representa la tabla de reporte field-crop
type FieldCrop struct {
	ProjectID    int64
	ProjectName  string
	CustomerID   int64
	CustomerName string
	CampaignID   int64
	CampaignName string
	Columns      []FieldCropColumn
	Rows         []FieldCropRow
}

// FieldCropColumn representa una columna en la tabla
type FieldCropColumn struct {
	ID        string // field_id-crop_id
	FieldID   int64
	FieldName string
	CropID    int64
	CropName  string
}

// FieldCropRow representa una fila en la tabla
type FieldCropRow struct {
	Key       string
	Unit      string
	ValueType string // "number" or "text"
	Values    map[string]FieldCropValue
}

// FieldCropValue representa un valor en la tabla
type FieldCropValue struct {
	Number decimal.Decimal
}

// LaborMetric representa una métrica de labor
type LaborMetric struct {
	LaborID        int64
	LaborName      string
	CategoryID     int64
	CategoryName   string
	SurfaceHa      decimal.Decimal
	CostUsd        decimal.Decimal
	CostPerHa      decimal.Decimal
	WorkOrderCount int64
}

// SupplyMetric representa una métrica de supply
type SupplyMetric struct {
	SupplyID       int64
	SupplyName     string
	CategoryID     int64
	CategoryName   string
	SurfaceHa      decimal.Decimal
	QuantityUsed   decimal.Decimal
	CostUsd        decimal.Decimal
	CostPerHa      decimal.Decimal
	WorkOrderCount int64
}
