// File: handler/dto/list_projects.go
package dto

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
)

type ListedProject struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type PageInfo struct {
	PerPage int   `json:"per_page"`
	Page    int   `json:"page"`
	MaxPage int   `json:"max_page"`
	Total   int64 `json:"total"`
}

type ListProjectsResponse struct {
	Data     []ListedProject `json:"data"`
	PageInfo PageInfo        `json:"page_info"`
}

func NewListProjectsResponse(
	items []domain.ListedProject,
	page, perPage int,
	total int64,
) ListProjectsResponse {
	out := make([]ListedProject, len(items))
	for i, p := range items {
		out[i] = ListedProject{ID: p.ID, Name: p.Name}
	}

	maxPage := int((total + int64(perPage) - 1) / int64(perPage))

	return ListProjectsResponse{
		Data: out,
		PageInfo: PageInfo{
			PerPage: perPage,
			Page:    page,
			MaxPage: maxPage,
			Total:   total,
		},
	}
}
