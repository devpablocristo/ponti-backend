package domain

// Order representa una orden de trabajo
type Workorder struct {
	Number       string          // e.g. "0000-0001"
	ProjectID    int64           // ID de proyecto
	FieldID      int64           // ID de campo
	LotID        int64           // ID de lote
	CropID       int64           // ID de cultivo
	LaborID      int64           // ID de labor
	Contractor   string          // Nombre del contratista
	Items        []WorkorderItem // Insumos y cantidades
	Observations string          // Observaciones adicionales
}

// OrderItem representa un insumo dentro de la orden
type WorkorderItem struct {
	SupplyID      int64   // ID de insumo
	TotalUsed     float64 // lts o kg totales usados
	EffectiveArea float64 // superficie efectiva
	FinalDose     float64 // dosis final (total/superficie)
}

