package excel

import (
	"github.com/devpablocristo/ponti-backend/internal/labor/usecases/domain"
)

type ExcelTableDto struct {
	Name           string  `excel:"NOMBRE"`
	CategoryName   string  `excel:"CATEGORÍA"`
	Price          float64 `excel:"PRECIO"`
	PriceStatus    string  `excel:"ESTADO_PRECIO"`
	ContractorName string  `excel:"CONTRATISTA"`
}

func BuildExcelTableDTO(items []domain.ListedLabor) []ExcelTableDto {
	out := make([]ExcelTableDto, 0, len(items))
	for _, it := range items {
		price, _ := it.Price.Float64()
		out = append(out, ExcelTableDto{
			Name:           it.Name,
			Price:          price,
			PriceStatus:    mapPriceStatus(it.IsPartialPrice),
			CategoryName:   it.CategoryName,
			ContractorName: it.ContractorName,
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
