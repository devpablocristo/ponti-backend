package list

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	domain "github.com/alphacodinggroup/ponti-backend/internal/supply/usecases/domain"
)

// Estructura de supply para listados con información completa
type ListedSupply struct {
	ID    int64           `json:"id"`
	Name  string          `json:"name"`
	Price decimal.Decimal `json:"price"` // Precio U$
	IsPartialPrice bool            `json:"is_partial_price"`
	TotalUSD       decimal.Decimal `json:"total_usd"`     // Total U$
	UnitID         int64           `json:"unit_id"`       // Unidad ID
	UnitName       string          `json:"unit_name"`     // Nombre de unidad
	CategoryName   string          `json:"category_name"` // Rubro
	CategoryID     int64           `json:"category_id"`   // Rubro ID
	TypeName       string          `json:"type_name"`     // Tipo/Clase
	TypeID         int64           `json:"type_id"`       // Tipo/Clase ID
	Origin         *SupplyOrigin   `json:"origin,omitempty"`
}

type SupplyOrigin struct {
	Type            string     `json:"type"`
	SourceProjectID *int64     `json:"source_project_id,omitempty"`
	SourceProject   string     `json:"source_project,omitempty"`
	MovementID      *int64     `json:"movement_id,omitempty"`
	ReferenceNumber string     `json:"reference_number,omitempty"`
	ProviderName    string     `json:"provider_name,omitempty"`
	MovementDate    *time.Time `json:"movement_date,omitempty"`
}

// MarshalJSON aplica redondeo: Precio U$ con 2 decimales, Total U$ al entero más próximo
func (s ListedSupply) MarshalJSON() ([]byte, error) {
	aux := struct {
		ID             int64         `json:"id"`
		Name           string        `json:"name"`
		Price          string        `json:"price"`
		IsPartialPrice bool          `json:"is_partial_price"`
		TotalUSD       string        `json:"total_usd"`
		UnitID         int64         `json:"unit_id"`
		UnitName       string        `json:"unit_name"`
		CategoryName   string        `json:"category_name"`
		CategoryID     int64         `json:"category_id"`
		TypeName       string        `json:"type_name"`
		TypeID         int64         `json:"type_id"`
		Origin         *SupplyOrigin `json:"origin,omitempty"`
	}{
		ID:             s.ID,
		Name:           s.Name,
		Price:          s.Price.StringFixed(2), // Precio U$: 2 decimales
		IsPartialPrice: s.IsPartialPrice,
		TotalUSD:       s.TotalUSD.StringFixed(0), // Total U$: entero más próximo
		UnitID:         s.UnitID,
		UnitName:       s.UnitName,
		CategoryName:   s.CategoryName,
		CategoryID:     s.CategoryID,
		TypeName:       s.TypeName,
		TypeID:         s.TypeID,
		Origin:         s.Origin,
	}
	return json.Marshal(aux)
}

// Respuesta principal de listado de supplies con métricas
type ListSuppliesResponse struct {
	Data        []ListedSupply  `json:"data"`
	PageInfo    types.PageInfo  `json:"page_info"`
	TotalKg     decimal.Decimal `json:"total_kg"`      // Total insumos invertidos (kg)
	TotalLts    decimal.Decimal `json:"total_lts"`     // Total insumos invertidos (lts)
	TotalNetUSD decimal.Decimal `json:"total_net_usd"` // Total U$/Neto
}

// MarshalJSON aplica redondeo al entero más próximo para las métricas
func (r ListSuppliesResponse) MarshalJSON() ([]byte, error) {
	aux := struct {
		Data        []ListedSupply `json:"data"`
		PageInfo    types.PageInfo `json:"page_info"`
		TotalKg     string         `json:"total_kg"`
		TotalLts    string         `json:"total_lts"`
		TotalNetUSD string         `json:"total_net_usd"`
	}{
		Data:        r.Data,
		PageInfo:    r.PageInfo,
		TotalKg:     r.TotalKg.StringFixed(0),     // Total insumos invertidos (kg): entero más próximo
		TotalLts:    r.TotalLts.StringFixed(0),    // Total insumos invertidos (lts): entero más próximo
		TotalNetUSD: r.TotalNetUSD.StringFixed(0), // Total U$/Neto: entero más próximo
	}
	return json.Marshal(aux)
}

// Constructor para la respuesta paginada de supplies con métricas
func NewListSuppliesResponse(
	items []domain.Supply,
	page, perPage int,
	total int64,
) ListSuppliesResponse {
	var totalKg decimal.Decimal
	var totalLts decimal.Decimal
	var totalNetUSD decimal.Decimal

	out := make([]ListedSupply, len(items))
	for i, s := range items {
		// Calcular Total U$ como precio * cantidad (asumiendo cantidad = 1 por ahora)
		totalUSD := s.Price // TODO: multiplicar por cantidad real si existe

		out[i] = ListedSupply{
			ID:    s.ID,
			Name:  s.Name,
			Price: s.Price, // Precio U$
			// Nuevo: enviamos el estado parcial/final al listado.
			IsPartialPrice: s.IsPartialPrice,
			TotalUSD:       totalUSD,       // Total U$
			UnitID:         s.UnitID,       // Unidad ID
			UnitName:       s.UnitName,     // Nombre de unidad
			CategoryName:   s.CategoryName, // Rubro
			CategoryID:     s.CategoryID,   // Rubro ID
			TypeName:       s.Type.Name,    // Tipo/Clase
			TypeID:         s.Type.ID,      // Tipo/Clase ID
			Origin:         mapOrigin(s.Origin),
		}

		// Acumular métricas según el tipo de unidad
		if isKG(s.UnitName) {
			totalKg = totalKg.Add(decimal.NewFromInt(1)) // TODO: usar cantidad real
		} else if isLt(s.UnitName) {
			totalLts = totalLts.Add(decimal.NewFromInt(1)) // TODO: usar cantidad real
		}
		totalNetUSD = totalNetUSD.Add(totalUSD)
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
		TotalKg:     totalKg,
		TotalLts:    totalLts,
		TotalNetUSD: totalNetUSD,
	}
}

func mapOrigin(origin *domain.SupplyOrigin) *SupplyOrigin {
	if origin == nil {
		return nil
	}
	return &SupplyOrigin{
		Type:            origin.Type,
		SourceProjectID: origin.SourceProjectID,
		SourceProject:   origin.SourceProject,
		MovementID:      origin.MovementID,
		ReferenceNumber: origin.ReferenceNumber,
		ProviderName:    origin.ProviderName,
		MovementDate:    origin.MovementDate,
	}
}

func isKG(unitName string) bool {
	return strings.Contains(strings.ToLower(unitName), "kg")
}

func isLt(unitName string) bool {
	return strings.Contains(strings.ToLower(unitName), "lt")
}
