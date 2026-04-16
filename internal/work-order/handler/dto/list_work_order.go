// Package dto define los DTOs HTTP para work orders.
package dto

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"

	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
	"github.com/devpablocristo/ponti-backend/internal/work-order/usecases/domain"
)

type WorkOrderListElement struct {
	ID                int64           `json:"id"`
	Number            string          `json:"number"`
	ProjectName       string          `json:"project_name"`
	FieldName         string          `json:"field_name"`
	LotName           string          `json:"lot_name"`
	Date              time.Time       `json:"date"`
	CropName          string          `json:"crop_name"`
	LaborName         string          `json:"labor_name"`
	LaborCategoryName string          `json:"labor_category_name"`
	TypeName          string          `json:"type_name"`
	Contractor        string          `json:"contractor"`
	SurfaceHa         decimal.Decimal `json:"surface_ha"`
	SupplyName        string          `json:"supply_name"`
	Consumption       decimal.Decimal `json:"consumption"`
	CategoryName      string          `json:"category_name"`
	Dose              decimal.Decimal `json:"dose"`
	CostPerHa         decimal.Decimal `json:"cost_per_ha"`
	UnitPrice         decimal.Decimal `json:"unit_price"`
	TotalCost         decimal.Decimal `json:"total_cost"`
	IsDigital         bool            `json:"is_digital"`
	Status            string          `json:"status"`
}


// MarshalJSON asegura 2 decimales en todos los campos decimal de salida.
func (w WorkOrderListElement) MarshalJSON() ([]byte, error) {
	aux := struct {
		ID                int64     `json:"id"`
		Number            string    `json:"number"`
		ProjectName       string    `json:"project_name"`
		FieldName         string    `json:"field_name"`
		LotName           string    `json:"lot_name"`
		Date              time.Time `json:"date"`
		CropName          string    `json:"crop_name"`
		LaborName         string    `json:"labor_name"`
		LaborCategoryName string    `json:"labor_category_name"`
		TypeName          string    `json:"type_name"`
		Contractor        string    `json:"contractor"`
		SurfaceHa         string    `json:"surface_ha"`
		SupplyName        string    `json:"supply_name"`
		Consumption       string    `json:"consumption"`
		CategoryName      string    `json:"category_name"`
		Dose              string    `json:"dose"`
		CostPerHa         string    `json:"cost_per_ha"`
		UnitPrice         string    `json:"unit_price"`
		TotalCost         string    `json:"total_cost"`
		IsDigital         bool      `json:"is_digital"`
		Status            string    `json:"status"`
	}{
		ID:                w.ID,
		Number:            w.Number,
		ProjectName:       w.ProjectName,
		FieldName:         w.FieldName,
		LotName:           w.LotName,
		Date:              w.Date,
		CropName:          w.CropName,
		LaborName:         w.LaborName,
		LaborCategoryName: w.LaborCategoryName,
		TypeName:          w.TypeName,
		Contractor:        w.Contractor,
		SurfaceHa:         w.SurfaceHa.StringFixed(2),
		SupplyName:        w.SupplyName,
		Consumption:       w.Consumption.StringFixed(2),
		CategoryName:      w.CategoryName,
		Dose:              w.Dose.StringFixed(2),
		CostPerHa:         w.CostPerHa.StringFixed(2),
		UnitPrice:         w.UnitPrice.StringFixed(2),
		TotalCost:         w.TotalCost.Round(0).String(),
		IsDigital:         w.IsDigital,
		Status:            w.Status,
	}

	return json.Marshal(aux)
}

type WorkOrderListResponse struct {
	PageInfo types.PageInfo         `json:"page_info"`
	Items    []WorkOrderListElement `json:"items"`
}

func FromDomainListElement(d *domain.WorkOrderListElement) *WorkOrderListElement {
	return &WorkOrderListElement{
		ID:                d.ID,
		Number:            d.Number,
		ProjectName:       d.ProjectName,
		FieldName:         d.FieldName,
		LotName:           d.LotName,
		Date:              d.Date,
		CropName:          d.CropName,
		LaborName:         d.LaborName,
		LaborCategoryName: d.LaborCategoryName,
		TypeName:          d.TypeName,
		Contractor:        d.Contractor,
		SurfaceHa:         d.SurfaceHa,
		SupplyName:        d.SupplyName,
		Consumption:       d.Consumption,
		CategoryName:      d.CategoryName,
		Dose:              d.Dose,
		CostPerHa:         d.CostPerHa,
		UnitPrice:         d.UnitPrice,
		TotalCost:         d.TotalCost,
		IsDigital:         d.IsDigital,
		Status:            d.Status,
	}
}

func FromDomainList(pageInfo types.PageInfo, list []domain.WorkOrderListElement) WorkOrderListResponse {
	items := make([]WorkOrderListElement, len(list))
	for i, d := range list {
		items[i] = *FromDomainListElement(&d)
	}
	return WorkOrderListResponse{PageInfo: pageInfo, Items: items}
}
