// Package report proporciona funcionalidades para generar reportes financieros y operativos
package report

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/repository/models"
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

// ===== FUNCIONES PRINCIPALES =====

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

	// Usar GORM para mapear automáticamente desde la vista v3_report_field_crop_metrics
	var modelResults []models.FieldCropMetricModel

	// Construir query base con GORM
	query := r.db.Client().Model(&models.FieldCropMetricModel{}).
		Where("project_id IN ?", projectIDs)

	// Aplicar filtros adicionales
	if filters.FieldID != nil {
		query = query.Where("field_id = ?", *filters.FieldID)
	}

	// Ejecutar query
	if err := query.Find(&modelResults).Error; err != nil {
		return nil, fmt.Errorf("error al obtener métricas: %w", err)
	}

	// Convertir modelos a dominio
	metrics := make([]domain.FieldCropMetric, len(modelResults))
	for i, model := range modelResults {
		metrics[i] = *model.ToDomainFieldCropMetric()
	}

	return metrics, nil
}

// getFieldCropColumns obtiene las columnas (field-crop combinations)
func (r *ReportRepository) getFieldCropColumns(filters domain.ReportFilter) ([]domain.FieldCropColumn, error) {
	var columns []domain.FieldCropColumn

	query := `
		SELECT DISTINCT
			CONCAT(f.id, '-', c.id) as id,
			f.id as field_id,
			f.name as field_name,
			c.id as crop_id,
			c.name as crop_name
		FROM fields f
		JOIN lots l ON f.id = l.field_id AND l.deleted_at IS NULL
		JOIN crops c ON l.current_crop_id = c.id AND c.deleted_at IS NULL
		WHERE f.project_id = ? AND f.deleted_at IS NULL AND l.current_crop_id IS NOT NULL
		ORDER BY f.id, c.id
	`

	err := r.db.Client().Raw(query, *filters.ProjectID).Scan(&columns).Error
	if err != nil {
		return nil, fmt.Errorf("error obteniendo columnas: %w", err)
	}

	return columns, nil
}

// buildReportRows construye las filas del reporte
func (r *ReportRepository) buildReportRows(metrics []domain.FieldCropMetric, columns []domain.FieldCropColumn) []domain.FieldCropRow {
	// Crear mapa de métricas por field_crop_key
	metricMap := make(map[string]domain.FieldCropMetric)
	for _, metric := range metrics {
		key := fmt.Sprintf("%d-%d", metric.FieldID, metric.CropID)
		metricMap[key] = metric
	}

	// Crear mapa de columnas
	columnMap := make(map[string]domain.FieldCropColumn)
	for _, col := range columns {
		columnMap[col.ID] = col
	}

	// Definir filas del reporte
	rows := []domain.FieldCropRow{
		// Información básica
		r.buildRow("surface", "ha", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.SurfaceHa }),
		r.buildRow("production", "tn", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.ProductionTn }),
		r.buildRow("yield", "tn/ha", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.YieldTnHa }),

		// Precios y comercialización
		r.buildRow("freight_cost", "usd/tn", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.FreightCostUsdTn }),
		r.buildRow("commercial_cost", "usd/tn", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.CommercialCostUsdTn }),
		r.buildRow("net_price", "usd/tn", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.NetPriceUsdTn }),
		r.buildRow("gross_price", "usd/tn", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.GrossPriceUsdTn }),
		r.buildRow("net_income", "usd/ha", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.NetIncomeUsdHa }),

		// Costos directos
		// TODO: Cambiar a usar LaborCostsUsdHa cuando esté disponible en la vista v3
		r.buildRow("labors_cost", "usd/ha", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.LaborCostsUsdHa }),
		// TODO: Cambiar a usar SupplyCostsUsdHa cuando esté disponible en la vista v3
		r.buildRow("supplies_cost", "usd/ha", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.SupplyCostsUsdHa }),
		r.buildRow("total_direct_costs", "usd/ha", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.DirectCostsUsdHa }),
		r.buildRow("gross_margin", "usd/ha", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.GrossMarginUsdHa }),

		// Costos adicionales
		r.buildRow("lease", "usd/ha", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.RentUsdHa }),
		r.buildRow("admin", "usd/ha", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.AdministrationUsdHa }),
		r.buildRow("operating_result", "usd/ha", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.OperatingResultUsdHa }),

		// Métricas adicionales
		r.buildRow("total_invested", "usd", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.TotalInvestedUsd }),
		r.buildRow("return_pct", "%", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.ReturnPct }),
		r.buildRow("indifference_yield", "tn/ha", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return decimal.Zero }),
		r.buildRow("indifference_price", "usd/tn", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.IndifferenceYieldUsdTn }),
	}

	// Agregar filas detalladas de supplies y labors
	rows = append(rows, r.buildSupplyDetailRows(columnMap)...)
	rows = append(rows, r.buildLaborDetailRows(columnMap)...)

	return rows
}

