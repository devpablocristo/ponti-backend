package dto

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
)

type UpdateRealStockRequest struct {
	RealStockUnits int64  `json:"real_stock_units"`
	UpdatedBy      *int64 `json:"updated_by"`
}

func (r *UpdateRealStockRequest) ToDomain() *domain.Stock {
	return &domain.Stock{
		RealStockUnits: r.RealStockUnits,
		Base: shareddomain.Base{
			UpdatedBy: r.UpdatedBy,
		},
	}
}

type UpdateRealStockResponse struct {
	Message string `json:"message"`
}
