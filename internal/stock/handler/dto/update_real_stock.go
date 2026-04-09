package dto

import (
	"github.com/shopspring/decimal"

	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	"github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
)

type UpdateRealStockRequest struct {
	RealStockUnits decimal.Decimal `json:"real_stock_units"`
}

func (r *UpdateRealStockRequest) ToDomain(updatedBy *string) *domain.Stock {
	return &domain.Stock{
		RealStockUnits: r.RealStockUnits,
		Base: shareddomain.Base{
			UpdatedBy: updatedBy,
		},
	}
}

type UpdateRealStockResponse struct {
	Message string `json:"message"`
}

func NewUpdateRealStockResponse(message string) *UpdateRealStockResponse {
	return &UpdateRealStockResponse{
		Message: message,
	}
}
