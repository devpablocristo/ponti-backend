// Package report proporciona funcionalidades para generar reportes financieros y operativos
package report

import (
	"context"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
)

// GormEnginePort define la interfaz para el motor GORM
type GormEnginePort interface {
	Client() *gorm.DB
}

// ReportRepository implementa la interfaz de repositorio para reportes
type ReportRepository struct {
	db GormEnginePort
}

// NewReportRepository crea una nueva instancia del repositorio
func NewReportRepository(db GormEnginePort) *ReportRepository {
	return &ReportRepository{
		db: db,
	}
}

// ===== REPORTE POR CAMPO/CULTIVO =====

// GetFieldCropMetrics obtiene las métricas por campo y cultivo
func (r *ReportRepository) GetFieldCropMetrics(filters domain.ReportFilter) ([]domain.FieldCropMetric, error) {
	// Obtener project IDs relacionados con los filtros
	projectIDs, err := r.getRelatedProjectIDs(filters)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo proyectos relacionados: %w", err)
	}

	if len(projectIDs) == 0 {
		return []domain.FieldCropMetric{}, nil
	}

	// Usar IN en lugar de ANY para evitar problemas de mapeo de arrays
	placeholders := make([]string, len(projectIDs))
	for i := range projectIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(`
		SELECT project_id, field_id, field_name, current_crop_id, crop_name,
		       income_usd, costos_directos_ejecutados_usd, costos_directos_invertidos_usd,
		       arriendo_invertidos_usd, estructura_invertidos_usd, operating_result_usd, operating_result_pct
		FROM v3_report_field_crop_metrics_view 
		WHERE project_id IN (%s)
	`, strings.Join(placeholders, ","))

	args := make([]any, len(projectIDs))
	for i, id := range projectIDs {
		args[i] = id
	}

	// Aplicar filtros adicionales
	if filters.FieldID != nil {
		query += " AND field_id = $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, *filters.FieldID)
	}

	// Ejecutar consulta con la nueva estructura v3
	rows, err := r.db.Client().WithContext(context.Background()).Raw(query, args...).Rows()
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar consulta: %w", err)
	}
	defer rows.Close()

	var metrics []domain.FieldCropMetric
	for rows.Next() {
		// Leer valores raw de la vista v3
		var rawProjectID, rawFieldID, rawCropID any
		var fieldName, cropName string
		var rawIncomeUsd, rawCostosDirectosEjecutadosUsd, rawCostosDirectosInvertidosUsd any
		var rawArriendoInvertidosUsd, rawEstructuraInvertidosUsd, rawOperatingResultUsd, rawOperatingResultPct any

		if err := rows.Scan(
			&rawProjectID, &rawFieldID, &fieldName, &rawCropID, &cropName,
			&rawIncomeUsd, &rawCostosDirectosEjecutadosUsd, &rawCostosDirectosInvertidosUsd,
			&rawArriendoInvertidosUsd, &rawEstructuraInvertidosUsd, &rawOperatingResultUsd, &rawOperatingResultPct,
		); err != nil {
			return nil, fmt.Errorf("error al escanear fila: %w", err)
		}

		// Convertir valores raw
		projectID := r.convertToInt64(rawProjectID)
		fieldID := r.convertToInt64(rawFieldID)
		cropID := r.convertToInt64(rawCropID)

		metric := domain.FieldCropMetric{
			ProjectID:               projectID,
			FieldID:                 fieldID,
			FieldName:               fieldName,
			CropID:                  cropID,
			CropName:                cropName,
			SuperficieHa:            decimal.Zero, // No disponible en v3
			ProduccionTn:            decimal.Zero, // No disponible en v3
			AreaSembradaHa:          decimal.Zero, // No disponible en v3
			AreaCosechadaHa:         decimal.Zero, // No disponible en v3
			RendimientoTnHa:         decimal.Zero, // No disponible en v3
			PrecioBrutoUsdTn:        decimal.Zero, // No disponible en v3
			GastoFleteUsdTn:         decimal.Zero, // No disponible en v3
			GastoComercialUsdTn:     decimal.Zero, // No disponible en v3
			PrecioNetoUsdTn:         decimal.Zero, // No disponible en v3
			IngresoNetoUsd:          r.convertToDecimal(rawIncomeUsd),
			IngresoNetoUsdHa:        decimal.Zero, // No disponible en v3
			CostosLaboresUsd:        decimal.Zero, // No disponible en v3
			CostosInsumosUsd:        decimal.Zero, // No disponible en v3
			TotalCostosDirectosUsd:  r.convertToDecimal(rawCostosDirectosEjecutadosUsd),
			CostosDirectosUsdHa:     decimal.Zero, // No disponible en v3
			MargenBrutoUsd:          decimal.Zero, // No disponible en v3
			MargenBrutoUsdHa:        decimal.Zero, // No disponible en v3
			ArriendoUsd:             r.convertToDecimal(rawArriendoInvertidosUsd),
			ArriendoUsdHa:           decimal.Zero, // No disponible en v3
			AdministracionUsd:       r.convertToDecimal(rawEstructuraInvertidosUsd),
			AdministracionUsdHa:     decimal.Zero, // No disponible en v3
			ResultadoOperativoUsd:   r.convertToDecimal(rawOperatingResultUsd),
			ResultadoOperativoUsdHa: decimal.Zero, // No disponible en v3
			TotalInvertidoUsd:       r.convertToDecimal(rawCostosDirectosInvertidosUsd),
			TotalInvertidoUsdHa:     decimal.Zero, // No disponible en v3
			RentaPct:                r.convertToDecimal(rawOperatingResultPct),
			RindeIndiferenciaUsdTn:  decimal.Zero, // No disponible en v3
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// ===== FUNCIONES AUXILIARES =====

// BuildFieldCrop construye la tabla completa del reporte field-crop
func (r *ReportRepository) BuildFieldCrop(filters domain.ReportFilter) (*domain.FieldCrop, error) {
	// Obtener información del proyecto
	projectInfo, err := r.getProjectInfo(filters)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo información del proyecto: %w", err)
	}

	// Obtener columnas (field-crop combinations)
	columns, err := r.getFieldCropColumns(filters)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo columnas: %w", err)
	}

	// Obtener métricas básicas
	metrics, err := r.GetFieldCropMetrics(filters)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo métricas: %w", err)
	}

	// Construir filas del reporte
	rows := r.buildReportRows(metrics, columns)

	// Convertir ProjectInfo a los tipos correctos
	var customerID *int64
	var customerName *string
	var campaignID *int64
	var campaignName *string

	if projectInfo.CustomerID != 0 {
		customerID = &projectInfo.CustomerID
	}
	if projectInfo.CustomerName != "" {
		customerName = &projectInfo.CustomerName
	}
	if projectInfo.CampaignID != 0 {
		campaignID = &projectInfo.CampaignID
	}
	if projectInfo.CampaignName != "" {
		campaignName = &projectInfo.CampaignName
	}

	return &domain.FieldCrop{
		ProjectID:    projectInfo.ProjectID,
		ProjectName:  projectInfo.ProjectName,
		CustomerID:   customerID,
		CustomerName: customerName,
		CampaignID:   campaignID,
		CampaignName: campaignName,
		Columns:      columns,
		Rows:         rows,
	}, nil
}

