package domain

// Lot representa el dominio para un lote (subdivisión de un campo).
type Lot struct {
	ID             int64   // Identificador único
	Name           string  // Identificador o código del lote
	FieldID        int64   // ID del campo al que pertenece
	ProjectID      int64   // ID del proyecto asociado
	CurrentCropID  *int64  // Puede ser nulo: ID del cultivo actual
	PreviousCropID *int64  // Puede ser nulo: ID del cultivo anterior
	Variety        string  // Variedad (por ejemplo, del cultivo actual)
	Area           float64 // Superficie en hectáreas
	Season         string  // Temporada de cultivo
}
