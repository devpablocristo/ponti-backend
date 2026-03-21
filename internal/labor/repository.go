// Package labor contiene la implementación del repositorio para el módulo de labor
package labor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/devpablocristo/ponti-backend/internal/labor/repository/models"
	"github.com/devpablocristo/ponti-backend/internal/labor/usecases/domain"
	shareddb "github.com/devpablocristo/ponti-backend/internal/shared/db"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	sharedfilters "github.com/devpablocristo/ponti-backend/internal/shared/filters"
	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"
	workOrderModels "github.com/devpablocristo/ponti-backend/internal/work-order/repository/models"
	"gorm.io/gorm"

	"github.com/devpablocristo/core/backend/go/domainerr"
	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
	"github.com/shopspring/decimal"
)

type GormEnginePort interface {
	Client() *gorm.DB
}
type Repository struct {
	db GormEnginePort
}

func NewRepository(db GormEnginePort) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreateLabor(ctx context.Context, labor *domain.Labor) (int64, error) {
	if err := sharedrepo.ValidateEntity(labor, "labor"); err != nil {
		return 0, err
	}
	model := models.FromDomain(labor)
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, domainerr.Internal("failed to create labor")
	}
	return model.ID, nil
}

func (r *Repository) GetLabor(ctx context.Context, laborID int64) (*domain.Labor, error) {
	var m models.Labor
	if err := r.db.Client().WithContext(ctx).
		Preload("Category").
		First(&m, laborID).Error; err != nil {
		return nil, sharedrepo.HandleGormError(err, "labor", laborID)
	}
	return m.ToDomain(), nil
}

func (r *Repository) GetWorkOrdersByLaborID(ctx context.Context, laborID int64) (int64, error) {
	var count int64
	if err := r.db.Client().WithContext(ctx).
		Model(&workOrderModels.WorkOrder{}).
		Joins("JOIN labors ON labors.id = workorders.labor_id").
		Where("labors.id = ? AND workorders.deleted_at IS NULL", laborID).
		Count(&count).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, domainerr.Internal("failed to get work order")
	}
	return count, nil
}

func (r *Repository) DeleteLabor(ctx context.Context, id int64) error {
	result := r.db.Client().WithContext(ctx).
		Delete(&models.Labor{}, "id = ?", id)
	if result.Error != nil {
		return domainerr.Internal("failed to delete labor")
	}
	if result.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("labor with id %d does not exist", id))
	}
	return nil
}

func (r *Repository) UpdateLabor(ctx context.Context, labor *domain.Labor) error {
	if err := sharedrepo.ValidateEntity(labor, "labor"); err != nil {
		return err
	}

	updates := map[string]any{
		"name":             labor.Name,
		"contractor_name":  labor.ContractorName,
		"price":            labor.Price,
		"is_partial_price": labor.IsPartialPrice,
		"project_id":       labor.ProjectId,
		"category_id":      labor.CategoryId,
		"updated_by":       labor.UpdatedBy,
	}

	result := r.db.Client().WithContext(ctx).
		Model(&models.Labor{}).
		Where("id = ?", labor.ID).
		Updates(updates)

	if result.Error != nil {
		return domainerr.Internal("failed to update labor")
	}
	if result.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("labor with id %d does not exist", labor.ID))
	}
	return nil
}

func (r *Repository) ListLabor(ctx context.Context, page, perPage int, projectID int64) ([]domain.ListedLabor, int64, error) {
	var list []models.Labor
	var total int64

	base := r.db.Client().WithContext(ctx).
		Model(&models.Labor{}).
		Where("project_id = ?", projectID)

	// Conteo total filtrado por proyecto
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count labors")
	}

	if err := base.
		Preload("Category").
		Select("id, name, contractor_name, price, is_partial_price, category_id, updated_at").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Find(&list).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to list labor")
	}

	// Mapear a dominio ligero
	labors := make([]domain.ListedLabor, len(list))
	for i, labor := range list {
		labors[i] = domain.ListedLabor{
			ID:             labor.ID,
			Name:           labor.Name,
			Price:          labor.Price,
			IsPartialPrice: labor.IsPartialPrice,
			ContractorName: labor.ContractorName,
			CategoryId:     labor.LaborCategoryID,
			CategoryName:   labor.Category.Name,
			Base:           shareddomain.Base{UpdatedAt: labor.UpdatedAt},
		}
	}

	return labors, total, nil
}

