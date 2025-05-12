package dto

import (
	cropdom "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/crop/usecases/domain"
	customerdom "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/customer/usecases/domain"
	fielddom "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/field/usecases/domain"
	investordom "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/investor/usecases/domain"
	lotdom "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/lot/usecases/domain"
	managerdom "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/manager/usecases/domain"
	projectdom "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/project/usecases/domain"
)

// Project DTO for create/update and response
type Project struct {
	ProjectName     string     `json:"project_name" binding:"required"`
	Customer        Customer   `json:"customer" binding:"required"`
	ProjectManagers []Manager  `json:"project_managers" binding:"required,dive,required"`
	Investors       []Investor `json:"investors" binding:"required,dive,required"`
	Fields          []Field    `json:"fields" binding:"required,dive,required"`
}

// Customer DTO
type Customer struct {
	ID   int64  `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}

// Manager DTO
type Manager struct {
	ID   int64  `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}

// Investor DTO with percentage
type Investor struct {
	ID         int64  `json:"id" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Percentage int    `json:"percentage" binding:"required"`
}

// Field DTO including nested lots
type Field struct {
	Name        string `json:"name" binding:"required"`
	LeaseTypeID int64  `json:"lease_type_id" binding:"required"`
	Lots        []Lot  `json:"lots" binding:"required,dive,required"`
}

// Lot DTO referencing crops by ID
type Lot struct {
	Name           string  `json:"name" binding:"required"`
	Hectares       float64 `json:"hectares" binding:"required"`
	PreviousCropID int64   `json:"previous_crop_id" binding:"required"`
	CurrentCropID  int64   `json:"current_crop_id" binding:"required"`
	Season         string  `json:"season" binding:"required"`
}

// ToDomain maps the DTO to the domain.Project
func (r *Project) ToDomain() *projectdom.Project {
	d := &projectdom.Project{
		Name: r.ProjectName,
		Customer: customerdom.Customer{
			ID:   r.Customer.ID,
			Name: r.Customer.Name,
		},
	}
	for _, mgr := range r.ProjectManagers {
		d.Managers = append(d.Managers, managerdom.Manager{ID: mgr.ID, Name: mgr.Name})
	}
	for _, inv := range r.Investors {
		d.Investors = append(d.Investors, investordom.Investor{ID: inv.ID, Name: inv.Name, Percentage: inv.Percentage})
	}
	for _, f := range r.Fields {
		fld := fielddom.Field{ID: 0, Name: f.Name, LeaseTypeID: f.LeaseTypeID}
		for _, lt := range f.Lots {
			fld.Lots = append(fld.Lots, lotdom.Lot{
				Name:         lt.Name,
				Hectares:     lt.Hectares,
				PreviousCrop: cropdom.Crop{ID: lt.PreviousCropID},
				CurrentCrop:  cropdom.Crop{ID: lt.CurrentCropID},
				Season:       lt.Season,
			})
		}
		d.Fields = append(d.Fields, fld)
	}
	return d
}

// FromDomain maps a domain.Project to the DTO
func FromDomain(d *projectdom.Project) *Project {
	r := &Project{
		ProjectName: d.Name,
		Customer:    Customer{ID: d.Customer.ID, Name: d.Customer.Name},
	}
	for _, mgr := range d.Managers {
		r.ProjectManagers = append(r.ProjectManagers, Manager{ID: mgr.ID, Name: mgr.Name})
	}
	for _, inv := range d.Investors {
		r.Investors = append(r.Investors, Investor{ID: inv.ID, Name: inv.Name, Percentage: inv.Percentage})
	}
	for _, fld := range d.Fields {
		dtoF := Field{Name: fld.Name, LeaseTypeID: fld.LeaseTypeID}
		for _, lt := range fld.Lots {
			dtoF.Lots = append(dtoF.Lots, Lot{Name: lt.Name, Hectares: lt.Hectares, PreviousCropID: lt.PreviousCrop.ID, CurrentCropID: lt.CurrentCrop.ID, Season: lt.Season})
		}
		r.Fields = append(r.Fields, dtoF)
	}
	return r
}
