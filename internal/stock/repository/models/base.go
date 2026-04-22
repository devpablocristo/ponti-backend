// Package models contiene modelos de persistencia para stock.
package models

import (
	"time"

	investormod "github.com/devpablocristo/ponti-backend/internal/investor/repository/models"
	projmod "github.com/devpablocristo/ponti-backend/internal/project/repository/models"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
	"github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
	supplymod "github.com/devpablocristo/ponti-backend/internal/supply/repository/models"
	"github.com/shopspring/decimal"
)

type Stock struct {
	ID                int64                      `gorm:"primaryKey;autoIncrement;column:id"`
	ProjectID         int64                      `gorm:"not null;index;column:project_id"`
	Project           projmod.Project            `gorm:"foreignKey:ProjectID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	SupplyID          int64                      `gorm:"not null;index;column:supply_id"`
	Supply            supplymod.Supply           `gorm:"foreignKey:SupplyID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	InvestorID        int64                      `gorm:"not null;index;column:investor_id"`
	Investor          investormod.Investor       `gorm:"foreignKey:InvestorID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	CloseDate         *time.Time                 `gorm:"null;column:close_date"`
	SupplyMovements   []supplymod.SupplyMovement `gorm:"foreignKey:StockId;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	InitialStock      decimal.Decimal            `gorm:"not null;column:initial_units"`
	YearPeriod        int64                      `gorm:"not null;column:year_period"`
	MonthPeriod       int64                      `gorm:"not null;column:month_period"`
	UnitsEntered      decimal.Decimal            `gorm:"not null;column:units_entered"`
	Consumed          decimal.Decimal            `gorm:"-"`
	UnitsConsumed     decimal.Decimal            `gorm:"not null;column:units_consumed"`
	RealStockUnits    decimal.Decimal            `gorm:"not null;column:real_stock_units"`
	HasRealStockCount bool                       `gorm:"not null;column:has_real_stock_count"`
	sharedmodels.Base
}

// ToDomain convierte el modelo Stock a la entidad de dominio
func (m *Stock) ToDomain() *domain.Stock {
	return &domain.Stock{
		ID:                m.SupplyID,
		ProjectID:         m.ProjectID,
		Supply:            m.Supply.ToDomain(),
		EntryStock:        m.UnitsEntered,
		Consumed:          m.Consumed,
		StockUnits:        m.UnitsEntered.Sub(m.Consumed),
		RealStockUnits:    m.RealStockUnits,
		HasRealStockCount: m.HasRealStockCount,
		Base: shareddomain.Base{
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
		},
	}
}

// FromDomain convierte la entidad de dominio Stock al modelo de persistencia
func FromDomain(d *domain.Stock) *Stock {
	investorID := int64(0)
	return &Stock{
		ID:                d.ID,
		ProjectID:         d.ProjectID,
		SupplyID:          d.Supply.ID,
		InvestorID:        investorID,
		RealStockUnits:    d.RealStockUnits,
		HasRealStockCount: d.HasRealStockCount,
		YearPeriod:        0,
		MonthPeriod:       0,
		InitialStock:      decimal.Zero,
		CloseDate:         nil,
		SupplyMovements:   []supplymod.SupplyMovement{},
		Base: sharedmodels.Base{
			CreatedAt: d.CreatedAt,
			UpdatedAt: d.UpdatedAt,
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}

}
