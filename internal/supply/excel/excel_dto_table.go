package excel

import "github.com/alphacodinggroup/ponti-backend/internal/supply/usecases/domain"

type SupplyTableDTO struct {
	Name     string  `excel:"NOMBRE"`
	UnitName string  `excel:"UNIDAD"`
	Price    float64 `excel:"PRECIO"`
	PriceStatus  string `excel:"ESTADO PRECIO"`
	CategoryName string `excel:"RUBRO"`
	TypeName     string `excel:"TIPO/CLASE"`
}

func BuildDTO(items []domain.Supply) []SupplyTableDTO {
	out := make([]SupplyTableDTO, 0, len(items))

	for _, it := range items {
		out = append(out, SupplyTableDTO{
			Name:         it.Name,
			UnitName:     it.UnitName,
			Price:        decToFloat(it.Price, 2),
			PriceStatus:  mapPriceStatus(it.IsPartialPrice),
			CategoryName: it.CategoryName,
			TypeName:     it.Type.Name,
		})
	}
	return out
}

func mapPriceStatus(isPartial bool) string {
	if isPartial {
		return "Parcial"
	}
	return "Final"
}
