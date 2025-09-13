// Package models contiene los modelos de base de datos para los reportes
package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
	"github.com/shopspring/decimal"
)

// ===== MODELOS DE BASE DE DATOS =====

// FieldCropMetricModel representa el modelo de base de datos para métricas por campo y cultivo
type FieldCropMetricModel struct {
	ProjectID int64  `gorm:"column:project_id"`
	FieldID   int64  `gorm:"column:field_id"`
	FieldName string `gorm:"column:field_name"`
	CropID    int64  `gorm:"column:current_crop_id"`
	CropName  string `gorm:"column:crop_name"`

	// Información general
	SuperficieHa    decimal.Decimal `gorm:"column:superficie_ha"`
	ProduccionTn    decimal.Decimal `gorm:"column:produccion_tn"`
	AreaSembradaHa  decimal.Decimal `gorm:"column:area_sembrada_ha"`
	AreaCosechadaHa decimal.Decimal `gorm:"column:area_cosechada_ha"`

	// Rendimiento
	RendimientoTnHa decimal.Decimal `gorm:"column:rendimiento_tn_ha"`

	// Precios y comercialización
	PrecioBrutoUsdTn    decimal.Decimal `gorm:"column:precio_bruto_usd_tn"`
	GastoFleteUsdTn     decimal.Decimal `gorm:"column:gasto_flete_usd_tn"`
	GastoComercialUsdTn decimal.Decimal `gorm:"column:gasto_comercial_usd_tn"`
	PrecioNetoUsdTn     decimal.Decimal `gorm:"column:precio_neto_usd_tn"`

	// Ingreso neto
	IngresoNetoUsd   decimal.Decimal `gorm:"column:ingreso_neto_usd"`
	IngresoNetoUsdHa decimal.Decimal `gorm:"column:ingreso_neto_usd_ha"`

	// Costos directos
	CostosLaboresUsd       decimal.Decimal `gorm:"column:costos_labores_usd"`
	CostosInsumosUsd       decimal.Decimal `gorm:"column:costos_insumos_usd"`
	TotalCostosDirectosUsd decimal.Decimal `gorm:"column:total_costos_directos_usd"`
	CostosDirectosUsdHa    decimal.Decimal `gorm:"column:costos_directos_usd_ha"`

	// Margen bruto
	MargenBrutoUsd   decimal.Decimal `gorm:"column:margen_bruto_usd"`
	MargenBrutoUsdHa decimal.Decimal `gorm:"column:margen_bruto_usd_ha"`

	// Arriendo
	ArriendoUsd   decimal.Decimal `gorm:"column:arriendo_usd"`
	ArriendoUsdHa decimal.Decimal `gorm:"column:arriendo_usd_ha"`

	// Costos administrativos
	AdministracionUsd   decimal.Decimal `gorm:"column:administracion_usd"`
	AdministracionUsdHa decimal.Decimal `gorm:"column:administracion_usd_ha"`

	// Resultado operativo
	ResultadoOperativoUsd   decimal.Decimal `gorm:"column:resultado_operativo_usd"`
	ResultadoOperativoUsdHa decimal.Decimal `gorm:"column:resultado_operativo_usd_ha"`

	// Total invertido
	TotalInvertidoUsd   decimal.Decimal `gorm:"column:total_invertido_usd"`
	TotalInvertidoUsdHa decimal.Decimal `gorm:"column:total_invertido_usd_ha"`

	// Métricas calculadas
	RentaPct               decimal.Decimal `gorm:"column:renta_pct"`
	RindeIndiferenciaUsdTn decimal.Decimal `gorm:"column:rinde_indiferencia_usd_tn"`
}

// TableName especifica el nombre de la tabla para GORM
// ACTUALIZADO: Usar vista v3 (SSOT)
func (FieldCropMetricModel) TableName() string {
	return "v3_report_field_crop_metrics_view"
}

// LaborMetricModel representa el modelo de base de datos para métricas de labores
type LaborMetricModel struct {
	ProjectID      int64           `gorm:"column:project_id"`
	FieldID        int64           `gorm:"column:field_id"`
	CropID         int64           `gorm:"column:crop_id"`
	CropName       string          `gorm:"column:crop_name"`
	LaborID        int64           `gorm:"column:labor_id"`
	LaborName      string          `gorm:"column:labor_name"`
	CategoryID     int64           `gorm:"column:category_id"`
	CategoryName   string          `gorm:"column:category_name"`
	SurfaceHa      decimal.Decimal `gorm:"column:surface_ha"`
	CostUsd        decimal.Decimal `gorm:"column:cost_usd"`
	CostPerHa      decimal.Decimal `gorm:"column:cost_per_ha"`
	WorkOrderCount int64           `gorm:"column:workorder_count"`
}

