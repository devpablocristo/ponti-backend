package dashboard

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
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
	// Obtener datos del módulo 1: Avance de Siembra
	sowingData, err := r.getSowingProgress(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Obtener datos del módulo 2: Avance de Costos
	costsData, err := r.getCostsProgress(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Obtener datos del módulo 3: Avance de Cosecha
	harvestData, err := r.getHarvestProgress(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Obtener datos del módulo 4: Avance de Aportes
	contributionsData, err := r.getContributionsProgress(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Obtener datos del módulo 5: Resultado Operativo
	operatingData, err := r.getOperatingResult(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Obtener datos del módulo 6: Balance de Gestión
	managementBalanceData, err := r.getManagementBalance(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Obtener datos del módulo 7: Incidencia de Costos por Cultivo
	cropIncidenceData, err := r.getCropIncidence(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Obtener datos del módulo 8: Indicadores Operativos
	operationalIndicatorsData, err := r.getOperationalIndicators(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Crear estructura temporal con datos de siembra, costos, cosecha, aportes y resultado operativo
	var sowingHectares, sowingTotalHectares, sowingProgressPercent decimal.Decimal
	var costsExecutedUSD, costsBudgetUSD, costsProgressPercent decimal.Decimal
	var harvestHectares, harvestTotalHectares, harvestProgressPercent decimal.Decimal
	var operatingResultUSD, operatingTotalCostsUSD, operatingResultPct decimal.Decimal

	if sowingData.Hectares != nil {
		sowingHectares = *sowingData.Hectares
	} else {
		sowingHectares = decimal.Zero
	}

	if sowingData.TotalHectares != nil {
		sowingTotalHectares = *sowingData.TotalHectares
	} else {
		sowingTotalHectares = decimal.Zero
	}

	if sowingData.ProgressPct != nil {
		sowingProgressPercent = *sowingData.ProgressPct
	} else {
		sowingProgressPercent = decimal.Zero
	}

	if costsData.ExecutedCostsUSD != nil {
		costsExecutedUSD = *costsData.ExecutedCostsUSD
	} else {
		costsExecutedUSD = decimal.Zero
	}

	if costsData.BudgetCostUSD != nil {
		costsBudgetUSD = *costsData.BudgetCostUSD
	} else {
		costsBudgetUSD = decimal.Zero
	}

	if costsData.ProgressPct != nil {
		costsProgressPercent = *costsData.ProgressPct
	} else {
		costsProgressPercent = decimal.Zero
	}

	if harvestData.Hectares != nil {
		harvestHectares = *harvestData.Hectares
	} else {
		harvestHectares = decimal.Zero
	}

	if harvestData.TotalHectares != nil {
		harvestTotalHectares = *harvestData.TotalHectares
	} else {
		harvestTotalHectares = decimal.Zero
	}

	if harvestData.ProgressPct != nil {
		harvestProgressPercent = *harvestData.ProgressPct
	} else {
		harvestProgressPercent = decimal.Zero
	}

	if operatingData.ResultUSD != nil {
		operatingResultUSD = *operatingData.ResultUSD
	} else {
		operatingResultUSD = decimal.Zero
	}

	if operatingData.TotalCostsUSD != nil {
		operatingTotalCostsUSD = *operatingData.TotalCostsUSD
	} else {
		operatingTotalCostsUSD = decimal.Zero
	}

	if operatingData.ResultPct != nil {
		operatingResultPct = *operatingData.ResultPct
	} else {
		operatingResultPct = decimal.Zero
	}

	tempData := &models.DashboardDataModel{
		SowingHectares:         sowingHectares,
		SowingTotalHectares:    sowingTotalHectares,
		SowingProgressPercent:  sowingProgressPercent,
		CostsExecutedUSD:       costsExecutedUSD,
		CostsBudgetUSD:         costsBudgetUSD,
		CostsProgressPct:       costsProgressPercent,
		HarvestHectares:        harvestHectares,
		HarvestTotalHectares:   harvestTotalHectares,
		HarvestProgressPercent: harvestProgressPercent,
		OperatingResultUSD:     operatingResultUSD,
		OperatingTotalCostsUSD: operatingTotalCostsUSD,
		OperatingResultPct:     operatingResultPct,
		// Los demás campos se dejan en cero hasta implementar los otros módulos
	}

	// Usar el mapper para convertir a dominio, pasando los datos de aportes, balance de gestión, crop incidence y operational indicators
	investorContributions := r.mapper.ContributionsProgressToInvestorContribution(contributionsData)
	return r.mapper.DashboardDataToDomain(tempData, cropIncidenceData, investorContributions, managementBalanceData, operationalIndicatorsData), nil
}

// getRelatedProjectIDs encuentra los IDs de proyectos relacionados con los filtros
// Esta función es utilizada por todos los módulos del dashboard ya que todos trabajan a nivel de proyecto
// NOTA: Si ya tenemos ProjectID, no se debe llamar a esta función
func (r *Repository) getRelatedProjectIDs(ctx context.Context, filter domain.DashboardFilter) ([]int64, error) {
	query := `
		SELECT DISTINCT p.id
		FROM projects p
		WHERE 1=1
	`

	args := []int64{}
	argIndex := 1

	// Aplicar filtros si están presentes (excluyendo ProjectID que se maneja directamente)
	if filter.CustomerID != nil {
		query += " AND p.customer_id = $" + string(rune(argIndex+'0'))
		args = append(args, *filter.CustomerID)
		argIndex++
	}
	if filter.CampaignID != nil {
		query += " AND p.campaign_id = $" + string(rune(argIndex+'0'))
		args = append(args, *filter.CampaignID)
		argIndex++
	}
	if filter.FieldID != nil {
		query += " AND EXISTS (SELECT 1 FROM fields f WHERE f.id = $" + string(rune(argIndex+'0')) + " AND f.project_id = p.id)"
		args = append(args, *filter.FieldID)
		argIndex++
	}

	// Convertir []int64 a []any para GORM
	interfaceArgs := make([]any, len(args))
	for i, v := range args {
		interfaceArgs[i] = v
	}

	// Ejecutar la consulta
	var projectIDs []int64
	err := r.db.Client().WithContext(ctx).Raw(query, interfaceArgs...).Scan(&projectIDs).Error
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get related project IDs", err)
	}

	return projectIDs, nil
}

// getSowingProgress obtiene los datos del avance de siembra
func (r *Repository) getSowingProgress(ctx context.Context, filter domain.DashboardFilter) (*models.SowingProgressModel, error) {
	var projectIDs []int64
	var err error

	// Si tenemos ProjectID directamente, usarlo sin buscar
	if filter.ProjectID != nil {
		projectIDs = []int64{*filter.ProjectID}
	} else {
		// Solo buscar proyectos relacionados si no tenemos ProjectID directo
		projectIDs, err = r.getRelatedProjectIDs(ctx, filter)
		if err != nil {
			return nil, err
		}
	}

	// Si no hay proyectos relacionados, retornar datos vacíos
	if len(projectIDs) == 0 {
		zero := decimal.Zero
		return &models.SowingProgressModel{
			Hectares:      &zero,
			TotalHectares: &zero,
			ProgressPct:   &zero,
		}, nil
	}

	query := `
		SELECT 
			sowing_hectares,
			sowing_total_hectares,
			sowing_progress_pct
		FROM v3_dashboard 
		WHERE project_id = ANY($1)
	`

	args := []any{projectIDs}

	// Ejecutar la consulta
	var result models.SowingProgressModel

	// Obtener las filas de la consulta
	rows, err := r.db.Client().WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get sowing progress data", err)
	}
	defer rows.Close()

	// Verificar si hay filas
	hasRows := rows.Next()

	if hasRows {
		// Leer los valores raw
		var rawHectares, rawTotalHectares, rawProgressPct any
		err = rows.Scan(&rawHectares, &rawTotalHectares, &rawProgressPct)
		if err != nil {
			return nil, types.NewError(types.ErrInternal, "failed to scan sowing progress data", err)
		}

		// Convertir los valores raw a decimal.Decimal
		var hectares, totalHectares, progressPct *decimal.Decimal

		// Convertir Hectares
		if rawHectares != nil {
			if strVal, ok := rawHectares.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					hectares = &dec
				}
			} else if floatVal, ok := rawHectares.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				hectares = &dec
			}
		}

		// Convertir TotalHectares
		if rawTotalHectares != nil {
			if strVal, ok := rawTotalHectares.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					totalHectares = &dec
				}
			} else if floatVal, ok := rawTotalHectares.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				totalHectares = &dec
			}
		}

		// Convertir ProgressPct
		if rawProgressPct != nil {
			if strVal, ok := rawProgressPct.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					progressPct = &dec
				}
			} else if floatVal, ok := rawProgressPct.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				progressPct = &dec
			}
		}

		// Crear el resultado
		result = models.SowingProgressModel{
			Hectares:      hectares,
			TotalHectares: totalHectares,
			ProgressPct:   progressPct,
		}

		return &result, nil
	}

	// Si no hay filas, retornar valores por defecto
	zero := decimal.Zero
	return &models.SowingProgressModel{
		Hectares:      &zero,
		TotalHectares: &zero,
		ProgressPct:   &zero,
	}, nil
}

