package dto

import (
	domain "github.com/devpablocristo/ponti-backend/internal/class-type/usecases/domain"
	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
)

type ClassTypeResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func ClassTypeFromDomain(d *domain.ClassType) ClassTypeResponse {
	return ClassTypeResponse{ID: d.ID, Name: d.Name}
}

type ListClassTypesResponse struct {
	Data     []ClassTypeResponse `json:"data"`
	PageInfo types.PageInfo      `json:"page_info"`
}

func NewListClassTypesResponse(items []domain.ClassType, page, perPage int, total int64) ListClassTypesResponse {
	data := make([]ClassTypeResponse, 0, len(items))
	for i := range items {
		data = append(data, ClassTypeFromDomain(&items[i]))
	}
	return ListClassTypesResponse{
		Data:     data,
		PageInfo: types.NewPageInfo(page, perPage, total),
	}
}
