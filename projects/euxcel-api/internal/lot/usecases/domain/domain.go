package domain

type Lot struct {
	ID             int64 // Auto-generated primary key
	FieldID        int64 // Foreign key to Field
	Name           string
	Hectares       float64
	PreviousCropID int64
	CurrentCropID  int64
	Season         string
}
