package excel

import (
	"fmt"
	"time"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/usecases/domain"
	"github.com/shopspring/decimal"
)

type SupplyMovementExcelDTO struct {
	EntryType       string          `excel:"INGRESO"`
	ReferenceNumber string          `excel:"N° REMITO"`
	EntryDate       time.Time       `excel:"FECHA"`
	InvestorName    string          `excel:"INVERSOR"`
	SupplyName      string          `excel:"INSUMO"`
	Quantity        string          `excel:"CANTIDAD"`
	Category        string          `excel:"RUBRO"`
	Type            string          `excel:"TIPO/CLASE"`
	ProviderName    string          `excel:"PROVEEDOR"`
	PriceUSD        decimal.Decimal `excel:"PRECIO U$"`
	TotalUSD        decimal.Decimal `excel:"TOTAL U$"`
}

func BuildSupplyMovementDTO(items []*domain.SupplyMovement) []SupplyMovementExcelDTO {
	out := make([]SupplyMovementExcelDTO, 0, len(items))

	for _, it := range items {
		out = append(out, SupplyMovementExcelDTO{
			EntryType:       it.MovementType,
			ReferenceNumber: it.ReferenceNumber,
			EntryDate:       *it.MovementDate,
			InvestorName:    it.Investor.Name,
			SupplyName:      it.Supply.Name,
			Quantity:        fmt.Sprintf("%s %s", it.Quantity.String(), it.Supply.UnitName),
			Category:        it.Supply.CategoryName,
			Type:            it.Supply.Type.Name,
			ProviderName:    it.Provider.Name,
			PriceUSD:        it.Supply.Price,
			TotalUSD:        it.Supply.Price.Mul(it.Quantity),
		})
	}

	return out
}
