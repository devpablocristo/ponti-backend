package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
	"github.com/shopspring/decimal"
)

// MetricCategory define las categorías de métricas
type MetricCategory string

const (
	CategorySurface    MetricCategory = "surface"
	CategoryProduction MetricCategory = "production"
	CategoryPrice      MetricCategory = "price"
	CategoryCost       MetricCategory = "cost"
	CategoryProfit     MetricCategory = "profit"
)

// MetricDefinition define una métrica del reporte
type MetricDefinition struct {
	Key      string
	Unit     string
	Category MetricCategory
	GetValue func(metric domain.FieldCropMetric) decimal.Decimal
}

// GetAvailableMetrics retorna todas las métricas disponibles
func GetAvailableMetrics() []MetricDefinition {
	return []MetricDefinition{
		// Superficie en hectáreas por tipo de cultivo
		{
			Key:      "surface",
			Unit:     "ha",
			Category: CategorySurface,
			GetValue: func(m domain.FieldCropMetric) decimal.Decimal { return m.SuperficieHa },
		},

		// Toneladas por tipo de cultivo
		{
			Key:      "production",
			Unit:     "tn",
			Category: CategoryProduction,
			GetValue: func(m domain.FieldCropMetric) decimal.Decimal { return m.ProduccionTn },
		},

		// Rendimiento por tipo de cultivo (Toneladas / Hectáreas)
		{
			Key:      "yield",
			Unit:     "tn/ha",
			Category: CategoryProduction,
			GetValue: func(m domain.FieldCropMetric) decimal.Decimal { return m.RendimientoTnHa },
		},

		// Gasto de flete por tipo de cultivo
		{
			Key:      "freight_cost",
			Unit:     "usd/tn",
			Category: CategoryPrice,
			GetValue: func(m domain.FieldCropMetric) decimal.Decimal { return m.GastoFleteUsdTn },
		},

		// Gasto comercial por tipo de cultivo
		{
			Key:      "commercial_cost",
			Unit:     "usd/tn",
			Category: CategoryPrice,
			GetValue: func(m domain.FieldCropMetric) decimal.Decimal { return m.GastoComercialUsdTn },
		},

		// Precio neto por tipo de cultivo
		{
			Key:      "net_price",
			Unit:     "usd/tn",
			Category: CategoryPrice,
			GetValue: func(m domain.FieldCropMetric) decimal.Decimal { return m.PrecioNetoUsdTn },
		},

		// Precio bruto por tonelada
		{
			Key:      "gross_price",
			Unit:     "usd/tn",
			Category: CategoryPrice,
			GetValue: func(m domain.FieldCropMetric) decimal.Decimal { return m.PrecioBrutoUsdTn },
		},

		// Ingreso neto por hectárea
		{
			Key:      "net_income",
			Unit:     "usd/ha",
			Category: CategoryProfit,
			GetValue: func(m domain.FieldCropMetric) decimal.Decimal { return m.IngresoNetoUsdHa },
		},

		// Total de labor por hectárea
		{
			Key:      "labors_cost",
			Unit:     "usd/ha",
			Category: CategoryCost,
			GetValue: func(m domain.FieldCropMetric) decimal.Decimal { return m.CostosLaboresUsd.Div(m.SuperficieHa) },
		},

		// Total de insumos por hectárea
		{
			Key:      "supplies_cost",
			Unit:     "usd/ha",
			Category: CategoryCost,
			GetValue: func(m domain.FieldCropMetric) decimal.Decimal { return m.CostosInsumosUsd.Div(m.SuperficieHa) },
		},

		// Total costos directos por hectárea
		{
			Key:      "total_direct_costs",
			Unit:     "usd/ha",
			Category: CategoryCost,
			GetValue: func(m domain.FieldCropMetric) decimal.Decimal { return m.CostosDirectosUsdHa },
		},

		// Margen bruto por hectárea
		{
			Key:      "gross_margin",
			Unit:     "usd/ha",
			Category: CategoryProfit,
			GetValue: func(m domain.FieldCropMetric) decimal.Decimal { return m.MargenBrutoUsdHa },
		},

		// Arriendo por hectárea
		{
			Key:      "lease",
			Unit:     "usd/ha",
			Category: CategoryCost,
			GetValue: func(m domain.FieldCropMetric) decimal.Decimal { return m.ArriendoUsdHa },
		},

		// Costo administrativo por hectárea
		{
			Key:      "admin",
			Unit:     "usd/ha",
			Category: CategoryCost,
			GetValue: func(m domain.FieldCropMetric) decimal.Decimal { return m.AdministracionUsdHa },
		},

		// Resultado operativo por hectárea
		{
			Key:      "operating_result",
			Unit:     "usd/ha",
			Category: CategoryProfit,
			GetValue: func(m domain.FieldCropMetric) decimal.Decimal { return m.ResultadoOperativoUsdHa },
		},

		// Total invertido (Costos directos + Arriendo + Administración)
		{
			Key:      "total_invested",
			Unit:     "usd",
			Category: CategoryCost,
			GetValue: func(m domain.FieldCropMetric) decimal.Decimal { return m.TotalInvertidoUsd },
		},

		// Renta (Resultado operativo / Total invertido)
		{
			Key:      "return_pct",
			Unit:     "%",
			Category: CategoryProfit,
			GetValue: func(m domain.FieldCropMetric) decimal.Decimal { return m.RentaPct },
		},

		// Rinde indiferencia
		{
			Key:      "indifference_yield",
			Unit:     "tn/ha",
			Category: CategoryProfit,
			GetValue: func(m domain.FieldCropMetric) decimal.Decimal { return m.RindeIndiferenciaUsdTn },
		},

		// Precio indiferencia (calculado como Total invertido / Rendimiento)
		{
			Key:      "indifference_price",
			Unit:     "usd/tn",
			Category: CategoryProfit,
			GetValue: func(m domain.FieldCropMetric) decimal.Decimal {
				if m.RendimientoTnHa.IsZero() {
					return decimal.Zero
				}
				return m.TotalInvertidoUsd.Div(m.RendimientoTnHa)
			},
		},
	}
}
