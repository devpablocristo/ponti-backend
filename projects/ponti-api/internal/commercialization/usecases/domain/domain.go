package domain

import (
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	"github.com/shopspring/decimal"
)

type CropCommercialization struct {
	ID             int64           // ID de cada registro
	ProjectID      int64           // ID del Proyecto
	CropName       string          // nombre del cultivo
	BoardPrice     decimal.Decimal // Precio en pizarra
	FreightCost    decimal.Decimal // Costo de flete
	CommercialCost float64         // Gasto comerciales (%)
	NetPrice       decimal.Decimal // Precio neto

	shareddomain.Base
}

func (cc *CropCommercialization) CalculateNetPrice() decimal.Decimal {
	// boardPrice * commercialCost(%) / 100
	DecimalCommercialCost := decimal.NewFromFloat(cc.CommercialCost)
	boardPercentage := cc.BoardPrice.Mul(DecimalCommercialCost).Div(decimal.NewFromInt(100)).Round(2)

	// boardPrice - freigthCost - boardPercentage
	netPrice := cc.BoardPrice.Sub(cc.FreightCost).Sub(boardPercentage).Round(2)

	return netPrice
}
