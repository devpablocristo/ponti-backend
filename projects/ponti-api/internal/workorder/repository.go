package workorder

import (
	"context"
	"errors"

	"gorm.io/gorm"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/repository/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/usecases/domain"
)

type GormEngine interface {
	Client() *gorm.DB
}

// Repository implementa domain.RepositoryPort
type Repository struct {
	db GormEngine
}

func NewRepository(db GormEngine) *Repository {
	return &Repository{db: db}
}

// CreateWorkOrder crea la orden y sus ítems en una transacción
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

// GetOrder obtiene la orden con sus ítems (solo lectura)
func (r *Repository) GetOrder(ctx context.Context, number string) (*domain.WorkOrder, error) {
	var ord models.WorkOrder
	if err := r.db.Client().WithContext(ctx).
		Preload("Items").
		First(&ord, "number = ?", number).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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

// UpdateWorkOrder actualiza cabecera e ítems en una sola transacción
func (r *Repository) UpdateWorkOrder(ctx context.Context, o *domain.WorkOrder) error {
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1) Verificar existencia
		var count int64
		if err := tx.Model(&models.WorkOrder{}).
			Where("number = ?", o.Number).
			Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check order existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, "order not found", nil)
		}

		// 2) Actualizar campos de la orden
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

		// 3) Reemplazar ítems: eliminar antiguos y crear nuevos
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

// DeleteWorkOrder borra la orden e ítems en una transacción
func (r *Repository) DeleteWorkOrder(ctx context.Context, number string) error {
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1) Verificar existencia
		var count int64
		if err := tx.Model(&models.WorkOrder{}).
			Where("number = ?", number).
			Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check order existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, "order not found", nil)
		}

		// 2) Eliminar ítems
		if err := tx.Where("order_number = ?", number).
			Delete(&models.WorkOrderItem{}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to delete order items", err)
		}

		// 3) Eliminar orden
		if err := tx.Delete(&models.WorkOrder{}, "number = ?", number).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to delete order", err)
		}
		return nil
	})
}

// DuplicateOrder reutiliza CreateWorkOrder, que internamente maneja su propia transacción
func (r *Repository) DuplicateOrder(ctx context.Context, number string) (string, error) {
	orig, err := r.GetOrder(ctx, number)
	if err != nil {
		return "", err
	}
	// TODO: lógica para generar nuevo número (p.ej. secuencia +1)
	newNumber := "0000-0002"
	orig.Number = newNumber
	return r.CreateWorkOrder(ctx, orig)
}