func (r *Repository) ListLaborCategoriesByTypeID(ctx context.Context, typeID int64) ([]domain.LaborCategory, error) {
	var laborCategoriesModels []models.LaborCategory
	db0 := r.db.
		Client().
		WithContext(ctx).
		Model(&models.LaborCategory{}).
		Where("type_id = ?", typeID)

	if err := db0.Find(&laborCategoriesModels).Error; err != nil {
		return nil, domainerr.Internal("failed to list labor categories")
	}

	laborCategories := make([]domain.LaborCategory, len(laborCategoriesModels))
	for i, laborCategory := range laborCategoriesModels {
		laborCategories[i] = domain.LaborCategory{
			ID:   laborCategory.ID,
			Name: laborCategory.Name,
			LaborType: domain.LaborType{
				ID:   laborCategory.LaborType.ID,
				Name: laborCategory.LaborType.Name,
			},
		}
	}
	return laborCategories, nil
}

func (r *Repository) ListByWorkOrder(ctx context.Context, workOrderID int64) ([]domain.LaborRawItem, error) {
	var v4Models []models.LaborListItem

	query := fmt.Sprintf(`
		SELECT
            v4.workorder_id,
            v4.workorder_number,
            v4.date,
            v4.project_name,
            v4.field_name,
            COALESCE(v4.crop_name, '') AS crop_name,
            v4.labor_name,
            v4.contractor,
            v4.surface_ha,
            v4.cost_per_ha,
            COALESCE(v4.labor_category_name, '') AS category_name,
            COALESCE(v4.investor_name, '') AS investor_name,
			v4.dollar_average_month,
			v4.dollar_average_month AS usd_avg_value,
			i.id AS invoice_id,
			i.number AS invoice_number,
			i.company AS invoice_company,
			i.date AS invoice_date,
			i.status AS invoice_status
		FROM %s AS v4
		LEFT JOIN invoices i ON i.work_order_id = v4.workorder_id
		WHERE v4.workorder_id = ?
	`, shareddb.ReportView("labor_list"))

	err := r.db.Client().WithContext(ctx).Raw(query, workOrderID).Scan(&v4Models).Error
	if err != nil {
		return nil, domainerr.Internal("failed to list labors by work order")
	}

	// Convertir a LaborRawItem para mantener compatibilidad
	raws := make([]domain.LaborRawItem, len(v4Models))
	for i, m := range v4Models {
		// Manejar InvoiceID de forma segura
		var invoiceID int64
		if m.InvoiceID != nil {
			invoiceID = *m.InvoiceID
		}

		raws[i] = domain.LaborRawItem{
			WorkOrderID:     m.WorkOrderID,
			WorkOrderNumber: m.WorkOrderNumber,
			Date:            m.Date,
			ProjectName:     m.ProjectName,
			FieldName:       m.FieldName,
			CropName:        safeStringPtr(m.CropName),
			LaborName:       m.LaborName,
			Contractor:      m.Contractor,
			SurfaceHa:       m.SurfaceHa,
			CostHa:          m.CostPerHa,
			CategoryName:    safeStringPtr(m.LaborCategoryName),
			InvestorName:    safeStringPtr(m.InvestorName),
			USDAvgValue:     m.USDAvgValue,
			InvoiceID:       invoiceID,
			InvoiceNumber:   safeStringPtr(m.InvoiceNumber),
			InvoiceCompany:  safeStringPtr(m.InvoiceCompany),
			InvoiceDate:     m.InvoiceDate,
			InvoiceStatus:   safeStringPtr(m.InvoiceStatus),
		}
	}
	return raws, nil
}

