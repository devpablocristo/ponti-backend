package dto

import (
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/usecases/domain"
	"github.com/shopspring/decimal"
)

// Estructura de supply para listados con información completa
type ListedSupply struct {
	ID           int64           `json:"id"`
	Name         string          `json:"name"`
	Price        decimal.Decimal `json:"price"`         // Precio
	CategoryName string          `json:"category_name"` // Rubro
	TypeName     string          `json:"type_name"`     // Tipo/Clase
}

// Respuesta principal de listado de supplies
type ListSuppliesResponse struct {
	Data     []ListedSupply `json:"data"`
	PageInfo types.PageInfo `json:"page_info"`
}

// Constructor para la respuesta paginada de supplies
func NewListSuppliesResponse(
	items []domain.Supply,
	page, perPage int,
	total int64,
) ListSuppliesResponse {
	out := make([]ListedSupply, len(items))
	for i, s := range items {
		out[i] = ListedSupply{
			ID:           s.ID,
			Name:         s.Name,
			Price:        s.Price,        // Precio
			CategoryName: s.CategoryName, // Rubro
			TypeName:     s.Type.Name,    // Tipo/Clase
		}
	}

	maxPage := int((total + int64(perPage) - 1) / int64(perPage))

	return ListSuppliesResponse{
		Data: out,
		PageInfo: types.PageInfo{
			PerPage: perPage,
			Page:    page,
			MaxPage: maxPage,
			Total:   total,
		},
	}
}
