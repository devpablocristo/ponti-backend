package domain

// Workorder representa una orden de trabajo
type Workorder struct {
	Number        string          // e.g. "0000-0001"
	ProjectID     int64           // ID de proyecto
	FieldID       int64           // ID de campo
	LotID         int64           // ID de lote
	CropID        int64           // ID de cultivo
	LaborID       int64           // ID de labor
	Contractor    string          // Nombre del contratista
	Observations  string          // Observaciones adicionales
	Date          string          // Fecha de la orden (YYYY-MM-DD)
	InvestorID    int64           // ID del inversor
	EffectiveArea float64         // Superficie efectiva total
	Items         []WorkorderItem // Insumos y cantidades
}

// WorkorderItem representa un insumo dentro de la orden
type WorkorderItem struct {
	SupplyID  int64   // ID de insumo
	TotalUsed float64 // litros o kg totales usados
	FinalDose float64 // dosis final (total/superficie)
}

// Hay cuatro estados posibles: e.g. "pending", "in_progress", "completed", "cancelled".
// Filtro para listar workorders.// WorkorderFilter para listar workorders
type WorkorderFilter struct {
	ProjectID *int64 // filtrar por proyecto
	FieldID   *int64 // filtrar por campo
}