// buildRow construye una fila del reporte
func (r *ReportRepository) buildRow(key, unit, valueType string, metricMap map[string]domain.FieldCropMetric, columnMap map[string]domain.FieldCropColumn, getValue func(domain.FieldCropMetric) decimal.Decimal) domain.FieldCropRow {
	values := make(map[string]domain.FieldCropValue)

	for colID := range columnMap {
		if metric, exists := metricMap[colID]; exists {
			values[colID] = domain.FieldCropValue{
				Number: getValue(metric),
			}
		} else {
			values[colID] = domain.FieldCropValue{
				Number: decimal.Zero,
			}
		}
	}

	return domain.FieldCropRow{
		Key:       key,
		Unit:      unit,
		ValueType: valueType,
		Values:    values,
	}
}

// buildSupplyDetailRows construye las filas detalladas de supplies desde v3_report_field_crop_insumos
func (r *ReportRepository) buildSupplyDetailRows(columnMap map[string]domain.FieldCropColumn) []domain.FieldCropRow {
	// Consultar vista v3_report_field_crop_insumos (migración 000130)
	var supplyDetails []models.FieldCropSupplyDetailModel

	// Extraer project_ids de las columnas
	projectIDs := make(map[int64]bool)
	for _, col := range columnMap {
		projectIDs[col.FieldID] = true
	}

	// Construir query
	query := r.db.Client().Model(&models.FieldCropSupplyDetailModel{})
	if len(projectIDs) > 0 {
		fieldIDs := make([]int64, 0, len(projectIDs))
		for fid := range projectIDs {
			fieldIDs = append(fieldIDs, fid)
		}
		query = query.Where("field_id IN ?", fieldIDs)
	}

	// Ejecutar query
	if err := query.Find(&supplyDetails).Error; err != nil {
		// Si hay error, retornar filas vacías
		return r.buildEmptySupplyRows(columnMap)
	}

	// Mapear resultados: field_id-crop_id -> SupplyDetailModel
	supplyMap := make(map[string]models.FieldCropSupplyDetailModel)
	for _, detail := range supplyDetails {
		key := fmt.Sprintf("%d-%d", detail.FieldID, detail.CropID)
		supplyMap[key] = detail
	}

	// Construir filas
	rows := []domain.FieldCropRow{
		r.buildSupplyRow("supply_semilla", "usd/ha", columnMap, supplyMap, func(d models.FieldCropSupplyDetailModel) decimal.Decimal { return d.SemillasUsdHa }),
		r.buildSupplyRow("supply_coadyuvantes", "usd/ha", columnMap, supplyMap, func(d models.FieldCropSupplyDetailModel) decimal.Decimal { return d.CoadyuvantesUsdHa }),
		r.buildSupplyRow("supply_curasemillas", "usd/ha", columnMap, supplyMap, func(d models.FieldCropSupplyDetailModel) decimal.Decimal { return d.CurasemillasUsdHa }),
		r.buildSupplyRow("supply_herbicidas", "usd/ha", columnMap, supplyMap, func(d models.FieldCropSupplyDetailModel) decimal.Decimal { return d.HerbicidasUsdHa }),
		r.buildSupplyRow("supply_insecticidas", "usd/ha", columnMap, supplyMap, func(d models.FieldCropSupplyDetailModel) decimal.Decimal { return d.InsecticidasUsdHa }),
		r.buildSupplyRow("supply_fungicidas", "usd/ha", columnMap, supplyMap, func(d models.FieldCropSupplyDetailModel) decimal.Decimal { return d.FungicidasUsdHa }),
		// Otros Insumos y Fertilizantes no están en la vista, dejarlos en 0
		r.buildSupplyRow("supply_otros_insumos", "usd/ha", columnMap, supplyMap, func(d models.FieldCropSupplyDetailModel) decimal.Decimal { return decimal.Zero }),
		r.buildSupplyRow("supply_fertilizantes", "usd/ha", columnMap, supplyMap, func(d models.FieldCropSupplyDetailModel) decimal.Decimal { return decimal.Zero }),
	}

	return rows
}

