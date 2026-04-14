// Package domain contiene modelos de dominio para stock.
package domain

import (
	"time"

	"github.com/devpablocristo/ponti-backend/internal/investor/usecases/domain"
	projdom "github.com/devpablocristo/ponti-backend/internal/project/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	supplydomain "github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
	"github.com/shopspring/decimal"
)

type Stock struct {
	ID                int64
	Project           *projdom.Project
	Supply            *supplydomain.Supply
	Investor          *domain.Investor
	CloseDate         *time.Time
	SupplyMovements   []supplydomain.SupplyMovement
	RealStockUnits    decimal.Decimal
	Consumed          decimal.Decimal
	UnitsTransferred  decimal.Decimal
	InitialStock      decimal.Decimal
	YearPeriod        int64
	MonthPeriod       int64
	HasRealStockCount bool
	shareddomain.Base
}

func (s *Stock) GetTotalUSD() decimal.Decimal {
	return s.GetStockUnits().Mul(s.Supply.Price)
}

func (s *Stock) GetStockUnits() decimal.Decimal {
	return s.GetNetMovementUnits().Sub(s.Consumed)
}

func (s *Stock) GetStockDifference() decimal.Decimal {
	return s.RealStockUnits.Sub(s.GetStockUnits())
}

func (s *Stock) GetStockDifferencePtr() *decimal.Decimal {
	if !s.HasRealStockCount {
		return nil
	}
	diff := s.GetStockDifference()
	return &diff
}

func (s *Stock) GetEntryStock() decimal.Decimal {
	var stockUnits decimal.Decimal
	for _, supplyMovement := range s.SupplyMovements {
		if supplyMovement.IsEntry && supplyMovement.Quantity.GreaterThanOrEqual(decimal.Zero) {
			stockUnits = stockUnits.Add(supplyMovement.Quantity)
		}
	}
	return stockUnits
}

func (s *Stock) GetOutStock() decimal.Decimal {
	var stockUnits decimal.Decimal
	for _, supplyMovement := range s.SupplyMovements {
		if supplyMovement.IsEntry && supplyMovement.Quantity.LessThan(decimal.Zero) {
			stockUnits = stockUnits.Add(supplyMovement.Quantity.Abs())
			continue
		}
		if !supplyMovement.IsEntry {
			stockUnits = stockUnits.Add(supplyMovement.Quantity.Abs())
		}
	}

	if s.UnitsTransferred.GreaterThan(decimal.Zero) {
		stockUnits = stockUnits.Add(s.UnitsTransferred)
	}

	return stockUnits
}

func (s *Stock) GetSupplyUnitName() string {
	return s.Supply.UnitName
}

func (s *Stock) GetNetMovementUnits() decimal.Decimal {
	var stockUnits decimal.Decimal
	for _, supplyMovement := range s.SupplyMovements {
		if supplyMovement.IsEntry {
			stockUnits = stockUnits.Add(supplyMovement.Quantity)
			continue
		}
		stockUnits = stockUnits.Sub(supplyMovement.Quantity.Abs())
	}

	if s.UnitsTransferred.GreaterThan(decimal.Zero) {
		stockUnits = stockUnits.Sub(s.UnitsTransferred)
	}

	return stockUnits
}
