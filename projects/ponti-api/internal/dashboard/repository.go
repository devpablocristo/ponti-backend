package dashboard

import (
	"context"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
	"gorm.io/gorm"
)

type GormEnginePort interface {
	Client() *gorm.DB
}

type Repository struct {
	db     GormEnginePort
	mapper *models.DashboardModelMapper
}

func NewRepository(db GormEnginePort) *Repository {
	return &Repository{
		db:     db,
		mapper: models.NewDashboardModelMapper(),
	}
}

func (r *Repository) GetDashboard(ctx context.Context, filter domain.DashboardFilter) (*domain.DashboardData, error) {
	// Construir la consulta SQL base usando todos los campos de la vista
	query := `
		SELECT 
			-- Métricas de siembra
			COALESCE(SUM(sowing_hectares), 0) as sowing_hectares,
			COALESCE(SUM(sowing_total_hectares), 0) as sowing_total_hectares,
			-- Tomar el porcentaje directamente sin sumar
			COALESCE(MAX(sowing_progress_percent), 0) as sowing_progress_percent,
			
			-- Métricas de cosecha
			COALESCE(SUM(harvest_hectares), 0) as harvest_hectares,
			COALESCE(SUM(harvest_total_hectares), 0) as harvest_total_hectares,
			-- Tomar el porcentaje directamente sin sumar
			COALESCE(MAX(harvest_progress_percent), 0) as harvest_progress_percent,
			
			-- Métricas de costos
			COALESCE(SUM(executed_costs_usd), 0) as costs_executed_usd,
			COALESCE(SUM(budget_total_usd), 0) as costs_budget_usd,
			-- Tomar el porcentaje directamente sin sumar
			COALESCE(MAX(costs_progress_pct), 0) as costs_progress_pct,
			COALESCE(SUM(executed_labors_usd), 0) as executed_labors_usd,
			COALESCE(SUM(executed_supplies_usd), 0) as executed_supplies_usd,
			COALESCE(SUM(budget_cost_usd), 0) as budget_cost_usd,
			
			-- Resultado operativo
			COALESCE(SUM(income_usd), 0) as operating_income_usd,
			COALESCE(SUM(operating_result_total_costs_usd), 0) as operating_total_costs_usd,
			COALESCE(SUM(operating_result_usd), 0) as operating_result_usd,
			-- Tomar el porcentaje directamente sin sumar
			COALESCE(MAX(operating_result_pct), 0) as operating_result_pct,
			
			-- Balance de gestión - Semilla
			COALESCE(SUM(semilla_ejecutados_usd), 0) as semilla_ejecutados_usd,
			COALESCE(SUM(semilla_invertidos_usd), 0) as semilla_invertidos_usd,
			COALESCE(SUM(semilla_stock_usd), 0) as semilla_stock_usd,
			
			-- Balance de gestión - Insumos
			COALESCE(SUM(insumos_ejecutados_usd), 0) as insumos_ejecutados_usd,
			COALESCE(SUM(insumos_invertidos_usd), 0) as insumos_invertidos_usd,
			COALESCE(SUM(insumos_stock_usd), 0) as insumos_stock_usd,
			
			-- Balance de gestión - Labores
			COALESCE(SUM(labores_ejecutados_usd), 0) as labores_ejecutados_usd,
			COALESCE(SUM(labores_invertidos_usd), 0) as labores_invertidos_usd,
			COALESCE(SUM(labores_stock_usd), 0) as labores_stock_usd,
			
			-- Balance de gestión - Arriendo
			COALESCE(SUM(arriendo_invertidos_usd), 0) as arriendo_invertidos_usd,
			
			-- Balance de gestión - Estructura
			COALESCE(SUM(estructura_invertidos_usd), 0) as estructura_invertidos_usd,
			
			-- Totales del Balance de Gestión
			COALESCE(SUM(costos_directos_ejecutados_usd), 0) as costos_directos_ejecutados_usd,
			COALESCE(SUM(costos_directos_invertidos_usd), 0) as costos_directos_invertidos_usd,
			COALESCE(SUM(costos_directos_stock_usd), 0) as costos_directos_stock_usd
		FROM dashboard_view 
		WHERE row_kind = 'metric'
		AND field_id IS NULL
	`

	// Aplicar filtros si están presentes
	args := []interface{}{}
	if len(filter.CustomerIDs) > 0 {
		query += " AND customer_id = ANY($1)"
		args = append(args, filter.CustomerIDs)
	}
	if len(filter.ProjectIDs) > 0 {
		query += " AND project_id = ANY($2)"
		args = append(args, filter.ProjectIDs)
	}
	if len(filter.CampaignIDs) > 0 {
		query += " AND campaign_id = ANY($3)"
		args = append(args, filter.CampaignIDs)
	}
	if len(filter.FieldIDs) > 0 {
		query += " AND field_id = ANY($4)"
		args = append(args, filter.FieldIDs)
	}

	// Ejecutar la consulta
	var result models.DashboardDataModel

	err := r.db.Client().WithContext(ctx).Raw(query, args...).Scan(&result).Error
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get dashboard data", err)
	}

	// Obtener datos de cultivos
	crops, err := r.getCropIncidence(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Obtener datos de inversores
	investors, err := r.getInvestorContributions(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Obtener indicadores operativos
	operational, err := r.getOperationalIndicators(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Usar el mapper para convertir a dominio
	return r.mapper.DashboardDataToDomain(&result, crops, investors, operational), nil
}

func (r *Repository) getCropIncidence(ctx context.Context, filter domain.DashboardFilter) ([]models.CropIncidenceModel, error) {
	query := `
		SELECT DISTINCT
			crop_name,
			crop_hectares,
			incidence_pct,
			cost_per_ha_usd
		FROM dashboard_view 
		WHERE row_kind = 'metric'
		AND crop_name IS NOT NULL
		AND crop_name != ''
	`

	args := []interface{}{}
	if len(filter.CustomerIDs) > 0 {
		query += " AND customer_id = ANY($1)"
		args = append(args, filter.CustomerIDs)
	}
	if len(filter.ProjectIDs) > 0 {
		query += " AND project_id = ANY($2)"
		args = append(args, filter.ProjectIDs)
	}
	if len(filter.CampaignIDs) > 0 {
		query += " AND campaign_id = ANY($3)"
		args = append(args, filter.CampaignIDs)
	}
	if len(filter.FieldIDs) > 0 {
		query += " AND field_id = ANY($4)"
		args = append(args, filter.FieldIDs)
	}

	query += " ORDER BY crop_name"

	var crops []models.CropIncidenceModel

	err := r.db.Client().WithContext(ctx).Raw(query, args...).Scan(&crops).Error
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get crop incidence data", err)
	}

	return crops, nil
}

func (r *Repository) getInvestorContributions(ctx context.Context, filter domain.DashboardFilter) ([]models.InvestorContributionModel, error) {
	query := `
		SELECT DISTINCT
			investor_id,
			investor_name,
			investor_percentage_pct
		FROM dashboard_view 
		WHERE row_kind = 'metric'
		AND investor_id IS NOT NULL
		AND investor_id > 0
		AND field_id IS NULL
	`

	args := []interface{}{}
	if len(filter.CustomerIDs) > 0 {
		query += " AND customer_id = ANY($1)"
		args = append(args, filter.CustomerIDs)
	}
	if len(filter.ProjectIDs) > 0 {
		query += " AND project_id = ANY($2)"
		args = append(args, filter.ProjectIDs)
	}
	if len(filter.CampaignIDs) > 0 {
		query += " AND campaign_id = ANY($3)"
		args = append(args, filter.CampaignIDs)
	}

	query += " ORDER BY investor_id"

	var investors []models.InvestorContributionModel

	err := r.db.Client().WithContext(ctx).Raw(query, args...).Scan(&investors).Error
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get investor contributions data", err)
	}

	return investors, nil
}

