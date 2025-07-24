package dto

import (
	"time"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/commercialization/usecases/domain"
	decimal "github.com/shopspring/decimal"
)

type CommercializationResponse struct {
	CropName       string          `json:"crop_name"`
	BoardPrice     decimal.Decimal `json:"board_price"`
	FreightCost    decimal.Decimal `json:"freight_cost"`
	CommercialCost float64         `json:"commercial_cost"`
	NetPrice       decimal.Decimal `json:"net_price"`

	CreatedAt time.Time `json:"created_at"`
}

func FromDomain(d *domain.CropCommercialization) CommercializationResponse {
	return CommercializationResponse{
		CropName:       d.CropName,
		BoardPrice:     d.BoardPrice,
		FreightCost:    d.FreightCost,
		CommercialCost: d.CommercialCost,
		NetPrice:       d.NetPrice,
		CreatedAt:      d.CreatedAt,
	}
}
