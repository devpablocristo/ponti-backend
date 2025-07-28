package dto

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/commercialization/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	decimal "github.com/shopspring/decimal"
)

type CropCommercialization struct {
	CropName       string          `json:"crop_name" binding:"required"`
	BoardPrice     decimal.Decimal `json:"board_price" binding:"required"`
	FreightCost    decimal.Decimal `json:"freight_cost" binding:"required"`
	CommercialCost float64         `json:"commercial_cost" binding:"required"`
}

type BulkCommercializationRequest struct {
	Values []CropCommercialization `json:"values" binding:"required,dive"`
}

func (b *BulkCommercializationRequest) ToDomainSlice(projecID int64, userID int64) []domain.CropCommercialization {
	out := make([]domain.CropCommercialization, len(b.Values))
	for i, item := range b.Values {
		out[i] = domain.CropCommercialization{
			ProjectID:      projecID,
			CropName:       item.CropName,
			BoardPrice:     item.BoardPrice,
			FreightCost:    item.FreightCost,
			CommercialCost: item.CommercialCost,
			Base: shareddomain.Base{
				CreatedBy: &userID,
			},
		}
	}
	return out
}
