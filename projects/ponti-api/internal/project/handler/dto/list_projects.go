// File: handler/dto/list_projects.go
package dto

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
)

// ListedProject es el DTO ligero con solo ID y Name para listados
type ListedProject struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// PageInfo contiene la metadata de paginación
type PageInfo struct {
	PerPage int   `json:"per_page"`
	Page    int   `json:"page"`
	MaxPage int   `json:"max_page"`
	Total   int64 `json:"total"`
}

// ListProjectsResponse es la respuesta estándar para listados paginados
type ListProjectsResponse struct {
	Data     []ListedProject `json:"data"`
	PageInfo PageInfo        `json:"page_info"`
}

// NewListProjectsResponse crea un ListProjectsResponse a partir de los proyectos ligeros,
// el número de página, tamaño por página y total de registros.
func NewListProjectsResponse(
	items []domain.ListedProject,
	page, perPage int,
	total int64,
) ListProjectsResponse {
	// Mapear domain.ListedProject a DTO ListedProject
	out := make([]ListedProject, len(items))
	for i, p := range items {
		out[i] = ListedProject{ID: p.ID, Name: p.Name}
	}

	// Calcular número máximo de páginas
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