// getCostsProgress obtiene los datos del avance de costos
func (r *Repository) getCostsProgress(ctx context.Context, filter domain.DashboardFilter) (*models.CostsProgressModel, error) {
	var projectIDs []int64
	var err error

	// Si tenemos ProjectID directamente, usarlo sin buscar
	if filter.ProjectID != nil {
		projectIDs = []int64{*filter.ProjectID}
	} else {
		// Solo buscar proyectos relacionados si no tenemos ProjectID directo
		// (por CustomerID, CampaignID o FieldID)
		projectIDs, err = r.getRelatedProjectIDs(ctx, filter)
		if err != nil {
			return nil, err
		}
	}

	// Si no hay proyectos relacionados, retornar datos vacíos
	if len(projectIDs) == 0 {
		zero := decimal.Zero
		return &models.CostsProgressModel{
			ExecutedCostsUSD: &zero,
			BudgetCostUSD:    &zero,
			ProgressPct:      &zero,
		}, nil
	}

	query := `
		SELECT 
		executed_costs_usd,
		budget_cost_usd,
		costs_progress_pct
	FROM v3_dashboard 
		WHERE project_id = ANY($1)
	`

	args := []any{projectIDs}

	// Ejecutar la consulta
	var result models.CostsProgressModel

	// Obtener las filas de la consulta
	rows, err := r.db.Client().WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get costs progress data", err)
	}
	defer rows.Close()

	// Verificar si hay filas
	hasRows := rows.Next()

	if hasRows {
		// Leer los valores raw
		var rawCostsUSD, rawBudgetCost, rawProgressPct any
		err = rows.Scan(&rawCostsUSD, &rawBudgetCost, &rawProgressPct)
		if err != nil {
			return nil, types.NewError(types.ErrInternal, "failed to scan costs progress data", err)
		}

		// Convertir los valores raw a decimal.Decimal
		var costsUSD, budgetCost, progressPct *decimal.Decimal

		// Convertir ExecutedCostsUSD
		if rawCostsUSD != nil {
			if strVal, ok := rawCostsUSD.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					costsUSD = &dec
				}
			} else if floatVal, ok := rawCostsUSD.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				costsUSD = &dec
			}
		}

		// Convertir BudgetCostUSD
		if rawBudgetCost != nil {
			if strVal, ok := rawBudgetCost.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					budgetCost = &dec
				}
			} else if floatVal, ok := rawBudgetCost.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				budgetCost = &dec
			}
		}

		// Convertir ProgressPct
		if rawProgressPct != nil {
			if strVal, ok := rawProgressPct.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					progressPct = &dec
				}
			} else if floatVal, ok := rawProgressPct.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				progressPct = &dec
			}
		}

		// Crear el resultado
		result = models.CostsProgressModel{
			ExecutedCostsUSD: costsUSD,
			BudgetCostUSD:    budgetCost,
			ProgressPct:      progressPct,
		}

		return &result, nil
	}

	// Si no hay filas, retornar valores por defecto
	zero := decimal.Zero
	return &models.CostsProgressModel{
		ExecutedCostsUSD: &zero,
		BudgetCostUSD:    &zero,
		ProgressPct:      &zero,
	}, nil
}

