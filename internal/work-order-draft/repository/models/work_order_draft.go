package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	campaignmod "github.com/devpablocristo/ponti-backend/internal/campaign/repository/models"
	cropmod "github.com/devpablocristo/ponti-backend/internal/crop/repository/models"
	customermod "github.com/devpablocristo/ponti-backend/internal/customer/repository/models"
	fieldmod "github.com/devpablocristo/ponti-backend/internal/field/repository/models"
	investormod "github.com/devpablocristo/ponti-backend/internal/investor/repository/models"
	labormod "github.com/devpablocristo/ponti-backend/internal/labor/repository/models"
	lotmod "github.com/devpablocristo/ponti-backend/internal/lot/repository/models"
	projectmod "github.com/devpablocristo/ponti-backend/internal/project/repository/models"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
	supplymod "github.com/devpablocristo/ponti-backend/internal/supply/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/work-order-draft/usecases/domain"
)

type WorkOrderDraft struct {
	ID                   int64                         `gorm:"primaryKey;column:id"`
	Number               string                        `gorm:"column:number;not null"`
	Date                 time.Time                     `gorm:"column:date;type:date;not null"`
	CustomerID           int64                         `gorm:"column:customer_id;not null"`
	Customer             customermod.Customer          `gorm:"foreignKey:CustomerID"`
	ProjectID            int64                         `gorm:"column:project_id;not null"`
	Project              projectmod.Project            `gorm:"foreignKey:ProjectID"`
	CampaignID           *int64                        `gorm:"column:campaign_id"`
	Campaign             *campaignmod.Campaign         `gorm:"foreignKey:CampaignID"`
	FieldID              int64                         `gorm:"column:field_id;not null"`
	Field                fieldmod.Field                `gorm:"foreignKey:FieldID"`
	LotID                int64                         `gorm:"column:lot_id;not null"`
	Lot                  lotmod.Lot                    `gorm:"foreignKey:LotID"`
	CropID               int64                         `gorm:"column:crop_id;not null"`
	Crop                 cropmod.Crop                  `gorm:"foreignKey:CropID"`
	LaborID              int64                         `gorm:"column:labor_id;not null"`
	Labor                labormod.Labor                `gorm:"foreignKey:LaborID"`
	Contractor           string                        `gorm:"column:contractor;size:100;not null"`
	EffectiveArea        decimal.Decimal               `gorm:"column:effective_area;not null"`
	Observations         string                        `gorm:"column:observations"`
	InvestorID           int64                         `gorm:"column:investor_id;not null"`
	Investor             investormod.Investor          `gorm:"foreignKey:InvestorID"`
	IsDigital            bool                          `gorm:"column:is_digital"`
	Status               string                        `gorm:"column:status;size:30;not null"`
	ReviewedBy           *int64                        `gorm:"column:reviewed_by"`
	PublishedWorkOrderID *int64                        `gorm:"column:published_work_order_id"`
	ReviewNotes          string                        `gorm:"column:review_notes"`
	DeletedAt            gorm.DeletedAt                `gorm:"index"`
	Items                []WorkOrderDraftItem          `gorm:"foreignKey:DraftID;references:ID;constraint:OnDelete:CASCADE"`
	InvestorSplits       []WorkOrderDraftInvestorSplit `gorm:"foreignKey:DraftID;references:ID;constraint:OnDelete:CASCADE"`

	sharedmodels.Base
}

func (WorkOrderDraft) TableName() string { return "work_order_drafts" }

type WorkOrderDraftItem struct {
	ID         int64            `gorm:"primaryKey;autoIncrement"`
	DraftID    int64            `gorm:"column:draft_id;index"`
	SupplyID   int64            `gorm:"column:supply_id;not null"`
	SupplyName string           `gorm:"column:supply_name;not null"`
	Supply     supplymod.Supply `gorm:"foreignKey:SupplyID"`
	TotalUsed  decimal.Decimal  `gorm:"column:total_used;not null"`
	FinalDose  decimal.Decimal  `gorm:"column:final_dose;not null"`
}

func (WorkOrderDraftItem) TableName() string { return "work_order_draft_items" }

