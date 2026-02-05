package excel

import "github.com/alphacodinggroup/ponti-backend/internal/supply/usecases/domain"

type SupplyTableDTO struct {
	Name         string  `excel:"NOMBRE"`
	Price        float64 `excel:"PRECIO"`
	CategoryName string  `excel:"RUBRO"`
	TypeName     string  `excel:"TIPO/CLASE"`
}

func BuildDTO(items []domain.Supply) []SupplyTableDTO {
	out := make([]SupplyTableDTO, 0, len(items))

	for _, it := range items {
		out = append(out, SupplyTableDTO{
			Name:         it.Name,
			Price:        decToFloat(it.Price, 2),
			CategoryName: it.CategoryName,
			TypeName:     it.Type.Name,
		})
	}
	return out
}
