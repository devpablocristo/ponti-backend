// Package models contiene los modelos de base de datos para los reportes
package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
	"github.com/shopspring/decimal"
)

// ===== MODELOS DE BASE DE DATOS =====

// SummaryResultsModel representa el modelo de base de datos para resumen de resultados por cultivo
type SummaryResultsModel struct {
	ProjectID int64  `gorm:"column:project_id"`
	CropID    int64  `gorm:"column:current_crop_id"`
	CropName  string `gorm:"column:crop_name"`

	// Métricas por cultivo
	SuperficieHa          decimal.Decimal `gorm:"column:surface_ha"`
	IngresoNetoUsd        decimal.Decimal `gorm:"column:net_income_usd"`
	CostosDirectosUsd     decimal.Decimal `gorm:"column:direct_costs_usd"`
	ArriendoUsd           decimal.Decimal `gorm:"column:rent_usd"`
	EstructuraUsd         decimal.Decimal `gorm:"column:structure_usd"`
	TotalInvertidoUsd     decimal.Decimal `gorm:"column:total_invested_usd"`
	ResultadoOperativoUsd decimal.Decimal `gorm:"column:operating_result_usd"`
	RentaCultivoPct       decimal.Decimal `gorm:"column:crop_return_pct"`

	// Totales del proyecto (para comparación)
	TotalSuperficieHa          decimal.Decimal `gorm:"column:total_surface_ha"`
	TotalIngresoNetoUsd        decimal.Decimal `gorm:"column:total_net_income_usd"`
	TotalCostosDirectosUsd     decimal.Decimal `gorm:"column:total_direct_costs_usd"`
	TotalArriendoUsd           decimal.Decimal `gorm:"column:total_rent_usd"`
	TotalEstructuraUsd         decimal.Decimal `gorm:"column:total_structure_usd"`
	TotalInvertidoProyectoUsd  decimal.Decimal `gorm:"column:total_invested_project_usd"`
	TotalResultadoOperativoUsd decimal.Decimal `gorm:"column:total_operating_result_usd"`
	RentaProyectoPct           decimal.Decimal `gorm:"column:project_return_pct"`
}

// TableName especifica el nombre de la tabla para GORM
// Usar vista v3 (SSOT)
func (SummaryResultsModel) TableName() string {
	return "v3_report_summary_results_view"
}

// ===== MAPPERS =====

// ToDomainSummaryResults convierte de modelo a dominio
func (m *SummaryResultsModel) ToDomainSummaryResults() *domain.SummaryResults {
	return &domain.SummaryResults{
		ProjectID:                  m.ProjectID,
		CropID:                     m.CropID,
		CropName:                   m.CropName,
		SuperficieHa:               m.SuperficieHa,
		IngresoNetoUsd:             m.IngresoNetoUsd,
		CostosDirectosUsd:          m.CostosDirectosUsd,
		ArriendoUsd:                m.ArriendoUsd,
		EstructuraUsd:              m.EstructuraUsd,
		TotalInvertidoUsd:          m.TotalInvertidoUsd,
		ResultadoOperativoUsd:      m.ResultadoOperativoUsd,
		RentaCultivoPct:            m.RentaCultivoPct,
		TotalSuperficieHa:          m.TotalSuperficieHa,
		TotalIngresoNetoUsd:        m.TotalIngresoNetoUsd,
		TotalCostosDirectosUsd:     m.TotalCostosDirectosUsd,
		TotalArriendoUsd:           m.TotalArriendoUsd,
		TotalEstructuraUsd:         m.TotalEstructuraUsd,
		TotalInvertidoProyectoUsd:  m.TotalInvertidoProyectoUsd,
		TotalResultadoOperativoUsd: m.TotalResultadoOperativoUsd,
		RentaProyectoPct:           m.RentaProyectoPct,
	}
}

// ===== FUNCIONES DE AGRUPACIÓN =====

// GroupSummaryResultsByCrop agrupa los resultados por cultivo
func GroupSummaryResultsByCrop(results []*domain.SummaryResults) map[string][]domain.SummaryResults {
	grouped := make(map[string][]domain.SummaryResults)

	for _, result := range results {
		cropName := result.CropName
		if cropName == "" {
			cropName = "Sin cultivo"
		}
		grouped[cropName] = append(grouped[cropName], *result)
	}

	return grouped
}

// CalculateProjectTotals calcula los totales del proyecto
func CalculateProjectTotals(results []*domain.SummaryResults) *domain.ProjectTotals {
	if len(results) == 0 {
		return &domain.ProjectTotals{}
	}

	// Usar los totales del primer resultado (ya vienen calculados de la vista)
	first := results[0]
	return &domain.ProjectTotals{
		TotalSuperficieHa:          first.TotalSuperficieHa,
		TotalIngresoNetoUsd:        first.TotalIngresoNetoUsd,
		TotalCostosDirectosUsd:     first.TotalCostosDirectosUsd,
		TotalArriendoUsd:           first.TotalArriendoUsd,
		TotalEstructuraUsd:         first.TotalEstructuraUsd,
		TotalInvertidoProyectoUsd:  first.TotalInvertidoProyectoUsd,
		TotalResultadoOperativoUsd: first.TotalResultadoOperativoUsd,
		RentaProyectoPct:           first.RentaProyectoPct,
	}
}
