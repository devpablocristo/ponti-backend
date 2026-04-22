// Package domain contiene modelos de dominio para stock continuo por proyecto.
package domain

import (
	"time"

	investordomain "github.com/devpablocristo/ponti-backend/internal/investor/usecases/domain"
	projdomain "github.com/devpablocristo/ponti-backend/internal/project/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	supplydomain "github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
	"github.com/shopspring/decimal"
)

// Stock representa el summary canónico de inventario por proyecto + insumo.
// El identificador operativo es supply_id.
type Stock struct {
	ID                int64
	ProjectID         int64
	Project           *projdomain.Project
	Supply            *supplydomain.Supply
	Investor          *investordomain.Investor
	SupplyMovements   []supplydomain.SupplyMovement
	EntryStock        decimal.Decimal
	OutStock          decimal.Decimal
	Consumed          decimal.Decimal
	StockUnits        decimal.Decimal
	RealStockUnits    decimal.Decimal
	HasRealStockCount bool
	LastCountAt       *time.Time
	CloseDate         *time.Time
	InitialStock      decimal.Decimal
	YearPeriod        int64
	MonthPeriod       int64
	shareddomain.Base
}

func (s *Stock) GetTotalUSD() decimal.Decimal {
	if s == nil || s.Supply == nil {
		return decimal.Zero
	}
	return s.GetStockUnits().Mul(s.Supply.Price)
}

func (s *Stock) GetStockUnits() decimal.Decimal {
	if s == nil {
		return decimal.Zero
	}
	if !s.StockUnits.IsZero() || !s.EntryStock.IsZero() || !s.OutStock.IsZero() {
		return s.StockUnits
	}
	return s.GetNetMovementUnits().Sub(s.Consumed)
}

func (s *Stock) GetEntryStock() decimal.Decimal {
	if s == nil {
		return decimal.Zero
	}
	if !s.EntryStock.IsZero() {
		return s.EntryStock
	}
	var total decimal.Decimal
	for _, supplyMovement := range s.SupplyMovements {
		if supplyMovement.IsEntry && supplyMovement.Quantity.GreaterThanOrEqual(decimal.Zero) {
			total = total.Add(supplyMovement.Quantity)
		}
	}
	return total
}

func (s *Stock) GetOutStock() decimal.Decimal {
	if s == nil {
		return decimal.Zero
	}
	if !s.OutStock.IsZero() {
		return s.OutStock
	}
	var total decimal.Decimal
	for _, supplyMovement := range s.SupplyMovements {
		if supplyMovement.Quantity.LessThan(decimal.Zero) {
			total = total.Add(supplyMovement.Quantity.Abs())
			continue
		}
		if !supplyMovement.IsEntry {
			total = total.Add(supplyMovement.Quantity.Abs())
		}
	}
	return total
}

func (s *Stock) GetNetMovementUnits() decimal.Decimal {
	if s == nil {
		return decimal.Zero
	}
	if !s.EntryStock.IsZero() || !s.OutStock.IsZero() {
		return s.EntryStock.Sub(s.OutStock)
	}
	var total decimal.Decimal
	for _, supplyMovement := range s.SupplyMovements {
		if supplyMovement.Quantity.LessThan(decimal.Zero) {
			total = total.Sub(supplyMovement.Quantity.Abs())
			continue
		}
		if supplyMovement.IsEntry {
			total = total.Add(supplyMovement.Quantity)
			continue
		}
		total = total.Sub(supplyMovement.Quantity.Abs())
	}
	return total
}

func (s *Stock) GetStockDifference() decimal.Decimal {
	if s == nil {
		return decimal.Zero
	}
	return s.RealStockUnits.Sub(s.GetStockUnits())
}

func (s *Stock) GetStockDifferencePtr() *decimal.Decimal {
	if s == nil || !s.HasRealStockCount {
		return nil
	}
	diff := s.GetStockDifference()
	return &diff
}

func (s *Stock) GetSupplyUnitName() string {
	if s == nil || s.Supply == nil {
		return ""
	}
	return s.Supply.UnitName
}