type WorkOrderDraftInvestorSplit struct {
	ID         int64                `gorm:"primaryKey;autoIncrement"`
	DraftID    int64                `gorm:"column:draft_id;index"`
	InvestorID int64                `gorm:"column:investor_id;not null"`
	Investor   investormod.Investor `gorm:"foreignKey:InvestorID"`
	Percentage decimal.Decimal      `gorm:"column:percentage;not null"`
}

func (WorkOrderDraftInvestorSplit) TableName() string { return "work_order_draft_investor_splits" }

func FromDomain(d *domain.WorkOrderDraft) *WorkOrderDraft {
	model := &WorkOrderDraft{
		Number:               d.Number,
		Date:                 d.Date,
		CustomerID:           d.CustomerID,
		ProjectID:            d.ProjectID,
		CampaignID:           d.CampaignID,
		FieldID:              d.FieldID,
		LotID:                d.LotID,
		CropID:               d.CropID,
		LaborID:              d.LaborID,
		Contractor:           d.Contractor,
		EffectiveArea:        d.EffectiveArea,
		Observations:         d.Observations,
		InvestorID:           d.InvestorID,
		IsDigital:            d.IsDigital,
		Status:               string(d.Status),
		ReviewedBy:           d.ReviewedBy,
		PublishedWorkOrderID: d.PublishedWorkOrderID,
		ReviewNotes:          d.ReviewNotes,
	}

	if d.ID > 0 {
		model.ID = d.ID
	}

	if len(d.Items) > 0 {
		items := make([]WorkOrderDraftItem, len(d.Items))
		for i, item := range d.Items {
			items[i] = WorkOrderDraftItem{
				SupplyID:   item.SupplyID,
				SupplyName: item.SupplyName,
				TotalUsed:  item.TotalUsed,
				FinalDose:  item.FinalDose,
			}
		}
		model.Items = items
	}

	if len(d.InvestorSplits) > 0 {
		splits := make([]WorkOrderDraftInvestorSplit, len(d.InvestorSplits))
		for i, split := range d.InvestorSplits {
			splits[i] = WorkOrderDraftInvestorSplit{
				InvestorID: split.InvestorID,
				Percentage: split.Percentage,
			}
		}
		model.InvestorSplits = splits
	}

	return model
}

func (m *WorkOrderDraft) ToDomain() *domain.WorkOrderDraft {
	items := make([]domain.WorkOrderDraftItem, len(m.Items))
	for i, item := range m.Items {
		items[i] = domain.WorkOrderDraftItem{
			SupplyID:   item.SupplyID,
			SupplyName: item.SupplyName,
			TotalUsed:  item.TotalUsed,
			FinalDose:  item.FinalDose,
		}
	}

	splits := make([]domain.WorkOrderDraftInvestorSplit, len(m.InvestorSplits))
	for i, split := range m.InvestorSplits {
		splits[i] = domain.WorkOrderDraftInvestorSplit{
			InvestorID: split.InvestorID,
			Percentage: split.Percentage,
		}
	}

	return &domain.WorkOrderDraft{
		ID:                   m.ID,
		Number:               m.Number,
		Date:                 m.Date,
		CustomerID:           m.CustomerID,
		CustomerName:         m.Customer.Name,
		ProjectID:            m.ProjectID,
		ProjectName:          m.Project.Name,
		CampaignID:           m.CampaignID,
		CampaignName:         draftCampaignName(m),
		FieldID:              m.FieldID,
		FieldName:            m.Field.Name,
		LotID:                m.LotID,
		LotName:              m.Lot.Name,
		CropID:               m.CropID,
		CropName:             m.Crop.Name,
		LaborID:              m.LaborID,
		LaborName:            m.Labor.Name,
		Contractor:           m.Contractor,
		EffectiveArea:        m.EffectiveArea,
		Observations:         m.Observations,
		InvestorID:           m.InvestorID,
		IsDigital:            m.IsDigital,
		Status:               domain.Status(m.Status),
		ReviewedBy:           m.ReviewedBy,
		PublishedWorkOrderID: m.PublishedWorkOrderID,
		ReviewNotes:          m.ReviewNotes,
		Items:                items,
		InvestorSplits:       splits,
		Base: shareddomain.Base{
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
		},
	}
}

func draftCampaignName(d *WorkOrderDraft) string {
	if d == nil {
		return ""
	}

	if d.Campaign != nil && d.Campaign.Name != "" {
		return d.Campaign.Name
	}

	return d.Project.Campaign.Name
}
