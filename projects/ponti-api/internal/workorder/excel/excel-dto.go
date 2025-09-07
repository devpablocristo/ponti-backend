package excel

import (
	"time"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/usecases/domain"
	"github.com/shopspring/decimal"
)

type WorkorderExcelDto struct {
	Number       string          `excel:"NUMERO DE ORDEN"`
	ProjectName  string          `excel:"PROYECTO"`
	FieldName    string          `excel:"CAMPO"`
	LotName      string          `excel:"LOTE"`
	Date         time.Time       `excel:"FECHA"`
	CropName     string          `excel:"CULTIVO"`
	LaborName    string          `excel:"LABOR"`
	TypeName     string          `excel:"TIPO/CLASE"`
	Contractor   string          `excel:"CONTRATISTA"`
	SurfaceHa    decimal.Decimal `excel:"SUPERFICIE"`
	SupplyName   string          `excel:"INSUMO"`
	Consumption  decimal.Decimal `excel:"CONSUMO"`
	CategoryName string          `excel:"RUBRO"`
	Dose         decimal.Decimal `excel:"DOSIS"`
	CostPerHa    decimal.Decimal `excel:"COST U$/HA"`
	UnitPrice    decimal.Decimal `excel:"PRECIO UNIDAD"`
	TotalCost    decimal.Decimal `excel:"TOTAL COSTO"`
}

func BuildWorkorderExcelDTO(items []domain.WorkorderListElement) []WorkorderExcelDto {
	out := make([]WorkorderExcelDto, 0, len(items))

	for _, it := range items {
		out = append(out, WorkorderExcelDto{
			Number:       it.Number,
			ProjectName:  it.ProjectName,
			FieldName:    it.FieldName,
			LotName:      it.LotName,
			Date:         it.Date,
			CropName:     it.CropName,
			LaborName:    it.LaborName,
			TypeName:     it.TypeName,
			Contractor:   it.Contractor,
			SurfaceHa:    it.SurfaceHa,
			SupplyName:   it.SupplyName,
			Consumption:  it.Consumption,
			CategoryName: it.CategoryName,
			Dose:         it.Dose,
			CostPerHa:    it.CostPerHa,
			UnitPrice:    it.UnitPrice,
			TotalCost:    it.TotalCost,
		})
	}

	return out
}