// getHarvestProgress obtiene los datos del avance de cosecha
func (r *Repository) getHarvestProgress(ctx context.Context, filter domain.DashboardFilter) (*models.HarvestProgressModel, error) {
	var projectIDs []int64
	var err error

	// Si tenemos ProjectID directamente, usarlo sin buscar
	if filter.ProjectID != nil {
		projectIDs = []int64{*filter.ProjectID}
	} else {
		// Solo buscar proyectos relacionados si no tenemos ProjectID directo
		// (por CustomerID, CampaignID o FieldID)
		projectIDs, err = r.getRelatedProjectIDs(ctx, filter)
		if err != nil {
			return nil, err
		}
	}

	// Si no hay proyectos relacionados, retornar datos vacíos
	if len(projectIDs) == 0 {
		zero := decimal.Zero
		return &models.HarvestProgressModel{
			Hectares:      &zero,
			TotalHectares: &zero,
			ProgressPct:   &zero,
		}, nil
	}

	query := `
		SELECT 
			harvest_hectares,
			harvest_total_hectares,
			harvest_progress_pct
		FROM v3_dashboard 
		WHERE project_id = ANY($1)
	`

	args := []any{projectIDs}

	// Ejecutar la consulta
	var result models.HarvestProgressModel

	// Obtener las filas de la consulta
	rows, err := r.db.Client().WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get harvest progress data", err)
	}
	defer rows.Close()

	// Verificar si hay filas
	hasRows := rows.Next()

	if hasRows {
		// Leer los valores raw
		var rawHectares, rawTotalHectares, rawProgressPct any
		err = rows.Scan(&rawHectares, &rawTotalHectares, &rawProgressPct)
		if err != nil {
			return nil, types.NewError(types.ErrInternal, "failed to scan harvest progress data", err)
		}

		// Convertir los valores raw a decimal.Decimal
		var hectares, totalHectares, progressPct *decimal.Decimal

		// Convertir Hectares
		if rawHectares != nil {
			if strVal, ok := rawHectares.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					hectares = &dec
				}
			} else if floatVal, ok := rawHectares.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				hectares = &dec
			}
		}

		// Convertir TotalHectares
		if rawTotalHectares != nil {
			if strVal, ok := rawTotalHectares.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					totalHectares = &dec
				}
			} else if floatVal, ok := rawTotalHectares.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				totalHectares = &dec
			}
		}

		// Convertir ProgressPct
		if rawProgressPct != nil {
			if strVal, ok := rawProgressPct.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					progressPct = &dec
				}
			} else if floatVal, ok := rawProgressPct.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				progressPct = &dec
			}
		}

		// Crear el resultado
		result = models.HarvestProgressModel{
			Hectares:      hectares,
			TotalHectares: totalHectares,
			ProgressPct:   progressPct,
		}

		return &result, nil
	}

	// Si no hay filas, retornar valores por defecto
	zero := decimal.Zero
	return &models.HarvestProgressModel{
		Hectares:      &zero,
		TotalHectares: &zero,
		ProgressPct:   &zero,
	}, nil
}

