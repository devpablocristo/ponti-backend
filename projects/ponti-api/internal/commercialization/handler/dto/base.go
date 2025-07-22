package dto

import (
	"time"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/commercialization/usecases/domain"
)

type CommercializationResponse struct {
	CropName       string  `json:"crop_name"`
	BoardPrice     float64 `json:"board_price"`
	FreightCost    float64 `json:"freight_cost"`
	CommercialCost float64 `json:"commercial_cost"`
	NetPrice       float64 `json:"net_price"`

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
