package dto

import (
	"github.com/shopspring/decimal"
)

// UpdateRealStockRequest es el body del conteo manual de stock de campo. El handler aplica
// last-write-wins (ignora cualquier updated_at del cliente), por eso no hay campo de versión.
type UpdateRealStockRequest struct {
	RealStockUnits decimal.Decimal `json:"real_stock_units"`
}

type UpdateRealStockResponse struct {
	Message string `json:"message"`
}

func NewUpdateRealStockResponse(message string) *UpdateRealStockResponse {
	return &UpdateRealStockResponse{
		Message: message,
	}
}