// getContributionsProgress obtiene los datos del avance de aportes
func (r *Repository) getContributionsProgress(ctx context.Context, filter domain.DashboardFilter) ([]models.ContributionsProgressModel, error) {
	var projectIDs []int64
	var err error

	// Si tenemos ProjectID directamente, usarlo sin buscar
	if filter.ProjectID != nil {
		projectIDs = []int64{*filter.ProjectID}
	} else {
		// Solo buscar proyectos relacionados si no tenemos ProjectID directo
		// (por CustomerID, CampaignID o FieldID)
		projectIDs, err = r.getRelatedProjectIDs(ctx, filter)
		if err != nil {
			return nil, err
		}
	}

	// Si no hay proyectos relacionados, retornar datos vacíos
	if len(projectIDs) == 0 {
		return []models.ContributionsProgressModel{}, nil
	}

	query := `
		SELECT 
			investor_id,
			investor_name,
			investor_percentage_pct,
			contributions_progress_pct
		FROM v3_dashboard_contributions_progress 
		WHERE project_id = ANY($1)
		ORDER BY investor_id
	`

	args := []any{projectIDs}

	// Ejecutar la consulta
	var results []models.ContributionsProgressModel

	// Obtener las filas de la consulta
	rows, err := r.db.Client().WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get contributions progress data", err)
	}
	defer rows.Close()

	// Leer todas las filas
	for i := 0; rows.Next(); i++ {
		// Leer los valores raw
		var rawInvestorID, rawInvestorName, rawPercentage, rawProgressPct any
		err = rows.Scan(&rawInvestorID, &rawInvestorName, &rawPercentage, &rawProgressPct)
		if err != nil {
			return nil, types.NewError(types.ErrInternal, "failed to scan contributions progress data", err)
		}

		// Convertir los valores raw a los tipos correctos
		var investorID *int64
		var investorName *string
		var percentage, progressPct *decimal.Decimal

		// Convertir InvestorID
		if rawInvestorID != nil {
			if intVal, ok := rawInvestorID.(int64); ok {
				investorID = &intVal
			} else if floatVal, ok := rawInvestorID.(float64); ok {
				intVal := int64(floatVal)
				investorID = &intVal
			} else if intVal, ok := rawInvestorID.(int); ok {
				int64Val := int64(intVal)
				investorID = &int64Val
			}
		}

		// Convertir InvestorName
		if rawInvestorName != nil {
			if strVal, ok := rawInvestorName.(string); ok {
				investorName = &strVal
			}
		}

		// Convertir InvestorPercentage
		if rawPercentage != nil {
			if strVal, ok := rawPercentage.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					percentage = &dec
				}
			} else if floatVal, ok := rawPercentage.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				percentage = &dec
			} else if intVal, ok := rawPercentage.(int64); ok {
				dec := decimal.NewFromInt(intVal)
				percentage = &dec
			} else if intVal, ok := rawPercentage.(int); ok {
				dec := decimal.NewFromInt(int64(intVal))
				percentage = &dec
			}
		}

		// Convertir ContributionsProgressPct
		if rawProgressPct != nil {
			if strVal, ok := rawProgressPct.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					progressPct = &dec
				}
			} else if floatVal, ok := rawProgressPct.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				progressPct = &dec
			} else if intVal, ok := rawProgressPct.(int64); ok {
				dec := decimal.NewFromInt(intVal)
				progressPct = &dec
			} else if intVal, ok := rawProgressPct.(int); ok {
				dec := decimal.NewFromInt(int64(intVal))
				progressPct = &dec
			}
		}

		// Procesar valores de la base de datos

		// Crear el resultado individual
		result := models.ContributionsProgressModel{
			InvestorID:               investorID,
			InvestorName:             investorName,
			InvestorPercentage:       percentage,
			ContributionsProgressPct: progressPct,
		}

		results = append(results, result)
	}

	return results, nil
}

// getOperatingResult obtiene los datos del resultado operativo
func (r *Repository) getOperatingResult(ctx context.Context, filter domain.DashboardFilter) (*models.OperatingResultModel, error) {
	var projectIDs []int64
	var err error

	// Si tenemos ProjectID directamente, usarlo sin buscar
	if filter.ProjectID != nil {
		projectIDs = []int64{*filter.ProjectID}
	} else {
		// Solo buscar proyectos relacionados si no tenemos ProjectID directo
		// (por CustomerID, CampaignID o FieldID)
		projectIDs, err = r.getRelatedProjectIDs(ctx, filter)
		if err != nil {
			return nil, err
		}
	}

	// Si no hay proyectos relacionados, retornar datos vacíos
	if len(projectIDs) == 0 {
		zero := decimal.Zero
		return &models.OperatingResultModel{
			IncomeUSD:     &zero,
			TotalCostsUSD: &zero,
			ResultUSD:     &zero,
			ResultPct:     &zero,
		}, nil
	}

	query := `
		SELECT 
			operating_result_total_costs_usd,
			operating_result_usd,
			operating_result_pct
		FROM v3_dashboard 
		WHERE project_id = ANY($1)
	`

	args := []any{projectIDs}

	// Ejecutar la consulta
	var result models.OperatingResultModel

	// Obtener las filas de la consulta
	rows, err := r.db.Client().WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get operating result data", err)
	}
	defer rows.Close()

	// Verificar si hay filas
	hasRows := rows.Next()

	if hasRows {
		// Leer los valores raw
		var rawTotalCostsUSD, rawResultUSD, rawResultPct any
		err = rows.Scan(&rawTotalCostsUSD, &rawResultUSD, &rawResultPct)
		if err != nil {
			return nil, types.NewError(types.ErrInternal, "failed to scan operating result data", err)
		}

		// Convertir los valores raw a decimal.Decimal
		var totalCostsUSD, resultUSD, resultPct *decimal.Decimal

		// Convertir TotalCostsUSD
		if rawTotalCostsUSD != nil {
			if strVal, ok := rawTotalCostsUSD.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					totalCostsUSD = &dec
				}
			} else if floatVal, ok := rawTotalCostsUSD.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				totalCostsUSD = &dec
			} else if intVal, ok := rawTotalCostsUSD.(int64); ok {
				dec := decimal.NewFromInt(intVal)
				totalCostsUSD = &dec
			} else if intVal, ok := rawTotalCostsUSD.(int); ok {
				dec := decimal.NewFromInt(int64(intVal))
				totalCostsUSD = &dec
			}
		}

		// Convertir ResultUSD
		if rawResultUSD != nil {
			if strVal, ok := rawResultUSD.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					resultUSD = &dec
				}
			} else if floatVal, ok := rawResultUSD.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				resultUSD = &dec
			} else if intVal, ok := rawResultUSD.(int64); ok {
				dec := decimal.NewFromInt(intVal)
				resultUSD = &dec
			} else if intVal, ok := rawResultUSD.(int); ok {
				dec := decimal.NewFromInt(int64(intVal))
				resultUSD = &dec
			}
		}

		// Convertir ResultPct
		if rawResultPct != nil {
			if strVal, ok := rawResultPct.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					resultPct = &dec
				}
			} else if floatVal, ok := rawResultPct.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				resultPct = &dec
			} else if intVal, ok := rawResultPct.(int64); ok {
				dec := decimal.NewFromInt(intVal)
				resultPct = &dec
			} else if intVal, ok := rawResultPct.(int); ok {
				dec := decimal.NewFromInt(int64(intVal))
				resultPct = &dec
			}
		}

		// Crear el resultado
		result = models.OperatingResultModel{
			TotalCostsUSD: totalCostsUSD,
			ResultUSD:     resultUSD,
			ResultPct:     resultPct,
		}

		return &result, nil
	}

	// Si no hay filas, retornar valores por defecto
	zero := decimal.Zero
	return &models.OperatingResultModel{
		IncomeUSD:     &zero,
		TotalCostsUSD: &zero,
		ResultUSD:     &zero,
		ResultPct:     &zero,
	}, nil
}