// buildLaborDetailRows construye las filas detalladas de labores desde v3_report_field_crop_labores
func (r *ReportRepository) buildLaborDetailRows(columnMap map[string]domain.FieldCropColumn) []domain.FieldCropRow {
	// Consultar vista v3_report_field_crop_labores (migración 000130)
	var laborDetails []models.FieldCropLaborDetailModel

	// Extraer field_ids de las columnas
	fieldIDs := make(map[int64]bool)
	for _, col := range columnMap {
		fieldIDs[col.FieldID] = true
	}

	// Construir query
	query := r.db.Client().Model(&models.FieldCropLaborDetailModel{})
	if len(fieldIDs) > 0 {
		fieldIDList := make([]int64, 0, len(fieldIDs))
		for fid := range fieldIDs {
			fieldIDList = append(fieldIDList, fid)
		}
		query = query.Where("field_id IN ?", fieldIDList)
	}

	// Ejecutar query
	if err := query.Find(&laborDetails).Error; err != nil {
		// Si hay error, retornar filas vacías
		return r.buildEmptyLaborRows(columnMap)
	}

	// Mapear resultados: field_id-crop_id -> LaborDetailModel
	laborMap := make(map[string]models.FieldCropLaborDetailModel)
	for _, detail := range laborDetails {
		key := fmt.Sprintf("%d-%d", detail.FieldID, detail.CropID)
		laborMap[key] = detail
	}

	// Construir filas
	rows := []domain.FieldCropRow{
		r.buildLaborRow("labor_siembra", "usd/ha", columnMap, laborMap, func(d models.FieldCropLaborDetailModel) decimal.Decimal { return d.SiembraUsdHa }),
		r.buildLaborRow("labor_pulverizacion", "usd/ha", columnMap, laborMap, func(d models.FieldCropLaborDetailModel) decimal.Decimal { return d.PulverizacionUsdHa }),
		r.buildLaborRow("labor_riego", "usd/ha", columnMap, laborMap, func(d models.FieldCropLaborDetailModel) decimal.Decimal { return d.RiegoUsdHa }),
		r.buildLaborRow("labor_cosecha", "usd/ha", columnMap, laborMap, func(d models.FieldCropLaborDetailModel) decimal.Decimal { return d.CosechaUsdHa }),
		r.buildLaborRow("labor_otras_labores", "usd/ha", columnMap, laborMap, func(d models.FieldCropLaborDetailModel) decimal.Decimal { return d.OtrasLaboresUsdHa }),
	}

	return rows
}

// buildSupplyRow construye una fila de insumo desde el mapa de detalles
func (r *ReportRepository) buildSupplyRow(
	key string,
	unit string,
	columnMap map[string]domain.FieldCropColumn,
	supplyMap map[string]models.FieldCropSupplyDetailModel,
	extractor func(models.FieldCropSupplyDetailModel) decimal.Decimal,
) domain.FieldCropRow {
	values := make(map[string]domain.FieldCropValue)

	for colID, col := range columnMap {
		// Buscar en el mapa de detalles
		mapKey := fmt.Sprintf("%d-%d", col.FieldID, col.CropID)
		if detail, found := supplyMap[mapKey]; found {
			values[colID] = domain.FieldCropValue{Number: extractor(detail)}
		} else {
			values[colID] = domain.FieldCropValue{Number: decimal.Zero}
		}
	}

	return domain.FieldCropRow{
		Key:       key,
		Unit:      unit,
		ValueType: "number",
		Values:    values,
	}
}

// buildLaborRow construye una fila de labor desde el mapa de detalles
func (r *ReportRepository) buildLaborRow(
	key string,
	unit string,
	columnMap map[string]domain.FieldCropColumn,
	laborMap map[string]models.FieldCropLaborDetailModel,
	extractor func(models.FieldCropLaborDetailModel) decimal.Decimal,
) domain.FieldCropRow {
	values := make(map[string]domain.FieldCropValue)

	for colID, col := range columnMap {
		// Buscar en el mapa de detalles
		mapKey := fmt.Sprintf("%d-%d", col.FieldID, col.CropID)
		if detail, found := laborMap[mapKey]; found {
			values[colID] = domain.FieldCropValue{Number: extractor(detail)}
		} else {
			values[colID] = domain.FieldCropValue{Number: decimal.Zero}
		}
	}

	return domain.FieldCropRow{
		Key:       key,
		Unit:      unit,
		ValueType: "number",
		Values:    values,
	}
}

