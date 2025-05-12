// ================================
// File: internal/project/repository/models/project.go
// ================================
package models

import (
	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/project/usecases/domain"
)

// Project es el modelo GORM para la entidad Project con todas sus relaciones.
type Project struct {
	ID         int64             `gorm:"primaryKey"`
	Name       string            `gorm:"type:varchar(100);not null"`
	CustomerID int64             `gorm:"not null"`
	Customer   Client            `gorm:"foreignKey:CustomerID"`
	Managers   []Manager         `gorm:"many2many:project_managers;"`
	Investors  []ProjectInvestor `gorm:"foreignKey:ProjectID"`
	Fields     []Field           `gorm:"foreignKey:ProjectID"`
}

// Client representa el cliente relacionado a un proyecto.
type Client struct {
	ID   int64  `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(100);not null"`
}

// Manager representa un manager asociado a uno o más proyectos.
type Manager struct {
	ID     int64 `gorm:"primaryKey"`
	UserID int64
	User   User   `gorm:"foreignKey:UserID"`
	Title  string `gorm:"type:varchar(100)"`
}

// User representa al usuario subyacente de un Manager o Investor.
type User struct {
	ID    int64  `gorm:"primaryKey"`
	Name  string `gorm:"type:varchar(100);not null"`
	Email string `gorm:"type:varchar(100);unique;not null"`
}

// ProjectInvestor es la tabla pivote con campo extra de porcentaje.
type ProjectInvestor struct {
	ProjectID  int64    `gorm:"primaryKey"`
	InvestorID int64    `gorm:"primaryKey"`
	Percentage int      `gorm:"not null"`
	Investor   Investor `gorm:"foreignKey:InvestorID"`
}

// Investor representa la entidad Investor.
type Investor struct {
	ID      int64 `gorm:"primaryKey"`
	UserID  int64
	User    User   `gorm:"foreignKey:UserID"`
	Company string `gorm:"type:varchar(100)"`
}

// Field representa un lote dentro de un proyecto.
type Field struct {
	ID        int64  `gorm:"primaryKey"`
	Name      string `gorm:"type:varchar(100);not null"`
	LeaseType string `gorm:"type:varchar(50);not null"`
	ProjectID int64  `gorm:"not null;index"`
	Lots      []Lot  `gorm:"foreignKey:FieldID"`
}

// Lot representa una parcela dentro de un lote.
type Lot struct {
	ID           int64  `gorm:"primaryKey"`
	Name         string `gorm:"type:varchar(100);not null"`
	Hectares     int    `gorm:"not null"`
	PreviousCrop string `gorm:"type:varchar(100)"`
	CurrentCrop  string `gorm:"type:varchar(100);not null"`
	Season       string `gorm:"type:varchar(50);not null"`
	FieldID      int64  `gorm:"not null;index"`
}

func (p *Project) ToDomain() *domain.Project {
	d := &domain.Project{
		ID:         p.ID,
		Name:       p.Name,
		CustomerID: p.CustomerID,
		Customer: domain.Client{
			ID:   p.Customer.ID,
			Name: p.Customer.Name,
		},
	}
	for _, m := range p.Managers {
		d.Managers = append(d.Managers, domain.Manager{
			ID:   m.ID,
			Name: m.User.Name,
		})
	}
	for _, pi := range p.Investors {
		d.Investors = append(d.Investors, domain.InvestorDetail{
			ID:         pi.InvestorID,
			Name:       pi.Investor.User.Name,
			Percentage: pi.Percentage,
		})
	}
	for _, f := range p.Fields {
		fld := domain.Field{
			Name:      f.Name,
			LeaseType: f.LeaseType,
		}
		for _, plt := range f.Lots {
			fld.Lots = append(fld.Lots, domain.Lot{
				Name:         plt.Name,
				Hectares:     plt.Hectares,
				PreviousCrop: plt.PreviousCrop,
				CurrentCrop:  plt.CurrentCrop,
				Season:       plt.Season,
			})
		}
		d.Fields = append(d.Fields, fld)
	}
	return d
}

// FromDomain convierte una entidad domain.Project a un modelo GORM Project.
func FromDomain(d *domain.Project) *Project {
	p := &Project{
		ID:         d.ID,
		Name:       d.Name,
		CustomerID: d.CustomerID,
	}
	// Relaciones many-to-many managers
	for _, m := range d.Managers {
		p.Managers = append(p.Managers, Manager{ID: m.ID})
	}
	// Pivote inversores
	for _, inv := range d.Investors {
		p.Investors = append(p.Investors, ProjectInvestor{
			ProjectID:  d.ID,
			InvestorID: inv.ID,
			Percentage: inv.Percentage,
		})
	}
	// Fields y plots
	for _, f := range d.Fields {
		fld := Field{
			Name:      f.Name,
			LeaseType: f.LeaseType,
		}
		for _, plt := range f.Lots {
			fld.Lots = append(fld.Lots, Lot{
				Name:         plt.Name,
				Hectares:     plt.Hectares,
				PreviousCrop: plt.PreviousCrop,
				CurrentCrop:  plt.CurrentCrop,
				Season:       plt.Season,
			})
		}
		p.Fields = append(p.Fields, fld)
	}
	return p
}
