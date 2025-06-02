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
	Customer  Customer          `gorm:"foreignKey:CustomerID;references:ID"`
	Campaign  Campaign          `gorm:"foreignKey:CampaignID;references:ID"`
	Managers  []Manager         `gorm:"many2many:project_managers;"`
	Investors []ProjectInvestor `gorm:"foreignKey:ProjectID;references:ID"`
	Fields    []Field           `gorm:"foreignKey:ProjectID"`
}

// Manager y Customer solo usan el ID para las relaciones
type Manager struct {
	ID int64 `gorm:"primaryKey;autoIncrement;column:id"`
}

type Customer struct {
	ID int64 `gorm:"primaryKey;autoIncrement;column:id"`
}

type Campaign struct {
	ID int64 `gorm:"primaryKey;autoIncrement:false;column:id"`
}

type Investor struct {
	ID   int64  `gorm:"primaryKey;autoIncrement:false;column:id"`
	Name string `gorm:"size:255;not null;column:name"`
}

// Tabla pivote con campo extra "Percentage"
type ProjectInvestor struct {
	ProjectID  int64 `gorm:"primaryKey;autoIncrement:false;column:project_id"`
	InvestorID int64 `gorm:"primaryKey;autoIncrement:false;column:investor_id"`
	Percentage int   `gorm:"not null;column:percentage"`

	Investor Investor `gorm:"foreignKey:InvestorID;references:ID"`
}

// Field como hijo real del proyecto (agregá más campos según tu dominio)
type Field struct {
	ID        int64  `gorm:"primaryKey;autoIncrement;column:id"`
	Name      string `gorm:"size:100;not null;column:name"`
	ProjectID int64  `gorm:"not null;index;column:project_id"`
	// Otros campos de Field si hay...
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
			ID: d.Customer.ID,
		},
		Campaign: Campaign{
			ID: d.Campaign.ID,
		},
		Managers:  make([]Manager, 0, len(d.Managers)),
		Investors: make([]ProjectInvestor, 0, len(d.Investors)),
		Fields:    make([]Field, 0, len(d.Fields)),
	}

	// Many2Many: Managers solo IDs
	for _, mgr := range d.Managers {
		m.Managers = append(m.Managers, Manager{ID: mgr.ID})
	}
	// ProjectInvestors: asignar percentage también
	for _, inv := range d.Investors {
		m.Investors = append(m.Investors, ProjectInvestor{
			ProjectID:  d.ID, // GORM la pone igual, pero por las dudas
			InvestorID: inv.ID,
			Percentage: inv.Percentage,
			Investor: Investor{
				ID: inv.ID,
				// Name no siempre disponible acá, solo para preload/join
			},
		})
	}
	// Fields hijos completos (sumá más campos si corresponde)
	for _, f := range d.Fields {
		m.Fields = append(m.Fields, Field{
			ID:        f.ID,
			Name:      f.Name,
			ProjectID: d.ID,
			// Otros campos si hay...
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
		Customer:  customerdom.Customer{ID: m.Customer.ID},
		Campaign:  campaigndom.Campaign{ID: m.Campaign.ID},
		Managers:  make([]managerdom.Manager, 0, len(m.Managers)),
		Investors: make([]investordom.Investor, 0, len(m.Investors)),
		Fields:    make([]fielddom.Field, 0, len(m.Fields)),
	}

	for _, mgr := range m.Managers {
		d.Managers = append(d.Managers, managerdom.Manager{ID: mgr.ID})
	}
	for _, piv := range m.Investors {
		d.Investors = append(d.Investors, investordom.Investor{
			ID:         piv.InvestorID,
			Name:       piv.Investor.Name, // solo si preload
			Percentage: piv.Percentage,
		})
	}
	for _, f := range m.Fields {
		d.Fields = append(d.Fields, fielddom.Field{
			ID:        f.ID,
			Name:      f.Name,
			ProjectID: f.ProjectID,
			// Otros campos...
		})
	}
	return d
}
