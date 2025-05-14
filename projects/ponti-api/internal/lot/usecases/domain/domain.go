package domain

type Lot struct {
	ID             int64
	Name           string
	FieldID        int64
	Hectares       float64
	PreviousCropID int64
	CurrentCropID  int64
	Season         string
}
