package report

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
	"github.com/shopspring/decimal"
)

// ===== FUNCIONES AUXILIARES MOVIDAS DESDE REPOSITORY.GO =====

// convertToInt64 convierte valores raw a int64 como en el dashboard
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

// convertToDecimal convierte valores raw a decimal.Decimal como en el dashboard
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
// Esta función es utilizada por todos los métodos del repositorio ya que todos trabajan a nivel de proyecto
// NOTA: Si ya tenemos ProjectID, no se debe llamar a esta función
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

	// Aplicar filtros si están presentes (excluyendo ProjectID que se maneja directamente)
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
		argIndex++
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
	// En el futuro se podría implementar una lógica de agregación más sofisticada
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
		return nil, fmt.Errorf("error obteniendo información del proyecto: %w", err)
	}

	return &projectInfo, nil
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
		JOIN crops c ON (l.current_crop_id = c.id OR l.previous_crop_id = c.id) AND c.deleted_at IS NULL
		WHERE f.project_id = ? AND f.deleted_at IS NULL
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
		r.buildRow("surface", "ha", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.SuperficieHa }),
		r.buildRow("production", "tn", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.ProduccionTn }),
		r.buildRow("yield", "tn/ha", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.RendimientoTnHa }),

		// Precios y comercialización
		r.buildRow("freight_cost", "usd/tn", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.GastoFleteUsdTn }),
		r.buildRow("commercial_cost", "usd/tn", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.GastoComercialUsdTn }),
		r.buildRow("net_price", "usd/tn", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.PrecioNetoUsdTn }),
		r.buildRow("gross_price", "usd/tn", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.PrecioBrutoUsdTn }),
		r.buildRow("net_income", "usd/ha", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.IngresoNetoUsdHa }),

		// Costos directos
		r.buildRow("labors_cost", "usd/ha", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.CostosLaboresUsd }),
		r.buildRow("supplies_cost", "usd/ha", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.CostosInsumosUsd }),
		r.buildRow("total_direct_costs", "usd/ha", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.CostosDirectosUsdHa }),
		r.buildRow("gross_margin", "usd/ha", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.MargenBrutoUsdHa }),

		// Costos adicionales
		r.buildRow("lease", "usd/ha", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.ArriendoUsdHa }),
		r.buildRow("admin", "usd/ha", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.AdministracionUsdHa }),
		r.buildRow("operating_result", "usd/ha", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.ResultadoOperativoUsdHa }),

		// Métricas adicionales
		r.buildRow("total_invested", "usd", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.TotalInvertidoUsd }),
		r.buildRow("return_pct", "%", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.RentaPct }),
		r.buildRow("indifference_yield", "tn/ha", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return decimal.Zero }),
		r.buildRow("indifference_price", "usd/tn", "number", metricMap, columnMap, func(m domain.FieldCropMetric) decimal.Decimal { return m.RindeIndiferenciaUsdTn }),
	}

	// Agregar filas detalladas de supplies y labors
	rows = append(rows, r.buildSupplyDetailRows(metricMap, columnMap)...)
	rows = append(rows, r.buildLaborDetailRows(metricMap, columnMap)...)

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

// buildSupplyDetailRows construye las filas detalladas de supplies
func (r *ReportRepository) buildSupplyDetailRows(metricMap map[string]domain.FieldCropMetric, columnMap map[string]domain.FieldCropColumn) []domain.FieldCropRow {
	// Retornar filas con valores en cero ya que no hay datos detallados disponibles
	return r.buildEmptySupplyRows(columnMap)

}

// buildLaborDetailRows construye las filas detalladas de labores
func (r *ReportRepository) buildLaborDetailRows(metricMap map[string]domain.FieldCropMetric, columnMap map[string]domain.FieldCropColumn) []domain.FieldCropRow {
	// Retornar filas con valores en cero ya que no hay datos detallados disponibles
	return r.buildEmptyLaborRows(columnMap)

}

