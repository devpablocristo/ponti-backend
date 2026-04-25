package domain

import (
	"time"

	"github.com/shopspring/decimal"

	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
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
	LotName              string
	CropID               int64
	CropName             string
	LaborID              int64
	LaborName            string
	Contractor           string
	EffectiveArea        decimal.Decimal
	Observations         string
	InvestorID           int64
	IsDigital            bool
	Status               Status
	ReviewedBy           *int64
	PublishedWorkOrderID *int64
	ReviewNotes          string
	Items                []WorkOrderDraftItem
	InvestorSplits       []WorkOrderDraftInvestorSplit
	Base                 shareddomain.Base
}

type WorkOrderDraftItem struct {
	SupplyID   int64
	SupplyName string
	TotalUsed  decimal.Decimal
	FinalDose  decimal.Decimal
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
	IsDigital   bool
	Status      Status
	Base        shareddomain.Base
}

type WorkOrderDraftBatchLotItem struct {
	SupplyID  int64
	TotalUsed decimal.Decimal
}

type WorkOrderDraftBatchLot struct {
	LotID         int64
	EffectiveArea decimal.Decimal
	Items         []WorkOrderDraftBatchLotItem
}

type WorkOrderDraftBatchCreate struct {
	Number         string
	Date           time.Time
	CustomerID     int64
	ProjectID      int64
	CampaignID     *int64
	FieldID        int64
	CropID         int64
	LaborID        int64
	Contractor     string
	Observations   string
	InvestorID     int64
	InvestorSplits []WorkOrderDraftInvestorSplit
	Lots           []WorkOrderDraftBatchLot
}

type WorkOrderDraftBatchCreateResultItem struct {
	ID            int64
	Number        string
	LotID         int64
	EffectiveArea decimal.Decimal
}
