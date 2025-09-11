// Package domain contiene los modelos de dominio para los reportes
package domain

import (
	"github.com/shopspring/decimal"
)

// ===== DOMAIN MODELS =====

// ReportFilter representa los filtros para los reportes
type ReportFilter struct {
	CustomerID *int64 `json:"customer_id"`
	ProjectID  *int64 `json:"project_id"`
	CampaignID *int64 `json:"campaign_id"`
	FieldID    *int64 `json:"field_id"`
}

// ProjectInfo representa la información básica de un proyecto
type ProjectInfo struct {
	ProjectID    int64  `json:"project_id" gorm:"column:project_id"`
	ProjectName  string `json:"project_name" gorm:"column:project_name"`
	CustomerID   int64  `json:"customer_id" gorm:"column:customer_id"`
	CustomerName string `json:"customer_name" gorm:"column:customer_name"`
	CampaignID   int64  `json:"campaign_id" gorm:"column:campaign_id"`
	CampaignName string `json:"campaign_name" gorm:"column:campaign_name"`
}

// FieldCropMetric representa una métrica por campo y cultivo
type FieldCropMetric struct {
	ProjectID int64  `json:"project_id"`
	FieldID   int64  `json:"field_id"`
	FieldName string `json:"field_name"`
	CropID    int64  `json:"crop_id"`
	CropName  string `json:"crop_name"`

	// Información general
	SuperficieHa    decimal.Decimal `json:"superficie_ha"`
	ProduccionTn    decimal.Decimal `json:"produccion_tn"`
	AreaSembradaHa  decimal.Decimal `json:"area_sembrada_ha"`
	AreaCosechadaHa decimal.Decimal `json:"area_cosechada_ha"`

	// Rendimiento
	RendimientoTnHa decimal.Decimal `json:"rendimiento_tn_ha"`

	// Precios y comercialización
	PrecioBrutoUsdTn    decimal.Decimal `json:"precio_bruto_usd_tn"`
	GastoFleteUsdTn     decimal.Decimal `json:"gasto_flete_usd_tn"`
	GastoComercialUsdTn decimal.Decimal `json:"gasto_comercial_usd_tn"`
	PrecioNetoUsdTn     decimal.Decimal `json:"precio_neto_usd_tn"`

	// Ingreso neto
	IngresoNetoUsd   decimal.Decimal `json:"ingreso_neto_usd"`
	IngresoNetoUsdHa decimal.Decimal `json:"ingreso_neto_usd_ha"`

	// Costos directos
	CostosLaboresUsd       decimal.Decimal `json:"costos_labores_usd"`
	CostosInsumosUsd       decimal.Decimal `json:"costos_insumos_usd"`
	TotalCostosDirectosUsd decimal.Decimal `json:"total_costos_directos_usd"`
	CostosDirectosUsdHa    decimal.Decimal `json:"costos_directos_usd_ha"`

	// Margen bruto
	MargenBrutoUsd   decimal.Decimal `json:"margen_bruto_usd"`
	MargenBrutoUsdHa decimal.Decimal `json:"margen_bruto_usd_ha"`

	// Arriendo
	ArriendoUsd   decimal.Decimal `json:"arriendo_usd"`
	ArriendoUsdHa decimal.Decimal `json:"arriendo_usd_ha"`

	// Costos administrativos
	AdministracionUsd   decimal.Decimal `json:"administracion_usd"`
	AdministracionUsdHa decimal.Decimal `json:"administracion_usd_ha"`

	// Resultado operativo
	ResultadoOperativoUsd   decimal.Decimal `json:"resultado_operativo_usd"`
	ResultadoOperativoUsdHa decimal.Decimal `json:"resultado_operativo_usd_ha"`

	// Total invertido
	TotalInvertidoUsd   decimal.Decimal `json:"total_invertido_usd"`
	TotalInvertidoUsdHa decimal.Decimal `json:"total_invertido_usd_ha"`

	// Métricas calculadas
	RentaPct               decimal.Decimal `json:"renta_pct"`
	RindeIndiferenciaUsdTn decimal.Decimal `json:"rinde_indiferencia_usd_tn"`
}

// LaborMetric representa una métrica de labor
type LaborMetric struct {
	LaborID        int64           `json:"labor_id"`
	LaborName      string          `json:"labor_name"`
	CategoryID     int64           `json:"category_id"`
	CategoryName   string          `json:"category_name"`
	SurfaceHa      decimal.Decimal `json:"surface_ha"`
	CostUsd        decimal.Decimal `json:"cost_usd"`
	CostPerHa      decimal.Decimal `json:"cost_per_ha"`
	WorkOrderCount int64           `json:"workorder_count"`
}

// SupplyMetric representa una métrica de supply
type SupplyMetric struct {
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

// ===== TABLE DOMAIN MODELS =====

// FieldCrop representa la tabla de reporte field-crop
type FieldCrop struct {
	ProjectID    int64             `json:"project_id"`
	ProjectName  string            `json:"project_name"`
	CustomerID   *int64            `json:"customer_id,omitempty"`
	CustomerName *string           `json:"customer_name,omitempty"`
	CampaignID   *int64            `json:"campaign_id,omitempty"`
	CampaignName *string           `json:"campaign_name,omitempty"`
	Columns      []FieldCropColumn `json:"columns"`
	Rows         []FieldCropRow    `json:"rows"`
}

// FieldCropColumn representa una columna en la tabla
type FieldCropColumn struct {
	ID        string `json:"id"` // field_id-crop_id
	FieldID   int64  `json:"field_id"`
	FieldName string `json:"field_name"`
	CropID    int64  `json:"crop_id"`
	CropName  string `json:"crop_name"`
}

// FieldCropRow representa una fila en la tabla
type FieldCropRow struct {
	Key       string                    `json:"key"`
	Unit      string                    `json:"unit"`
	ValueType string                    `json:"value_type"` // "number" or "text"
	Values    map[string]FieldCropValue `json:"values"`
}

// FieldCropValue representa un valor en la tabla
type FieldCropValue struct {
	Number decimal.Decimal `json:"number"`
}
