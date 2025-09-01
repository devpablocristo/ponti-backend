package excel

import (
	"time"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/usecases/domain"
	"github.com/shopspring/decimal"
)

type ExcelDto struct {
	WorkorderNumber string          `excel:"OT N°"`
	Date            time.Time       `excel:"FECHA"`
	ProjectName     string          `excel:"PROYECTO"`
	FieldName       string          `excel:"CAMPO"`
	CropName        string          `excel:"CULTIVO"`
	LaborName       string          `excel:"LABOR"`
	Contractor      string          `excel:"CONTRATISTA"`
	SurfaceHa       decimal.Decimal `excel:"SUPÉRFICIE"`
	CostHa          decimal.Decimal `excel:"COSTO $/HECTÁREA"`
	InvestorName    string          `excel:"INVERSOR"`
	USDAvgValue     decimal.Decimal `excel:"U$ PROM"`
	NetTotal        decimal.Decimal `excel:"TOTAL $ / NETO"`
	TotalIVA        decimal.Decimal `excel:"TOTAL $ / IVA"`
	USDCostHa       decimal.Decimal `excel:"COSTO U$ /HA"`
	USDNetTotal     decimal.Decimal `excel:"TOTAL U$ NETO"`

	InvoiceNumber  string     `excel:"N° FACTURA"`
	InvoiceCompany string     `excel:"EMPRESA"`
	InvoiceDate    *time.Time `excel:"FECHA DE FACTURACIÓN"`
	InvoiceStatus  string     `excel:"ESTADO DE FACTURA"`
}

func BuildExcelDTO(items []domain.LaborListItem) []ExcelDto {
	out := make([]ExcelDto, 0, len(items))

	for _, it := range items {

		var invDate *time.Time
		if !it.InvoiceDate.IsZero() {
			d := it.InvoiceDate
			invDate = &d
		}

		out = append(out, ExcelDto{
			WorkorderNumber: it.WorkorderNumber,
			Date:            it.Date,
			ProjectName:     it.ProjectName,
			FieldName:       it.FieldName,
			CropName:        it.CropName,
			LaborName:       it.LaborName,
			Contractor:      it.Contractor,
			SurfaceHa:       it.SurfaceHa,
			CostHa:          it.CostHa,
			InvestorName:    it.InvestorName,
			USDAvgValue:     it.USDAvgValue,
			NetTotal:        it.NetTotal,
			TotalIVA:        it.TotalIVA,
			USDCostHa:       it.USDCostHa,
			USDNetTotal:     it.USDNetTotal,
			InvoiceNumber:   it.InvoiceNumber,
			InvoiceCompany:  it.InvoiceCompany,
			InvoiceDate:     invDate,
			InvoiceStatus:   it.InvoiceStatus,
		})
	}

	return out
}
