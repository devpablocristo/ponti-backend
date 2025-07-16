package domain

import (
	"time"

	cropdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
)

type Lot struct {
	ID            int64
	Name          string
	FieldID       int64
	Hectares      float64
	PreviousCrop  cropdom.Crop
	CurrentCrop   cropdom.Crop
	Variety       string
	Season        string
	Status        string
	Dates         []LotDates
	Cost          float64 // Costo por hectárea
	HarvestedTons float64 // Toneladas cosechadas
	UpdatedAt     time.Time
}

type LotKPIs struct {
	SeededArea     float64
	HarvestedArea  float64
	YieldTnPerHa   float64
	CostPerHectare float64
}
