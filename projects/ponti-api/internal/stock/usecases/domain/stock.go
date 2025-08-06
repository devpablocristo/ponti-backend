package domain

import (
	fielddom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
	projdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	supplydomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/usecases/domain"
	supplymovementdomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/usecases/domain"
	"time"
)

type Stock struct {
	ID              int64
	Project         *projdom.Project
	Field           *fielddom.Field
	Supply          *supplydomain.Supply
	Investor        *domain.Investor
	CloseDate       *time.Time
	SupplyMovements []supplymovementdomain.SupplyMovement
	RealStockUnits  float64
	InitialStock    float64
	YearPeriod      int64
	MonthPeriod     int64
	shareddomain.Base
}

func (s *Stock) GetTotalUSD() float64 {
	return s.GetStockUnits() * s.Supply.Price
}

func (s *Stock) GetStockUnits() float64 {
	var stockUnits float64
	for _, supplyMovement := range s.SupplyMovements {
		stockUnits += supplyMovement.Quantity
	}
	return stockUnits + s.InitialStock
}

func (s *Stock) GetStockDifference() float64 {
	return s.RealStockUnits - s.GetStockUnits()
}