// getManagementBalance obtiene los datos del balance de gestión
func (r *Repository) getManagementBalance(ctx context.Context, filter domain.DashboardFilter) (*models.ManagementBalanceModel, error) {
	var projectIDs []int64
	var err error

	// Si tenemos ProjectID directamente, usarlo sin buscar
	if filter.ProjectID != nil {
		projectIDs = []int64{*filter.ProjectID}
	} else {
		// Solo buscar proyectos relacionados si no tenemos ProjectID directo
		// (por CustomerID, CampaignID o FieldID)
		projectIDs, err = r.getRelatedProjectIDs(ctx, filter)
		if err != nil {
			return nil, err
		}
	}

	// Si no hay proyectos relacionados, retornar datos vacíos
	if len(projectIDs) == 0 {
		return &models.ManagementBalanceModel{}, nil
	}

	query := `
		SELECT 
			income_usd,
			costos_directos_ejecutados_usd,
			costos_directos_invertidos_usd,
			arriendo_invertidos_usd,
			estructura_invertidos_usd,
			operating_result_usd,
			operating_result_pct,
			-- Campos de ejecutados (corregidos en migración 000110)
			semilla_cost,
			insumos_cost,
			labores_cost,
			-- Campos de invertidos (agregados en migración 000110)
			semillas_invertidos_usd,
			agroquimicos_invertidos_usd,
			labores_invertidos_usd
		FROM v3_dashboard_management_balance p
		WHERE p.project_id = ANY($1)
	`

	args := []any{projectIDs}

	// Ejecutar la consulta
	var result models.ManagementBalanceModel

	// Obtener las filas de la consulta
	rows, err := r.db.Client().WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get management balance data", err)
	}
	defer rows.Close()

	// Verificar si hay filas
	hasRows := rows.Next()

	if hasRows {
		// Leer los valores raw
		var rawIncomeUSD, rawDirectCostsExecuted, rawDirectCostsInvested, rawRent, rawStructure, rawOperatingResult, rawOperatingResultPct any
		var rawSemillaCost, rawInsumosCost, rawLaboresCost, rawSemillasInvertidos, rawAgroquimicosInvertidos, rawLaboresInvertidos any
		err = rows.Scan(&rawIncomeUSD, &rawDirectCostsExecuted, &rawDirectCostsInvested, &rawRent, &rawStructure, &rawOperatingResult, &rawOperatingResultPct,
			&rawSemillaCost, &rawInsumosCost, &rawLaboresCost, &rawSemillasInvertidos, &rawAgroquimicosInvertidos, &rawLaboresInvertidos)
		if err != nil {
			return nil, types.NewError(types.ErrInternal, "failed to scan management balance data", err)
		}

		// Convertir los valores raw a decimal.Decimal
		var incomeUSD, directCostsExecuted, directCostsInvested, rent, structure, operatingResult, operatingResultPct *decimal.Decimal
		var semillaCost, insumosCost, laboresCost, semillasInvertidos, agroquimicosInvertidos, laboresInvertidos *decimal.Decimal

		// Convertir IncomeUSD
		if rawIncomeUSD != nil {
			if strVal, ok := rawIncomeUSD.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					incomeUSD = &dec
				}
			} else if floatVal, ok := rawIncomeUSD.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				incomeUSD = &dec
			} else if intVal, ok := rawIncomeUSD.(int64); ok {
				dec := decimal.NewFromInt(intVal)
				incomeUSD = &dec
			} else if intVal, ok := rawIncomeUSD.(int); ok {
				dec := decimal.NewFromInt(int64(intVal))
				incomeUSD = &dec
			}
		}

		// Convertir DirectCostsExecuted
		if rawDirectCostsExecuted != nil {
			if strVal, ok := rawDirectCostsExecuted.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					directCostsExecuted = &dec
				}
			} else if floatVal, ok := rawDirectCostsExecuted.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				directCostsExecuted = &dec
			} else if intVal, ok := rawDirectCostsExecuted.(int64); ok {
				dec := decimal.NewFromInt(intVal)
				directCostsExecuted = &dec
			} else if intVal, ok := rawDirectCostsExecuted.(int); ok {
				dec := decimal.NewFromInt(int64(intVal))
				directCostsExecuted = &dec
			}
		}

		// Convertir DirectCostsInvested
		if rawDirectCostsInvested != nil {
			if strVal, ok := rawDirectCostsInvested.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					directCostsInvested = &dec
				}
			} else if floatVal, ok := rawDirectCostsInvested.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				directCostsInvested = &dec
			} else if intVal, ok := rawDirectCostsInvested.(int64); ok {
				dec := decimal.NewFromInt(intVal)
				directCostsInvested = &dec
			} else if intVal, ok := rawDirectCostsInvested.(int); ok {
				dec := decimal.NewFromInt(int64(intVal))
				directCostsInvested = &dec
			}
		}

		// Convertir Rent
		if rawRent != nil {
			if strVal, ok := rawRent.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					rent = &dec
				}
			} else if floatVal, ok := rawRent.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				rent = &dec
			} else if intVal, ok := rawRent.(int64); ok {
				dec := decimal.NewFromInt(intVal)
				rent = &dec
			} else if intVal, ok := rawRent.(int); ok {
				dec := decimal.NewFromInt(int64(intVal))
				rent = &dec
			}
		}

		// Convertir Structure
		if rawStructure != nil {
			if strVal, ok := rawStructure.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					structure = &dec
				}
			} else if floatVal, ok := rawStructure.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				structure = &dec
			} else if intVal, ok := rawStructure.(int64); ok {
				dec := decimal.NewFromInt(intVal)
				structure = &dec
			} else if intVal, ok := rawStructure.(int); ok {
				dec := decimal.NewFromInt(int64(intVal))
				structure = &dec
			}
		}

		// Convertir OperatingResult
		if rawOperatingResult != nil {
			if strVal, ok := rawOperatingResult.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					operatingResult = &dec
				}
			} else if floatVal, ok := rawOperatingResult.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				operatingResult = &dec
			} else if intVal, ok := rawOperatingResult.(int64); ok {
				dec := decimal.NewFromInt(intVal)
				operatingResult = &dec
			} else if intVal, ok := rawOperatingResult.(int); ok {
				dec := decimal.NewFromInt(int64(intVal))
				operatingResult = &dec
			}
		}

		// Convertir OperatingResultPct
		if rawOperatingResultPct != nil {
			if strVal, ok := rawOperatingResultPct.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					operatingResultPct = &dec
				}
			} else if floatVal, ok := rawOperatingResultPct.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				operatingResultPct = &dec
			} else if intVal, ok := rawOperatingResultPct.(int64); ok {
				dec := decimal.NewFromInt(intVal)
				operatingResultPct = &dec
			} else if intVal, ok := rawOperatingResultPct.(int); ok {
				dec := decimal.NewFromInt(int64(intVal))
				operatingResultPct = &dec
			}
		}

		// Convertir SemillaCost
		if rawSemillaCost != nil {
			if strVal, ok := rawSemillaCost.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					semillaCost = &dec
				}
			} else if floatVal, ok := rawSemillaCost.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				semillaCost = &dec
			} else if intVal, ok := rawSemillaCost.(int64); ok {
				dec := decimal.NewFromInt(intVal)
				semillaCost = &dec
			} else if intVal, ok := rawSemillaCost.(int); ok {
				dec := decimal.NewFromInt(int64(intVal))
				semillaCost = &dec
			}
		}

		// Convertir InsumosCost
		if rawInsumosCost != nil {
			if strVal, ok := rawInsumosCost.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					insumosCost = &dec
				}
			} else if floatVal, ok := rawInsumosCost.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				insumosCost = &dec
			} else if intVal, ok := rawInsumosCost.(int64); ok {
				dec := decimal.NewFromInt(intVal)
				insumosCost = &dec
			} else if intVal, ok := rawInsumosCost.(int); ok {
				dec := decimal.NewFromInt(int64(intVal))
				insumosCost = &dec
			}
		}

		// Convertir LaboresCost
		if rawLaboresCost != nil {
			if strVal, ok := rawLaboresCost.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					laboresCost = &dec
				}
			} else if floatVal, ok := rawLaboresCost.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				laboresCost = &dec
			} else if intVal, ok := rawLaboresCost.(int64); ok {
				dec := decimal.NewFromInt(intVal)
				laboresCost = &dec
			} else if intVal, ok := rawLaboresCost.(int); ok {
				dec := decimal.NewFromInt(int64(intVal))
				laboresCost = &dec
			}
		}

		// Convertir SemillasInvertidos
		if rawSemillasInvertidos != nil {
			if strVal, ok := rawSemillasInvertidos.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					semillasInvertidos = &dec
				}
			} else if floatVal, ok := rawSemillasInvertidos.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				semillasInvertidos = &dec
			} else if intVal, ok := rawSemillasInvertidos.(int64); ok {
				dec := decimal.NewFromInt(intVal)
				semillasInvertidos = &dec
			} else if intVal, ok := rawSemillasInvertidos.(int); ok {
				dec := decimal.NewFromInt(int64(intVal))
				semillasInvertidos = &dec
			}
		}

		// Convertir AgroquimicosInvertidos
		if rawAgroquimicosInvertidos != nil {
			if strVal, ok := rawAgroquimicosInvertidos.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					agroquimicosInvertidos = &dec
				}
			} else if floatVal, ok := rawAgroquimicosInvertidos.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				agroquimicosInvertidos = &dec
			} else if intVal, ok := rawAgroquimicosInvertidos.(int64); ok {
				dec := decimal.NewFromInt(intVal)
				agroquimicosInvertidos = &dec
			} else if intVal, ok := rawAgroquimicosInvertidos.(int); ok {
				dec := decimal.NewFromInt(int64(intVal))
				agroquimicosInvertidos = &dec
			}
		}

		// Convertir LaboresInvertidos
		if rawLaboresInvertidos != nil {
			if strVal, ok := rawLaboresInvertidos.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					laboresInvertidos = &dec
				}
			} else if floatVal, ok := rawLaboresInvertidos.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				laboresInvertidos = &dec
			} else if intVal, ok := rawLaboresInvertidos.(int64); ok {
				dec := decimal.NewFromInt(intVal)
				laboresInvertidos = &dec
			} else if intVal, ok := rawLaboresInvertidos.(int); ok {
				dec := decimal.NewFromInt(int64(intVal))
				laboresInvertidos = &dec
			}
		}

		// Crear el resultado con estructura anidada
		result = models.ManagementBalanceModel{
			Summary: &models.ManagementBalanceSummary{
				IncomeUSD:                 *incomeUSD,
				DirectCostsExecutedUSD:    *directCostsExecuted,
				DirectCostsInvestedUSD:    *directCostsInvested,
				StockUSD:                  decimal.Zero,
				RentUSD:                   *rent,
				StructureUSD:              *structure,
				OperatingResultUSD:        *operatingResult,
				OperatingResultPct:        *operatingResultPct,
				SemillaCostUSD:            *semillaCost,
				InsumosCostUSD:            *insumosCost,
				LaboresCostUSD:            *laboresCost,
				SemillasInvertidosUSD:     *semillasInvertidos,
				AgroquimicosInvertidosUSD: *agroquimicosInvertidos,
				LaboresInvertidosUSD:      *laboresInvertidos,
			},
			Breakdown: []models.ManagementBalanceBreakdown{}, // TODO: Implementar cuando se requiera
			TotalsRow: &models.ManagementBalanceTotals{
				TotalExecutedUSD: *directCostsExecuted,
				TotalInvestedUSD: *directCostsInvested,
				TotalStockUSD:    decimal.Zero,
			},
		}

		return &result, nil
	}

	// Si no hay filas, retornar valores por defecto
	zero := decimal.Zero
	return &models.ManagementBalanceModel{
		Summary: &models.ManagementBalanceSummary{
			IncomeUSD:              zero,
			DirectCostsExecutedUSD: zero,
			DirectCostsInvestedUSD: zero,
			StockUSD:               zero,
			RentUSD:                zero,
			StructureUSD:           zero,
			OperatingResultUSD:     zero,
			OperatingResultPct:     zero,
		},
		Breakdown: []models.ManagementBalanceBreakdown{},
		TotalsRow: &models.ManagementBalanceTotals{
			TotalExecutedUSD: zero,
			TotalInvestedUSD: zero,
			TotalStockUSD:    zero,
		},
	}, nil
}

