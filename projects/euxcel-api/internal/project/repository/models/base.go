package models

import (
	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/project/usecases/domain"
)

// Project represents the GORM model for a project.
type Project struct {
	ID               int64   `gorm:"primaryKey"`
	Name             string  `gorm:"type:varchar(100);not null"`
	CustomerID       int64   `gorm:"not null"`
	ProjectAdminCost float64 `gorm:"not null"`
	AdminResponsible string  `gorm:"type:varchar(100);not null"`
}

// ToDomain converts the Project model to the domain entity.
func (p Project) ToDomain() *domain.Project {
	return &domain.Project{
		ID:               p.ID,
		Name:             p.Name,
		CustomerID:       p.CustomerID,
		ProjectAdminCost: p.ProjectAdminCost,
		AdminResponsible: p.AdminResponsible,
	}
}

// FromDomainProject converts a domain Project entity to the GORM model.
func FromDomainProject(d *domain.Project) *Project {
	return &Project{
		ID:               d.ID,
		Name:             d.Name,
		CustomerID:       d.CustomerID,
		ProjectAdminCost: d.ProjectAdminCost,
		AdminResponsible: d.AdminResponsible,
	}
}