func (r *Repository) ListGroupLabor(
	ctx context.Context,
	inp types.Input,
	projectID int64,
	fieldID int64,
) ([]domain.LaborListItem, types.PageInfo, error) {
	if err := sharedfilters.ValidateFieldBelongsToProject(ctx, r.db.Client(), projectID, fieldID); err != nil {
		return nil, types.PageInfo{}, err
	}
	where := []string{}
	args := []any{}
	if fieldID != 0 {
		where = append(where, "v4.field_id = ?")
		args = append(args, fieldID)
	} else if projectID != 0 {
		where = append(where, "v4.project_id = ?")
		args = append(args, projectID)
	} else {
		return nil, types.PageInfo{}, domainerr.Validation("fieldID or projectID is required")
	}
	whereSQL := strings.Join(where, " AND ")
	view := shareddb.ReportView("labor_list")

	selectCols := `
			v4.workorder_id,
			v4.workorder_number,
			v4.date,
			v4.project_id,
			v4.field_id,
			v4.project_name,
			v4.field_name,
			v4.lot_id,
			v4.lot_name,
			v4.crop_id,
			COALESCE(v4.crop_name, '') AS crop_name,
			v4.labor_id,
			v4.labor_name,
			v4.labor_category_id,
			COALESCE(v4.labor_category_name, '') AS labor_category_name,
			v4.contractor,
			v4.contractor_name,
			v4.surface_ha,
			v4.cost_per_ha,
			v4.total_labor_cost,
			v4.dollar_average_month,
			v4.dollar_average_month AS usd_avg_value,
			v4.usd_cost_ha,
			v4.usd_net_total,
			v4.investor_id,
			COALESCE(v4.investor_name, '') AS investor_name,
			i.id AS invoice_id,
			i.number AS invoice_number,
			i.company AS invoice_company,
			i.date AS invoice_date,
			i.status AS invoice_status
	`

	// Conteo para paginación
	var total int64
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s AS v4
		LEFT JOIN invoices i ON i.work_order_id = v4.workorder_id
		WHERE %s
	`, view, whereSQL)
	if err := r.db.Client().WithContext(ctx).Raw(countQuery, args...).Scan(&total).Error; err != nil {
		return nil, types.PageInfo{}, domainerr.Internal("failed to count labors for work order")
	}

	offset := (int(inp.Page) - 1) * int(inp.PageSize)

	// Datos
	var rows []models.LaborListItem
	dataQuery := fmt.Sprintf(`
		SELECT %s
		FROM %s AS v4
		LEFT JOIN invoices i ON i.work_order_id = v4.workorder_id
		WHERE %s
		ORDER BY v4.workorder_number DESC
		LIMIT ? OFFSET ?
	`, selectCols, view, whereSQL)
	dataArgs := append(append([]any{}, args...), int(inp.PageSize), offset)
	if err := r.db.Client().WithContext(ctx).Raw(dataQuery, dataArgs...).Scan(&rows).Error; err != nil {
		return nil, types.PageInfo{}, domainerr.Internal("failed to list grouped labors")
	}

	// IVA (tasa 0.105; si viene 1.105 se normaliza en getIVAPercentage)
	ivaRate, _ := r.getIVAPercentage(ctx)        // 0.105
	ivaMul := decimal.NewFromInt(1).Add(ivaRate) // 1.105

	// Mapear a dominio
	list := make([]domain.LaborListItem, len(rows))
	for i, m := range rows {
		// Costo ARS/ha SIN IVA = USD/ha × dólar prom
		costHaARS := m.USDCostHa.Mul(m.USDAvgValue)

		// Total neto ARS (SIN IVA)
		netTotal := costHaARS.Mul(m.SurfaceHa)

		// "Total IVA" (histórico): mostramos TOTAL CON IVA para que coincida con la UI actual
		totalConIVA := netTotal.Mul(ivaMul) // net × 1.105

		// USD (vienen de la vista)
		usdCostHa := m.USDCostHa
		usdNetTotal := m.USDNetTotal

		// Invoice seguro
		var invoiceID int64
		if m.InvoiceID != nil {
			invoiceID = *m.InvoiceID
		}

		list[i] = domain.LaborListItem{
			WorkOrderID:     m.WorkOrderID,
			WorkOrderNumber: m.WorkOrderNumber,
			Date:            m.Date,
			ProjectName:     m.ProjectName,
			FieldName:       m.FieldName,
			CropName:        safeStringPtr(m.CropName),
			LaborName:       m.LaborName,
			Contractor:      m.Contractor,
			SurfaceHa:       m.SurfaceHa,
			CostHa:          costHaARS, // ARS/ha SIN IVA (10 × 1000 = 10.000)
			CategoryName:    safeStringPtr(m.LaborCategoryName),
			InvestorName:    safeStringPtr(m.InvestorName),
			USDAvgValue:     m.USDAvgValue,
			NetTotal:        netTotal,    // 10.000 × 100 = 1.000.000
			TotalIVA:        totalConIVA, // MOSTRAMOS TOTAL CON IVA: 1.000.000 × 1.105 = 1.105.000
			USDCostHa:       usdCostHa,   // 10
			USDNetTotal:     usdNetTotal, // 1000
			InvoiceID:       invoiceID,
			InvoiceNumber:   safeStringPtr(m.InvoiceNumber),
			InvoiceCompany:  safeStringPtr(m.InvoiceCompany),
			InvoiceDate:     m.InvoiceDate,
			InvoiceStatus:   safeStringPtr(m.InvoiceStatus),
		}
	}

	pageInfo := types.NewPageInfo(int(inp.Page), int(inp.PageSize), total)
	return list, pageInfo, nil
}

// getIVAPercentage obtiene el porcentaje de IVA desde bparams
func (r *Repository) getIVAPercentage(ctx context.Context) (decimal.Decimal, error) {
	var value string
	err := r.db.Client().WithContext(ctx).
		// business_parameters es la tabla real (bparams era un nombre legacy).
		Table("business_parameters").
		Select("value").
		Where("key = ? AND deleted_at IS NULL", "iva_percentage").
		Scan(&value).Error
	if err != nil || value == "" {
		slog.Warn("IVA percentage not found in business_parameters, using fallback 0.105",
			"error", err, "key", "iva_percentage")
		return decimal.NewFromFloat(0.105), nil
	}
	v, err := decimal.NewFromString(value)
	if err != nil {
		slog.Warn("IVA percentage value is not a valid decimal, using fallback 0.105",
			"error", err, "raw_value", value)
		return decimal.NewFromFloat(0.105), nil
	}
	if v.GreaterThan(decimal.NewFromInt(1)) {
		v = v.Sub(decimal.NewFromInt(1)) // 1.105 -> 0.105
	}
	return v, nil
}

// ListGroupLaborOld MÉTODO VIEJO COMPLETAMENTE COMENTADO PARA REFERENCIA.
// TODO: Eliminar este método.
// Este método implementa la lógica original con cálculos en Go y join con project_dollar_values.
func (r *Repository) ListGroupLaborOld(ctx context.Context, inp types.Input, projectID int64, fieldID int64, usdMonth string) ([]domain.LaborRawItem, types.PageInfo, error) {
	if err := sharedfilters.ValidateFieldBelongsToProject(ctx, r.db.Client(), projectID, fieldID); err != nil {
		return nil, types.PageInfo{}, err
	}
	where := []string{}
	args := []any{usdMonth}
	if fieldID != 0 {
		where = append(where, "v4.field_id = ?")
		args = append(args, fieldID)
	} else if projectID != 0 {
		where = append(where, "v4.project_id = ?")
		args = append(args, projectID)
	} else {
		return nil, types.PageInfo{}, domainerr.Validation(
			"fieldID or projectID is required")
	}
	whereSQL := strings.Join(where, " AND ")
	view := shareddb.ReportView("labor_list")

	selectCols := `
				v4.workorder_id,
				v4.workorder_number,
				v4.date,
				v4.project_id,
				v4.field_id,
				v4.project_name,
				v4.field_name,
				COALESCE(v4.crop_name, '') AS crop_name,
				v4.labor_name,
				COALESCE(v4.labor_category_name, '') AS category_name,
				v4.contractor,
				v4.surface_ha,
				v4.cost_per_ha,
				v4.contractor_name,
				COALESCE(v4.investor_name, '') AS investor_name,
				pdv.average_value AS usd_avg_value,
				i.id AS invoice_id,
				i.number AS invoice_number,
				i.company AS invoice_company,
				i.date AS invoice_date,
				i.status AS invoice_status
			`

	var total int64
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s AS v4
		LEFT JOIN invoices i ON i.work_order_id = v4.workorder_id
		INNER JOIN project_dollar_values pdv
			ON pdv.project_id = v4.project_id AND pdv.month = ? AND pdv.deleted_at IS NULL
		WHERE %s
	`, view, whereSQL)
	if err := r.db.Client().WithContext(ctx).Raw(countQuery, args...).Scan(&total).Error; err != nil {
		return nil, types.PageInfo{}, domainerr.Internal(
			"failed to count labors for work order")
	}

	offset := (int(inp.Page) - 1) * int(inp.PageSize)

	var rows []models.LaborListItem
	dataQuery := fmt.Sprintf(`
		SELECT %s
		FROM %s AS v4
		LEFT JOIN invoices i ON i.work_order_id = v4.workorder_id
		INNER JOIN project_dollar_values pdv
			ON pdv.project_id = v4.project_id AND pdv.month = ? AND pdv.deleted_at IS NULL
		WHERE %s
		ORDER BY v4.workorder_number DESC
		LIMIT ? OFFSET ?
	`, selectCols, view, whereSQL)
	dataArgs := append(append([]any{}, args...), int(inp.PageSize), offset)
	if err := r.db.Client().WithContext(ctx).Raw(dataQuery, dataArgs...).Scan(&rows).Error; err != nil {
		return nil, types.PageInfo{}, domainerr.Internal(
			"failed to list grouped labors")
	}

	list := make([]domain.LaborRawItem, len(rows))
	for i, m := range rows {
		// Calcular valores de USD dinámicamente
		netTotal := m.SurfaceHa.Mul(m.CostPerHa)

		// Obtener porcentaje de IVA dinámicamente desde bparams
		ivaPercentage, err := r.getIVAPercentage(ctx)
		if err != nil {
			slog.Warn("failed to get IVA percentage from bparams, using fallback 0.105",
				"error", err)
			ivaPercentage = decimal.NewFromFloat(0.105)
		}
		totalIVA := netTotal.Mul(ivaPercentage)

		usdCostHa := m.CostPerHa.Div(m.USDAvgValue)
		usdNetTotal := netTotal.Div(m.USDAvgValue)

		// Manejar InvoiceID de forma segura
		var invoiceID int64
		if m.InvoiceID != nil {
			invoiceID = *m.InvoiceID
		}

		list[i] = domain.LaborRawItem{
			WorkOrderID:     m.WorkOrderID,
			WorkOrderNumber: m.WorkOrderNumber,
			Date:            m.Date,
			ProjectName:     m.ProjectName,
			FieldName:       m.FieldName,
			CropName:        safeStringPtr(m.CropName),
			LaborName:       m.LaborName,
			Contractor:      m.Contractor,
			SurfaceHa:       m.SurfaceHa,
			CostHa:          m.CostPerHa,
			CategoryName:    safeStringPtr(m.LaborCategoryName),
			InvestorName:    safeStringPtr(m.InvestorName),
			USDAvgValue:     m.USDAvgValue,
			NetTotal:        netTotal,
			TotalIVA:        totalIVA,
			USDCostHa:       usdCostHa,
			USDNetTotal:     usdNetTotal,
			InvoiceID:       invoiceID,
			InvoiceNumber:   safeStringPtr(m.InvoiceNumber),
			InvoiceCompany:  safeStringPtr(m.InvoiceCompany),
			InvoiceDate:     m.InvoiceDate,
			InvoiceStatus:   safeStringPtr(m.InvoiceStatus),
		}
	}

	pageInfo := types.NewPageInfo(int(inp.Page), int(inp.PageSize), total)
	return list, pageInfo, nil
}

