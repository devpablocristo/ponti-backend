package dto

import (
	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/project/usecases/domain"
)

type Project struct {
	ProjectName     string     `json:"projectName" binding:"required"`
	CustomerID      int64      `json:"customerId" binding:"required"`
	ProjectManagers []Manager  `json:"projectManagers" binding:"required,dive"`
	Investors       []Investor `json:"investors" binding:"required,dive"`
	Fields          []Field    `json:"fields" binding:"required,dive"`
}

// Manager represents a project manager.
type Manager struct {
	ID   int64  `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}

// Investor represents an investor with percentage.
type Investor struct {
	ID         int64  `json:"id" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Percentage int    `json:"percentage" binding:"required"`
}

// Field represents a field containing plots.
type Field struct {
	Name      string `json:"name" binding:"required"`
	LeaseType string `json:"leaseType" binding:"required"`
	Lots      []Lot  `json:"plots" binding:"required,dive"`
}

// Lot represents a lot within a field.
type Lot struct {
	Name         string `json:"name" binding:"required"`
	Hectares     int    `json:"hectares" binding:"required"`
	PreviousCrop string `json:"previousCrop"`
	CurrentCrop  string `json:"currentCrop" binding:"required"`
	Season       string `json:"season" binding:"required"`
}

// ToDomain transforms the  into a domain Project.
func (r *CreateProject) ToDomain() *domain.Project {
	d := &domain.Project{
		Name:       r.ProjectName,
		CustomerID: r.CustomerID,
	}
	for _, mgr := range r.ProjectManagers {
		d.Managers = append(d.Managers, domain.Manager{ID: mgr.ID, Name: mgr.Name})
	}
	for _, inv := range r.Investors {
		d.Investors = append(d.Investors, domain.InvestorDetail{ID: inv.ID, Name: inv.Name, Percentage: inv.Percentage})
	}
	for _, f := range r.Fields {
		field := domain.Field{Name: f.Name, LeaseType: f.LeaseType}
		for _, p := range f.Lots {
			field.Lots = append(field.Lots, domain.Lot{
				Name:         p.Name,
				Hectares:     p.Hectares,
				PreviousCrop: p.PreviousCrop,
				CurrentCrop:  p.CurrentCrop,
				Season:       p.Season,
			})
		}
		d.Fields = append(d.Fields, field)
	}
	return d
}

// FromDomain transforms a domain Project into a .
func FromDomain(d *domain.Project) *Project {
	r := &Project{
		ProjectName: d.Name,
		CustomerID:  d.CustomerID,
	}
	for _, mgr := range d.Managers {
		r.ProjectManagers = append(r.ProjectManagers, Manager{ID: mgr.ID, Name: mgr.Name})
	}
	for _, inv := range d.Investors {
		r.Investors = append(r.Investors, Investor{ID: inv.ID, Name: inv.Name, Percentage: inv.Percentage})
	}
	for _, fld := range d.Fields {
		field := Field{Name: fld.Name, LeaseType: fld.LeaseType}
		for _, plt := range fld.Lots {
			field.Lots = append(field.Lots, Lot{
				Name:         plt.Name,
				Hectares:     plt.Hectares,
				PreviousCrop: plt.PreviousCropID,
				CurrentCrop:  plt.CurrentCropID,
				Season:       plt.Season,
			})
		}
		r.Fields = append(r.Fields, field)
	}
	return r
}

// type Lot struct {
// 	ID             int64   // Identificador único
// 	Identifier     string  // Identificador o código del lote
// 	FieldID        int64   // ID del campo al que pertenece
// 	ProjectID      int64   // ID del proyecto asociado
// 	CurrentCropID  *int64  // Puede ser nulo: ID del cultivo actual
// 	PreviousCropID *int64  // Puede ser nulo: ID del cultivo anterior
// 	Variety        string  // Variedad (por ejemplo, del cultivo actual)
// 	Area           float64 // Superficie en hectáreas
// }
