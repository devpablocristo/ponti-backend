package dto

import (
	"time"

	"github.com/devpablocristo/core/errors/go/domainerr"
	"github.com/shopspring/decimal"

	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	stockdomain "github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
)

type CreateStockCountRequest struct {
	CountedUnits decimal.Decimal `json:"counted_units"`
	CountedAt    time.Time       `json:"counted_at"`
	Note         string          `json:"note"`
}

func (r *CreateStockCountRequest) Validate() error {
	if r.CountedAt.IsZero() {
		return domainerr.Validation("counted_at is required")
	}
	if r.CountedUnits.LessThan(decimal.Zero) {
		return domainerr.Validation("counted_units must be greater than or equal to 0")
	}
	return nil
}

func (r *CreateStockCountRequest) ToDomain(createdBy *string) *stockdomain.StockCount {
	now := time.Now().UTC()
	return &stockdomain.StockCount{
		CountedUnits: r.CountedUnits,
		CountedAt:    r.CountedAt,
		Note:         r.Note,
		Base: shareddomain.Base{
			CreatedAt: now,
			UpdatedAt: now,
			CreatedBy: createdBy,
			UpdatedBy: createdBy,
		},
	}
}

type CreateStockCountResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func NewCreateStockCountResponse(id int64, message string) *CreateStockCountResponse {
	return &CreateStockCountResponse{
		ID:      id,
		Message: message,
	}
}
