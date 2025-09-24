// Package labor contiene la implementación del repositorio para el módulo de labor
package labor

import (
	"context"
	"errors"
	"fmt"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/repository/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/usecases/domain"
	workordermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/repository/models"
	"gorm.io/gorm"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
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
	if labor == nil {
		return 0, types.NewError(types.ErrValidation, "labor is nil", nil)
	}
	model := models.FromDomain(labor)
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create labor", err)
	}
	return model.ID, nil
}

func (r *Repository) GetWorkordersByLaborID(ctx context.Context, laborID int64) (int64, error) {
	var count int64
	if err := r.db.Client().WithContext(ctx).
		Model(&workordermodels.Workorder{}).
		Joins("JOIN labors ON labors.id = workorders.labor_id").
		Where("labors.id = ?", laborID).
		Count(&count).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, types.NewError(types.ErrInternal, "failed to get workorder", err)
	}
	return count, nil
}

func (r *Repository) DeleteLabor(ctx context.Context, id int64) error {
	result := r.db.Client().WithContext(ctx).
		Delete(&models.Labor{}, "id = ?", id)
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to delete labor", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("labor with id %d does not exist", id), nil)
	}
	return nil
}

func (r *Repository) UpdateLabor(ctx context.Context, labor *domain.Labor) error {
	if labor == nil {
		return types.NewError(types.ErrValidation, "labor is nil", nil)
	}
	result := r.db.Client().WithContext(ctx).
		Model(&models.Labor{}).
		Where("id = ?", labor.ID).
		Updates(models.FromDomain(labor))
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to update labor", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("labor with id %d does not exist", labor.ID), nil)
	}
	return nil
}

func (r *Repository) ListLabor(ctx context.Context, page, perPage int, projectID int64) ([]domain.ListedLabor, int64, error) {
	var list []models.Labor
	var total int64

	db0 := r.db.Client().WithContext(ctx).Model(&models.Labor{})

	// Conteo total
	if err := db0.Count(&total).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to count labors", err)
	}

	if err := db0.
		Preload("Category").
		Select("id, name, contractor_name, price, category_id").
		Where("project_id = ?", projectID).
		Limit(perPage).
		Offset((page - 1) * perPage).
		Find(&list).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to list labor", err)
	}

	// Mapear a dominio ligero
	labors := make([]domain.ListedLabor, len(list))
	for i, labor := range list {
		labors[i] = domain.ListedLabor{
			ID:             labor.ID,
			Name:           labor.Name,
			Price:          labor.Price,
			ContractorName: labor.ContractorName,
			CategoryId:     labor.LaborCategoryID,
			CategoryName:   labor.Category.Name,
		}
	}

	return labors, total, nil
}

func (r *Repository) ListLaborCategoriesByTypeID(ctx context.Context, typeID int64) ([]domain.LaborCategory, error) {
	var laborCategoriesModels []models.LaborCategory
	db0 :=
		r.db.
			Client().
			WithContext(ctx).
			Model(&models.LaborCategory{}).
			Where("type_id = ?", typeID)

	if err := db0.Find(&laborCategoriesModels).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list labor categories", err)
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

func (r *Repository) ListByWorkorder(ctx context.Context, workorderID int64, usdMonth string) ([]domain.LaborRawItem, error) {
	var v3Models []models.LaborListItem

	// Usar la vista v3_labor_list como base y agregar campos adicionales
	err := r.db.Client().
		WithContext(ctx).
		Table("v3_labor_list AS v3").
		Select(`
            v3.workorder_id,
            v3.workorder_number,
            v3.date,
            v3.project_name,
            v3.field_name,
            COALESCE(v3.crop_name, '') AS crop_name,
            v3.labor_name,
            v3.contractor,
            v3.surface_ha,
            v3.cost_per_ha,
            COALESCE(v3.labor_category_name, '') AS category_name,
            COALESCE(v3.investor_name, '') AS investor_name,
			pdv.average_value AS usd_avg_value,
			i.id AS invoice_id,
			i.number AS invoice_number,
			i.company AS invoice_company,
			i.date AS invoice_date,
			i.status AS invoice_status
        `).
		Joins("LEFT JOIN invoices i ON i.work_order_id = v3.workorder_id").
		Joins("INNER JOIN project_dollar_values pdv ON pdv.project_id = v3.project_id AND pdv.month = ? AND pdv.deleted_at IS NULL", usdMonth).
		Where("v3.workorder_id = ?", workorderID).
		Scan(&v3Models).Error

	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list labors by workorder", err)
	}

	// Convertir a LaborRawItem para mantener compatibilidad
	raws := make([]domain.LaborRawItem, len(v3Models))
	for i, m := range v3Models {
		raws[i] = domain.LaborRawItem{
			WorkorderID:     m.WorkorderID,
			WorkorderNumber: m.WorkorderNumber,
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
			InvoiceID:       *m.InvoiceID,
			InvoiceNumber:   safeStringPtr(m.InvoiceNumber),
			InvoiceCompany:  safeStringPtr(m.InvoiceCompany),
			InvoiceDate:     m.InvoiceDate,
			InvoiceStatus:   safeStringPtr(m.InvoiceStatus),
		}
	}
	return raws, nil
}

