package domain

import (
	"time"

	"github.com/shopspring/decimal"

	shareddomain "github.com/alphacodinggroup/ponti-backend/internal/shared/domain"
)

type Status string

const (
	StatusDraft         Status = "draft"
	StatusPendingReview Status = "pending_review"
	StatusPublished     Status = "published"
	StatusRejected      Status = "rejected"
)

type WorkOrderDraft struct {
	ID                   int64
	Number               string
	Date                 time.Time
	CustomerID           int64
	CustomerName         string
	ProjectID            int64
	ProjectName          string
	CampaignID           *int64
	CampaignName         string
	FieldID              int64
	FieldName            string
	LotID                int64
	CropID               int64
	LaborID              int64
	Contractor           string
	EffectiveArea        decimal.Decimal
	Observations         string
	InvestorID           int64
	Status               Status
	ReviewedBy           *int64
	PublishedWorkOrderID *int64
	ReviewNotes          string
	Items                []WorkOrderDraftItem
	InvestorSplits       []WorkOrderDraftInvestorSplit
	Base                 shareddomain.Base
}

type WorkOrderDraftItem struct {
	SupplyID  int64
	TotalUsed decimal.Decimal
	FinalDose decimal.Decimal
}

type WorkOrderDraftInvestorSplit struct {
	InvestorID int64
	Percentage decimal.Decimal
}

type WorkOrderDraftListItem struct {
	ID          int64
	Number      string
	Date        time.Time
	ProjectID   int64
	ProjectName string
	FieldID     int64
	FieldName   string
	Status      Status
	Base        shareddomain.Base
}