// GetProjectInfo obtiene información del proyecto por ID
func (r *ReportRepository) GetProjectInfo(projectID int64) (*domain.ProjectInfo, error) {
	filters := domain.ReportFilter{
		ProjectID: &projectID,
	}
	return r.getProjectInfo(filters)
}

// GetInvestorContributionReport obtiene el reporte de aportes de inversores
func (r *ReportRepository) GetInvestorContributionReport(ctx context.Context, filter domain.ReportFilter) (*domain.InvestorContributionReport, error) {
	// Obtener project IDs relacionados con los filtros
	projectIDs, err := r.getRelatedProjectIDs(filter)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo proyectos relacionados: %w", err)
	}

	if len(projectIDs) == 0 {
		return &domain.InvestorContributionReport{}, nil
	}

	// Usar IN en lugar de un solo project_id para soportar todos los filtros
	placeholders := make([]string, len(projectIDs))
	for i := range projectIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(`
		SELECT 
			project_id,
			project_name,
			customer_id,
			customer_name,
			campaign_id,
			campaign_name,
			surface_total_ha,
			lease_fixed_usd,
			lease_is_fixed,
			admin_per_ha_usd,
			admin_total_usd,
			contributions_data,
			comparison_data,
			harvest_data
		FROM investor_contribution_data_view
		WHERE project_id IN (%s)
	`, strings.Join(placeholders, ","))

	args := make([]any, len(projectIDs))
	for i, id := range projectIDs {
		args[i] = id
	}

	var results []struct {
		ProjectID         int64           `gorm:"column:project_id"`
		ProjectName       string          `gorm:"column:project_name"`
		CustomerID        int64           `gorm:"column:customer_id"`
		CustomerName      string          `gorm:"column:customer_name"`
		CampaignID        int64           `gorm:"column:campaign_id"`
		CampaignName      string          `gorm:"column:campaign_name"`
		SurfaceTotalHa    decimal.Decimal `gorm:"column:surface_total_ha"`
		LeaseFixedUsd     decimal.Decimal `gorm:"column:lease_fixed_usd"`
		LeaseIsFixed      bool            `gorm:"column:lease_is_fixed"`
		AdminPerHaUsd     decimal.Decimal `gorm:"column:admin_per_ha_usd"`
		AdminTotalUsd     decimal.Decimal `gorm:"column:admin_total_usd"`
		ContributionsData string          `gorm:"column:contributions_data"`
		ComparisonData    string          `gorm:"column:comparison_data"`
		HarvestData       string          `gorm:"column:harvest_data"`
	}

	err = r.db.Client().Raw(query, args...).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("error consultando vista de aportes de inversores: %w", err)
	}

	if len(results) == 0 {
		return &domain.InvestorContributionReport{}, nil
	}

	// Si hay múltiples proyectos, usar el primero como base y agregar los demás
	// En el futuro se podría implementar una lógica de agregación más sofisticada
	firstResult := results[0]

	// Construir reporte desde los datos de la vista
	report := &domain.InvestorContributionReport{
		ProjectID:    firstResult.ProjectID,
		ProjectName:  firstResult.ProjectName,
		CustomerID:   firstResult.CustomerID,
		CustomerName: firstResult.CustomerName,
		CampaignID:   firstResult.CampaignID,
		CampaignName: firstResult.CampaignName,
		General: domain.GeneralProjectData{
			SurfaceTotalHa: firstResult.SurfaceTotalHa,
			LeaseFixedUsd:  firstResult.LeaseFixedUsd,
			LeaseIsFixed:   firstResult.LeaseIsFixed,
			AdminPerHaUsd:  firstResult.AdminPerHaUsd,
			AdminTotalUsd:  firstResult.AdminTotalUsd,
		},
	}

	// Parsear JSON de contributions
	if firstResult.ContributionsData != "" {
		contributions, err := r.parseContributionsFromJSON(firstResult.ContributionsData)
		if err != nil {
			return nil, fmt.Errorf("error parseando contributions: %w", err)
		}
		report.Contributions = contributions
	}

	// Parsear JSON de comparison
	if firstResult.ComparisonData != "" {
		comparison, err := r.parseComparisonFromJSON(firstResult.ComparisonData)
		if err != nil {
			return nil, fmt.Errorf("error parseando comparison: %w", err)
		}
		report.Comparison = comparison
	}

	// Parsear JSON de harvest
	if firstResult.HarvestData != "" {
		harvest, err := r.parseHarvestFromJSON(firstResult.HarvestData)
		if err != nil {
			return nil, fmt.Errorf("error parseando harvest: %w", err)
		}
		report.Harvest = harvest
	}

	return report, nil
}
