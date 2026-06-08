// Package dto define los contratos HTTP del registry.
package dto

import (
	domain "github.com/devpablocristo/ponti-backend/internal/registry/usecases/domain"
)

// SetAliasesRequest reemplaza el conjunto de alias de un actor.
type SetAliasesRequest struct {
	Aliases []string `json:"aliases"`
}

type rowResponse struct {
	EntityType string   `json:"entity_type"`
	ID         int64    `json:"id"`
	Name       string   `json:"name"`
	Tax        string   `json:"tax,omitempty"`
	Roles      []string `json:"roles"`
	Archived   bool     `json:"archived"`
}

type pageInfo struct {
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	Total   int64 `json:"total"`
	MaxPage int   `json:"max_page"`
}

// SearchResponse es la página de resultados del registry.
type SearchResponse struct {
	Data     []rowResponse `json:"data"`
	PageInfo pageInfo      `json:"page_info"`
}

// NewSearchResponse arma la respuesta paginada desde el resultado de dominio.
func NewSearchResponse(res domain.RegistryResult, page, perPage int) SearchResponse {
	rows := make([]rowResponse, 0, len(res.Rows))
	for _, r := range res.Rows {
		roles := r.Roles
		if roles == nil {
			roles = []string{}
		}
		rows = append(rows, rowResponse{
			EntityType: r.EntityType, ID: r.ID, Name: r.Name,
			Tax: r.Tax, Roles: roles, Archived: r.Archived,
		})
	}
	maxPage := 1
	if perPage > 0 {
		maxPage = int((res.Total + int64(perPage) - 1) / int64(perPage))
		if maxPage < 1 {
			maxPage = 1
		}
	}
	return SearchResponse{
		Data:     rows,
		PageInfo: pageInfo{Page: page, PerPage: perPage, Total: res.Total, MaxPage: maxPage},
	}
}
