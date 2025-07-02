package domain

import cropdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"

type Lot struct {
	ID           int64
	Name         string
	FieldID      int64
	Hectares     float64
	PreviousCrop cropdom.Crop
	CurrentCrop  cropdom.Crop
	Season       string
}

type LotKPIs struct {
	SeededArea     float64
	HarvestedArea  float64
	YieldTnPerHa   float64
	CostPerHectare float64
}
