package dto

import "github.com/devpablocristo/ponti-backend/internal/data-integrity/usecases/domain"

type TentativePriceItemResponse struct {
	SupplyID     int64  `json:"supply_id"`
	Name         string `json:"name"`
	CategoryName string `json:"category_name"`
	Price        string `json:"price"`
}

type TentativePricesResponse struct {
	Count int64                        `json:"count"`
	Items []TentativePriceItemResponse `json:"items"`
}

func ToTentativePricesResponse(report *domain.TentativePricesReport) TentativePricesResponse {
	if report == nil {
		return TentativePricesResponse{Items: []TentativePriceItemResponse{}}
	}

	items := make([]TentativePriceItemResponse, len(report.Items))
	for i := range report.Items {
		items[i] = TentativePriceItemResponse{
			SupplyID:     report.Items[i].SupplyID,
			Name:         report.Items[i].Name,
			CategoryName: report.Items[i].CategoryName,
			Price:        report.Items[i].Price.StringFixed(2),
		}
	}

	return TentativePricesResponse{
		Count: report.Count,
		Items: items,
	}
}