// SupplyMetricModel representa el modelo de base de datos para métricas de insumos
type SupplyMetricModel struct {
	ProjectID      int64           `gorm:"column:project_id"`
	FieldID        int64           `gorm:"column:field_id"`
	CropID         int64           `gorm:"column:crop_id"`
	CropName       string          `gorm:"column:crop_name"`
	SupplyID       int64           `gorm:"column:supply_id"`
	SupplyName     string          `gorm:"column:supply_name"`
	CategoryID     int64           `gorm:"column:category_id"`
	CategoryName   string          `gorm:"column:category_name"`
	SurfaceHa      decimal.Decimal `gorm:"column:surface_ha"`
	QuantityUsed   decimal.Decimal `gorm:"column:quantity_used"`
	CostUsd        decimal.Decimal `gorm:"column:cost_usd"`
	CostPerHa      decimal.Decimal `gorm:"column:cost_per_ha"`
	WorkOrderCount int64           `gorm:"column:workorder_count"`
}

// ===== MAPPERS =====

// ToDomainFieldCropMetric convierte de modelo a dominio
func (m *FieldCropMetricModel) ToDomainFieldCropMetric() *domain.FieldCropMetric {
	return &domain.FieldCropMetric{
		ProjectID:               m.ProjectID,
		FieldID:                 m.FieldID,
		FieldName:               m.FieldName,
		CropID:                  m.CropID,
		CropName:                m.CropName,
		SuperficieHa:            m.SuperficieHa,
		ProduccionTn:            m.ProduccionTn,
		AreaSembradaHa:          m.AreaSembradaHa,
		AreaCosechadaHa:         m.AreaCosechadaHa,
		RendimientoTnHa:         m.RendimientoTnHa,
		PrecioBrutoUsdTn:        m.PrecioBrutoUsdTn,
		GastoFleteUsdTn:         m.GastoFleteUsdTn,
		GastoComercialUsdTn:     m.GastoComercialUsdTn,
		PrecioNetoUsdTn:         m.PrecioNetoUsdTn,
		IngresoNetoUsd:          m.IngresoNetoUsd,
		IngresoNetoUsdHa:        m.IngresoNetoUsdHa,
		CostosLaboresUsd:        m.CostosLaboresUsd,
		CostosInsumosUsd:        m.CostosInsumosUsd,
		TotalCostosDirectosUsd:  m.TotalCostosDirectosUsd,
		CostosDirectosUsdHa:     m.CostosDirectosUsdHa,
		MargenBrutoUsd:          m.MargenBrutoUsd,
		MargenBrutoUsdHa:        m.MargenBrutoUsdHa,
		ArriendoUsd:             m.ArriendoUsd,
		ArriendoUsdHa:           m.ArriendoUsdHa,
		AdministracionUsd:       m.AdministracionUsd,
		AdministracionUsdHa:     m.AdministracionUsdHa,
		ResultadoOperativoUsd:   m.ResultadoOperativoUsd,
		ResultadoOperativoUsdHa: m.ResultadoOperativoUsdHa,
		TotalInvertidoUsd:       m.TotalInvertidoUsd,
		TotalInvertidoUsdHa:     m.TotalInvertidoUsdHa,
		RentaPct:                m.RentaPct,
		RindeIndiferenciaUsdTn:  m.RindeIndiferenciaUsdTn,
	}
}

// ToDomainLaborMetric convierte de modelo a dominio
func (m *LaborMetricModel) ToDomainLaborMetric() *domain.LaborMetric {
	return &domain.LaborMetric{
		LaborID:        m.LaborID,
		LaborName:      m.LaborName,
		CategoryID:     m.CategoryID,
		CategoryName:   m.CategoryName,
		SurfaceHa:      m.SurfaceHa,
		CostUsd:        m.CostUsd,
		CostPerHa:      m.CostPerHa,
		WorkOrderCount: m.WorkOrderCount,
	}
}

// ToDomainSupplyMetric convierte de modelo a dominio
func (m *SupplyMetricModel) ToDomainSupplyMetric() *domain.SupplyMetric {
	return &domain.SupplyMetric{
		SupplyID:       m.SupplyID,
		SupplyName:     m.SupplyName,
		CategoryID:     m.CategoryID,
		CategoryName:   m.CategoryName,
		SurfaceHa:      m.SurfaceHa,
		QuantityUsed:   m.QuantityUsed,
		CostUsd:        m.CostUsd,
		CostPerHa:      m.CostPerHa,
		WorkOrderCount: m.WorkOrderCount,
	}
}

