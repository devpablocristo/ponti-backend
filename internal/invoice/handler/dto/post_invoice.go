package dto

import (
	"strings"
	"time"

	"github.com/devpablocristo/core/backend/go/domainerr"

	domain "github.com/devpablocristo/ponti-backend/internal/invoice/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
)

type InvoiceRequest struct {
	Number  string    `json:"number" binding:"required"`
	Company string    `json:"company" binding:"required"`
	Date    time.Time `json:"date" binding:"required"`
	Status  string    `json:"status" binding:"required"`
}

func (ir *InvoiceRequest) Validate() error {
	if strings.TrimSpace(ir.Number) == "" {
		return domainerr.Validation("The field 'number' is required")
	}
	if strings.TrimSpace(ir.Company) == "" {
		return domainerr.Validation("The field 'company' is required")
	}
	var timeZero time.Time
	if ir.Date.Equal(timeZero) {
		return domainerr.Validation("The field 'date' is required")
	}
	if strings.TrimSpace(ir.Status) == "" {
		return domainerr.Validation("The field 'status' is required")
	}
	return nil
}

func (ir *InvoiceRequest) ToDomain(workOrderID int64, userID string) *domain.Invoice {
	return &domain.Invoice{
		WorkOrderID: workOrderID,
		Number:      ir.Number,
		Company:     ir.Company,
		Date:        ir.Date,
		Status:      ir.Status,
		Base: shareddomain.Base{
			CreatedBy: &userID,
			UpdatedBy: &userID,
		},
	}
}
