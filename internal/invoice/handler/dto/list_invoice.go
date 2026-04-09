package dto

import (
	domain "github.com/devpablocristo/ponti-backend/internal/invoice/usecases/domain"
)

type PageInfo struct {
	PerPage int   `json:"per_page"`
	Page    int   `json:"page"`
	MaxPage int   `json:"max_page"`
	Total   int64 `json:"total"`
}

type ListInvoicesResponse struct {
	Data     []InvoiceResponse `json:"data"`
	PageInfo PageInfo          `json:"page_info"`
}

func NewListInvoicesResponse(
	items []domain.Invoice,
	page, perPage int,
	total int64,
) ListInvoicesResponse {
	out := make([]InvoiceResponse, len(items))
	for i, item := range items {
		out[i] = FromDomain(&item)
	}

	maxPage := int((total + int64(perPage) - 1) / int64(perPage))
	return ListInvoicesResponse{
		Data: out,
		PageInfo: PageInfo{
			PerPage: perPage,
			Page:    page,
			MaxPage: maxPage,
			Total:   total,
		},
	}
}
