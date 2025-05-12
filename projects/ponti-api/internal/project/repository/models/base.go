package models

import (
	customerdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/usecases/domain"
	fielddom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	investordom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
	managerdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/usecases/domain"
	projectdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
)

// Project es el modelo GORM para proyectos.
// Sólo persiste su propio ID, nombre y los IDs de las relaciones.
type Project struct {
	ID         int64  `gorm:"primaryKey"`
	Name       string `gorm:"type:varchar(100);not null"`
	CustomerID int64  `gorm:"not null;index;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	Managers  []Manager  `gorm:"many2many:project_managers;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Investors []Investor `gorm:"many2many:project_investors;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Fields    []Field    `gorm:"foreignKey:ProjectID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// Manager sólo expone el ID para la tabla pivote project_managers.
type Manager struct {
	ID int64 `gorm:"primaryKey"`
}

// Investor sólo expone el ID para la tabla pivote project_investors.
type Investor struct {
	ID int64 `gorm:"primaryKey"`
}

// Field tiene su propio ID y la FK ProjectID; no almacena nada más.
type Field struct {
	ID        int64 `gorm:"primaryKey"`
	ProjectID int64 `gorm:"not null;index"`
}

// FromDomain convierte el dominio a modelo GORM, guardando solo IDs.
func FromDomain(d *projectdom.Project) *Project {
	m := &Project{
		ID:         d.ID,
		Name:       d.Name,
		CustomerID: d.Customer.ID,
	}
	// Managers
	for _, mgr := range d.Managers {
		m.Managers = append(m.Managers, Manager{ID: mgr.ID})
	}
	// Investors
	for _, inv := range d.Investors {
		m.Investors = append(m.Investors, Investor{ID: inv.ID})
	}
	// Fields
	for _, fld := range d.Fields {
		m.Fields = append(m.Fields, Field{ID: fld.ID, ProjectID: d.ID})
	}
	return m
}

// ToDomain convierte el modelo GORM a dominio, cargando solo IDs.
func (m *Project) ToDomain() *projectdom.Project {
	d := &projectdom.Project{
		ID:   m.ID,
		Name: m.Name,
		Customer: customerdom.Customer{
			ID: m.CustomerID,
		},
	}
	// Managers
	for _, mgr := range m.Managers {
		d.Managers = append(d.Managers, managerdom.Manager{ID: mgr.ID})
	}
	// Investors
	for _, inv := range m.Investors {
		d.Investors = append(d.Investors, investordom.Investor{ID: inv.ID})
	}
	// Fields
	for _, fld := range m.Fields {
		d.Fields = append(d.Fields, fielddom.Field{ID: fld.ID})
	}
	return d
}
