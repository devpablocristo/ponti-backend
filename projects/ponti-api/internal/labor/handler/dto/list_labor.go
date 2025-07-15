package dto

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/usecases/domain"
	"time"
)

type ListedLabor struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	CategoryId     int64     `json:"category_id"`
	Price          float64   `json:"price"`
	ContractorName string    `json:"contractor_name"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// PageInfo contiene metadata de paginación.
type PageInfo struct {
	PerPage int   `json:"per_page"`
	Page    int   `json:"page"`
	MaxPage int   `json:"max_page"`
	Total   int64 `json:"total"`
}

type ListLaborResponse struct {
	Data     []ListedLabor `json:"data"`
	PageInfo PageInfo      `json:"page_info"`
}

func NewListLaborsResponse(
	items []domain.ListedLabor,
	page, perPage int,
	total int64,
) ListLaborResponse {
	out := make([]ListedLabor, len(items))
	for i, l := range items {
		out[i] = ListedLabor{
			ID:             l.ID,
			Name:           l.Name,
			CategoryId:     l.LaborCategoryId,
			Price:          l.Price,
			ContractorName: l.ContractorName,
			UpdatedAt:      l.UpdatedAt,
		}
	}

	maxPage := int((total + int64(perPage) - 1) / int64(perPage))
	return ListLaborResponse{
		Data: out,
		PageInfo: PageInfo{
			PerPage: perPage,
			Page:    page,
			MaxPage: maxPage,
			Total:   total,
		},
	}
}
