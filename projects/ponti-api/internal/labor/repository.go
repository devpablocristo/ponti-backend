package labor

import (
	"context"
	"fmt"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/repository/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/usecases/domain"
	"gorm.io/gorm"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
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

func (r *Repository) deleteLabor(ctx context.Context, id int64) error {
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

	base := r.db.Client().
		WithContext(ctx).
		Table("workorders AS w").
		Joins("INNER JOIN projects p    ON w.project_id   = p.id").
		Joins("INNER JOIN fields f      ON w.field_id     = f.id").
		Joins("INNER JOIN crops c       ON w.crop_id      = c.id").
		Joins("INNER JOIN labors lb     ON w.labor_id     = lb.id").
		Joins("INNER JOIN categories lc ON lb.category_id = lc.id").
		Joins("INNER JOIN investors inv ON w.investor_id  = inv.id").
		Joins("LEFT JOIN invoices i ON i.work_order_id = w.id").
		Joins("LEFT JOIN project_dollar_values pdv ON pdv.project_id = w.project_id AND pdv.month = ? AND pdv.deleted_at IS NULL", usdMonth)

	if fieldID != 0 {
		base = base.Where("w.field_id = ?", fieldID)
	} else if projectID != 0 {
		base = base.Where("w.project_id = ?", projectID)
	} else {
		return nil, types.PageInfo{}, types.NewError(types.ErrValidation,
			"fieldID or projectID is required", nil)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, types.PageInfo{}, types.NewError(types.ErrInternal,
			"failed to count labors for workorder", err)
	}

	offset := (int(inp.Page) - 1) * int(inp.PageSize)

	var rows []models.LaborRawItem
	if err := base.Select(`
			w.id AS workorder_id,
            w.number                AS workorder_number,
            w.date                  AS date,
            p.name                  AS project_name,
            f.name                  AS field_name,
            c.name                  AS crop_name,
            lb.name                 AS labor_name,
            lc.name                 AS category_name,
            w.contractor            AS contractor,
            w.effective_area        AS effective_area,
            lb.price                AS price,
            lb.contractor_name      AS contractor_name,
            inv.name                AS investor_name,
			pdv.average_value       AS usd_avg_value,
			i.id                    AS invoice_id,
			i.number                AS invoice_number,
			i.company               AS invoice_company,
			i.date                  AS invoice_date,
			i.status                AS invoice_status
        `).Order("w.number DESC").
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
			InvoiceID:       m.InvoiceID,
			InvoiceNumber:   m.InvoiceNumber,
			InvoiceCompany:  m.InvoiceCompany,
			InvoiceDate:     m.InvoiceDate,
			InvoiceStatus:   m.InvoiceStatus,
		}
	}

	pageInfo := types.NewPageInfo(int(inp.Page), int(inp.PageSize), total)
	return list, pageInfo, nil
}
