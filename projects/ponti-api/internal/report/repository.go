// Package report proporciona funcionalidades para generar reportes financieros y operativos
package report

import (
	"context"
	"fmt"
	"strings"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
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
		       superficie_ha, produccion_tn, area_sembrada_ha, area_cosechada_ha,
		       rendimiento_tn_ha, precio_bruto_usd_tn, gasto_flete_usd_tn, gasto_comercial_usd_tn, precio_neto_usd_tn,
		       ingreso_neto_usd, ingreso_neto_usd_ha, costos_labores_usd, costos_insumos_usd, 
		       total_costos_directos_usd, costos_directos_usd_ha, margen_bruto_usd, margen_bruto_usd_ha,
		       arriendo_usd, arriendo_usd_ha, administracion_usd, administracion_usd_ha,
		       resultado_operativo_usd, resultado_operativo_usd_ha, total_invertido_usd, total_invertido_usd_ha,
		       renta_pct, rinde_indiferencia_usd_tn
		FROM report_field_crop_metrics_view_v2 
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

	// Ejecutar consulta EXACTAMENTE como en el dashboard
	rows, err := r.db.Client().WithContext(context.Background()).Raw(query, args...).Rows()
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar consulta: %w", err)
	}
	defer rows.Close()

	var metrics []domain.FieldCropMetric
	for rows.Next() {
		// Leer valores raw EXACTAMENTE como en el dashboard
		var rawProjectID, rawFieldID, rawCropID any
		var fieldName, cropName string
		var rawSuperficieHa, rawProduccionTn, rawAreaSembradaHa, rawAreaCosechadaHa any
		var rawRendimientoTnHa, rawPrecioBrutoUsdTn, rawGastoFleteUsdTn, rawGastoComercialUsdTn, rawPrecioNetoUsdTn any
		var rawIngresoNetoUsd, rawIngresoNetoUsdHa, rawCostosLaboresUsd, rawCostosInsumosUsd any
		var rawTotalCostosDirectosUsd, rawCostosDirectosUsdHa, rawMargenBrutoUsd, rawMargenBrutoUsdHa any
		var rawArriendoUsd, rawArriendoUsdHa, rawAdministracionUsd, rawAdministracionUsdHa any
		var rawResultadoOperativoUsd, rawResultadoOperativoUsdHa, rawTotalInvertidoUsd, rawTotalInvertidoUsdHa any
		var rawRentaPct, rawRindeIndiferenciaUsdTn any

		if err := rows.Scan(
			&rawProjectID, &rawFieldID, &fieldName, &rawCropID, &cropName,
			&rawSuperficieHa, &rawProduccionTn, &rawAreaSembradaHa, &rawAreaCosechadaHa,
			&rawRendimientoTnHa, &rawPrecioBrutoUsdTn, &rawGastoFleteUsdTn, &rawGastoComercialUsdTn, &rawPrecioNetoUsdTn,
			&rawIngresoNetoUsd, &rawIngresoNetoUsdHa, &rawCostosLaboresUsd, &rawCostosInsumosUsd,
			&rawTotalCostosDirectosUsd, &rawCostosDirectosUsdHa, &rawMargenBrutoUsd, &rawMargenBrutoUsdHa,
			&rawArriendoUsd, &rawArriendoUsdHa, &rawAdministracionUsd, &rawAdministracionUsdHa,
			&rawResultadoOperativoUsd, &rawResultadoOperativoUsdHa, &rawTotalInvertidoUsd, &rawTotalInvertidoUsdHa,
			&rawRentaPct, &rawRindeIndiferenciaUsdTn,
		); err != nil {
			return nil, fmt.Errorf("error al escanear fila: %w", err)
		}

		// Convertir valores raw EXACTAMENTE como en el dashboard
		projectID := r.convertToInt64(rawProjectID)
		fieldID := r.convertToInt64(rawFieldID)
		cropID := r.convertToInt64(rawCropID)

		metric := domain.FieldCropMetric{
			ProjectID:               projectID,
			FieldID:                 fieldID,
			FieldName:               fieldName,
			CropID:                  cropID,
			CropName:                cropName,
			SuperficieHa:            r.convertToDecimal(rawSuperficieHa),
			ProduccionTn:            r.convertToDecimal(rawProduccionTn),
			AreaSembradaHa:          r.convertToDecimal(rawAreaSembradaHa),
			AreaCosechadaHa:         r.convertToDecimal(rawAreaCosechadaHa),
			RendimientoTnHa:         r.convertToDecimal(rawRendimientoTnHa),
			PrecioBrutoUsdTn:        r.convertToDecimal(rawPrecioBrutoUsdTn),
			GastoFleteUsdTn:         r.convertToDecimal(rawGastoFleteUsdTn),
			GastoComercialUsdTn:     r.convertToDecimal(rawGastoComercialUsdTn),
			PrecioNetoUsdTn:         r.convertToDecimal(rawPrecioNetoUsdTn),
			IngresoNetoUsd:          r.convertToDecimal(rawIngresoNetoUsd),
			IngresoNetoUsdHa:        r.convertToDecimal(rawIngresoNetoUsdHa),
			CostosLaboresUsd:        r.convertToDecimal(rawCostosLaboresUsd),
			CostosInsumosUsd:        r.convertToDecimal(rawCostosInsumosUsd),
			TotalCostosDirectosUsd:  r.convertToDecimal(rawTotalCostosDirectosUsd),
			CostosDirectosUsdHa:     r.convertToDecimal(rawCostosDirectosUsdHa),
			MargenBrutoUsd:          r.convertToDecimal(rawMargenBrutoUsd),
			MargenBrutoUsdHa:        r.convertToDecimal(rawMargenBrutoUsdHa),
			ArriendoUsd:             r.convertToDecimal(rawArriendoUsd),
			ArriendoUsdHa:           r.convertToDecimal(rawArriendoUsdHa),
			AdministracionUsd:       r.convertToDecimal(rawAdministracionUsd),
			AdministracionUsdHa:     r.convertToDecimal(rawAdministracionUsdHa),
			ResultadoOperativoUsd:   r.convertToDecimal(rawResultadoOperativoUsd),
			ResultadoOperativoUsdHa: r.convertToDecimal(rawResultadoOperativoUsdHa),
			TotalInvertidoUsd:       r.convertToDecimal(rawTotalInvertidoUsd),
			TotalInvertidoUsdHa:     r.convertToDecimal(rawTotalInvertidoUsdHa),
			RentaPct:                r.convertToDecimal(rawRentaPct),
			RindeIndiferenciaUsdTn:  r.convertToDecimal(rawRindeIndiferenciaUsdTn),
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
