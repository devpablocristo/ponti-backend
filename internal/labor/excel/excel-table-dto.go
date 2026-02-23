package excel

import (
	"github.com/alphacodinggroup/ponti-backend/internal/labor/usecases/domain"
)

type ExcelTableDto struct {
	Name           string  `excel:"NOMBRE"`
	CategoryName   string  `excel:"CATEGORÍA"`
	Price          float64 `excel:"PRECIO"`
	ContractorName string  `excel:"CONTRATISTA"`
}

func BuildExcelTableDTO(items []domain.ListedLabor) []ExcelTableDto {
	out := make([]ExcelTableDto, 0, len(items))
	for _, it := range items {
		price, _ := it.Price.Float64()
		out = append(out, ExcelTableDto{
			Name:           it.Name,
			Price:          price,
			CategoryName:   it.CategoryName,
			ContractorName: it.ContractorName,
		})
	}
	return out
}
