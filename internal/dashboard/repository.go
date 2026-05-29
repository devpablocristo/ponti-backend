package dashboard

import (
	"context"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/devpablocristo/core/errors/go/domainerr"
	sharedfilters "github.com/devpablocristo/ponti-backend/internal/shared/filters"

	models "github.com/devpablocristo/ponti-backend/internal/dashboard/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/dashboard/usecases/domain"
	db "github.com/devpablocristo/ponti-backend/internal/shared/db"
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
	// Determinar projectIDs una sola vez
	projectIDs, err := r.resolveProjectIDs(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Si no hay proyectos, retornar datos vacíos
	if len(projectIDs) == 0 {
		return r.createEmptyDashboardData(), nil
	}

	// Obtener todas las métricas principales en una sola query
	metricsData, err := r.getMetrics(ctx, projectIDs, filter)
	if err != nil {
		return nil, err
	}

	// Obtener datos 1:N en paralelo (contribuciones, cultivos)
	contributionsData, err := r.getContributionsProgress(ctx, projectIDs, filter)
	if err != nil {
		return nil, err
	}

	managementBalanceData, err := r.getManagementBalance(ctx, projectIDs, filter)
	if err != nil {
		return nil, err
	}

	cropIncidenceData, err := r.getCropIncidence(ctx, projectIDs, filter)
	if err != nil {
		return nil, err
	}

	operationalIndicatorsData, err := r.getOperationalIndicators(ctx, projectIDs, filter)
	if err != nil {
		return nil, err
	}

	// Construir modelo de datos consolidado
	tempData := &models.DashboardDataModel{
		SowingHectares:         metricsData.SowingHectares,
		SowingTotalHectares:    metricsData.SowingTotalHectares,
		SowingProgressPercent:  metricsData.SowingProgressPct,
		CostsExecutedUSD:       metricsData.ExecutedCostsUSD,
		CostsBudgetUSD:         metricsData.BudgetCostUSD,
		CostsProgressPct:       metricsData.CostsProgressPct,
		HarvestHectares:        metricsData.HarvestHectares,
		HarvestTotalHectares:   metricsData.HarvestTotalHectares,
		HarvestProgressPercent: metricsData.HarvestProgressPct,
		OperatingResultUSD:     metricsData.OperatingResultUSD,
		OperatingTotalCostsUSD: metricsData.OperatingResultTotalCostsUSD,
		OperatingResultPct:     metricsData.OperatingResultPct,
	}

	// Retornar datos mapeados a dominio (contributionsData ya contiene contributions_progress_pct)
	return r.mapper.DashboardDataToDomain(tempData, cropIncidenceData, contributionsData, managementBalanceData, operationalIndicatorsData), nil
}

func (r *Repository) applyWorkspaceFilters(
	q *gorm.DB,
	filter domain.DashboardFilter,
	cols sharedfilters.WorkspaceFilterColumns,
) *gorm.DB {
	return sharedfilters.ApplyWorkspaceFilters(q, sharedfilters.WorkspaceFilter{
		CustomerID: filter.CustomerID,
		ProjectID:  filter.ProjectID,
		CampaignID: filter.CampaignID,
		FieldID:    filter.FieldID,
	}, cols)
}

func (r *Repository) metricsView(filter domain.DashboardFilter) string {
	if filter.FieldID != nil {
		return db.DashboardView("metrics_field")
	}
	return db.DashboardView("metrics")
}

func (r *Repository) managementBalanceView(filter domain.DashboardFilter) string {
	if filter.FieldID != nil {
		return db.DashboardView("management_balance_field")
	}
	return db.DashboardView("management_balance")
}

func (r *Repository) cropIncidenceView(filter domain.DashboardFilter) string {
	if filter.FieldID != nil {
		return db.DashboardView("crop_incidence_field")
	}
	return db.DashboardView("crop_incidence")
}

func (r *Repository) operationalIndicatorsView(filter domain.DashboardFilter) string {
	if filter.FieldID != nil {
		return db.DashboardView("operational_indicators_field")
	}
	return db.DashboardView("operational_indicators")
}

// resolveProjectIDs determina los IDs de proyectos a consultar basándose en los filtros
func (r *Repository) resolveProjectIDs(ctx context.Context, filter domain.DashboardFilter) ([]int64, error) {
	// Si tenemos ProjectID directamente, usarlo
	if filter.ProjectID != nil {
		return []int64{*filter.ProjectID}, nil
	}

	// Buscar proyectos relacionados con los otros filtros
	return r.getRelatedProjectIDs(ctx, filter)
}

// getRelatedProjectIDs encuentra los IDs de proyectos relacionados con los filtros
func (r *Repository) getRelatedProjectIDs(ctx context.Context, filter domain.DashboardFilter) ([]int64, error) {
	query := r.db.Client().WithContext(ctx).
		Table("projects p").
		Select("DISTINCT p.id").
		Where("p.deleted_at IS NULL")

	// Aplicar filtros dinámicamente
	if filter.CustomerID != nil {
		query = query.Where("p.customer_id = ?", *filter.CustomerID)
	}
	if filter.CampaignID != nil {
		query = query.Where("p.campaign_id = ?", *filter.CampaignID)
	}
	if filter.FieldID != nil {
		query = query.Where("EXISTS (SELECT 1 FROM fields f WHERE f.id = ? AND f.project_id = p.id AND f.deleted_at IS NULL)", *filter.FieldID)
	}

	var projectIDs []int64
	if err := query.Pluck("p.id", &projectIDs).Error; err != nil {
		return nil, domainerr.Internal("failed to get related project IDs")
	}

	return projectIDs, nil
}

// createEmptyDashboardData retorna una estructura de dashboard vacía con valores por defecto
func (r *Repository) createEmptyDashboardData() *domain.DashboardData {
	zero := decimal.Zero
	tempData := &models.DashboardDataModel{
		SowingHectares:         zero,
		SowingTotalHectares:    zero,
		SowingProgressPercent:  zero,
		CostsExecutedUSD:       zero,
		CostsBudgetUSD:         zero,
		CostsProgressPct:       zero,
		HarvestHectares:        zero,
		HarvestTotalHectares:   zero,
		HarvestProgressPercent: zero,
		OperatingResultUSD:     zero,
		OperatingTotalCostsUSD: zero,
		OperatingResultPct:     zero,
	}

	return r.mapper.DashboardDataToDomain(
		tempData,
		[]models.CropIncidenceModel{},
		[]models.ContributionsProgressModel{},
		&models.ManagementBalanceModel{
			Summary: &models.ManagementBalanceSummary{
				IncomeUSD:                 zero,
				DirectCostsExecutedUSD:    zero,
				DirectCostsInvestedUSD:    zero,
				StockUSD:                  zero,
				RentExecutedUSD:           zero,
				RentUSD:                   zero,
				StructureExecutedUSD:      zero,
				StructureUSD:              zero,
				OperatingResultUSD:        zero,
				OperatingResultPct:        zero,
				SemillaCostUSD:            zero,
				InsumosCostUSD:            zero,
				LaboresCostUSD:            zero,
				SemillasInvertidosUSD:     zero,
				SemillasStockUSD:          zero,
				AgroquimicosInvertidosUSD: zero,
				AgroquimicosStockUSD:      zero,
				LaboresInvertidosUSD:      zero,
			},
			Breakdown: []models.ManagementBalanceBreakdown{},
			TotalsRow: &models.ManagementBalanceTotals{
				TotalExecutedUSD: zero,
				TotalInvestedUSD: zero,
				TotalStockUSD:    zero,
			},
		},
		&models.OperationalIndicatorModel{},
	)
}

// getMetrics obtiene todas las métricas principales del dashboard en una sola query
func (r *Repository) getMetrics(ctx context.Context, projectIDs []int64, filter domain.DashboardFilter) (*models.DashboardMetricsModel, error) {
	var result models.DashboardMetricsModel

	viewName := r.metricsView(filter)
	columns := sharedfilters.WorkspaceFilterColumns{
		CustomerID: "customer_id",
		ProjectID:  "project_id",
		CampaignID: "campaign_id",
	}
	if filter.FieldID != nil {
		columns.FieldID = "field_id"
	}

	err := r.applyWorkspaceFilters(
		r.db.Client().WithContext(ctx).
			Table(viewName).
			Select("*").
			Where("project_id IN ?", projectIDs),
		filter,
		columns,
	).Scan(&result).Error

	if err != nil {
		return nil, domainerr.Internal("failed to get dashboard metrics")
	}

	return &result, nil
}

// getContributionsProgress obtiene los datos del avance de aportes por inversor
func (r *Repository) getContributionsProgress(ctx context.Context, projectIDs []int64, filter domain.DashboardFilter) ([]models.ContributionsProgressModel, error) {
	var results []models.ContributionsProgressModel

	viewName := db.DashboardView("contributions_progress")
	columns := sharedfilters.WorkspaceFilterColumns{
		ProjectID: "project_id",
	}
	err := r.applyWorkspaceFilters(
		r.db.Client().WithContext(ctx).
			Table(viewName).
			Select("*").
			Where("project_id IN ?", projectIDs).
			Order("investor_id"),
		filter,
		columns,
	).Scan(&results).Error

	if err != nil {
		return nil, domainerr.Internal("failed to get contributions progress data")
	}

	return results, nil
}

// getManagementBalance obtiene los datos del balance de gestión
func (r *Repository) getManagementBalance(ctx context.Context, projectIDs []int64, filter domain.DashboardFilter) (*models.ManagementBalanceModel, error) {
	var summary models.ManagementBalanceSummary

	viewName := r.managementBalanceView(filter)
	columns := sharedfilters.WorkspaceFilterColumns{
		ProjectID: "project_id",
	}
	if filter.FieldID != nil {
		columns.FieldID = "field_id"
		columns.CustomerID = "customer_id"
		columns.CampaignID = "campaign_id"
	}

	err := r.applyWorkspaceFilters(
		r.db.Client().WithContext(ctx).
			Table(viewName).
			Select("*").
			Where("project_id IN ?", projectIDs),
		filter,
		columns,
	).Scan(&summary).Error

	if err != nil {
		return nil, domainerr.Internal("failed to get management balance data")
	}

	breakdown := []models.ManagementBalanceBreakdown{
		{Category: "Semillas", ExecutedUSD: summary.SemillaCostUSD, InvestedUSD: summary.SemillasInvertidosUSD, StockUSD: summary.SemillasStockUSD},
		{Category: "Agroquímicos", ExecutedUSD: summary.InsumosCostUSD, InvestedUSD: summary.AgroquimicosInvertidosUSD, StockUSD: summary.AgroquimicosStockUSD},
		{Category: "Fertilizantes", ExecutedUSD: summary.FertilizantesCostUSD, InvestedUSD: summary.FertilizantesInvertidosUSD, StockUSD: summary.FertilizantesStockUSD},
		{Category: "Labores", ExecutedUSD: summary.LaboresCostUSD, InvestedUSD: summary.LaboresInvertidosUSD, StockUSD: summary.LaboresStockUSD},
	}

	return &models.ManagementBalanceModel{
		Summary:   &summary,
		Breakdown: breakdown,
		TotalsRow: &models.ManagementBalanceTotals{
			TotalExecutedUSD: summary.DirectCostsExecutedUSD,
			TotalInvestedUSD: summary.DirectCostsInvestedUSD,
			TotalStockUSD:    summary.StockUSD,
		},
	}, nil
}

// getCropIncidence obtiene los datos de incidencia de costos por cultivo
func (r *Repository) getCropIncidence(ctx context.Context, projectIDs []int64, filter domain.DashboardFilter) ([]models.CropIncidenceModel, error) {
	var results []models.CropIncidenceModel

	viewName := r.cropIncidenceView(filter)
	columns := sharedfilters.WorkspaceFilterColumns{
		ProjectID: "project_id",
	}
	if filter.FieldID != nil {
		columns.FieldID = "field_id"
		columns.CustomerID = "customer_id"
		columns.CampaignID = "campaign_id"
	}

	err := r.applyWorkspaceFilters(
		r.db.Client().WithContext(ctx).
			Table(viewName).
			Select("*").
			Where("project_id IN ?", projectIDs).
			Order("crop_name"),
		filter,
		columns,
	).Scan(&results).Error

	if err != nil {
		return nil, domainerr.Internal("failed to get crop incidence data")
	}

	return results, nil
}

// getOperationalIndicators obtiene los indicadores operativos
func (r *Repository) getOperationalIndicators(ctx context.Context, projectIDs []int64, filter domain.DashboardFilter) (*models.OperationalIndicatorModel, error) {
	var result models.OperationalIndicatorModel

	viewName := r.operationalIndicatorsView(filter)
	columns := sharedfilters.WorkspaceFilterColumns{
		ProjectID: "project_id",
	}
	if filter.FieldID != nil {
		columns.FieldID = "field_id"
		columns.CustomerID = "customer_id"
		columns.CampaignID = "campaign_id"
	}

	err := r.applyWorkspaceFilters(
		r.db.Client().WithContext(ctx).
			Table(viewName).
			Select("*").
			Where("project_id IN ?", projectIDs).
			Limit(1),
		filter,
		columns,
	).Scan(&result).Error

	if err != nil {
		return nil, domainerr.Internal("failed to get operational indicators data")
	}

	return &result, nil
}
