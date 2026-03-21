package dto

import (
	"time"

	"github.com/devpablocristo/core/backend/go/domainerr"

	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	"github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
)

type UpdateCloseDateRequest struct {
	CloseDate time.Time `json:"close_date"`
}

func (r *UpdateCloseDateRequest) ToDomain(updateBy *string) *domain.Stock {
	return &domain.Stock{
		CloseDate: &r.CloseDate,
		Base: shareddomain.Base{
			UpdatedBy: updateBy,
		},
	}
}

type UpdateCloseDateResponse struct {
	Message string `json:"message"`
}

func (r *UpdateCloseDateRequest) Validate() error {
	var timeZero time.Time
	if r.CloseDate.Equal(timeZero) {
		return domainerr.Validation("close_date is required")
	}
	return nil
}
