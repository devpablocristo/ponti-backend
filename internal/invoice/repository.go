package invoice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alphacodinggroup/ponti-backend/internal/invoice/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/internal/invoice/usecases/domain"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"gorm.io/gorm"
)

type GormEnginePort interface {
	Client() *gorm.DB
}

type Repository struct {
	db GormEnginePort
}

func NewRepository(db GormEnginePort) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByWorkOrderID(ctx context.Context, workOrderID int64) (*domain.Invoice, error) {
	if workOrderID == 0 {
		return nil, types.NewError(types.ErrInvalidID, "invalid WorkOrderID", nil)
	}

	var row models.Invoice
	if err := r.db.Client().WithContext(ctx).Where("work_order_id = ?", workOrderID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, "There is no invoice for this work order", err)
		}
		return nil, types.NewError(types.ErrInternal, "Failed to find invoice", err)

	}
	out := row.ToDomain()

	return out, nil
}

func (r *Repository) Create(ctx context.Context, item *domain.Invoice) (int64, error) {
	if item.WorkOrderID == 0 {
		return 0, types.NewError(types.ErrInvalidID, "invalid WorkOrderID", nil)
	}

	m := models.FromDomain(item)
	if err := r.db.Client().WithContext(ctx).Create(&m).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "fail to create invoice", err)
	}
	return m.ID, nil
}

func (r *Repository) Update(ctx context.Context, item *domain.Invoice) error {
	if item.WorkOrderID == 0 {
		return types.NewError(types.ErrInvalidID, "invalid WorkOrderID", nil)
	}

	result := r.db.Client().WithContext(ctx).
		Where("work_order_id = ?", item.WorkOrderID).
		Model(models.Invoice{}).
		Updates(map[string]any{
			"number":     item.Number,
			"company":    item.Company,
			"date":       item.Date,
			"status":     item.Status,
			"updated_at": time.Now(),
			"updated_by": item.UpdatedBy,
		})

	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to update invoice", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("invoice for work order %d does not exist", item.WorkOrderID), nil)
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, workOrderID int64) error {
	if workOrderID == 0 {
		return types.NewError(types.ErrInvalidID, "invalid WorkOrderID", nil)
	}

	result := r.db.Client().WithContext(ctx).Where("work_order_id = ?", workOrderID).Delete(&models.Invoice{})
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to delete invoice", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("invoice for work order %d does not exist", workOrderID), nil)
	}
	return nil
}