func (r *Repository) getOperationalIndicators(ctx context.Context, filter domain.DashboardFilter) (*models.OperationalIndicatorModel, error) {
	// Consultar indicadores operativos reales desde la vista
	query := `
		SELECT 
			primera_orden_fecha,
			primera_orden_id,
			ultima_orden_fecha,
			ultima_orden_id,
			arqueo_stock_fecha,
			cierre_campana_fecha
		FROM dashboard_view 
		WHERE row_kind = 'metric'
	`

	args := []interface{}{}
	if len(filter.CustomerIDs) > 0 {
		query += " AND customer_id = ANY($1)"
		args = append(args, filter.CustomerIDs)
	}
	if len(filter.ProjectIDs) > 0 {
		query += " AND project_id = ANY($2)"
		args = append(args, filter.ProjectIDs)
	}
	if len(filter.CampaignIDs) > 0 {
		query += " AND campaign_id = ANY($3)"
		args = append(args, filter.CampaignIDs)
	}
	if len(filter.FieldIDs) > 0 {
		query += " AND field_id = ANY($4)"
		args = append(args, filter.FieldIDs)
	}

	query += " ORDER BY customer_id, project_id, campaign_id, field_id LIMIT 1"

	var result models.OperationalIndicatorModel

	err := r.db.Client().WithContext(ctx).Raw(query, args...).Scan(&result).Error
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get operational indicators data", err)
	}

	return &result, nil
}
