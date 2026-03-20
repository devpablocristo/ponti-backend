package dto

import (
	"github.com/shopspring/decimal"

	"github.com/devpablocristo/ponti-backend/internal/labor/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
)

type Labor struct {
	ID             int64           `json:"id"`
	Name           string          `json:"name"`
	ContractorName string          `json:"contractor_name"`
	Price          decimal.Decimal `json:"price"`
	IsPartialPrice *bool           `json:"is_partial_price"`
	CategoryId     int64           `json:"category_id"`
}

func (l Labor) ToDomain(projectId int64, userId string) *domain.Labor {
	return &domain.Labor{
		ID:             l.ID,
		Name:           l.Name,
		ContractorName: l.ContractorName,
		Price:          l.Price,
		IsPartialPrice: boolOrDefault(l.IsPartialPrice, false),
		ProjectId:      projectId,
		CategoryId:     l.CategoryId,
		Base: shareddomain.Base{
			UpdatedBy: &userId,
		},
	}
}

func FromDomain(d domain.Labor) *Labor {
	return &Labor{
		ID:             d.ID,
		Name:           d.Name,
		ContractorName: d.ContractorName,
		Price:          d.Price,
		IsPartialPrice: boolPtr(d.IsPartialPrice),
	}
}