// ===== FUNCIONES DE AGRUPACIÓN =====

// GroupFieldCropMetricsByCrop agrupa las métricas por cultivo
func GroupFieldCropMetricsByCrop(metrics []*domain.FieldCropMetric) map[string][]domain.FieldCropMetric {
	grouped := make(map[string][]domain.FieldCropMetric)

	for _, metric := range metrics {
		cropName := metric.CropName
		if cropName == "" {
			cropName = "Sin cultivo"
		}
		grouped[cropName] = append(grouped[cropName], *metric)
	}

	return grouped
}

// CalculateCropTotals calcula los totales por cultivo
func CalculateCropTotals(metrics []*domain.FieldCropMetric) map[string]domain.FieldCropMetric {
	totals := make(map[string]domain.FieldCropMetric)

	// Agrupar por cultivo
	grouped := make(map[string][]*domain.FieldCropMetric)
	for _, metric := range metrics {
		cropName := metric.CropName
		if cropName == "" {
			cropName = "Sin cultivo"
		}
		grouped[cropName] = append(grouped[cropName], metric)
	}

	// Calcular totales para cada cultivo
	for cropName, cropMetrics := range grouped {
		if len(cropMetrics) == 0 {
			continue
		}

		total := domain.FieldCropMetric{
			CropName: cropName,
		}

		for _, metric := range cropMetrics {
			// Sumar valores numéricos
			total.SuperficieHa = total.SuperficieHa.Add(metric.SuperficieHa)
			total.ProduccionTn = total.ProduccionTn.Add(metric.ProduccionTn)
			total.AreaSembradaHa = total.AreaSembradaHa.Add(metric.AreaSembradaHa)
			total.AreaCosechadaHa = total.AreaCosechadaHa.Add(metric.AreaCosechadaHa)
			total.IngresoNetoUsd = total.IngresoNetoUsd.Add(metric.IngresoNetoUsd)
			total.CostosLaboresUsd = total.CostosLaboresUsd.Add(metric.CostosLaboresUsd)
			total.CostosInsumosUsd = total.CostosInsumosUsd.Add(metric.CostosInsumosUsd)
			total.TotalCostosDirectosUsd = total.TotalCostosDirectosUsd.Add(metric.TotalCostosDirectosUsd)
			total.MargenBrutoUsd = total.MargenBrutoUsd.Add(metric.MargenBrutoUsd)
			total.ArriendoUsd = total.ArriendoUsd.Add(metric.ArriendoUsd)
			total.AdministracionUsd = total.AdministracionUsd.Add(metric.AdministracionUsd)
			total.ResultadoOperativoUsd = total.ResultadoOperativoUsd.Add(metric.ResultadoOperativoUsd)
			total.TotalInvertidoUsd = total.TotalInvertidoUsd.Add(metric.TotalInvertidoUsd)
		}

		// Calcular promedios y ratios
		if total.AreaCosechadaHa.GreaterThan(decimal.Zero) {
			total.RendimientoTnHa = total.ProduccionTn.Div(total.AreaCosechadaHa)
		}

		if total.AreaSembradaHa.GreaterThan(decimal.Zero) {
			total.IngresoNetoUsdHa = total.IngresoNetoUsd.Div(total.AreaSembradaHa)
			total.CostosDirectosUsdHa = total.TotalCostosDirectosUsd.Div(total.AreaSembradaHa)
			total.MargenBrutoUsdHa = total.MargenBrutoUsd.Div(total.AreaSembradaHa)
			total.ArriendoUsdHa = total.ArriendoUsd.Div(total.AreaSembradaHa)
			total.AdministracionUsdHa = total.AdministracionUsd.Div(total.AreaSembradaHa)
			total.ResultadoOperativoUsdHa = total.ResultadoOperativoUsd.Div(total.AreaSembradaHa)
			total.TotalInvertidoUsdHa = total.TotalInvertidoUsd.Div(total.AreaSembradaHa)
		}

		// Calcular renta
		if total.TotalInvertidoUsd.GreaterThan(decimal.Zero) {
			total.RentaPct = total.ResultadoOperativoUsd.Div(total.TotalInvertidoUsd)
		}

		// Calcular rinde indiferencia
		if total.RendimientoTnHa.GreaterThan(decimal.Zero) {
			total.RindeIndiferenciaUsdTn = total.TotalInvertidoUsd.Div(total.RendimientoTnHa)
		}

		// Promedio de precios (usar el primer valor no cero)
		for _, metric := range cropMetrics {
			if metric.PrecioBrutoUsdTn.GreaterThan(decimal.Zero) {
				total.PrecioBrutoUsdTn = metric.PrecioBrutoUsdTn
				break
			}
		}
		for _, metric := range cropMetrics {
			if metric.GastoFleteUsdTn.GreaterThan(decimal.Zero) {
				total.GastoFleteUsdTn = metric.GastoFleteUsdTn
				break
			}
		}
		for _, metric := range cropMetrics {
			if metric.GastoComercialUsdTn.GreaterThan(decimal.Zero) {
				total.GastoComercialUsdTn = metric.GastoComercialUsdTn
				break
			}
		}
		for _, metric := range cropMetrics {
			if metric.PrecioNetoUsdTn.GreaterThan(decimal.Zero) {
				total.PrecioNetoUsdTn = metric.PrecioNetoUsdTn
				break
			}
		}

		totals[cropName] = total
	}

	return totals
}

