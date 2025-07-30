package domain

import (
	fielddom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
	projdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	supplydomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/usecases/domain"
	"time"
)

type Stock struct {
	ID             int64
	Project        *projdom.Project
	Field          *fielddom.Field
	Supply         *supplydomain.Supply
	Investor       *domain.Investor
	CloseDate      *time.Time
	UnitsEntered   int64
	UnitsConsumed  int64
	RealStockUnits int64
	YearPeriod     int64
	MonthPeriod    int64
	shareddomain.Base
}

func (s *Stock) GetTotalUSD() float64 {
	return float64(s.UnitsEntered) * s.Supply.Price
}

func (s *Stock) GetStockUnits() int64 {
	return s.UnitsEntered - s.UnitsConsumed
}

func (s *Stock) GetStockDifference() int64 {
	return s.RealStockUnits - s.GetStockUnits()
}
