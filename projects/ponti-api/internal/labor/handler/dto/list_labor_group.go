package dto

import (
	"time"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
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
	USDAvgValue     decimal.Decimal `json:"usd_avg_value"`
	NetTotal        decimal.Decimal `json:"net_total"`
	TotalIVA        decimal.Decimal `json:"total_iva"`
	USDCostHa       decimal.Decimal `json:"usd_cost_ha"`
	USDNetTotal     decimal.Decimal `json:"usd_net_total"`
	InvoiceNumber   string          `json:"invoice_number"`
	InvoiceCompany  string          `json:"invoice_company"`
	InvoiceDate     time.Time       `json:"invoice_date"`
	InvoiceStatus   string          `json:"invoice_status"`
}

type ListLaborGroupResponse struct {
	PageInfo types.PageInfo  `json:"page_info"`
	Data     []LaborListItem `json:"data"`
}

func FromDomainListGroup(d *domain.LaborListItem) *LaborListItem {
	return &LaborListItem{
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
		USDAvgValue:     d.USDAvgValue,
		NetTotal:        d.NetTotal,
		TotalIVA:        d.TotalIVA,
		USDCostHa:       d.USDCostHa,
		USDNetTotal:     d.USDNetTotal,
		InvoiceNumber:   d.InvoiceNumber,
		InvoiceCompany:  d.InvoiceCompany,
		InvoiceDate:     d.InvoiceDate,
		InvoiceStatus:   d.InvoiceStatus,
	}
}

func FromDomainList(pageInfo types.PageInfo, list []domain.LaborListItem) ListLaborGroupResponse {
	items := make([]LaborListItem, len(list))
	for i, d := range list {
		items[i] = *FromDomainListGroup(&d)
	}

	return ListLaborGroupResponse{PageInfo: pageInfo, Data: items}
}
