package dto

import (
	"time"

	domain "github.com/devpablocristo/ponti-backend/internal/invoice/usecases/domain"
)

type InvoiceResponse struct {
	ID          int64     `json:"id"`
	WorkOrderID int64     `json:"work_order_id"`
	InvestorID  int64     `json:"investor_id"`
	Number      string    `json:"number"`
	Company     string    `json:"company"`
	Date        time.Time `json:"date"`
	Status      string    `json:"status"`

	CreateAt  time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func FromDomain(d *domain.Invoice) InvoiceResponse {
	return InvoiceResponse{
		ID:          d.ID,
		WorkOrderID: d.WorkOrderID,
		InvestorID:  d.InvestorID,
		Number:      d.Number,
		Company:     d.Company,
		Date:        d.Date,
		Status:      d.Status,
		CreateAt:    d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}