// safeStringPtr convierte un string pointer a string seguro
func safeStringPtr(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func (r *Repository) GetMetrics(ctx context.Context, f domain.LaborFilter) (*domain.LaborMetrics, error) {
	projectIDs, err := sharedfilters.ResolveProjectIDs(ctx, r.db.Client(), sharedfilters.WorkspaceFilter{
		CustomerID: f.CustomerID,
		ProjectID:  f.ProjectID,
		CampaignID: f.CampaignID,
		FieldID:    f.FieldID,
	})
	if err != nil {
		return nil, err
	}
	if len(projectIDs) == 0 && (f.ProjectID != nil || f.CustomerID != nil || f.CampaignID != nil || f.FieldID != nil) {
		return &domain.LaborMetrics{
			SurfaceHa:    decimal.Zero,
			NetTotalCost: decimal.Zero,
			AvgCostPerHa: decimal.Zero,
		}, nil
	}

	var row struct {
		SurfaceHa    decimal.Decimal `gorm:"column:surface_ha"`
		NetTotalCost decimal.Decimal `gorm:"column:total_labor_cost"`
		AvgCostPerHa decimal.Decimal `gorm:"column:avg_labor_cost_per_ha"`
	}

	// Caso 1: project_id Y field_id → devolver métricas de un campo específico
	if len(projectIDs) > 0 && f.FieldID != nil {
		q := fmt.Sprintf(`
			SELECT 
				surface_ha,
				total_labor_cost,
				avg_labor_cost_per_ha
			FROM %s
			WHERE project_id IN ? AND field_id = ?
		`, shareddb.ReportView("labor_metrics"))
		if err := r.db.Client().WithContext(ctx).Raw(q, projectIDs, *f.FieldID).Scan(&row).Error; err != nil {
			return nil, domainerr.Internal("failed to get labor metrics")
		}

		return &domain.LaborMetrics{
			SurfaceHa:    row.SurfaceHa,
			NetTotalCost: row.NetTotalCost,
			AvgCostPerHa: row.AvgCostPerHa,
		}, nil
	}

	// Caso 2: SOLO project_id (sin field_id) → sumar métricas de todos los campos del proyecto
	if len(projectIDs) > 0 {
		q := fmt.Sprintf(`
			SELECT 
				COALESCE(SUM(surface_ha), 0) as surface_ha,
				COALESCE(SUM(total_labor_cost), 0) as total_labor_cost,
				CASE 
					WHEN COALESCE(SUM(surface_ha), 0) > 0 
					THEN COALESCE(SUM(total_labor_cost), 0) / COALESCE(SUM(surface_ha), 0)
					ELSE 0 
				END as avg_labor_cost_per_ha
			FROM %s
			WHERE project_id IN ?
		`, shareddb.ReportView("labor_metrics"))
		if err := r.db.Client().WithContext(ctx).Raw(q, projectIDs).Scan(&row).Error; err != nil {
			return nil, domainerr.Internal("failed to get labor metrics")
		}

		return &domain.LaborMetrics{
			SurfaceHa:    row.SurfaceHa,
			NetTotalCost: row.NetTotalCost,
			AvgCostPerHa: row.AvgCostPerHa,
		}, nil
	}

	// Caso 3: SOLO field_id → devolver métricas de ese campo específico
	if f.FieldID != nil {
		q := fmt.Sprintf(`
			SELECT 
				surface_ha,
				total_labor_cost,
				avg_labor_cost_per_ha
			FROM %s
			WHERE field_id = ?
		`, shareddb.ReportView("labor_metrics"))
		if err := r.db.Client().WithContext(ctx).Raw(q, *f.FieldID).Scan(&row).Error; err != nil {
			return nil, domainerr.Internal("failed to get labor metrics")
		}

		return &domain.LaborMetrics{
			SurfaceHa:    row.SurfaceHa,
			NetTotalCost: row.NetTotalCost,
			AvgCostPerHa: row.AvgCostPerHa,
		}, nil
	}

	// Si no hay filtros, devolver ceros
	return &domain.LaborMetrics{
		SurfaceHa:    decimal.Zero,
		NetTotalCost: decimal.Zero,
		AvgCostPerHa: decimal.Zero,
	}, nil
}

func (r *Repository) ListAllGroupLabor(ctx context.Context) ([]domain.LaborRawItem, error) {
	base := r.db.Client().
		WithContext(ctx).
		Table(shareddb.ReportView("labor_list") + " AS v4").
		Select(`
            v4.workorder_id,
			v4.workorder_number,
			v4.date,
			v4.project_id,
			v4.field_id,
			v4.project_name,
			v4.field_name,
			COALESCE(v4.crop_name, '') AS crop_name,
			v4.labor_name,
			COALESCE(v4.labor_category_name, '') AS category_name,
			v4.contractor,
			v4.surface_ha,
			v4.cost_per_ha,
			v4.contractor_name,
			COALESCE(v4.investor_name, '') AS investor_name,
			v4.dollar_average_month,
			v4.dollar_average_month AS usd_avg_value,
			i.id AS invoice_id,
			i.number AS invoice_number,
			i.company AS invoice_company,
			i.date AS invoice_date,
			i.status AS invoice_status
        `).
		Joins(`LEFT JOIN invoices i ON i.work_order_id = v4.workorder_id AND i.deleted_at IS NULL`)

	var rows []models.LaborListItem

	if err := base.Order("v4.workorder_number DESC").Scan(&rows).Error; err != nil {
		return nil, domainerr.Internal("failed to list grouped labors")
	}

	list := make([]domain.LaborRawItem, len(rows))
	for i, m := range rows {
		// Calcular valores de USD dinámicamente
		netTotal := m.SurfaceHa.Mul(m.CostPerHa)

		// Obtener porcentaje de IVA dinámicamente desde bparams
		ivaPercentage, err := r.getIVAPercentage(ctx)
		if err != nil {
			slog.Warn("failed to get IVA percentage from bparams, using fallback 0.105",
				"error", err)
			ivaPercentage = decimal.NewFromFloat(0.105)
		}
		totalIVA := netTotal.Mul(ivaPercentage)

		usd := m.USDAvgValue
		var usdCostHa, usdNetTotal decimal.Decimal
		if usd.GreaterThan(decimal.Zero) {
			usdCostHa = m.CostPerHa.Div(usd)
			usdNetTotal = netTotal.Div(usd)
		} else {
			// usd <= 0 dejamos USD en 0 (sin dividir)
			usdCostHa = decimal.Zero
			usdNetTotal = decimal.Zero
		}

		// Manejar campos opcionales
		cropName := ""
		if m.CropName != nil {
			cropName = *m.CropName
		}
		categoryName := ""
		if m.LaborCategoryName != nil {
			categoryName = *m.LaborCategoryName
		}
		investorName := ""
		if m.InvestorName != nil {
			investorName = *m.InvestorName
		}
		invoiceNumber := ""
		if m.InvoiceNumber != nil {
			invoiceNumber = *m.InvoiceNumber
		}
		invoiceCompany := ""
		if m.InvoiceCompany != nil {
			invoiceCompany = *m.InvoiceCompany
		}
		invoiceStatus := ""
		if m.InvoiceStatus != nil {
			invoiceStatus = *m.InvoiceStatus
		}

		list[i] = domain.LaborRawItem{
			WorkOrderID:     m.WorkOrderID,
			WorkOrderNumber: m.WorkOrderNumber,
			Date:            m.Date,
			ProjectName:     m.ProjectName,
			FieldName:       m.FieldName,
			CropName:        cropName,
			LaborName:       m.LaborName,
			Contractor:      m.Contractor,
			SurfaceHa:       m.SurfaceHa,
			CostHa:          m.CostPerHa,
			CategoryName:    categoryName,
			InvestorName:    investorName,
			USDAvgValue:     m.USDAvgValue,
			NetTotal:        netTotal,
			TotalIVA:        totalIVA,
			USDCostHa:       usdCostHa,
			USDNetTotal:     usdNetTotal,
			InvoiceID:       0,
			InvoiceNumber:   invoiceNumber,
			InvoiceCompany:  invoiceCompany,
			InvoiceDate:     m.InvoiceDate,
			InvoiceStatus:   invoiceStatus,
		}
	}
	return list, nil
}
