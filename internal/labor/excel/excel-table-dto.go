package excel

import (
	"github.com/alphacodinggroup/ponti-backend/internal/labor/usecases/domain"
)

type ExcelTableDto struct {
	ID             int64   `excel:"ID"`
	Name           string  `excel:"NOMBRE"`
	Price          float64 `excel:"PRECIO"`
	CategoryName   string  `excel:"CATEGORÍA"`
	ContractorName string  `excel:"CONTRATISTA"`
}

func BuildExcelTableDTO(items []domain.ListedLabor) []ExcelTableDto {
	out := make([]ExcelTableDto, 0, len(items))
	for _, it := range items {
		price, _ := it.Price.Float64()
		out = append(out, ExcelTableDto{
			ID:             it.ID,
			Name:           it.Name,
			Price:          price,
			CategoryName:   it.CategoryName,
			ContractorName: it.ContractorName,
		})
	}
	return out
}
