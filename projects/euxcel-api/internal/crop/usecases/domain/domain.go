package domain

type Crop struct {
	ID     int64  // Primary key (INT)
	Name   string // Crop name (VARCHAR)
	Season string // Season (VARCHAR), e.g., "Summer", "Winter"
	LotID  int64  // Foreign key referencing the lot (INT)
}
