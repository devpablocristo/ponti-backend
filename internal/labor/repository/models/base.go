package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	catmod "github.com/devpablocristo/ponti-backend/internal/category/repository/models"
	"github.com/devpablocristo/ponti-backend/internal/labor/usecases/domain"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
)

type Labor struct {
	ID              int64           `gorm:"primaryKey;autoIncrement"`
	TenantID        uuid.UUID       `gorm:"column:tenant_id;type:uuid;index"`
	Name            string          `gorm:"type:varchar(255);not null;column:name"`
	ContractorName  string          `gorm:"type:varchar(255);not null;column:contractor_name"`
	Price           decimal.Decimal `gorm:"not null;column:price"`
	IsPartialPrice  bool            `gorm:"not null;default:false;column:is_partial_price"`
	ProjectId       int64           `gorm:"not null;column:project_id"`
	LaborCategoryID int64           `gorm:"not null;column:category_id"`
	Category        catmod.Category `gorm:"foreignKey:LaborCategoryID;references:ID" json:"category"`

	sharedmodels.Base
}

func (l Labor) ToDomain() *domain.Labor {
	return &domain.Labor{
		ID:             l.ID,
		Name:           l.Name,
		ContractorName: l.ContractorName,
		Price:          l.Price,
		ProjectId:      l.ProjectId,
		CategoryId:     l.LaborCategoryID,
		IsPartialPrice: l.IsPartialPrice,
	}
}

func FromDomain(d *domain.Labor) *Labor {
	return &Labor{
		ID:              d.ID,
		Name:            d.Name,
		ContractorName:  d.ContractorName,
		Price:           d.Price,
		ProjectId:       d.ProjectId,
		LaborCategoryID: d.CategoryId,
		IsPartialPrice:  d.IsPartialPrice,
		Base: sharedmodels.Base{
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}
