package models

import (
	fieldmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/repository/models"
	investormod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/repository/models"
	projmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/repository/models"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/usecases/domain"
	supplymod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/repository/models"
	"time"
)

type Stock struct {
	ID             int64                `gorm:"primaryKey;autoIncrement;column:id"`
	ProjectID      int64                `gorm:"not null;index;column:project_id"`
	Project        projmod.Project      `gorm:"foreignKey:ProjectID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	FieldID        int64                `gorm:"not null;index;column:field_id"`
	Field          fieldmod.Field       `gorm:"foreignKey:FieldID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	SupplyID       int64                `gorm:"not null;index;column:supply_id"`
	Supply         supplymod.Supply     `gorm:"foreignKey:SupplyID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	InvestorID     int64                `gorm:"not null;index;column:investor_id"`
	Investor       investormod.Investor `gorm:"foreignKey:InvestorID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	CloseDate      time.Time            `gorm:"not null;column:close_date"`
	YearPeriod     int64                `gorm:"not null;column:year_period"`
	MonthPeriod    int64                `gorm:"not null;column:month_period"`
	UnitsEntered   int64                `gorm:"not null;column:units_entered"`
	UnitsConsumed  int64                `gorm:"not null;column:units_consumed"`
	RealStockUnits int64                `gorm:"not null;column:real_stock_units"`
	sharedmodels.Base
}

// ToDomain convierte el modelo Stock a la entidad de dominio
func (m *Stock) ToDomain() *domain.Stock {
	var timeZero time.Time
	var closeDateNil *time.Time
	if m.CloseDate == timeZero {
		closeDateNil = nil
	} else {
		closeDateNil = &m.CloseDate
	}
	return &domain.Stock{
		ID:             m.ID,
		Project:        m.Project.ToDomain(),
		Field:          m.Field.ToDomain(),
		Supply:         m.Supply.ToDomain(),
		CloseDate:      closeDateNil,
		UnitsEntered:   m.UnitsEntered,
		UnitsConsumed:  m.UnitsConsumed,
		RealStockUnits: m.RealStockUnits,
		YearPeriod:     m.YearPeriod,
		MonthPeriod:    m.MonthPeriod,
		Investor:       m.Investor.ToDomain(),
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
	return &Stock{
		ID:        d.ID,
		ProjectID: d.Project.ID,
		FieldID:   d.Field.ID,
		Field: fieldmod.Field{
			ID: d.Field.ID,
		},
		SupplyID:       d.Supply.ID,
		InvestorID:     d.Investor.ID,
		RealStockUnits: d.RealStockUnits,
		UnitsEntered:   d.UnitsEntered,
		UnitsConsumed:  d.UnitsConsumed,
		YearPeriod:     d.YearPeriod,
		MonthPeriod:    d.MonthPeriod,
		Base: sharedmodels.Base{
			CreatedAt: d.CreatedAt,
			UpdatedAt: d.UpdatedAt,
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}