// getCropIncidence obtiene los datos de incidencia de costos por cultivo
func (r *Repository) getCropIncidence(ctx context.Context, filter domain.DashboardFilter) ([]models.CropIncidenceModel, error) {
	var projectIDs []int64
	var err error

	// Si tenemos ProjectID directamente, usarlo sin buscar
	if filter.ProjectID != nil {
		projectIDs = []int64{*filter.ProjectID}
	} else {
		// Solo buscar proyectos relacionados si no tenemos ProjectID directo
		// (por CustomerID, CampaignID o FieldID)
		projectIDs, err = r.getRelatedProjectIDs(ctx, filter)
		if err != nil {
			return nil, err
		}
	}

	// Si no hay proyectos relacionados, retornar datos vacíos
	if len(projectIDs) == 0 {
		return []models.CropIncidenceModel{}, nil
	}

	// Consultar la vista v3_dashboard_crop_incidence
	query := `
		SELECT 
			current_crop_id,
			crop_name,
			crop_hectares,
			crop_incidence_pct,
			cost_per_ha_usd
		FROM v3_dashboard_crop_incidence 
		WHERE project_id = ANY($1)
		ORDER BY crop_name
	`

	args := []any{projectIDs}

	// Ejecutar la consulta
	rows, err := r.db.Client().WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get crop incidence data", err)
	}
	defer rows.Close()

	var result []models.CropIncidenceModel

	// Leer todas las filas
	for rows.Next() {
		var rawCropID, rawCropName, rawHectares, rawIncidencePct, rawCostPerHa any
		err = rows.Scan(&rawCropID, &rawCropName, &rawHectares, &rawIncidencePct, &rawCostPerHa)
		if err != nil {
			return nil, types.NewError(types.ErrInternal, "failed to scan crop incidence data", err)
		}

		// Convertir los valores raw a los tipos correctos
		var cropID int64
		var cropName string
		var hectares, incidencePct, costPerHa decimal.Decimal

		// Convertir cropID
		if rawCropID != nil {
			if intVal, ok := rawCropID.(int64); ok {
				cropID = intVal
			} else if intVal, ok := rawCropID.(int); ok {
				cropID = int64(intVal)
			} else if floatVal, ok := rawCropID.(float64); ok {
				cropID = int64(floatVal)
			}
		}

		// Convertir cropName
		if rawCropName != nil {
			if strVal, ok := rawCropName.(string); ok {
				cropName = strVal
			}
		}

		// Convertir hectares
		if rawHectares != nil {
			if strVal, ok := rawHectares.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					hectares = dec
				}
			} else if floatVal, ok := rawHectares.(float64); ok {
				hectares = decimal.NewFromFloat(floatVal)
			} else if intVal, ok := rawHectares.(int64); ok {
				hectares = decimal.NewFromInt(intVal)
			} else if intVal, ok := rawHectares.(int); ok {
				hectares = decimal.NewFromInt(int64(intVal))
			}
		}

		// Convertir incidencePct
		if rawIncidencePct != nil {
			if strVal, ok := rawIncidencePct.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					incidencePct = dec
				}
			} else if floatVal, ok := rawIncidencePct.(float64); ok {
				incidencePct = decimal.NewFromFloat(floatVal)
			} else if intVal, ok := rawIncidencePct.(int64); ok {
				incidencePct = decimal.NewFromInt(intVal)
			} else if intVal, ok := rawIncidencePct.(int); ok {
				incidencePct = decimal.NewFromInt(int64(intVal))
			}
		}

		// Convertir costPerHa
		if rawCostPerHa != nil {
			if strVal, ok := rawCostPerHa.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					costPerHa = dec
				}
			} else if floatVal, ok := rawCostPerHa.(float64); ok {
				costPerHa = decimal.NewFromFloat(floatVal)
			} else if intVal, ok := rawCostPerHa.(int64); ok {
				costPerHa = decimal.NewFromInt(intVal)
			} else if intVal, ok := rawCostPerHa.(int); ok {
				costPerHa = decimal.NewFromInt(int64(intVal))
			}
		}

		// Crear el modelo
		cropModel := models.CropIncidenceModel{
			CropID:       cropID,
			Name:         cropName,
			Hectares:     hectares,
			IncidencePct: incidencePct,
			CostPerHa:    costPerHa,
		}

		result = append(result, cropModel)
	}

	return result, nil
}

