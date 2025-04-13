package dto

import (
	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/project/usecases/domain"
)

// Project is the DTO for a specific project.
type Project struct {
	ID               int64   `json:"id"`
	Name             string  `json:"name"`
	CustomerID       int64   `json:"customer_id"`
	ProjectAdminCost float64 `json:"project_admin_cost"`
	AdminResponsible string  `json:"admin_responsible"`
}

// ToDomain converts the DTO Project to the domain entity.
func (p Project) ToDomain() *domain.Project {
	return &domain.Project{
		ID:               p.ID,
		Name:             p.Name,
		CustomerID:       p.CustomerID,
		ProjectAdminCost: p.ProjectAdminCost,
		AdminResponsible: p.AdminResponsible,
	}
}

// FromDomain converts a domain Project to the DTO.
func FromDomain(d domain.Project) *Project {
	return &Project{
		ID:               d.ID,
		Name:             d.Name,
		CustomerID:       d.CustomerID,
		ProjectAdminCost: d.ProjectAdminCost,
		AdminResponsible: d.AdminResponsible,
	}
}