// buildEmptySupplyRows construye filas vacías de supplies
func (r *ReportRepository) buildEmptySupplyRows(columnMap map[string]domain.FieldCropColumn) []domain.FieldCropRow {
	// Cargar categorías de insumos desde la base de datos
	supplyCategories, err := r.getSupplyCategories()
	if err != nil {
		// Fallback a categorías por defecto si hay error
		supplyCategories = map[string]string{
			"supply_semillas":      "Semillas",
			"supply_curasemillas":  "Curásemillas",
			"supply_herbicidas":    "Herbicidas",
			"supply_insecticidas":  "Insecticidas",
			"supply_coadyuvantes":  "Coadyuvantes",
			"supply_fertilizantes": "Fertilizantes",
			"supply_fungicidas":    "Fungicidas",
			"supply_otros":         "Otros insumos",
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
		// Fallback a categorías por defecto si hay error
		laborCategories = map[string]string{
			"labor_siembra":       "Siembra",
			"labor_pulverizacion": "Pulverización",
			"labor_otras":         "Otras labores",
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

// parseContributionsFromJSON parsea el JSON de contributions
func (r *ReportRepository) parseContributionsFromJSON(jsonData string) ([]domain.ContributionCategory, error) {
	var contributions []domain.ContributionCategory
	err := json.Unmarshal([]byte(jsonData), &contributions)
	if err != nil {
		return nil, fmt.Errorf("error parseando JSON de contributions: %w", err)
	}
	return contributions, nil
}

// parseComparisonFromJSON parsea el JSON de comparison
func (r *ReportRepository) parseComparisonFromJSON(jsonData string) ([]domain.InvestorContributionComparison, error) {
	var comparison []domain.InvestorContributionComparison
	err := json.Unmarshal([]byte(jsonData), &comparison)
	if err != nil {
		return nil, fmt.Errorf("error parseando JSON de comparison: %w", err)
	}
	return comparison, nil
}

// parseHarvestFromJSON parsea el JSON de harvest
func (r *ReportRepository) parseHarvestFromJSON(jsonData string) (domain.HarvestSettlement, error) {
	var harvest domain.HarvestSettlement
	err := json.Unmarshal([]byte(jsonData), &harvest)
	if err != nil {
		return domain.HarvestSettlement{}, fmt.Errorf("error parseando JSON de harvest: %w", err)
	}
	return harvest, nil
}

// getSupplyCategories obtiene las categorías de insumos desde la base de datos
func (r *ReportRepository) getSupplyCategories() (map[string]string, error) {
	query := `
		SELECT c.id, c.name, t.name as type_name
		FROM categories c
		JOIN types t ON c.type_id = t.id
		WHERE c.deleted_at IS NULL 
		AND t.name IN ('Semilla', 'Agroquímicos', 'Fertilizantes')
		ORDER BY t.name, c.name
	`

	rows, err := r.db.Client().Raw(query).Rows()
	if err != nil {
		return nil, fmt.Errorf("error consultando categorías de insumos: %w", err)
	}
	defer rows.Close()

	categories := make(map[string]string)
	for rows.Next() {
		var id int64
		var name, typeName string
		if err := rows.Scan(&id, &name, &typeName); err != nil {
			continue
		}

		// Crear clave basada en el tipo y nombre
		var key string
		switch typeName {
		case "Semilla":
			key = "supply_semillas"
		case "Agroquímicos":
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
				key = "supply_otros"
			default:
				key = fmt.Sprintf("supply_%d", id) // Fallback usando ID
			}
		case "Fertilizantes":
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
		SELECT c.id, c.name, t.name as type_name
		FROM categories c
		JOIN types t ON c.type_id = t.id
		WHERE c.deleted_at IS NULL 
		AND t.name = 'Labores'
		ORDER BY c.name
	`

	rows, err := r.db.Client().Raw(query).Rows()
	if err != nil {
		return nil, fmt.Errorf("error consultando categorías de labores: %w", err)
	}
	defer rows.Close()

	categories := make(map[string]string)
	for rows.Next() {
		var id int64
		var name, typeName string
		if err := rows.Scan(&id, &name, &typeName); err != nil {
			continue
		}

		// Crear clave basada en el nombre
		var key string
		switch name {
		case "Siembra":
			key = "labor_siembra"
		case "Pulverización":
			key = "labor_pulverizacion"
		case "Otras Labores":
			key = "labor_otras"
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
