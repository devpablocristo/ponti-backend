package dto

import (
	"time"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/usecases/domain"
	"github.com/shopspring/decimal"
)

// WorkorderListElementDTO para serializar la respuesta JSON de listados.
type WorkorderListElement struct {
	Number        string          `json:"number"`
	ProjectName   string          `json:"project_name"`
	FieldName     string          `json:"field_name"`
	LotName       string          `json:"lot_name"`
	Date          time.Time       `json:"date"`
	CropName      string          `json:"crop_name"`
	LaborName     string          `json:"labor_name"`
	ClassTypeName string          `json:"class_type_name"`
	Contractor    string          `json:"contractor"`
	SurfaceHa     decimal.Decimal `json:"surface_ha"`
	SupplyName    string          `json:"supply_name"`
	Consumption   decimal.Decimal `json:"consumption"`
	CategoryName  string          `json:"category_name"`
	Dose          decimal.Decimal `json:"dose"`
	CostPerHa     decimal.Decimal `json:"cost_per_ha"`
	UnitPrice     decimal.Decimal `json:"unit_price"`
	TotalCost     decimal.Decimal `json:"total_cost"`
}

// WorkorderListResponse agrupa la página y los items.
type WorkorderListResponse struct {
	PageInfo types.PageInfo         `json:"page_info"`
	Items    []WorkorderListElement `json:"items"`
}

// FromDomainListElement convierte dominio a DTO.
func FromDomainListElement(d *domain.WorkorderListElement) *WorkorderListElement {
	return &WorkorderListElement{
		Number:        d.Number,
		ProjectName:   d.ProjectName,
		FieldName:     d.FieldName,
		LotName:       d.LotName,
		Date:          d.Date,
		CropName:      d.CropName,
		LaborName:     d.LaborName,
		ClassTypeName: d.ClassTypeName,
		Contractor:    d.Contractor,
		SurfaceHa:     d.SurfaceHa,
		SupplyName:    d.SupplyName,
		Consumption:   d.Consumption,
		CategoryName:  d.CategoryName,
		Dose:          d.Dose,
		CostPerHa:     d.CostPerHa,
		UnitPrice:     d.UnitPrice,
		TotalCost:     d.TotalCost,
	}
}

// FromDomainList convierte slice de dominio a respuesta.
func FromDomainList(pageInfo types.PageInfo, list []domain.WorkorderListElement) WorkorderListResponse {
	items := make([]WorkorderListElement, len(list))
	for i, d := range list {
		items[i] = *FromDomainListElement(&d)
	}
	return WorkorderListResponse{PageInfo: pageInfo, Items: items}
}
