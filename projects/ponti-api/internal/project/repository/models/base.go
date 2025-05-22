// File: repository/models/base.go
package models

import (
	"time"

	customerdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/usecases/domain"
	fieldmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/repository/models"
	fielddom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	invmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/repository/models"
	managerdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/usecases/domain"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
)

// Project es el modelo GORM para proyectos.
type Project struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;column:id"`
	Name       string    `gorm:"size:100;not null;column:name"`
	CustomerID int64     `gorm:"not null;index;column:customer_id;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	CreatedAt  time.Time `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime;column:updated_at"`

	Managers  []Manager         `gorm:"many2many:project_managers;association_autocreate:false;association_autoupdate:false;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Investors []ProjectInvestor `gorm:"foreignKey:ProjectID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Fields    []fieldmod.Field  `gorm:"foreignKey:ProjectID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// Manager sólo expone el ID para la tabla pivote project_managers.
type Manager struct {
	ID int64 `gorm:"primaryKey;column:id;autoIncrement:false"`
}

type ProjectInvestor struct {
	ProjectID  int64           `gorm:"primaryKey;column:project_id"`
	InvestorID int64           `gorm:"primaryKey;column:investor_id"`
	Percentage float64         `gorm:"not null"`
	Investor   invmod.Investor `gorm:"foreignKey:InvestorID"`
}

// FromDomain convierte un domain.Project al modelo GORM Project.
func FromDomain(d *domain.Project) *Project {
	m := &Project{
		Name:       d.Name,
		CustomerID: d.Customer.ID,
	}
	for _, mgr := range d.Managers {
		m.Managers = append(m.Managers, Manager{ID: mgr.ID})
	}
	for _, inv := range d.Investors {
		m.Investors = append(m.Investors, ProjectInvestor{
			InvestorID: inv.ID,
			Percentage: float64(inv.Percentage),
		})
	}
	return m
}

// ToDomain convierte el modelo GORM Project a domain.Project.
func (m *Project) ToDomain() *domain.Project {
	d := &domain.Project{
		ID:   m.ID,
		Name: m.Name,
		Customer: customerdom.Customer{
			ID: m.CustomerID,
		},
	}
	for _, mgr := range m.Managers {
		d.Managers = append(d.Managers, managerdom.Manager{ID: mgr.ID})
	}
	for _, fld := range m.Fields {
		d.Fields = append(d.Fields, fielddom.Field{
			ID:          fld.ID,
			ProjectID:   fld.ProjectID,
			Name:        fld.Name,
			LeaseTypeID: fld.LeaseTypeID,
			Lots:        fld.ToDomain().Lots,
		})
	}
	return d
}
