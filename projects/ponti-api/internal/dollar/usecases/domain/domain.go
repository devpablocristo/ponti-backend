package domain

import "time"

type DollarAverage struct {
	ID         int64     // ID de cada registro
	ProjectID  int64     // ID del proyecto
	Year       int64     // Año correspondiente al promedio
	Month      string    // Mes correspondiente al promedio
	StartValue float64   // Valor de inicio
	EndValue   float64   // Valor final
	AvgValue   float64   // Valor Promedio
	CreatedAt  time.Time // Cuando se creo el registro
	UpdatedAt  time.Time // Cuando se actualizo el registro
}
