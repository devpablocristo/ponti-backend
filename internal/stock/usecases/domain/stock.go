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
	ID               int64
	Project          *projdom.Project
	Supply           *supplydomain.Supply
	Investor         *domain.Investor
	CloseDate        *time.Time
	SupplyMovements  []supplydomain.SupplyMovement
	RealStockUnits   decimal.Decimal
	Consumed         decimal.Decimal
	UnitsTransferred decimal.Decimal
	InitialStock     decimal.Decimal
	YearPeriod       int64
	MonthPeriod      int64
	shareddomain.Base
}

func (s *Stock) GetTotalUSD() decimal.Decimal {
	return s.GetStockUnits().Mul(s.Supply.Price)
}

func (s *Stock) GetStockUnits() decimal.Decimal {

	return s.GetEntryStock().Sub(s.Consumed)
}

func (s *Stock) GetStockDifference() decimal.Decimal {
	return s.RealStockUnits.Sub(s.GetStockUnits())
}

func (s *Stock) GetEntryStock() decimal.Decimal {
	var stockUnits decimal.Decimal
	for _, supplyMovement := range s.SupplyMovements {
		if supplyMovement.IsEntry {
			stockUnits = stockUnits.Add(supplyMovement.Quantity)
		}
	}

	if s.UnitsTransferred.GreaterThan(decimal.Zero) {
		stockUnits = stockUnits.Sub(s.UnitsTransferred)
	}

	return stockUnits
}

func (s *Stock) GetOutStock() decimal.Decimal {
	var stockUnits decimal.Decimal
	for _, supplyMovement := range s.SupplyMovements {
		if !supplyMovement.IsEntry {
			stockUnits = stockUnits.Add(supplyMovement.Quantity)
		}
	}

	return stockUnits
}

func (s *Stock) GetSupplyUnitName() string {
	return s.Supply.UnitName
}
