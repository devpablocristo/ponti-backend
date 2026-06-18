package models

import (
	catmod "github.com/devpablocristo/ponti-backend/internal/category/repository/models"
	"github.com/devpablocristo/ponti-backend/internal/labor/usecases/domain"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
	"github.com/shopspring/decimal"
)

type Labor struct {
	ID              int64           `gorm:"primaryKey;autoIncrement"`
	Name            string          `gorm:"type:varchar(255);not null;column:name"`
	ContractorName  string          `gorm:"type:varchar(255);not null;column:contractor_name"`
	Price           decimal.Decimal `gorm:"column:price;default:0"`
	IsPartialPrice  bool            `gorm:"not null;default:false;column:is_partial_price"`
	ProjectId       int64           `gorm:"not null;column:project_id"`
	LaborCategoryID *int64          `gorm:"column:category_id"`
	Category        catmod.Category `gorm:"foreignKey:LaborCategoryID;references:ID" json:"category"`
	IsPending       bool            `gorm:"not null;default:false;column:is_pending"`
	sharedmodels.Base
}

func (l Labor) ToDomain() *domain.Labor {
	var categoryId int64
	if l.LaborCategoryID != nil {
		categoryId = *l.LaborCategoryID
	}
	return &domain.Labor{
		ID:             l.ID,
		Name:           l.Name,
		ContractorName: l.ContractorName,
		Price:          l.Price,
		ProjectId:      l.ProjectId,
		CategoryId:     categoryId,
		IsPartialPrice: l.IsPartialPrice,
		IsPending:      l.IsPending,
	}
}

func FromDomain(d *domain.Labor) *Labor {
	var categoryId *int64
	if d.CategoryId != 0 {
		categoryId = &d.CategoryId
	}
	return &Labor{
		ID:              d.ID,
		Name:            d.Name,
		ContractorName:  d.ContractorName,
		Price:           d.Price,
		ProjectId:       d.ProjectId,
		LaborCategoryID: categoryId,
		IsPartialPrice:  d.IsPartialPrice,
		IsPending:       d.IsPending,
		Base: sharedmodels.Base{
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}
