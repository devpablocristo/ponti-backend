package dashboard

import (
	"context"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/devpablocristo/core/errors/go/domainerr"
	sharedfilters "github.com/devpablocristo/ponti-backend/internal/shared/filters"

	models "github.com/devpablocristo/ponti-backend/internal/dashboard/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/dashboard/usecases/domain"
	"github.com/devpablocristo/ponti-backend/internal/shared/authz"
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
	projectIDs, err := sharedfilters.ResolveProjectIDs(ctx, r.db.Client(), sharedfilters.WorkspaceFilter{
		CustomerID: filter.CustomerID,
		ProjectID:  filter.ProjectID,
		CampaignID: filter.CampaignID,
		FieldID:    filter.FieldID,
	})
	if err != nil {
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

	if len(results) == 0 {
		return r.getContributionFallbacks(ctx, projectIDs)
	}

	return results, nil
}

func (r *Repository) getContributionFallbacks(ctx context.Context, projectIDs []int64) ([]models.ContributionsProgressModel, error) {
	type row struct {
		InvestorID         int64           `gorm:"column:investor_id"`
		InvestorName       string          `gorm:"column:investor_name"`
		InvestorPercentage decimal.Decimal `gorm:"column:investor_percentage_pct"`
	}

	var rows []row
	query := r.db.Client().WithContext(ctx).
		Table("project_investors pi").
		Select("pi.investor_id, i.name AS investor_name, SUM(pi.percentage)::numeric AS investor_percentage_pct").
		Joins("JOIN investors i ON i.id = pi.investor_id AND i.tenant_id = pi.tenant_id AND i.deleted_at IS NULL").
		Where("pi.project_id IN ? AND pi.deleted_at IS NULL", projectIDs)
	if tenantID, ok := authz.TenantFromContext(ctx); ok {
		query = query.Where("pi.tenant_id = ?", tenantID)
	}
	err := query.
		Group("pi.investor_id, i.name").
		Order("pi.investor_id").
		Scan(&rows).Error
	if err != nil {
		return nil, domainerr.Internal("failed to get investor contribution fallback data")
	}

	results := make([]models.ContributionsProgressModel, 0, len(rows))
	zero := decimal.Zero
	for _, item := range rows {
		investorID := item.InvestorID
		investorName := item.InvestorName
		percentage := item.InvestorPercentage
		results = append(results, models.ContributionsProgressModel{
			InvestorID:               &investorID,
			InvestorName:             &investorName,
			InvestorPercentage:       &percentage,
			ContributionsProgressPct: &zero,
		})
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

	return aggregateCropIncidence(results), nil
}

func aggregateCropIncidence(rows []models.CropIncidenceModel) []models.CropIncidenceModel {
	if len(rows) <= 1 {
		return rows
	}

	type aggregate struct {
		model   models.CropIncidenceModel
		costUSD decimal.Decimal
	}

	byCrop := make(map[int64]*aggregate, len(rows))
	order := make([]int64, 0, len(rows))
	totalHectares := decimal.Zero

	for _, row := range rows {
		if row.CropID == 0 {
			continue
		}

		totalHectares = totalHectares.Add(row.Hectares)
		current, exists := byCrop[row.CropID]
		if !exists {
			byCrop[row.CropID] = &aggregate{model: row}
			order = append(order, row.CropID)
			current = byCrop[row.CropID]
		} else {
			current.model.Hectares = current.model.Hectares.Add(row.Hectares)
			if current.model.Name == "" {
				current.model.Name = row.Name
			}
		}

		current.costUSD = current.costUSD.Add(row.CostPerHa.Mul(row.Hectares))
	}

	if len(order) == 0 {
		return []models.CropIncidenceModel{}
	}

	result := make([]models.CropIncidenceModel, 0, len(order))
	for _, cropID := range order {
		current := byCrop[cropID]
		if current.model.Hectares.IsPositive() {
			current.model.CostPerHa = current.costUSD.Div(current.model.Hectares)
		}
		if totalHectares.IsPositive() {
			current.model.IncidencePct = current.model.Hectares.Div(totalHectares).Mul(decimal.NewFromInt(100))
		} else {
			current.model.IncidencePct = decimal.Zero
		}
		result = append(result, current.model)
	}

	return result
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
