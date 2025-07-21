package domain

import "time"

type CropCommercialization struct {
	ID             int64     // ID de cada registro
	ProjectID      int64     // ID del Proyecto
	CropName       string    // Nombre del cultivo
	BoardPrice     float64   // Precio en pizarra
	FreightCost    float64   // Costo de flete
	CommercialCost float64   // Gasto comerciales
	NetPrice       float64   // Precio neto
	CreatedAt      time.Time // Cuando se creo el registro
	UpdatedAt      time.Time // Cuando se actualizo el registro
}
