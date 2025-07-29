package dto

import (
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/usecases/domain"
	"time"
)

type UpdateCloseDateRequest struct {
	CloseDate time.Time `json:"close_date"`
	UpdatedBy *int64    `json:"updated_by"`
}

func (r *UpdateCloseDateRequest) ToDomain() *domain.Stock {
	return &domain.Stock{
		CloseDate: &r.CloseDate,
		Base: shareddomain.Base{
			UpdatedBy: r.UpdatedBy,
		},
	}
}

type UpdateCloseDateResponse struct {
	Message string `json:"message"`
}
