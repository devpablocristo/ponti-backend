package domain

type Lot struct {
	ID      int64   // Primary key (INT)
	FieldID int64   // Foreign key referencing the field (INT)
	LotName string  // Lot name (VARCHAR)
	Area    float64 // Surface area (DECIMAL)
}
