package dto

import (
	"time"

	"github.com/shopspring/decimal"

	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	"github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
)

type UpdateRealStockRequest struct {
	RealStockUnits decimal.Decimal `json:"real_stock_units"`
	UpdatedAt      *time.Time      `json:"updated_at,omitempty"`
}

func (r *UpdateRealStockRequest) ToDomain(updatedBy *string) *domain.Stock {
	stock := &domain.Stock{
		RealStockUnits: r.RealStockUnits,
		Base: shareddomain.Base{
			UpdatedBy: updatedBy,
		},
	}
	if r.UpdatedAt != nil {
		stock.UpdatedAt = *r.UpdatedAt
	}
	return stock
}

type UpdateRealStockResponse struct {
	Message string `json:"message"`
}

func NewUpdateRealStockResponse(message string) *UpdateRealStockResponse {
	return &UpdateRealStockResponse{
		Message: message,
	}
}
