package domain

import (
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
)

type CropCommercialization struct {
	ID             int64   // ID de cada registro
	ProjectID      int64   // ID del Proyecto
	CropName       string  // Nombre del cultivo
	BoardPrice     float64 // Precio en pizarra
	FreightCost    float64 // Costo de flete
	CommercialCost float64 // Gasto comerciales
	NetPrice       float64 // Precio neto

	shareddomain.Base
}
