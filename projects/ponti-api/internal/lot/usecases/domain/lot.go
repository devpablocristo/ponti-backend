package domain

import (
	cropdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	"github.com/shopspring/decimal"
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
	Tons          int     // Toneladas cosechadas

	shareddomain.Base // <-- embebe campos de auditoría
}

type LotKPIs struct {
	SeededArea     decimal.Decimal
	HarvestedArea  decimal.Decimal
	YieldTnPerHa   decimal.Decimal
	CostPerHectare decimal.Decimal
}
