package dto

import (
	"time"

	domain "github.com/devpablocristo/ponti-backend/internal/invoice/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
)

type InvoiceRequest struct {
	Number  string    `json:"number" binding:"required"`
	Company string    `json:"company" binding:"required"`
	Date    time.Time `json:"date" binding:"required"`
	Status  string    `json:"status" binding:"required"`
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
