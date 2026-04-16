package models

import (
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	"github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
	"time"
)

type StockUpdate struct {
	CloseDate time.Time `gorm:"column:close_date"`
	UpdatedBy string    `gorm:"column:updated_by"`
}

func StockUpdateCloseDateFromDomain(d *domain.Stock) *StockUpdate {
	return &StockUpdate{
		CloseDate: *d.CloseDate,
		UpdatedBy: *d.UpdatedBy,
	}
}

// ToDomain convierte StockUpdateFields a dominio
func (s *StockUpdate) ToDomain() *domain.Stock {
	return &domain.Stock{
		CloseDate: &s.CloseDate,
		Base: shareddomain.Base{
			UpdatedBy: &s.UpdatedBy,
		},
	}
}
