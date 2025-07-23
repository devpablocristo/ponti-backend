package workorder

import (
	"context"
	"fmt"

	"gorm.io/gorm"

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

func (r *Repository) CreateWorkOrder(ctx context.Context, o *domain.WorkOrder) (string, error) {
	ord := models.WorkOrder{
		Number:       o.Number,
		ProjectID:    o.ProjectID,
		FieldID:      o.FieldID,
		LotID:        o.LotID,
		CropID:       o.CropID,
		LaborID:      o.LaborID,
		Contractor:   o.Contractor,
		Observations: o.Observations,
	}
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&ord).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to create order", err)
		}
		for _, it := range o.Items {
			item := models.WorkOrderItem{
				WorkOrderNumber: ord.Number,
				SupplyID:        it.SupplyID,
				TotalUsed:       it.TotalUsed,
				EffectiveArea:   it.EffectiveArea,
				FinalDose:       it.FinalDose,
			}
			if err := tx.Create(&item).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to create order item", err)
			}
		}
		return nil
	})
	return ord.Number, err
}

func (r *Repository) GetWorkOrder(ctx context.Context, number string) (*domain.WorkOrder, error) {
	var ord models.WorkOrder
	if err := r.db.Client().WithContext(ctx).
		Preload("Items").
		First(&ord, "number = ?", number).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.NewError(types.ErrNotFound, "order not found", err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get order", err)
	}
	items := make([]domain.WorkOrderItem, len(ord.Items))
	for i, it := range ord.Items {
		items[i] = domain.WorkOrderItem{
			SupplyID:      it.SupplyID,
			TotalUsed:     it.TotalUsed,
			EffectiveArea: it.EffectiveArea,
			FinalDose:     it.FinalDose,
		}
	}
	return &domain.WorkOrder{
		Number:       ord.Number,
		ProjectID:    ord.ProjectID,
		FieldID:      ord.FieldID,
		LotID:        ord.LotID,
		CropID:       ord.CropID,
		LaborID:      ord.LaborID,
		Contractor:   ord.Contractor,
		Observations: ord.Observations,
		Items:        items,
	}, nil
}

func (r *Repository) UpdateWorkOrder(ctx context.Context, o *domain.WorkOrder) error {
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.WorkOrder{}).
			Where("number = ?", o.Number).
			Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check order existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, "order not found", nil)
		}

		if err := tx.Model(&models.WorkOrder{}).
			Where("number = ?", o.Number).
			Updates(map[string]any{
				"project_id":   o.ProjectID,
				"field_id":     o.FieldID,
				"lot_id":       o.LotID,
				"crop_id":      o.CropID,
				"labor_id":     o.LaborID,
				"contractor":   o.Contractor,
				"observations": o.Observations,
			}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to update order", err)
		}

		if err := tx.Where("order_number = ?", o.Number).
			Delete(&models.WorkOrderItem{}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to delete old order items", err)
		}
		for _, it := range o.Items {
			item := models.WorkOrderItem{
				WorkOrderNumber: o.Number,
				SupplyID:        it.SupplyID,
				TotalUsed:       it.TotalUsed,
				EffectiveArea:   it.EffectiveArea,
				FinalDose:       it.FinalDose,
			}
			if err := tx.Create(&item).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to create updated order item", err)
			}
		}
		return nil
	})
}

func (r *Repository) DeleteWorkOrder(ctx context.Context, number string) error {
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.WorkOrder{}).
			Where("number = ?", number).
			Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check order existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, "order not found", nil)
		}

		if err := tx.Where("order_number = ?", number).
			Delete(&models.WorkOrderItem{}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to delete order items", err)
		}
		if err := tx.Delete(&models.WorkOrder{}, "number = ?", number).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to delete order", err)
		}
		return nil
	})
}

func (r *Repository) DuplicateWorkOrder(ctx context.Context, number string) (string, error) {
	orig, err := r.GetWorkOrder(ctx, number)
	if err != nil {
		return "", err
	}
	newSeq, err := getNextNumber(ctx, r.db.Client())
	if err != nil {
		return "", err
	}
	orig.Number = newSeq
	return r.CreateWorkOrder(ctx, orig)
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
