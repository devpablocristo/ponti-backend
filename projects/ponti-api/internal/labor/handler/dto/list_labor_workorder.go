package dto

import (
	"time"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/usecases/domain"
	"github.com/shopspring/decimal"
)

type LaborListItem struct {
	WorkorderNumber string          `json:"workorder_number"`
	Date            time.Time       `json:"date"`
	ProjectName     string          `json:"project_name"`
	FieldName       string          `json:"field_name"`
	CropName        string          `json:"crop_name"`
	LaborName       string          `json:"labor_name"`
	Contractor      string          `json:"contractor"`
	SurfaceHa       decimal.Decimal `json:"surface_ha"`
	CostHa          decimal.Decimal `json:"cost_ha"`
	CategoryName    string          `json:"category_name"`
	InvestorName    string          `json:"investor_name"`
	NetTotal        decimal.Decimal `json:"net_total"`
	TotalIVA        decimal.Decimal `json:"total_iva"`
}

type LaborByWorkorderListResponse struct {
	Data []LaborListItem `json:"data"`
}

func ToLaborListResponse(items []domain.LaborListItem) LaborByWorkorderListResponse {
	dtos := make([]LaborListItem, len(items))
	for i, d := range items {
		dtos[i] = LaborListItem{
			WorkorderNumber: d.WorkorderNumber,
			Date:            d.Date,
			ProjectName:     d.ProjectName,
			FieldName:       d.FieldName,
			CropName:        d.CropName,
			LaborName:       d.LaborName,
			Contractor:      d.Contractor,
			SurfaceHa:       d.SurfaceHa,
			CostHa:          d.CostHa,
			CategoryName:    d.CategoryName,
			InvestorName:    d.InvestorName,
			NetTotal:        d.NetTotal,
			TotalIVA:        d.TotalIVA,
		}
	}
	return LaborByWorkorderListResponse{Data: dtos}
}