func (r *Repository) ListGroupLabor(ctx context.Context, inp types.Input, projectID int64, fieldID int64, usdMonth string) ([]domain.LaborRawItem, types.PageInfo, error) {

	// Usar la vista v3_labor_list como base y agregar campos adicionales
	base := r.db.Client().
		WithContext(ctx).
		Table("v3_labor_list AS v3").
		Select(`
			v3.workorder_id,
			v3.workorder_number,
			v3.date,
			v3.project_id,
			v3.field_id,
			v3.project_name,
			v3.field_name,
			COALESCE(v3.crop_name, '') AS crop_name,
			v3.labor_name,
			COALESCE(v3.labor_category_name, '') AS category_name,
			v3.contractor,
			v3.surface_ha,
			v3.cost_per_ha,
			v3.contractor_name,
			COALESCE(v3.investor_name, '') AS investor_name,
			pdv.average_value AS usd_avg_value,
			i.id AS invoice_id,
			i.number AS invoice_number,
			i.company AS invoice_company,
			i.date AS invoice_date,
			i.status AS invoice_status
		`).
		Joins("LEFT JOIN invoices i ON i.work_order_id = v3.workorder_id").
		Joins("INNER JOIN project_dollar_values pdv ON pdv.project_id = v3.project_id AND pdv.month = ? AND pdv.deleted_at IS NULL", usdMonth)

	if fieldID != 0 {
		base = base.Where("v3.field_id = ?", fieldID)
	} else if projectID != 0 {
		base = base.Where("v3.project_id = ?", projectID)
	} else {
		return nil, types.PageInfo{}, types.NewError(types.ErrValidation,
			"fieldID or projectID is required", nil)
	}

	var total int64
	countQuery := base.Session(&gorm.Session{})
	if err := countQuery.Select("COUNT(*)").Count(&total).Error; err != nil {
		return nil, types.PageInfo{}, types.NewError(types.ErrInternal,
			"failed to count labors for workorder", err)
	}

	offset := (int(inp.Page) - 1) * int(inp.PageSize)

	var rows []models.LaborListItem
	if err := base.Order("v3.workorder_number DESC").
		Limit(int(inp.PageSize)).
		Offset(offset).
		Scan(&rows).Error; err != nil {
		return nil, types.PageInfo{}, types.NewError(types.ErrInternal,
			"failed to list grouped labors", err)
	}

	list := make([]domain.LaborRawItem, len(rows))
	for i, m := range rows {
		// Calcular valores de USD dinámicamente
		netTotal := m.SurfaceHa.Mul(m.CostPerHa)

		// Obtener porcentaje de IVA dinámicamente desde app_parameters
		ivaPercentage, err := r.getIVAPercentage(ctx)
		if err != nil {
			// Si hay error, usar valor por defecto y logear el error
			// TODO: Implementar logging apropiado
			ivaPercentage = decimal.NewFromFloat(0.105) // 10.5%
		}
		totalIVA := netTotal.Mul(ivaPercentage)

		usdCostHa := m.CostPerHa.Div(m.USDAvgValue)
		usdNetTotal := netTotal.Div(m.USDAvgValue)

		list[i] = domain.LaborRawItem{
			WorkorderID:     m.WorkorderID,
			WorkorderNumber: m.WorkorderNumber,
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
			InvoiceID:       *m.InvoiceID,
			InvoiceNumber:   safeStringPtr(m.InvoiceNumber),
			InvoiceCompany:  safeStringPtr(m.InvoiceCompany),
			InvoiceDate:     m.InvoiceDate,
			InvoiceStatus:   safeStringPtr(m.InvoiceStatus),
		}
	}

	pageInfo := types.NewPageInfo(int(inp.Page), int(inp.PageSize), total)
	return list, pageInfo, nil
}

// getIVAPercentage obtiene el porcentaje de IVA desde app_parameters
func (r *Repository) getIVAPercentage(ctx context.Context) (decimal.Decimal, error) {
	var value string
	err := r.db.Client().WithContext(ctx).
		Table("app_parameters").
		Select("value").
		Where("key = ?", "iva_percentage").
		Scan(&value).Error

	if err != nil {
		// Si no se encuentra el parámetro, usar valor por defecto
		return decimal.NewFromFloat(0.105), nil
	}

	ivaDecimal, err := decimal.NewFromString(value)
	if err != nil {
		// Si hay error al parsear, usar valor por defecto
		return decimal.NewFromFloat(0.105), nil
	}

	return ivaDecimal, nil
}

// safeStringPtr convierte un string pointer a string seguro
func safeStringPtr(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func (r *Repository) GetMetrics(ctx context.Context, f domain.LaborFilter) (*domain.LaborMetrics, error) {
	q := `
        SELECT 
          surface_ha,
          total_labor_cost,
          avg_labor_cost_per_ha
        FROM v3_labor_metrics
        WHERE 1=1
    `
	var args []any

	// Filtros: se decide el nivel de agrupación
	if f.ProjectID != nil && f.FieldID != nil {
		q += " AND project_id = ? AND field_id = ?"
		args = append(args, *f.ProjectID, *f.FieldID)
	} else if f.ProjectID != nil {
		q += " AND project_id = ?"
		args = append(args, *f.ProjectID)
	} else if f.FieldID != nil {
		q += " AND field_id = ?"
		args = append(args, *f.FieldID)
	}

	var row struct {
		SurfaceHa    decimal.Decimal `gorm:"column:surface_ha"`
		NetTotalCost decimal.Decimal `gorm:"column:total_labor_cost"`
		AvgCostPerHa decimal.Decimal `gorm:"column:avg_labor_cost_per_ha"`
	}

	if err := r.db.Client().WithContext(ctx).Raw(q, args...).Scan(&row).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get labor metrics", err)
	}

	return &domain.LaborMetrics{
		SurfaceHa:    row.SurfaceHa,
		NetTotalCost: row.NetTotalCost,
		AvgCostPerHa: row.AvgCostPerHa,
	}, nil
}
