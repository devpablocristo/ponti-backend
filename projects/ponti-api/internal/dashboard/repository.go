package dashboard

import (
	"context"
	"fmt"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
	"github.com/shopspring/decimal"
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

	// TODO: Implementar los otros 4 módulos cuando se requiera
	// - Módulo 5: Resultado Operativo
	// - Módulo 6: Balance de Gestión
	// - Módulo 7: Incidencia de Costos por Cultivo
	// - Módulo 8: Indicadores Operativos

	// Crear estructura temporal con datos de siembra, costos, cosecha y aportes
	var sowingHectares, sowingTotalHectares, sowingProgressPercent decimal.Decimal
	var costsExecutedUSD, costsBudgetUSD, costsProgressPercent decimal.Decimal
	var harvestHectares, harvestTotalHectares, harvestProgressPercent decimal.Decimal

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

	if costsData.BudgetTotalUSD != nil {
		costsBudgetUSD = *costsData.BudgetTotalUSD
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
		// Los demás campos se dejan en cero hasta implementar los otros módulos
	}

	// Usar el mapper para convertir a dominio, pasando los datos de aportes
	investorContributions := r.mapper.ContributionsProgressToInvestorContribution(contributionsData)
	return r.mapper.DashboardDataToDomain(tempData, nil, investorContributions, nil), nil
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

	// Convertir []int64 a []interface{} para GORM
	interfaceArgs := make([]interface{}, len(args))
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
		FROM dashboard_sowing_progress_view 
		WHERE project_id = ANY($1)
	`

	args := []interface{}{projectIDs}

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
		var rawHectares, rawTotalHectares, rawProgressPct interface{}
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
			ExecutedLaborsUSD:   &zero,
			ExecutedSuppliesUSD: &zero,
			ExecutedCostsUSD:    &zero,
			BudgetCostUSD:       &zero,
			BudgetTotalUSD:      &zero,
			ProgressPct:         &zero,
		}, nil
	}

	query := `
		SELECT 
			executed_labors_usd,
			executed_supplies_usd,
			executed_costs_usd,
			budget_cost_usd,
			budget_total_usd,
			costs_progress_pct
		FROM dashboard_costs_progress_view 
		WHERE project_id = ANY($1)
	`

	args := []interface{}{projectIDs}

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
		var rawLaborsUSD, rawSuppliesUSD, rawCostsUSD, rawBudgetCost, rawBudgetTotal, rawProgressPct interface{}
		err = rows.Scan(&rawLaborsUSD, &rawSuppliesUSD, &rawCostsUSD, &rawBudgetCost, &rawBudgetTotal, &rawProgressPct)
		if err != nil {
			return nil, types.NewError(types.ErrInternal, "failed to scan costs progress data", err)
		}

		// Convertir los valores raw a decimal.Decimal
		var laborsUSD, suppliesUSD, costsUSD, budgetCost, budgetTotal, progressPct *decimal.Decimal

		// Convertir ExecutedLaborsUSD
		if rawLaborsUSD != nil {
			if strVal, ok := rawLaborsUSD.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					laborsUSD = &dec
				}
			} else if floatVal, ok := rawLaborsUSD.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				laborsUSD = &dec
			}
		}

		// Convertir ExecutedSuppliesUSD
		if rawSuppliesUSD != nil {
			if strVal, ok := rawSuppliesUSD.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					suppliesUSD = &dec
				}
			} else if floatVal, ok := rawSuppliesUSD.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				suppliesUSD = &dec
			}
		}

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

		// Convertir BudgetTotalUSD
		if rawBudgetTotal != nil {
			if strVal, ok := rawBudgetTotal.(string); ok {
				if dec, err := decimal.NewFromString(strVal); err == nil {
					budgetTotal = &dec
				}
			} else if floatVal, ok := rawBudgetTotal.(float64); ok {
				dec := decimal.NewFromFloat(floatVal)
				budgetTotal = &dec
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
			ExecutedLaborsUSD:   laborsUSD,
			ExecutedSuppliesUSD: suppliesUSD,
			ExecutedCostsUSD:    costsUSD,
			BudgetCostUSD:       budgetCost,
			BudgetTotalUSD:      budgetTotal,
			ProgressPct:         progressPct,
		}

		return &result, nil
	}

	// Si no hay filas, retornar valores por defecto
	zero := decimal.Zero
	return &models.CostsProgressModel{
		ExecutedLaborsUSD:   &zero,
		ExecutedSuppliesUSD: &zero,
		ExecutedCostsUSD:    &zero,
		BudgetCostUSD:       &zero,
		BudgetTotalUSD:      &zero,
		ProgressPct:         &zero,
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
		FROM dashboard_harvest_progress_view 
		WHERE project_id = ANY($1)
	`

	args := []interface{}{projectIDs}

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
		var rawHectares, rawTotalHectares, rawProgressPct interface{}
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
		FROM dashboard_contributions_progress_view 
		WHERE project_id = ANY($1)
		ORDER BY investor_id
	`

	args := []interface{}{projectIDs}

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
		var rawInvestorID, rawInvestorName, rawPercentage, rawProgressPct interface{}
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

		// Debug: Imprimir valores raw para ver qué llega
		fmt.Printf("DEBUG: Raw values for row %d - InvestorID: %v (%T), InvestorName: %v (%T), Percentage: %v (%T), ProgressPct: %v (%T)\n",
			i+1, rawInvestorID, rawInvestorID, rawInvestorName, rawInvestorName, rawPercentage, rawPercentage, rawProgressPct, rawProgressPct)
		fmt.Printf("DEBUG: Converted values for row %d - InvestorID: %v, InvestorName: %v, Percentage: %v, ProgressPct: %v\n",
			i+1, investorID, investorName, percentage, progressPct)

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
// func (r *Repository) getOperatingResult(ctx context.Context, filter domain.DashboardFilter) (*models.OperatingResultModel, error) {
// 	var projectIDs []int64
// 	var err error
//
// 	// Si tenemos ProjectID directamente, usarlo sin buscar
// 	if filter.ProjectID != nil {
// 		projectIDs = []int64{*filter.ProjectID}
// 	} else {
// 		// Solo buscar proyectos relacionados si no tenemos ProjectID directo
// 		// (por CustomerID, CampaignID o FieldID)
// 		projectIDs, err = r.getRelatedProjectIDs(ctx, filter)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
//
// 	// Si no hay proyectos relacionados, retornar datos vacíos
// 	if len(projectIDs) == 0 {
// 		return &models.OperatingResultModel{}, nil
// 	}
//
// 	// Implementar consulta a dashboard_operating_result_view usando project_id = ANY($1)
// }

// getManagementBalance obtiene los datos del balance de gestión
// func (r *Repository) getManagementBalance(ctx context.Context, filter domain.DashboardFilter) (*models.ManagementBalanceModel, error) {
// 	var projectIDs []int64
// 	var err error
//
// 	// Si tenemos ProjectID directamente, usarlo sin buscar
// 	if filter.ProjectID != nil {
// 		projectIDs = []int64{*filter.ProjectID}
// 	} else {
// 		// Solo buscar proyectos relacionados si no tenemos ProjectID directo
// 		// (por CustomerID, CampaignID o FieldID)
// 		projectIDs, err = r.getRelatedProjectIDs(ctx, filter)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
//
// 	// Si no hay proyectos relacionados, retornar datos vacíos
// 	if len(projectIDs) == 0 {
// 		return &models.ManagementBalanceModel{}, nil
// 	}
//
// 	// Implementar consulta a dashboard_management_balance_view usando project_id = ANY($1)
// }

// getCropIncidence obtiene los datos de incidencia de costos por cultivo
// func (r *Repository) getCropIncidence(ctx context.Context, filter domain.DashboardFilter) ([]models.CropIncidenceModel, error) {
// 	var projectIDs []int64
// 	var err error
//
// 	// Si tenemos ProjectID directamente, usarlo sin buscar
// 	if filter.ProjectID != nil {
// 		projectIDs = []int64{*filter.ProjectID}
// 	} else {
// 		// Solo buscar proyectos relacionados si no tenemos ProjectID directo
// 		// (por CustomerID, CampaignID o FieldID)
// 		projectIDs, err = r.getRelatedProjectIDs(ctx, filter)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
//
// 	// Si no hay proyectos relacionados, retornar datos vacíos
// 	if len(projectIDs) == 0 {
// 		return []models.CropIncidenceModel{}, nil
// 	}
//
// 	// Implementar consulta a dashboard_crop_incidence_view usando project_id = ANY($1)
// }

// getOperationalIndicators obtiene los indicadores operativos
// func (r *Repository) getOperationalIndicators(ctx context.Context, filter domain.DashboardFilter) (*models.OperationalIndicatorModel, error) {
// 	var projectIDs []int64
// 	var err error
//
// 	// Si tenemos ProjectID directamente, usarlo sin buscar
// 	if filter.ProjectID != nil {
// 		projectIDs = []int64{*filter.ProjectID}
// 	} else {
// 		// Solo buscar proyectos relacionados si no tenemos ProjectID directo
// 		// (por CustomerID, CampaignID o FieldID)
// 		projectIDs, err = r.getRelatedProjectIDs(ctx, filter)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
//
// 	// Si no hay proyectos relacionados, retornar datos vacíos
// 	if len(projectIDs) == 0 {
// 		return &models.OperationalIndicatorModel{}, nil
// 	}
//
// 	// Implementar consulta a dashboard_operational_indicators_view usando project_id = ANY($1)
// }
