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
	SurfaceHa       decimal.Decimal `gorm:"column:superficie_ha"`
	ProductionTn    decimal.Decimal `gorm:"column:produccion_tn"`
	SownAreaHa      decimal.Decimal `gorm:"column:area_sembrada_ha"`
	HarvestedAreaHa decimal.Decimal `gorm:"column:area_cosechada_ha"`

	// Rendimiento
	YieldTnHa decimal.Decimal `gorm:"column:rendimiento_tn_ha"`

	// Precios y comercialización
	GrossPriceUsdTn     decimal.Decimal `gorm:"column:precio_bruto_usd_tn"`
	FreightCostUsdTn    decimal.Decimal `gorm:"column:gasto_flete_usd_tn"`
	CommercialCostUsdTn decimal.Decimal `gorm:"column:gasto_comercial_usd_tn"`
	NetPriceUsdTn       decimal.Decimal `gorm:"column:precio_neto_usd_tn"`

	// Ingreso neto
	NetIncomeUsd   decimal.Decimal `gorm:"column:ingreso_neto_usd"`
	NetIncomeUsdHa decimal.Decimal `gorm:"column:ingreso_neto_usd_ha"`

	// Costos directos
	LaborCostsUsd       decimal.Decimal `gorm:"column:costos_labores_usd"`
	SupplyCostsUsd      decimal.Decimal `gorm:"column:costos_insumos_usd"`
	TotalDirectCostsUsd decimal.Decimal `gorm:"column:total_costos_directos_usd"`
	DirectCostsUsdHa    decimal.Decimal `gorm:"column:costos_directos_usd_ha"`

	// Margen bruto
	GrossMarginUsd   decimal.Decimal `gorm:"column:margen_bruto_usd"`
	GrossMarginUsdHa decimal.Decimal `gorm:"column:margen_bruto_usd_ha"`

	// Arriendo
	RentUsd   decimal.Decimal `gorm:"column:arriendo_usd"`
	RentUsdHa decimal.Decimal `gorm:"column:arriendo_usd_ha"`

	// Costos administrativos
	AdministrationUsd   decimal.Decimal `gorm:"column:administracion_usd"`
	AdministrationUsdHa decimal.Decimal `gorm:"column:administracion_usd_ha"`

	// Resultado operativo
	OperatingResultUsd   decimal.Decimal `gorm:"column:resultado_operativo_usd"`
	OperatingResultUsdHa decimal.Decimal `gorm:"column:resultado_operativo_usd_ha"`

	// Total invertido
	TotalInvestedUsd   decimal.Decimal `gorm:"column:total_invertido_usd"`
	TotalInvestedUsdHa decimal.Decimal `gorm:"column:total_invertido_usd_ha"`

	// Métricas calculadas
	ReturnPct              decimal.Decimal `gorm:"column:renta_pct"`
	IndifferenceYieldUsdTn decimal.Decimal `gorm:"column:rinde_indiferencia_usd_tn"`
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
		ProjectID:              m.ProjectID,
		FieldID:                m.FieldID,
		FieldName:              m.FieldName,
		CropID:                 m.CropID,
		CropName:               m.CropName,
		SurfaceHa:              m.SurfaceHa,
		ProductionTn:           m.ProductionTn,
		SownAreaHa:             m.SownAreaHa,
		HarvestedAreaHa:        m.HarvestedAreaHa,
		YieldTnHa:              m.YieldTnHa,
		GrossPriceUsdTn:        m.GrossPriceUsdTn,
		FreightCostUsdTn:       m.FreightCostUsdTn,
		CommercialCostUsdTn:    m.CommercialCostUsdTn,
		NetPriceUsdTn:          m.NetPriceUsdTn,
		NetIncomeUsd:           m.NetIncomeUsd,
		NetIncomeUsdHa:         m.NetIncomeUsdHa,
		LaborCostsUsd:          m.LaborCostsUsd,
		SupplyCostsUsd:         m.SupplyCostsUsd,
		TotalDirectCostsUsd:    m.TotalDirectCostsUsd,
		DirectCostsUsdHa:       m.DirectCostsUsdHa,
		GrossMarginUsd:         m.GrossMarginUsd,
		GrossMarginUsdHa:       m.GrossMarginUsdHa,
		RentUsd:                m.RentUsd,
		RentUsdHa:              m.RentUsdHa,
		AdministrationUsd:      m.AdministrationUsd,
		AdministrationUsdHa:    m.AdministrationUsdHa,
		OperatingResultUsd:     m.OperatingResultUsd,
		OperatingResultUsdHa:   m.OperatingResultUsdHa,
		TotalInvestedUsd:       m.TotalInvestedUsd,
		TotalInvestedUsdHa:     m.TotalInvestedUsdHa,
		ReturnPct:              m.ReturnPct,
		IndifferenceYieldUsdTn: m.IndifferenceYieldUsdTn,
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
			total.SurfaceHa = total.SurfaceHa.Add(metric.SurfaceHa)
			total.ProductionTn = total.ProductionTn.Add(metric.ProductionTn)
			total.SownAreaHa = total.SownAreaHa.Add(metric.SownAreaHa)
			total.HarvestedAreaHa = total.HarvestedAreaHa.Add(metric.HarvestedAreaHa)
			total.NetIncomeUsd = total.NetIncomeUsd.Add(metric.NetIncomeUsd)
			total.LaborCostsUsd = total.LaborCostsUsd.Add(metric.LaborCostsUsd)
			total.SupplyCostsUsd = total.SupplyCostsUsd.Add(metric.SupplyCostsUsd)
			total.TotalDirectCostsUsd = total.TotalDirectCostsUsd.Add(metric.TotalDirectCostsUsd)
			total.GrossMarginUsd = total.GrossMarginUsd.Add(metric.GrossMarginUsd)
			total.RentUsd = total.RentUsd.Add(metric.RentUsd)
			total.AdministrationUsd = total.AdministrationUsd.Add(metric.AdministrationUsd)
			total.OperatingResultUsd = total.OperatingResultUsd.Add(metric.OperatingResultUsd)
			total.TotalInvestedUsd = total.TotalInvestedUsd.Add(metric.TotalInvestedUsd)
		}

		// Calcular promedios y ratios
		if total.HarvestedAreaHa.GreaterThan(decimal.Zero) {
			total.YieldTnHa = total.ProductionTn.Div(total.HarvestedAreaHa)
		}

		if total.SownAreaHa.GreaterThan(decimal.Zero) {
			total.NetIncomeUsdHa = total.NetIncomeUsd.Div(total.SownAreaHa)
			total.DirectCostsUsdHa = total.TotalDirectCostsUsd.Div(total.SownAreaHa)
			total.GrossMarginUsdHa = total.GrossMarginUsd.Div(total.SownAreaHa)
			total.RentUsdHa = total.RentUsd.Div(total.SownAreaHa)
			total.AdministrationUsdHa = total.AdministrationUsd.Div(total.SownAreaHa)
			total.OperatingResultUsdHa = total.OperatingResultUsd.Div(total.SownAreaHa)
			total.TotalInvestedUsdHa = total.TotalInvestedUsd.Div(total.SownAreaHa)
		}

		// Calcular renta
		if total.TotalInvestedUsd.GreaterThan(decimal.Zero) {
			total.ReturnPct = total.OperatingResultUsd.Div(total.TotalInvestedUsd)
		}

		// Calcular rinde indiferencia
		if total.YieldTnHa.GreaterThan(decimal.Zero) {
			total.IndifferenceYieldUsdTn = total.TotalInvestedUsd.Div(total.YieldTnHa)
		}

		// Promedio de precios (usar el primer valor no cero)
		for _, metric := range cropMetrics {
			if metric.GrossPriceUsdTn.GreaterThan(decimal.Zero) {
				total.GrossPriceUsdTn = metric.GrossPriceUsdTn
				break
			}
		}
		for _, metric := range cropMetrics {
			if metric.FreightCostUsdTn.GreaterThan(decimal.Zero) {
				total.FreightCostUsdTn = metric.FreightCostUsdTn
				break
			}
		}
		for _, metric := range cropMetrics {
			if metric.CommercialCostUsdTn.GreaterThan(decimal.Zero) {
				total.CommercialCostUsdTn = metric.CommercialCostUsdTn
				break
			}
		}
		for _, metric := range cropMetrics {
			if metric.NetPriceUsdTn.GreaterThan(decimal.Zero) {
				total.NetPriceUsdTn = metric.NetPriceUsdTn
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
		total.SurfaceHa = total.SurfaceHa.Add(metric.SurfaceHa)
		total.ProductionTn = total.ProductionTn.Add(metric.ProductionTn)
		total.SownAreaHa = total.SownAreaHa.Add(metric.SownAreaHa)
		total.HarvestedAreaHa = total.HarvestedAreaHa.Add(metric.HarvestedAreaHa)
		total.NetIncomeUsd = total.NetIncomeUsd.Add(metric.NetIncomeUsd)
		total.LaborCostsUsd = total.LaborCostsUsd.Add(metric.LaborCostsUsd)
		total.SupplyCostsUsd = total.SupplyCostsUsd.Add(metric.SupplyCostsUsd)
		total.TotalDirectCostsUsd = total.TotalDirectCostsUsd.Add(metric.TotalDirectCostsUsd)
		total.GrossMarginUsd = total.GrossMarginUsd.Add(metric.GrossMarginUsd)
		total.RentUsd = total.RentUsd.Add(metric.RentUsd)
		total.AdministrationUsd = total.AdministrationUsd.Add(metric.AdministrationUsd)
		total.OperatingResultUsd = total.OperatingResultUsd.Add(metric.OperatingResultUsd)
		total.TotalInvestedUsd = total.TotalInvestedUsd.Add(metric.TotalInvestedUsd)
	}

	// Calcular promedios y ratios
	if total.HarvestedAreaHa.GreaterThan(decimal.Zero) {
		total.YieldTnHa = total.ProductionTn.Div(total.HarvestedAreaHa)
	}

	if total.SownAreaHa.GreaterThan(decimal.Zero) {
		total.NetIncomeUsdHa = total.NetIncomeUsd.Div(total.SownAreaHa)
		total.DirectCostsUsdHa = total.TotalDirectCostsUsd.Div(total.SownAreaHa)
		total.GrossMarginUsdHa = total.GrossMarginUsd.Div(total.SownAreaHa)
		total.RentUsdHa = total.RentUsd.Div(total.SownAreaHa)
		total.AdministrationUsdHa = total.AdministrationUsd.Div(total.SownAreaHa)
		total.OperatingResultUsdHa = total.OperatingResultUsd.Div(total.SownAreaHa)
		total.TotalInvestedUsdHa = total.TotalInvestedUsd.Div(total.SownAreaHa)
	}

	// Calcular renta
	if total.TotalInvestedUsd.GreaterThan(decimal.Zero) {
		total.ReturnPct = total.OperatingResultUsd.Div(total.TotalInvestedUsd)
	}

	// Calcular rinde indiferencia
	if total.YieldTnHa.GreaterThan(decimal.Zero) {
		total.IndifferenceYieldUsdTn = total.TotalInvestedUsd.Div(total.YieldTnHa)
	}

	return total
}
