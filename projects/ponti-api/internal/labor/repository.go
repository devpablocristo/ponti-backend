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

func (r *Repository) ListLabor(ctx context.Context, page, perPage int, projectId int64) ([]domain.ListedLabor, int64, error) {
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
		Where("project_id = ?", projectId).
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

func (r *Repository) ListLaborCategoriesByTypeId(ctx context.Context, typeId int64) ([]domain.LaborCategory, error) {
	var laborCategoriesModels []models.LaborCategory
	db0 :=
		r.db.
			Client().
			WithContext(ctx).
			Model(&models.LaborCategory{}).
			Where("type_id = ?", typeId)

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
	var rawModels []models.LaborRawItem

	err := r.db.Client().
		WithContext(ctx).
		Table("workorders AS w").
		Select(`
            w.number                AS workorder_number,
            w.date                  AS date,
            p.name                  AS project_name,
            f.name                  AS field_name,
            c.name                  AS crop_name,
            lb.name                 AS labor_name,
            w.contractor            AS contractor,
            w.effective_area        AS effective_area,
            lb.price                AS price,
            lb.contractor_name      AS contractor_name,
            inv.name                AS investor_name,
			pdv.average_value       AS usd_avg_value,
			i.number                AS invoice_number,
			i.company               AS invoice_company,
			i.date                  AS invoice_date,
			i.status                AS invoice_status
        `).
		Joins("INNER JOIN projects p   ON w.project_id    = p.id").
		Joins("INNER JOIN fields f     ON w.field_id      = f.id").
		Joins("INNER JOIN crops c      ON w.crop_id       = c.id").
		Joins("INNER JOIN labors lb    ON w.labor_id      = lb.id").
		Joins("INNER JOIN investors inv ON w.investor_id  = inv.id").
		Joins("LEFT JOIN invoices i ON i.work_order_id = w.id").
		Joins("INNER JOIN project_dollar_values pdv ON pdv.project_id = w.project_id AND pdv.month = ? AND pdv.deleted_at IS NULL", usdMonth).
		Where("w.id = ?", workorderID).
		Scan(&rawModels).Error

	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list labors by workorder", err)
	}

	raws := make([]domain.LaborRawItem, len(rawModels))
	for i, m := range rawModels {
		raws[i] = domain.LaborRawItem{
			WorkorderNumber: m.WorkorderNumber,
			Date:            m.Date,
			ProjectName:     m.ProjectName,
			FieldName:       m.FieldName,
			CropName:        m.CropName,
			LaborName:       m.LaborName,
			Contractor:      m.Contractor,
			SurfaceHa:       m.SurfaceHa,
			CostHa:          m.CostHa,
			CategoryName:    m.CategoryName,
			InvestorName:    m.InvestorName,
			USDAvgValue:     m.USDAvgValue,
			InvoiceNumber:   m.InvoiceNumber,
			InvoiceCompany:  m.InvoiceCompany,
			InvoiceDate:     m.InvoiceDate,
			InvoiceStatus:   m.InvoiceStatus,
		}
	}
	return raws, nil
}

func (r *Repository) ListGroupLabor(ctx context.Context, inp types.Input, projectID int64, fieldID int64, usdMonth string) ([]domain.LaborRawItem, types.PageInfo, error) {

	// Usar la nueva vista fix_labors_list que ya tiene todos los cálculos correctos
	base := r.db.Client().
		WithContext(ctx).
		Table("fix_labors_list AS fll")

	if fieldID != 0 {
		base = base.Where("fll.field_id = ?", fieldID)
	} else if projectID != 0 {
		base = base.Where("fll.project_id = ?", projectID)
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

	var rows []models.LaborRawItem
	// Usar la vista que ya tiene todos los cálculos correctos
	if err := base.Select(`
		workorder_id,
		workorder_number,
		date,
		project_name,
		field_name,
		crop_name,
		labor_name,
		category_name,
		contractor,
		surface_ha,
		cost_ha,
		contractor_name,
		investor_name,
		usd_avg_value,
		net_total,
		total_iva,
		usd_cost_ha,
		usd_net_total
	`).Where("1=1").
		Order("workorder_number DESC").
		Limit(int(inp.PageSize)).
		Offset(offset).
		Scan(&rows).Error; err != nil {
		return nil, types.PageInfo{}, types.NewError(types.ErrInternal,
			"failed to list grouped labors", err)
	}

	list := make([]domain.LaborRawItem, len(rows))
	for i, m := range rows {
		list[i] = domain.LaborRawItem{
			WorkorderID:     m.WorkorderID,
			WorkorderNumber: m.WorkorderNumber,
			Date:            m.Date,
			ProjectName:     m.ProjectName,
			FieldName:       m.FieldName,
			CropName:        m.CropName,
			LaborName:       m.LaborName,
			Contractor:      m.Contractor,
			SurfaceHa:       m.SurfaceHa,
			CostHa:          m.CostHa,
			CategoryName:    m.CategoryName,
			InvestorName:    m.InvestorName,
			USDAvgValue:     m.USDAvgValue,
			NetTotal:        m.NetTotal,    // ✅ Viene de la vista (ya calculado)
			TotalIVA:        m.TotalIVA,    // ✅ Viene de la vista (ya calculado)
			USDCostHa:       m.USDCostHa,   // ✅ Viene de la vista (ya calculado)
			USDNetTotal:     m.USDNetTotal, // ✅ Viene de la vista (ya calculado)
		}
	}

	pageInfo := types.NewPageInfo(int(inp.Page), int(inp.PageSize), total)
	return list, pageInfo, nil
}

func (r *Repository) GetMetrics(ctx context.Context, f domain.LaborFilter) (*domain.LaborMetrics, error) {
	q := `
        SELECT 
          surface_ha,
          total_labor_cost AS net_total_cost,
          labor_cost_per_ha AS avg_cost_per_ha
        FROM labor_cards_cube_view
        WHERE 1=1
    `
	var args []any

	// Filtros: se decide el nivel de agrupación
	if f.ProjectID != nil && f.FieldID != nil {
		q += " AND project_id = ? AND field_id = ? AND level = 'project+field'"
		args = append(args, *f.ProjectID, *f.FieldID)
	} else if f.ProjectID != nil {
		q += " AND project_id = ? AND level = 'project'"
		args = append(args, *f.ProjectID)
	} else if f.FieldID != nil {
		q += " AND field_id = ? AND level = 'field'"
		args = append(args, *f.FieldID)
	} else {
		q += " AND level = 'global'"
	}

	var row struct {
		SurfaceHa    decimal.Decimal `gorm:"column:surface_ha"`
		NetTotalCost decimal.Decimal `gorm:"column:net_total_cost"`
		AvgCostPerHa decimal.Decimal `gorm:"column:avg_cost_per_ha"`
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
