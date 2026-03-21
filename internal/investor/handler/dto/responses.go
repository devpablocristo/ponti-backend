package dto

import (
	"time"

	domain "github.com/devpablocristo/ponti-backend/internal/investor/usecases/domain"
	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
)

// InvestorResponse es el DTO de salida para un inversor individual.
type InvestorResponse struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	Percentage int        `json:"percentage"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`
}

// InvestorFromDomain convierte un domain.Investor a InvestorResponse.
func InvestorFromDomain(d *domain.Investor) InvestorResponse {
	return InvestorResponse{
		ID:         d.ID,
		Name:       d.Name,
		Percentage: d.Percentage,
		ArchivedAt: d.ArchivedAt,
	}
}

// ListInvestorsResponse es la respuesta paginada para el listado de inversores.
type ListInvestorsResponse struct {
	Data     []InvestorResponse `json:"data"`
	PageInfo types.PageInfo     `json:"page_info"`
}

// NewListInvestorsResponse construye la respuesta paginada de inversores.
func NewListInvestorsResponse(items []domain.Investor, page, perPage int, total int64) ListInvestorsResponse {
	data := make([]InvestorResponse, 0, len(items))
	for i := range items {
		data = append(data, InvestorFromDomain(&items[i]))
	}
	return ListInvestorsResponse{
		Data:     data,
		PageInfo: types.NewPageInfo(page, perPage, total),
	}
}
