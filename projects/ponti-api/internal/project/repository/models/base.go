package models

import (
	"time"

	campaigndom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign/usecases/domain"
	customerdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/usecases/domain"
	fielddom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	investordom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
	managerdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/usecases/domain"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
)

// --------- MODELOS ---------

type Project struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;column:id"`
	Name       string    `gorm:"size:100;not null;column:name"`
	CustomerID int64     `gorm:"not null;index;column:customer_id"`
	CampaignID int64     `gorm:"not null;index;column:campaign_id"`
	AdminCost  int64     `gorm:"not null;column:admin_cost"`
	CreatedAt  time.Time `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime;column:updated_at"`

	// Relaciones
	Customer  Customer   `gorm:"foreignKey:CustomerID;references:ID"`
	Campaign  Campaign   `gorm:"foreignKey:CampaignID;references:ID"`
	Managers  []Manager  `gorm:"many2many:project_managers;"`
	Investors []Investor `gorm:"many2many:project_investors;"`
	Fields    []Field    `gorm:"foreignKey:ProjectID"`
}

// Solo el ID para las relaciones many-to-many
type Manager struct {
	ID int64 `gorm:"primaryKey;autoIncrement:false;column:id"`
}

type Investor struct {
	ID int64 `gorm:"primaryKey;autoIncrement:false;column:id"`
}

// Solo el ID y el Name para relaciones 1-N simples
type Customer struct {
	ID   int64  `gorm:"primaryKey;autoIncrement:false;column:id"`
	Name string `gorm:"size:100;not null;column:name"`
}

type Campaign struct {
	ID   int64  `gorm:"primaryKey;autoIncrement:false;column:id"`
	Name string `gorm:"size:100;not null;column:name"`
}

// Field como hijo real del proyecto
type Field struct {
	ID        int64  `gorm:"primaryKey;autoIncrement;column:id"`
	Name      string `gorm:"size:100;not null;column:name"`
	ProjectID int64  `gorm:"not null;index;column:project_id"`
	// Agrega aquí más campos propios de Field si los necesitas
}

// --- FROM DOMAIN ---

func FromDomain(d *domain.Project) *Project {
	m := &Project{
		ID:         d.ID,
		Name:       d.Name,
		CustomerID: d.Customer.ID,
		CampaignID: d.Campaign.ID,
		AdminCost:  d.AdminCost,
		Customer: Customer{
			ID:   d.Customer.ID,
			Name: d.Customer.Name,
		},
		Campaign: Campaign{
			ID:   d.Campaign.ID,
			Name: d.Campaign.Name,
		},
		Managers:  make([]Manager, 0, len(d.Managers)),
		Investors: make([]Investor, 0, len(d.Investors)),
		Fields:    make([]Field, 0, len(d.Fields)),
	}

	// Many2Many sólo ID
	for _, mgr := range d.Managers {
		m.Managers = append(m.Managers, Manager{ID: mgr.ID})
	}
	for _, inv := range d.Investors {
		m.Investors = append(m.Investors, Investor{ID: inv.ID})
	}
	// Fields hijos completos
	for _, f := range d.Fields {
		m.Fields = append(m.Fields, Field{
			ID:        f.ID,
			Name:      f.Name,
			ProjectID: m.ID,
			// Si Field tiene más campos, agregalos acá.
		})
	}
	return m
}

// --- TO DOMAIN ---

func (m *Project) ToDomain() *domain.Project {
	d := &domain.Project{
		ID:        m.ID,
		Name:      m.Name,
		AdminCost: m.AdminCost,
		Customer:  customerdom.Customer{ID: m.Customer.ID, Name: m.Customer.Name},
		Campaign:  campaigndom.Campaign{ID: m.Campaign.ID, Name: m.Campaign.Name},
		Managers:  make([]managerdom.Manager, 0, len(m.Managers)),
		Investors: make([]investordom.Investor, 0, len(m.Investors)),
		Fields:    make([]fielddom.Field, 0, len(m.Fields)),
	}

	for _, mgr := range m.Managers {
		d.Managers = append(d.Managers, managerdom.Manager{ID: mgr.ID})
	}
	for _, inv := range m.Investors {
		d.Investors = append(d.Investors, investordom.Investor{ID: inv.ID})
	}
	for _, f := range m.Fields {
		d.Fields = append(d.Fields, fielddom.Field{
			ID:        f.ID,
			Name:      f.Name,
			ProjectID: f.ProjectID,
			// Si Field tiene más campos, agregalos acá.
		})
	}
	return d
}