// buildEmptySupplyRows construye filas vacías de supplies
func (r *ReportRepository) buildEmptySupplyRows(columnMap map[string]domain.FieldCropColumn) []domain.FieldCropRow {
	// Cargar categorías de insumos desde la base de datos
	supplyCategories, err := r.getSupplyCategories()
	if err != nil {
		// Fallback a categorías por defecto si hay error (usando solo categorías de 000013)
		supplyCategories = map[string]string{
			"supply_semilla":       "Semilla",
			"supply_coadyuvantes":  "Coadyuvantes",
			"supply_curasemillas":  "Curasemillas",
			"supply_herbicidas":    "Herbicidas",
			"supply_insecticidas":  "Insecticidas",
			"supply_fungicidas":    "Fungicidas",
			"supply_otros_insumos": "Otros Insumos",
			"supply_fertilizantes": "Fertilizantes",
		}
	}

	var rows []domain.FieldCropRow
	for key := range supplyCategories {
		values := make(map[string]domain.FieldCropValue)
		for colID := range columnMap {
			values[colID] = domain.FieldCropValue{Number: decimal.Zero}
		}

		rows = append(rows, domain.FieldCropRow{
			Key:       key,
			Unit:      "usd/ha",
			ValueType: "number",
			Values:    values,
		})
	}

	return rows
}

// buildEmptyLaborRows construye filas vacías de labores
func (r *ReportRepository) buildEmptyLaborRows(columnMap map[string]domain.FieldCropColumn) []domain.FieldCropRow {
	// Cargar categorías de labores desde la base de datos
	laborCategories, err := r.getLaborCategories()
	if err != nil {
		// Fallback a categorías por defecto si hay error (usando solo categorías de 000013)
		laborCategories = map[string]string{
			"labor_siembra":       "Siembra",
			"labor_pulverizacion": "Pulverización",
			"labor_otras_labores": "Otras Labores",
			"labor_riego":         "Riego",
			"labor_cosecha":       "Cosecha",
		}
	}

	var rows []domain.FieldCropRow
	for key := range laborCategories {
		values := make(map[string]domain.FieldCropValue)
		for colID := range columnMap {
			values[colID] = domain.FieldCropValue{Number: decimal.Zero}
		}

		rows = append(rows, domain.FieldCropRow{
			Key:       key,
			Unit:      "usd/ha",
			ValueType: "number",
			Values:    values,
		})
	}

	return rows
}

// getSupplyCategories obtiene las categorías de insumos desde la base de datos
func (r *ReportRepository) getSupplyCategories() (map[string]string, error) {
	query := `
		SELECT c.id, c.name, c.type_id
		FROM categories c
		WHERE c.deleted_at IS NULL 
		AND c.type_id IN (1, 2, 3)
		ORDER BY c.type_id, c.name
	`

	rows, err := r.db.Client().Raw(query).Rows()
	if err != nil {
		return nil, fmt.Errorf("error querying supply categories: %w", err)
	}
	defer rows.Close()

	categories := make(map[string]string)
	for rows.Next() {
		var id int64
		var name string
		var typeID int64
		if err := rows.Scan(&id, &name, &typeID); err != nil {
			continue
		}

		// Crear clave basada en el tipo y nombre (usando solo categorías de 000013)
		var key string
		switch typeID {
		case 1: // Semilla
			key = "supply_semilla"
		case 2: // Agroquímicos
			switch name {
			case "Coadyuvantes":
				key = "supply_coadyuvantes"
			case "Curasemillas":
				key = "supply_curasemillas"
			case "Herbicidas":
				key = "supply_herbicidas"
			case "Insecticidas":
				key = "supply_insecticidas"
			case "Fungicidas":
				key = "supply_fungicidas"
			case "Otros Insumos":
				key = "supply_otros_insumos"
			default:
				key = fmt.Sprintf("supply_%d", id) // Fallback usando ID
			}
		case 3: // Fertilizantes
			key = "supply_fertilizantes"
		default:
			key = fmt.Sprintf("supply_%d", id) // Fallback usando ID
		}

		categories[key] = name
	}

	return categories, nil
}

