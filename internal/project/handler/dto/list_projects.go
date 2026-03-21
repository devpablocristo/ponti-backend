// Package dto define respuestas HTTP para proyectos.
package dto

import (
	"encoding/json"

	"github.com/shopspring/decimal"

	types "github.com/devpablocristo/ponti-backend/internal/shared/types"

	"github.com/devpablocristo/ponti-backend/internal/project/usecases/domain"
)

type ListedProject struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ListProjectsResponse struct {
	Items    []ListedProject `json:"items"`
	PageInfo types.PageInfo  `json:"page_info"`
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

	return ListProjectsResponse{
		Items:    out,
		PageInfo: types.NewPageInfo(page, perPage, total),
	}
}

type ProjectsResponse struct {
	Items         []Project       `json:"items"`
	TotalHectares decimal.Decimal `json:"total_hectares"`
	PageInfo      types.PageInfo  `json:"page_info"`
}

// MarshalJSON aplica redondeo de 3 decimales al campo TotalHectares
func (p ProjectsResponse) MarshalJSON() ([]byte, error) {
	aux := struct {
		Items         []Project      `json:"items"`
		TotalHectares string         `json:"total_hectares"`
		PageInfo      types.PageInfo `json:"page_info"`
	}{
		Items:         p.Items,
		TotalHectares: p.TotalHectares.Round(3).String(),
		PageInfo:      p.PageInfo,
	}
	return json.Marshal(aux)
}

func NewProjectsResponse(
	items []domain.Project,
	page, perPage int,
	total int64,
) ProjectsResponse {
	out := make([]Project, len(items))
	for i, p := range items {
		out[i] = Project{
			ID:          p.ID,
			ProjectName: p.Name,
			Customer: Customer{
				ID:   p.Customer.ID,
				Name: p.Customer.Name,
			},
			Campaign: Campaign{
				ID:   p.Campaign.ID,
				Name: p.Campaign.Name,
			},
		}

		for _, mgr := range p.Managers {
			out[i].ProjectManagers = append(out[i].ProjectManagers,
				Manager{ID: mgr.ID, Name: mgr.Name},
			)
		}

		for _, inv := range p.Investors {
			out[i].Investors = append(out[i].Investors,
				Investor{ID: inv.ID, Name: inv.Name, Percentage: inv.Percentage},
			)
		}

		for _, aci := range p.AdminCostInvestors {
			out[i].AdminCostInvestors = append(out[i].AdminCostInvestors,
				AdminCostInvestor{ID: aci.ID, Name: aci.Name, Percentage: aci.Percentage},
			)
		}
	}

	return ProjectsResponse{
		Items:    out,
		PageInfo: types.NewPageInfo(page, perPage, total),
	}
}
