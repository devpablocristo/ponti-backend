package invoice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/devpablocristo/core/errors/go/domainerr"
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

func (r *Repository) GetByWorkOrderAndInvestor(ctx context.Context, workOrderID int64, investorID int64) (*domain.Invoice, error) {
	if workOrderID == 0 {
		return nil, domainerr.Validation("invalid WorkOrderID")
	}
	if investorID == 0 {
		return nil, domainerr.Validation("invalid InvestorID")
	}

	var row models.Invoice
	if err := r.db.Client().WithContext(ctx).
	Where("work_order_id = ? AND (investor_id = ? OR investor_id IS NULL)", workOrderID, investorID).
	Order(fmt.Sprintf("CASE WHEN investor_id = %d THEN 0 ELSE 1 END, id DESC", investorID)).
	First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerr.NotFound("There is no invoice for this work order and investor")
		}
		return nil, domainerr.Internal("Failed to find invoice")
	}

	return row.ToDomain(), nil
}

func (r *Repository) Create(ctx context.Context, item *domain.Invoice) (int64, error) {
	if item.WorkOrderID == 0 {
		return 0, domainerr.Validation("invalid WorkOrderID")
	}
	if item.InvestorID == 0 {
		return 0, domainerr.Validation("invalid InvestorID")
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
	if item.InvestorID == 0 {
		return domainerr.Validation("invalid InvestorID")
	}

	result := r.db.Client().WithContext(ctx).
		Where("work_order_id = ? AND investor_id = ?", item.WorkOrderID, item.InvestorID).
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
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf(
			"invoice for work order %d and investor %d does not exist",
			item.WorkOrderID, item.InvestorID,
		))
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
		Joins("JOIN workorders ON workorders.id = invoices.work_order_id").
		Where("workorders.project_id = ?", projectID)

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

func (r *Repository) Delete(ctx context.Context, workOrderID int64, investorID int64) error {
	if workOrderID == 0 {
		return domainerr.Validation("invalid WorkOrderID")
	}
	if investorID == 0 {
		return domainerr.Validation("invalid InvestorID")
	}

	result := r.db.Client().WithContext(ctx).
		Where("work_order_id = ? AND investor_id = ?", workOrderID, investorID).
		Delete(&models.Invoice{})

	if result.Error != nil {
		return domainerr.Internal("failed to delete invoice")
	}
	if result.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf(
			"invoice for work order %d and investor %d does not exist",
			workOrderID, investorID,
		))
	}
	return nil
}

func (r *Repository) InvestorBelongsToWorkOrder(ctx context.Context, workOrderID int64, investorID int64) (bool, error) {
	if workOrderID == 0 {
		return false, domainerr.Validation("invalid WorkOrderID")
	}
	if investorID == 0 {
		return false, domainerr.Validation("invalid InvestorID")
	}

	type resultRow struct {
		IsValid bool `gorm:"column:is_valid"`
	}

	var row resultRow
	query := `
		WITH split_count AS (
			SELECT COUNT(*) AS total
			FROM workorder_investor_splits
			WHERE workorder_id = ?
			  AND deleted_at IS NULL
		)
		SELECT CASE
			WHEN (SELECT total FROM split_count) > 0 THEN EXISTS (
				SELECT 1
				FROM workorder_investor_splits
				WHERE workorder_id = ?
				  AND investor_id = ?
				  AND deleted_at IS NULL
			)
			ELSE EXISTS (
				SELECT 1
				FROM workorders
				WHERE id = ?
				  AND investor_id = ?
				  AND deleted_at IS NULL
			)
		END AS is_valid
	`

	if err := r.db.Client().WithContext(ctx).
		Raw(query, workOrderID, workOrderID, investorID, workOrderID, investorID).
		Scan(&row).Error; err != nil {
		return false, domainerr.Internal("failed to validate invoice investor")
	}

	return row.IsValid, nil
}

