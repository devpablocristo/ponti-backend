package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/commercialization/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	decimal "github.com/shopspring/decimal"
)

type CropCommercialization struct {
	ID             int64           `gorm:"primaryKey;autoIncrement;column:id"`
	ProjectID      int64           `gorm:"not null;index;column:project_id"`
	CropName       string          `gorm:"type:varchar(100);not null;column:crop_name"`
	BoardPrice     decimal.Decimal `gorm:"type:numeric(12,2);not null;column:board_price"`
	FreightCost    decimal.Decimal `gorm:"type:numeric(12,2);not null;column:freight_cost"`
	CommercialCost float64         `gorm:"not null;column:commercial_cost"`
	NetPrice       decimal.Decimal `gorm:"type:numeric(12,2);not null;column:net_price"`

	sharedmodels.Base
}

func FromDomain(cc *domain.CropCommercialization) *CropCommercialization {
	return &CropCommercialization{
		ID:             cc.ID,
		ProjectID:      cc.ProjectID,
		CropName:       cc.CropName,
		BoardPrice:     cc.BoardPrice,
		FreightCost:    cc.FreightCost,
		CommercialCost: cc.CommercialCost,
		NetPrice:       cc.NetPrice,
		Base: sharedmodels.Base{
			CreatedBy: cc.CreatedBy,
		},
	}
}

func (m *CropCommercialization) ToDomain() *domain.CropCommercialization {
	return &domain.CropCommercialization{
		ID:             m.ID,
		ProjectID:      m.ProjectID,
		CropName:       m.CropName,
		BoardPrice:     m.BoardPrice,
		FreightCost:    m.FreightCost,
		CommercialCost: m.CommercialCost,
		NetPrice:       m.NetPrice,
		Base: shareddomain.Base{
			CreatedAt: m.CreatedAt,
			CreatedBy: m.CreatedBy,
		},
	}
}