// getLaborCategories obtiene las categorías de labores desde la base de datos
func (r *ReportRepository) getLaborCategories() (map[string]string, error) {
	query := `
		SELECT c.id, c.name, c.type_id
		FROM categories c
		WHERE c.deleted_at IS NULL 
		AND c.type_id = 4
		ORDER BY c.name
	`

	rows, err := r.db.Client().Raw(query).Rows()
	if err != nil {
		return nil, fmt.Errorf("error querying labor categories: %w", err)
	}
	defer rows.Close()

	categories := make(map[string]string)
	for rows.Next() {
		var id int64
		var name string
		var typeID int64
		if err := rows.Scan(&id, &name, &typeID); err != nil {
			continue
		}

		// Crear clave basada en el nombre (usando solo categorías de 000013)
		var key string
		switch name {
		case "Siembra":
			key = "labor_siembra"
		case "Pulverización":
			key = "labor_pulverizacion"
		case "Otras Labores":
			key = "labor_otras_labores"
		case "Riego":
			key = "labor_riego"
		case "Cosecha":
			key = "labor_cosecha"
		default:
			key = fmt.Sprintf("labor_%d", id) // Fallback usando ID
		}

		categories[key] = name
	}

	return categories, nil
}

// ===== FUNCIONES AUXILIARES =====

