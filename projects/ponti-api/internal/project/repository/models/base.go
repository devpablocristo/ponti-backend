package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/base"
	campaigndom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign/usecases/domain"
	cropmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/repository/models"
	cropdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
	customerdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/usecases/domain"
	fieldmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/repository/models"
	fielddom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	investordom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
	leasetypedom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/leasetype/usecases/domain"
	lotmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/repository/models"
	lotdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
	managerdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/usecases/domain"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
)

// --------- MODELOS ---------

type Project struct {
	ID         int64  `gorm:"primaryKey;autoIncrement;column:id"`
	Name       string `gorm:"size:100;not null;column:name"`
	CustomerID int64  `gorm:"not null;index;column:customer_id"`
	CampaignID int64  `gorm:"not null;index;column:campaign_id"`
	AdminCost  int64  `gorm:"not null;column:admin_cost"`
	base.BaseModel

	// Relaciones (SOLO para preload/query, no setear manual en insert)
	Customer  Customer          `gorm:"foreignKey:CustomerID;references:ID"`
	Campaign  Campaign          `gorm:"foreignKey:CampaignID;references:ID"`
	Managers  []Manager         `gorm:"many2many:project_managers;"`
	Investors []ProjectInvestor `gorm:"foreignKey:ProjectID;references:ID"`
	Fields    []fieldmod.Field  `gorm:"foreignKey:ProjectID"`
}

type Manager struct {
	ID   int64  `gorm:"primaryKey;autoIncrement;column:id"`
	Name string `gorm:"type:varchar(255);not null;unique"`
	base.BaseModel
}

type Customer struct {
	ID   int64  `gorm:"primaryKey;autoIncrement;column:id"`
	Name string `gorm:"type:varchar(255);not null;unique"`
}

type Campaign struct {
	ID   int64  `gorm:"primaryKey;autoIncrement;column:id"`
	Name string `gorm:"type:varchar(255);not null;unique"`
}

type Investor struct {
	ID   int64  `gorm:"primaryKey;autoIncrement;column:id"`
	Name string `gorm:"type:varchar(255);not null;unique"`
}

type ProjectInvestor struct {
	ProjectID  int64 `gorm:"primaryKey;autoIncrement:false;column:project_id"`
	InvestorID int64 `gorm:"primaryKey;autoIncrement:false;column:investor_id"`
	Percentage int   `gorm:"not null;column:percentage"`
	base.BaseModel
	Investor Investor `gorm:"foreignKey:InvestorID;references:ID"`
}

// --- FROM DOMAIN (para INSERT, no setees relaciones embebidas) ---

func FromDomain(d *domain.Project) *Project {
	m := &Project{
		ID:         d.ID,
		Name:       d.Name,
		CustomerID: d.Customer.ID,
		CampaignID: d.Campaign.ID,
		AdminCost:  d.AdminCost,
		Managers:   make([]Manager, 0, len(d.Managers)),
		Investors:  make([]ProjectInvestor, 0, len(d.Investors)),
		Fields:     make([]fieldmod.Field, 0, len(d.Fields)),
		BaseModel:  base.BaseModel{CreatedBy: d.CreatedBy, UpdatedBy: d.UpdatedBy},
	}

	for _, mgr := range d.Managers {
		m.Managers = append(m.Managers, Manager{
			ID: mgr.ID,
			BaseModel: base.BaseModel{
				CreatedBy: d.CreatedBy,
				UpdatedBy: d.UpdatedBy,
			},
		})
	}
	for _, inv := range d.Investors {
		m.Investors = append(m.Investors, ProjectInvestor{
			InvestorID: inv.ID,
			Percentage: inv.Percentage,
			BaseModel: base.BaseModel{
				CreatedBy: d.CreatedBy,
				UpdatedBy: d.UpdatedBy,
			},
		})
	}
	for key, f := range d.Fields {
		m.Fields = append(m.Fields, fieldmod.Field{
			ID:               f.ID,
			Name:             f.Name,
			ProjectID:        d.ID,
			LeaseTypeID:      f.LeaseType.ID,
			LeaseTypePercent: f.LeaseTypePercent,
			LeaseTypeValue:   f.LeaseTypeValue,
			Lots:             make([]lotmod.Lot, 0, len(f.Lots)),
			BaseModel:        base.BaseModel{CreatedBy: d.CreatedBy, UpdatedBy: d.UpdatedBy},
		})

		for _, l := range f.Lots {
			m.Fields[key].Lots = append(m.Fields[key].Lots, lotmod.Lot{
				ID:             l.ID,
				Name:           l.Name,
				FieldID:        d.ID,
				Hectares:       l.Hectares,
				Season:         l.Season,
				PreviousCropID: l.PreviousCrop.ID,
				CurrentCropID:  l.CurrentCrop.ID,
				PreviousCrop:   cropmod.Crop{ID: l.PreviousCrop.ID, Name: l.PreviousCrop.Name},
				CurrentCrop:    cropmod.Crop{ID: l.CurrentCrop.ID, Name: l.CurrentCrop.Name},
				BaseModel:      base.BaseModel{CreatedBy: d.CreatedBy, UpdatedBy: d.UpdatedBy},
			})
		}
	}
	return m
}

func (m *Project) ToDomain() *domain.Project {
	d := &domain.Project{
		ID:        m.ID,
		Name:      m.Name,
		AdminCost: m.AdminCost,
		Customer: customerdom.Customer{
			ID:   m.CustomerID,
			Name: m.Customer.Name,
		},
		Campaign: campaigndom.Campaign{
			ID:   m.CampaignID,
			Name: m.Campaign.Name,
		},
		Managers:  make([]managerdom.Manager, 0, len(m.Managers)),
		Investors: make([]investordom.Investor, 0, len(m.Investors)),
		Fields:    make([]fielddom.Field, 0, len(m.Fields)),
		UpdatedAt: &m.UpdatedAt,
		CreatedBy: m.CreatedBy,
		UpdatedBy: m.UpdatedBy,
	}

	for _, mgr := range m.Managers {
		d.Managers = append(d.Managers, managerdom.Manager{
			ID:   mgr.ID,
			Name: mgr.Name, // Solo si preload
		})
	}
	for _, piv := range m.Investors {
		d.Investors = append(d.Investors, investordom.Investor{
			ID:         piv.InvestorID,
			Name:       piv.Investor.Name,
			Percentage: piv.Percentage,
		})
	}
	for _, f := range m.Fields {
		field := fielddom.Field{
			ID:               f.ID,
			Name:             f.Name,
			ProjectID:        f.ProjectID,
			LeaseType:        &leasetypedom.LeaseType{ID: f.LeaseTypeID, Name: f.LeaseType.Name},
			LeaseTypePercent: f.LeaseTypePercent,
			LeaseTypeValue:   f.LeaseTypeValue,
			Lots:             make([]lotdom.Lot, 0, len(f.Lots)),
		}

		for _, l := range f.Lots {
			field.Lots = append(field.Lots, lotdom.Lot{
				ID:       l.ID,
				Name:     l.Name,
				FieldID:  l.FieldID,
				Hectares: l.Hectares,
				Season:   l.Season,
				PreviousCrop: cropdom.Crop{
					ID:   l.PreviousCrop.ID,
					Name: l.PreviousCrop.Name,
				},
				CurrentCrop: cropdom.Crop{
					ID:   l.CurrentCrop.ID,
					Name: l.CurrentCrop.Name,
				},
			})
		}

		d.Fields = append(d.Fields, field)
	}
	return d
}
