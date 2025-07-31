package workorder

import (
	"context"
	"errors"
	"fmt"

	gorm "gorm.io/gorm"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/repository/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/usecases/domain"
)

type GormEngine interface {
	Client() *gorm.DB
}

type Repository struct {
	db GormEngine
}

func NewRepository(db GormEngine) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateWorkorder(
	ctx context.Context,
	o *domain.Workorder,
) (string, error) {
	ordModel := models.FromDomain(o)
	err := r.db.Client().WithContext(ctx).
		Transaction(func(tx *gorm.DB) error {
			if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).
				Create(ordModel).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to create order", err)
			}
			return nil
		})
	return ordModel.Number, err
}

func (r *Repository) GetWorkorder(ctx context.Context, number string) (*domain.Workorder, error) {
	var ordModel models.Workorder
	if err := r.db.Client().WithContext(ctx).
		Preload("Items").
		First(&ordModel, "number = ?", number).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, "order not found", err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get order", err)
	}

	return ordModel.ToDomain(), nil
}

func (r *Repository) UpdateWorkorder(ctx context.Context, o *domain.Workorder) error {
	ordModel := models.FromDomain(o)

	return r.db.Client().WithContext(ctx).
		Transaction(func(tx *gorm.DB) error {
			// verificamos existencia
			var count int64
			if err := tx.Model(&models.Workorder{}).
				Where("number = ?", ordModel.Number).
				Count(&count).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to check order existence", err)
			}
			if count == 0 {
				return types.NewError(types.ErrNotFound, "order not found", nil)
			}

			// actualizamos campos y asociaciones de una vez
			if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).
				Updates(ordModel).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to update order", err)
			}
			return nil
		})
}

func (r *Repository) DeleteWorkorder(ctx context.Context, number string) error {
	return r.db.Client().WithContext(ctx).
		Transaction(func(tx *gorm.DB) error {
			// verificar si existe
			var count int64
			if err := tx.Model(&models.Workorder{}).
				Where("number = ?", number).
				Count(&count).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to check order existence", err)
			}
			if count == 0 {
				return types.NewError(types.ErrNotFound, "order not found", nil)
			}

			// eliminamos items y luego la orden
			if err := tx.Where("order_number = ?", number).
				Delete(&models.WorkorderItem{}).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to delete order items", err)
			}
			if err := tx.Delete(&models.Workorder{}, "number = ?", number).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to delete order", err)
			}
			return nil
		})
}

func (r *Repository) DuplicateWorkorder(ctx context.Context, number string) (string, error) {
	// reuso Get + Create
	orig, err := r.GetWorkorder(ctx, number)
	if err != nil {
		return "", err
	}
	newNum, err := getNextNumber(ctx, r.db.Client())
	if err != nil {
		return "", err
	}
	orig.Number = newNum
	return r.CreateWorkorder(ctx, orig)
}

func getNextNumber(ctx context.Context, db *gorm.DB) (string, error) {
	var seq int64
	if err := db.WithContext(ctx).
		Raw("SELECT nextval('workorder_number_seq')").
		Scan(&seq).Error; err != nil {
		return "", types.NewError(types.ErrInternal, "failed to get next work order number from sequence", err)
	}
	return fmt.Sprintf("%04d", seq), nil
}

func (r *Repository) ListWorkorders(
	ctx context.Context,
	filt domain.WorkorderFilter,
	inp types.Input,
) ([]domain.WorkorderListElement, types.PageInfo, error) {
	db := r.db.Client().WithContext(ctx)
	if filt.ProjectID != nil {
		db = db.Where("project_id = ?", *filt.ProjectID)
	}
	if filt.FieldID != nil {
		db = db.Where("field_id = ?", *filt.FieldID)
	}

	// 1) contar
	var total int64
	if err := db.Model(&models.WorkorderListElement{}).
		Count(&total).Error; err != nil {
		return nil, types.PageInfo{}, types.NewError(types.ErrInternal,
			"failed to count workorder list elements", err)
	}

	// 2) obtener página desde la vista
	offset := (int(inp.Page) - 1) * int(inp.PageSize)
	var rows []models.WorkorderListElement
	if err := db.
		Limit(int(inp.PageSize)).
		Offset(offset).
		Order("number desc").
		Find(&rows).Error; err != nil {
		return nil, types.PageInfo{}, types.NewError(types.ErrInternal,
			"failed to list workorder list elements", err)
	}

	// 3) mapear al domain
	list := make([]domain.WorkorderListElement, len(rows))
	for i, m := range rows {
		list[i] = domain.WorkorderListElement{
			Number:        m.Number,
			ProjectName:   m.ProjectName,
			FieldName:     m.FieldName,
			LotName:       m.LotName,
			Date:          m.Date,
			CropName:      m.CropName,
			LaborName:     m.LaborName,
			ClassTypeName: m.ClassTypeName,
			Contractor:    m.Contractor,
			SurfaceHa:     m.SurfaceHa,
			SupplyName:    m.SupplyName,
			Consumption:   m.Consumption,
			CategoryName:  m.CategoryName,
			Dose:          m.Dose,
			CostPerHa:     m.CostPerHa,
			UnitPrice:     m.UnitPrice,
			TotalCost:     m.TotalCost,
		}
	}

	pageInfo := types.NewPageInfo(int(inp.Page), int(inp.PageSize), total)
	return list, pageInfo, nil
}