// CalculateGrandTotal calcula el total general
func CalculateGrandTotal(metrics []*domain.FieldCropMetric) domain.FieldCropMetric {
	total := domain.FieldCropMetric{
		CropName: "TOTAL GENERAL",
	}

	for _, metric := range metrics {
		// Sumar valores numéricos
		total.SuperficieHa = total.SuperficieHa.Add(metric.SuperficieHa)
		total.ProduccionTn = total.ProduccionTn.Add(metric.ProduccionTn)
		total.AreaSembradaHa = total.AreaSembradaHa.Add(metric.AreaSembradaHa)
		total.AreaCosechadaHa = total.AreaCosechadaHa.Add(metric.AreaCosechadaHa)
		total.IngresoNetoUsd = total.IngresoNetoUsd.Add(metric.IngresoNetoUsd)
		total.CostosLaboresUsd = total.CostosLaboresUsd.Add(metric.CostosLaboresUsd)
		total.CostosInsumosUsd = total.CostosInsumosUsd.Add(metric.CostosInsumosUsd)
		total.TotalCostosDirectosUsd = total.TotalCostosDirectosUsd.Add(metric.TotalCostosDirectosUsd)
		total.MargenBrutoUsd = total.MargenBrutoUsd.Add(metric.MargenBrutoUsd)
		total.ArriendoUsd = total.ArriendoUsd.Add(metric.ArriendoUsd)
		total.AdministracionUsd = total.AdministracionUsd.Add(metric.AdministracionUsd)
		total.ResultadoOperativoUsd = total.ResultadoOperativoUsd.Add(metric.ResultadoOperativoUsd)
		total.TotalInvertidoUsd = total.TotalInvertidoUsd.Add(metric.TotalInvertidoUsd)
	}

	// Calcular promedios y ratios
	if total.AreaCosechadaHa.GreaterThan(decimal.Zero) {
		total.RendimientoTnHa = total.ProduccionTn.Div(total.AreaCosechadaHa)
	}

	if total.AreaSembradaHa.GreaterThan(decimal.Zero) {
		total.IngresoNetoUsdHa = total.IngresoNetoUsd.Div(total.AreaSembradaHa)
		total.CostosDirectosUsdHa = total.TotalCostosDirectosUsd.Div(total.AreaSembradaHa)
		total.MargenBrutoUsdHa = total.MargenBrutoUsd.Div(total.AreaSembradaHa)
		total.ArriendoUsdHa = total.ArriendoUsd.Div(total.AreaSembradaHa)
		total.AdministracionUsdHa = total.AdministracionUsd.Div(total.AreaSembradaHa)
		total.ResultadoOperativoUsdHa = total.ResultadoOperativoUsd.Div(total.AreaSembradaHa)
		total.TotalInvertidoUsdHa = total.TotalInvertidoUsd.Div(total.AreaSembradaHa)
	}

	// Calcular renta
	if total.TotalInvertidoUsd.GreaterThan(decimal.Zero) {
		total.RentaPct = total.ResultadoOperativoUsd.Div(total.TotalInvertidoUsd)
	}

	// Calcular rinde indiferencia
	if total.RendimientoTnHa.GreaterThan(decimal.Zero) {
		total.RindeIndiferenciaUsdTn = total.TotalInvertidoUsd.Div(total.RendimientoTnHa)
	}

	return total
}