// BuildFieldCrop construye la tabla completa del reporte field-crop
func (r *ReportRepository) BuildFieldCrop(filters domain.ReportFilter) (*domain.FieldCrop, error) {
	// Obtener información del proyecto
	projectInfo, err := r.getProjectInfo(filters)
	if err != nil {
		return nil, fmt.Errorf("error getting project information: %w", err)
	}

	// Obtener columnas (field-crop combinations)
	columns, err := r.getFieldCropColumns(filters)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo columnas: %w", err)
	}

	// Obtener métricas básicas
	metrics, err := r.GetFieldCropMetrics(filters)
	if err != nil {
		return nil, fmt.Errorf("error getting metrics: %w", err)
	}

	// Construir filas del reporte
	rows := r.buildReportRows(metrics, columns)

	// Asignar valores directamente (no son opcionales en el frontend)
	return &domain.FieldCrop{
		ProjectID:    projectInfo.ProjectID,
		ProjectName:  projectInfo.ProjectName,
		CustomerID:   projectInfo.CustomerID,
		CustomerName: projectInfo.CustomerName,
		CampaignID:   projectInfo.CampaignID,
		CampaignName: projectInfo.CampaignName,
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
		FROM v3_investor_contribution_data_view
		WHERE project_id IN (%s)
	`, strings.Join(placeholders, ","))

	// Usar vista v3

	args := make([]any, len(projectIDs))
	for i, id := range projectIDs {
		args[i] = id
	}

	var results []models.InvestorContributionDataModel

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

	// Usar el mapper del modelo para convertir al domain
	report, err := firstResult.ToDomainInvestorContributionReport()
	if err != nil {
		return nil, fmt.Errorf("error convirtiendo modelo a domain: %w", err)
	}

	return report, nil
}

// ===== REPORTE DE RESUMEN DE RESULTADOS =====

// GetSummaryResults obtiene el resumen de resultados por cultivo
func (r *ReportRepository) GetSummaryResults(filters domain.SummaryResultsFilter) ([]domain.SummaryResults, error) {
	// Obtener project IDs relacionados con los filtros
	projectIDs, err := r.getRelatedProjectIDs(domain.ReportFilter{
		ProjectID:  filters.ProjectID,
		CustomerID: filters.CustomerID,
		CampaignID: filters.CampaignID,
		FieldID:    filters.FieldID,
	})
	if err != nil {
		return nil, fmt.Errorf("error obteniendo proyectos relacionados: %w", err)
	}

	if len(projectIDs) == 0 {
		return []domain.SummaryResults{}, nil
	}

	// Usar IN en lugar de ANY para evitar problemas de mapeo de arrays
	placeholders := make([]string, len(projectIDs))
	for i := range projectIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	// Construir query con filtros
	query := fmt.Sprintf(`
		SELECT 
			project_id,
			current_crop_id,
			crop_name,
			surface_ha,
			net_income_usd,
			direct_costs_usd,
			rent_usd,
			structure_usd,
			total_invested_usd,
			operating_result_usd,
			crop_return_pct,
			total_surface_ha,
			total_net_income_usd,
			total_direct_costs_usd,
			total_rent_usd,
			total_structure_usd,
			total_invested_project_usd,
			total_operating_result_usd,
			project_return_pct
		FROM v3_report_summary_results_view 
		WHERE project_id IN (%s)
		ORDER BY project_id, current_crop_id
	`, strings.Join(placeholders, ","))

	// Convertir projectIDs a []interface{} para la query
	args := make([]interface{}, len(projectIDs))
	for i, id := range projectIDs {
		args[i] = id
	}

	// Ejecutar query
	var models []models.SummaryResultsModel
	if err := r.db.Client().Raw(query, args...).Scan(&models).Error; err != nil {
		return nil, fmt.Errorf("error ejecutando query de resumen de resultados: %w", err)
	}

	// Convertir a dominio
	results := make([]domain.SummaryResults, len(models))
	for i, model := range models {
		results[i] = *model.ToDomainSummaryResults()
	}

	return results, nil
}

// ===== FUNCIONES AUXILIARES =====

// convertToInt64 convierte valores raw a int64
func (r *ReportRepository) convertToInt64(raw any) int64 {
	if raw == nil {
		return 0
	}

	switch v := raw.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	case string:
		if val, err := strconv.ParseInt(v, 10, 64); err == nil {
			return val
		}
	}
	return 0
}

// convertToDecimal convierte valores raw a decimal.Decimal
func (r *ReportRepository) convertToDecimal(raw any) decimal.Decimal {
	if raw == nil {
		return decimal.Zero
	}

	switch v := raw.(type) {
	case string:
		if dec, err := decimal.NewFromString(v); err == nil {
			return dec
		}
	case float64:
		return decimal.NewFromFloat(v)
	case int64:
		return decimal.NewFromInt(v)
	case int:
		return decimal.NewFromInt(int64(v))
	}
	return decimal.Zero
}

// getRelatedProjectIDs encuentra los IDs de proyectos relacionados con los filtros
func (r *ReportRepository) getRelatedProjectIDs(filter domain.ReportFilter) ([]int64, error) {
	// Si tenemos ProjectID directamente, usarlo sin buscar
	if filter.ProjectID != nil {
		return []int64{*filter.ProjectID}, nil
	}

	// Si no hay filtros, devolver todos los proyectos
	if filter.CustomerID == nil && filter.CampaignID == nil && filter.FieldID == nil {
		query := `SELECT DISTINCT p.id FROM projects p WHERE p.deleted_at IS NULL`
		var projectIDs []int64
		if err := r.db.Client().Raw(query).Scan(&projectIDs).Error; err != nil {
			return nil, fmt.Errorf("error al obtener todos los proyectos: %w", err)
		}
		return projectIDs, nil
	}

	query := `
		SELECT DISTINCT p.id
		FROM projects p
		WHERE p.deleted_at IS NULL
	`

	var args []any
	argIndex := 1

	// Aplicar filtros si están presentes
	if filter.CustomerID != nil {
		query += " AND p.customer_id = $" + fmt.Sprintf("%d", argIndex)
		args = append(args, *filter.CustomerID)
		argIndex++
	}
	if filter.CampaignID != nil {
		query += " AND p.campaign_id = $" + fmt.Sprintf("%d", argIndex)
		args = append(args, *filter.CampaignID)
		argIndex++
	}
	if filter.FieldID != nil {
		query += " AND EXISTS (SELECT 1 FROM fields f WHERE f.id = $" + fmt.Sprintf("%d", argIndex) + " AND f.project_id = p.id)"
		args = append(args, *filter.FieldID)
	}

	var projectIDs []int64
	if err := r.db.Client().Raw(query, args...).Scan(&projectIDs).Error; err != nil {
		return nil, fmt.Errorf("error al obtener proyectos relacionados: %w", err)
	}

	return projectIDs, nil
}

// getProjectInfo obtiene la información básica del proyecto
func (r *ReportRepository) getProjectInfo(filters domain.ReportFilter) (*domain.ProjectInfo, error) {
	// Obtener project IDs relacionados con los filtros
	projectIDs, err := r.getRelatedProjectIDs(filters)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo proyectos relacionados: %w", err)
	}

	if len(projectIDs) == 0 {
		return &domain.ProjectInfo{}, nil
	}

	// Usar el primer proyecto para la información básica
	projectID := projectIDs[0]

	var projectInfo domain.ProjectInfo

	query := `
		SELECT 
			p.id as project_id,
			p.name as project_name,
			c.id as customer_id,
			c.name as customer_name,
			camp.id as campaign_id,
			camp.name as campaign_name
		FROM projects p
		LEFT JOIN customers c ON p.customer_id = c.id
		LEFT JOIN campaigns camp ON p.campaign_id = camp.id
		WHERE p.id = ? AND p.deleted_at IS NULL
	`

	err = r.db.Client().Raw(query, projectID).Scan(&projectInfo).Error
	if err != nil {
		return nil, fmt.Errorf("error getting project information: %w", err)
	}

	return &projectInfo, nil
}
