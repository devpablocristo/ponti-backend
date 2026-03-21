package invoice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/devpablocristo/core/saas/go/shared/domainerr"
	"github.com/devpablocristo/ponti-backend/internal/invoice/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/invoice/usecases/domain"
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
		return nil, domainerr.Validation("invalid WorkOrderID")
	}

	var row models.Invoice
	if err := r.db.Client().WithContext(ctx).Where("work_order_id = ?", workOrderID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerr.NotFound("There is no invoice for this work order")
		}
		return nil, domainerr.Internal("Failed to find invoice")

	}
	out := row.ToDomain()

	return out, nil
}

func (r *Repository) Create(ctx context.Context, item *domain.Invoice) (int64, error) {
	if item.WorkOrderID == 0 {
		return 0, domainerr.Validation("invalid WorkOrderID")
	}

	m := models.FromDomain(item)
	if err := r.db.Client().WithContext(ctx).Create(&m).Error; err != nil {
		return 0, domainerr.Internal("fail to create invoice")
	}
	return m.ID, nil
}

func (r *Repository) Update(ctx context.Context, item *domain.Invoice) error {
	if item.WorkOrderID == 0 {
		return domainerr.Validation("invalid WorkOrderID")
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
		return domainerr.Internal("failed to update invoice")
	}
	if result.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("invoice for work order %d does not exist", item.WorkOrderID))
	}

	return nil
}

func (r *Repository) ListByProjectID(ctx context.Context, projectID int64, page, perPage int) ([]domain.Invoice, int64, error) {
	if projectID == 0 {
		return nil, 0, domainerr.Validation("invalid projectID")
	}

	var total int64
	query := r.db.Client().WithContext(ctx).
		Model(&models.Invoice{}).
		Joins("JOIN work_orders ON work_orders.id = invoices.work_order_id").
		Where("work_orders.project_id = ?", projectID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count invoices")
	}

	var rows []models.Invoice
	offset := (page - 1) * perPage
	if err := query.Offset(offset).Limit(perPage).Order("invoices.id DESC").Find(&rows).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to list invoices")
	}

	out := make([]domain.Invoice, len(rows))
	for i, row := range rows {
		out[i] = *row.ToDomain()
	}

	return out, total, nil
}

func (r *Repository) Delete(ctx context.Context, workOrderID int64) error {
	if workOrderID == 0 {
		return domainerr.Validation("invalid WorkOrderID")
	}

	result := r.db.Client().WithContext(ctx).Where("work_order_id = ?", workOrderID).Delete(&models.Invoice{})
	if result.Error != nil {
		return domainerr.Internal("failed to delete invoice")
	}
	if result.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("invoice for work order %d does not exist", workOrderID))
	}
	return nil
}
