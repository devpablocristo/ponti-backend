package domain

import (
	cropdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
)

// Lot represents a piece of land within a field.

type Lot struct {
	ID           int64
	Name         string
	Hectares     float64
	PreviousCrop cropdom.Crop
	CurrentCrop  cropdom.Crop
	Season       string
}
