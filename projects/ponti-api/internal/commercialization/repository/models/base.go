package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/commercialization/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
)

type CropCommercialization struct {
	ID             int64   `gorm:"primaryKey;autoIncrement"`
	ProjectID      int64   `gorm:"not null;index"`
	CropName       string  `gorm:"type:varchar(100);not null"`
	BoardPrice     float64 `gorm:"not null"`
	FreightCost    float64 `gorm:"not null"`
	CommercialCost float64 `gorm:"not null"`
	NetPrice       float64 `gorm:"not null"`

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