// getOperationalIndicators obtiene los indicadores operativos
func (r *Repository) getOperationalIndicators(ctx context.Context, filter domain.DashboardFilter) (*models.OperationalIndicatorModel, error) {
	var projectIDs []int64
	var err error

	// Si tenemos ProjectID directamente, usarlo sin buscar
	if filter.ProjectID != nil {
		projectIDs = []int64{*filter.ProjectID}
	} else {
		// Solo buscar proyectos relacionados si no tenemos ProjectID directo
		// (por CustomerID, CampaignID o FieldID)
		projectIDs, err = r.getRelatedProjectIDs(ctx, filter)
		if err != nil {
			return nil, err
		}
	}

	// Si no hay proyectos relacionados, retornar datos vacíos
	if len(projectIDs) == 0 {
		return &models.OperationalIndicatorModel{}, nil
	}

	query := `
		SELECT
			start_date,
			end_date,
			campaign_closing_date,
			first_workorder_id,
			last_workorder_id,
			last_stock_count_date
		FROM v3_dashboard 
		WHERE project_id = ANY($1)
		LIMIT 1
	`

	args := []any{projectIDs}

	rows, err := r.db.Client().WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get operational indicators data", err)
	}
	defer rows.Close()

	if rows.Next() {
		// Leer los valores raw
		var rawStartDate, rawEndDate, rawCampaignClosingDate, rawFirstWorkorderID, rawLastWorkorderID, rawLastStockCountDate any
		err = rows.Scan(&rawStartDate, &rawEndDate, &rawCampaignClosingDate, &rawFirstWorkorderID, &rawLastWorkorderID, &rawLastStockCountDate)
		if err != nil {
			return nil, types.NewError(types.ErrInternal, "failed to scan operational indicators data", err)
		}

		// Convertir los valores raw a los tipos correctos
		var startDate, endDate, campaignClosingDate, lastStockCountDate *time.Time
		var firstWorkorderID, lastWorkorderID *int64

		// Convertir startDate
		if rawStartDate != nil {
			if dateVal, ok := rawStartDate.(time.Time); ok {
				startDate = &dateVal
			}
		}

		// Convertir endDate
		if rawEndDate != nil {
			if dateVal, ok := rawEndDate.(time.Time); ok {
				endDate = &dateVal
			}
		}

		// Convertir campaignClosingDate
		if rawCampaignClosingDate != nil {
			if dateVal, ok := rawCampaignClosingDate.(time.Time); ok {
				campaignClosingDate = &dateVal
			}
		}

		// Convertir firstWorkorderID
		if rawFirstWorkorderID != nil {
			if idVal, ok := rawFirstWorkorderID.(int64); ok {
				firstWorkorderID = &idVal
			}
		}

		// Convertir lastWorkorderID
		if rawLastWorkorderID != nil {
			if idVal, ok := rawLastWorkorderID.(int64); ok {
				lastWorkorderID = &idVal
			}
		}

		// Convertir lastStockCountDate
		if rawLastStockCountDate != nil {
			if dateVal, ok := rawLastStockCountDate.(time.Time); ok {
				lastStockCountDate = &dateVal
			}
		}

		// Crear el modelo
		result := models.OperationalIndicatorModel{
			FirstWorkorderDate:   startDate,
			FirstWorkorderNumber: firstWorkorderID,
			LastWorkorderDate:    endDate,
			LastWorkorderNumber:  lastWorkorderID,
			LastStockCountDate:   lastStockCountDate,
			CampaignClosingDate:  campaignClosingDate,
		}

		return &result, nil
	}

	// Si no hay filas, retornar modelo vacío
	return &models.OperationalIndicatorModel{}, nil
}
