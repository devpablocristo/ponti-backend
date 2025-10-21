package dto

import (
	"encoding/json"

	"github.com/shopspring/decimal"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/usecases/domain"
)

// Estructura de supply para listados con información completa
type ListedSupply struct {
	ID           int64           `json:"id"`
	Name         string          `json:"name"`
	Price        decimal.Decimal `json:"price"`         // Precio
	UnitID       int64           `json:"unit_id"`       // Unidad ID
	CategoryName string          `json:"category_name"` // Rubro
	CategoryID   int64           `json:"category_id"`   // Rubro ID
	TypeName     string          `json:"type_name"`     // Tipo/Clase
	TypeID       int64           `json:"type_id"`       // Tipo/Clase ID
}

// MarshalJSON aplica redondeo de 2 decimales al precio
func (s ListedSupply) MarshalJSON() ([]byte, error) {
	aux := struct {
		ID           int64  `json:"id"`
		Name         string `json:"name"`
		Price        string `json:"price"`
		UnitID       int64  `json:"unit_id"`
		CategoryName string `json:"category_name"`
		CategoryID   int64  `json:"category_id"`
		TypeName     string `json:"type_name"`
		TypeID       int64  `json:"type_id"`
	}{
		ID:           s.ID,
		Name:         s.Name,
		Price:        s.Price.Round(2).String(), // Precio: 2 decimales
		UnitID:       s.UnitID,
		CategoryName: s.CategoryName,
		CategoryID:   s.CategoryID,
		TypeName:     s.TypeName,
		TypeID:       s.TypeID,
	}
	return json.Marshal(aux)
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
			UnitID:       s.UnitID,       // Unidad ID
			CategoryName: s.CategoryName, // Rubro
			CategoryID:   s.CategoryID,   // Rubro ID
			TypeName:     s.Type.Name,    // Tipo/Clase
			TypeID:       s.Type.ID,      // Tipo/Clase ID
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
