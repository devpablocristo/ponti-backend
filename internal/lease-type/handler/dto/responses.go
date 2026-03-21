package dto

import (
	domain "github.com/devpablocristo/ponti-backend/internal/lease-type/usecases/domain"
	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
)

type LeaseTypeResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func LeaseTypeFromDomain(d *domain.LeaseType) LeaseTypeResponse {
	return LeaseTypeResponse{ID: d.ID, Name: d.Name}
}

type ListLeaseTypesResponse struct {
	Data     []LeaseTypeResponse `json:"data"`
	PageInfo types.PageInfo      `json:"page_info"`
}

func NewListLeaseTypesResponse(items []domain.LeaseType, page, perPage int, total int64) ListLeaseTypesResponse {
	data := make([]LeaseTypeResponse, 0, len(items))
	for i := range items {
		data = append(data, LeaseTypeFromDomain(&items[i]))
	}
	return ListLeaseTypesResponse{
		Data:     data,
		PageInfo: types.NewPageInfo(page, perPage, total),
	}
}
