package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
)

type Dashboard struct {
	ID int64 `gorm:"primaryKey;autoIncrement;column:id" json:"id"`

	sharedmodels.Base
}

func (Dashboard) TableName() string {
	return "dashboard_view"
}

func (d Dashboard) ToDomain() *domain.Dashboard {
	return &domain.Dashboard{
		ID: d.ID,
		Base: shareddomain.Base{
			CreatedAt: d.CreatedAt,
			UpdatedAt: d.UpdatedAt,
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}

func FromDomainDashboard(d *domain.Dashboard) *Dashboard {
	return &Dashboard{
		ID: d.ID,
		Base: sharedmodels.Base{
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}
