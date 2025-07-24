package domain

import (
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	"github.com/shopspring/decimal"
)

type CropCommercialization struct {
	ID             int64           // ID de cada registro
	ProjectID      int64           // ID del Proyecto
	CropName       string          // Nombre del cultivo
	BoardPrice     decimal.Decimal // Precio en pizarra
	FreightCost    decimal.Decimal // Costo de flete
	CommercialCost float64         // Gasto comerciales (%)
	NetPrice       decimal.Decimal // Precio neto

	shareddomain.Base
}
