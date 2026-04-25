package dto

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"

	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
	"github.com/devpablocristo/ponti-backend/internal/work-order-draft/usecases/domain"
)

const dateLayout = "2006-01-02"

type WorkOrderDraftItem struct {
	SupplyID  int64           `json:"supply_id" binding:"required"`
	TotalUsed decimal.Decimal `json:"total_used" binding:"required"`
	FinalDose decimal.Decimal `json:"final_dose" binding:"required"`
}

type InvestorSplit struct {
	InvestorID int64           `json:"investor_id" binding:"required"`
	Percentage decimal.Decimal `json:"percentage" binding:"required"`
}

type WorkOrderDraft struct {
	Number         string               `json:"number"`
	Date           string               `json:"date" binding:"required"`
	CustomerID     int64                `json:"customer_id" binding:"required"`
	ProjectID      int64                `json:"project_id" binding:"required"`
	CampaignID     *int64               `json:"campaign_id"`
	FieldID        int64                `json:"field_id" binding:"required"`
	LotID          int64                `json:"lot_id" binding:"required"`
	CropID         int64                `json:"crop_id" binding:"required"`
	LaborID        int64                `json:"labor_id" binding:"required"`
	Contractor     string               `json:"contractor" binding:"required"`
	EffectiveArea  decimal.Decimal      `json:"effective_area" binding:"required"`
	Observations   string               `json:"observations"`
	InvestorID     int64                `json:"investor_id"`
	IsDigital      bool                 `json:"is_digital"`
	InvestorSplits []InvestorSplit      `json:"investor_splits,omitempty"`
	Items          []WorkOrderDraftItem `json:"items" binding:"required"`
}

type WorkOrderDraftBatchLotItem struct {
	SupplyID  int64           `json:"supply_id" binding:"required"`
	TotalUsed decimal.Decimal `json:"total_used" binding:"required"`
}

type WorkOrderDraftBatchLot struct {
	LotID         int64                        `json:"lot_id" binding:"required"`
	EffectiveArea decimal.Decimal              `json:"effective_area" binding:"required"`
	Items         []WorkOrderDraftBatchLotItem `json:"items" binding:"required"`
}

type WorkOrderDraftBatchCreateRequest struct {
	Number         string                   `json:"number"`
	Date           string                   `json:"date" binding:"required"`
	CustomerID     int64                    `json:"customer_id" binding:"required"`
	ProjectID      int64                    `json:"project_id" binding:"required"`
	CampaignID     *int64                   `json:"campaign_id"`
	FieldID        int64                    `json:"field_id" binding:"required"`
	CropID         int64                    `json:"crop_id" binding:"required"`
	LaborID        int64                    `json:"labor_id" binding:"required"`
	Contractor     string                   `json:"contractor" binding:"required"`
	Observations   string                   `json:"observations"`
	InvestorID     int64                    `json:"investor_id"`
	InvestorSplits []InvestorSplit          `json:"investor_splits,omitempty"`
	Lots           []WorkOrderDraftBatchLot `json:"lots" binding:"required"`
}

type WorkOrderDraftNumberPreviewRequest struct {
	ProjectID int64  `json:"project_id" binding:"required"`
	Number    string `json:"number"`
}

type WorkOrderDraftNumberPreviewResponse struct {
	Number string `json:"number"`
}

type WorkOrderDraftBatchCreateResponseItem struct {
	ID            int64           `json:"id"`
	Number        string          `json:"number"`
	LotID         int64           `json:"lot_id"`
	EffectiveArea decimal.Decimal `json:"effective_area"`
}

type WorkOrderDraftBatchCreateResponse struct {
	Items []WorkOrderDraftBatchCreateResponseItem `json:"items"`
}

func (r *WorkOrderDraft) ToDomain() (*domain.WorkOrderDraft, error) {
	dateValue, err := time.Parse(dateLayout, strings.TrimSpace(r.Date))
	if err != nil {
		return nil, types.NewError(types.ErrBadRequest, "date must have format YYYY-MM-DD", err)
	}

	items := make([]domain.WorkOrderDraftItem, len(r.Items))
	for i, item := range r.Items {
		items[i] = domain.WorkOrderDraftItem{
			SupplyID:  item.SupplyID,
			TotalUsed: item.TotalUsed,
			FinalDose: item.FinalDose,
		}
	}

	splits := make([]domain.WorkOrderDraftInvestorSplit, len(r.InvestorSplits))
	for i, split := range r.InvestorSplits {
		splits[i] = domain.WorkOrderDraftInvestorSplit{
			InvestorID: split.InvestorID,
			Percentage: split.Percentage,
		}
	}

	return &domain.WorkOrderDraft{
		Number:         r.Number,
		Date:           dateValue,
		CustomerID:     r.CustomerID,
		ProjectID:      r.ProjectID,
		CampaignID:     r.CampaignID,
		FieldID:        r.FieldID,
		LotID:          r.LotID,
		CropID:         r.CropID,
		LaborID:        r.LaborID,
		Contractor:     r.Contractor,
		EffectiveArea:  r.EffectiveArea,
		Observations:   r.Observations,
		InvestorID:     r.InvestorID,
		IsDigital:      r.IsDigital,
		Items:          items,
		InvestorSplits: splits,
	}, nil
}

func (r *WorkOrderDraftBatchCreateRequest) ToDomain() (*domain.WorkOrderDraftBatchCreate, error) {
	dateValue, err := time.Parse(dateLayout, strings.TrimSpace(r.Date))
	if err != nil {
		return nil, types.NewError(types.ErrBadRequest, "date must have format YYYY-MM-DD", err)
	}

	splits := make([]domain.WorkOrderDraftInvestorSplit, len(r.InvestorSplits))
	for i, split := range r.InvestorSplits {
		splits[i] = domain.WorkOrderDraftInvestorSplit{
			InvestorID: split.InvestorID,
			Percentage: split.Percentage,
		}
	}

	lots := make([]domain.WorkOrderDraftBatchLot, len(r.Lots))
	for i, lot := range r.Lots {
		items := make([]domain.WorkOrderDraftBatchLotItem, len(lot.Items))
		for j, item := range lot.Items {
			items[j] = domain.WorkOrderDraftBatchLotItem{
				SupplyID:  item.SupplyID,
				TotalUsed: item.TotalUsed,
			}
		}

		lots[i] = domain.WorkOrderDraftBatchLot{
			LotID:         lot.LotID,
			EffectiveArea: lot.EffectiveArea,
			Items:         items,
		}
	}

	return &domain.WorkOrderDraftBatchCreate{
		Number:         r.Number,
		Date:           dateValue,
		CustomerID:     r.CustomerID,
		ProjectID:      r.ProjectID,
		CampaignID:     r.CampaignID,
		FieldID:        r.FieldID,
		CropID:         r.CropID,
		LaborID:        r.LaborID,
		Contractor:     r.Contractor,
		Observations:   r.Observations,
		InvestorID:     r.InvestorID,
		InvestorSplits: splits,
		Lots:           lots,
	}, nil
}

type WorkOrderDraftResponseItem struct {
	SupplyID   int64           `json:"supply_id"`
	SupplyName string          `json:"supply_name"`
	TotalUsed  decimal.Decimal `json:"total_used"`
	FinalDose  decimal.Decimal `json:"final_dose"`
}

type WorkOrderDraftResponseInvestorSplit struct {
	InvestorID int64           `json:"investor_id"`
	Percentage decimal.Decimal `json:"percentage"`
}

type WorkOrderDraftResponse struct {
	ID                   int64                                 `json:"id"`
	Number               string                                `json:"number"`
	Date                 string                                `json:"date"`
	CustomerID           int64                                 `json:"customer_id"`
	CustomerName         string                                `json:"customer_name"`
	ProjectID            int64                                 `json:"project_id"`
	ProjectName          string                                `json:"project_name"`
	CampaignID           *int64                                `json:"campaign_id,omitempty"`
	CampaignName         string                                `json:"campaign_name"`
	FieldID              int64                                 `json:"field_id"`
	FieldName            string                                `json:"field_name"`
	LotID                int64                                 `json:"lot_id"`
	CropID               int64                                 `json:"crop_id"`
	LaborID              int64                                 `json:"labor_id"`
	Contractor           string                                `json:"contractor"`
	EffectiveArea        decimal.Decimal                       `json:"effective_area"`
	Observations         string                                `json:"observations"`
	InvestorID           int64                                 `json:"investor_id"`
	IsDigital            bool                                  `json:"is_digital"`
	Status               string                                `json:"status"`
	ReviewedBy           *int64                                `json:"reviewed_by,omitempty"`
	PublishedWorkOrderID *int64                                `json:"published_work_order_id,omitempty"`
	ReviewNotes          string                                `json:"review_notes"`
	Items                []WorkOrderDraftResponseItem          `json:"items"`
	InvestorSplits       []WorkOrderDraftResponseInvestorSplit `json:"investor_splits,omitempty"`
	CreatedAt            string                                `json:"created_at"`
	UpdatedAt            string                                `json:"updated_at"`
}

type WorkOrderDraftListItem struct {
	ID          int64  `json:"id"`
	Number      string `json:"number"`
	Date        string `json:"date"`
	ProjectID   int64  `json:"project_id"`
	ProjectName string `json:"project_name"`
	FieldID     int64  `json:"field_id"`
	FieldName   string `json:"field_name"`
	IsDigital   bool   `json:"is_digital"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

type WorkOrderDraftListResponse struct {
	PageInfo types.PageInfo           `json:"page_info"`
	Data     []WorkOrderDraftListItem `json:"data"`
}

func FromDomain(d *domain.WorkOrderDraft) *WorkOrderDraftResponse {
	items := make([]WorkOrderDraftResponseItem, len(d.Items))
	for i, item := range d.Items {
		items[i] = WorkOrderDraftResponseItem{
			SupplyID:   item.SupplyID,
			SupplyName: item.SupplyName,
			TotalUsed:  item.TotalUsed,
			FinalDose:  item.FinalDose,
		}
	}

	splits := make([]WorkOrderDraftResponseInvestorSplit, len(d.InvestorSplits))
	for i, split := range d.InvestorSplits {
		splits[i] = WorkOrderDraftResponseInvestorSplit{
			InvestorID: split.InvestorID,
			Percentage: split.Percentage,
		}
	}

	return &WorkOrderDraftResponse{
		ID:                   d.ID,
		Number:               d.Number,
		Date:                 d.Date.Format(dateLayout),
		CustomerID:           d.CustomerID,
		CustomerName:         d.CustomerName,
		ProjectID:            d.ProjectID,
		ProjectName:          d.ProjectName,
		CampaignID:           d.CampaignID,
		CampaignName:         d.CampaignName,
		FieldID:              d.FieldID,
		FieldName:            d.FieldName,
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
		Items:                items,
		InvestorSplits:       splits,
		CreatedAt:            d.Base.CreatedAt.Format(time.RFC3339),
		UpdatedAt:            d.Base.UpdatedAt.Format(time.RFC3339),
	}
}

func NewListResponse(pageInfo types.PageInfo, items []domain.WorkOrderDraftListItem) *WorkOrderDraftListResponse {
	resp := make([]WorkOrderDraftListItem, len(items))
	for i, item := range items {
		resp[i] = WorkOrderDraftListItem{
			ID:          item.ID,
			Number:      item.Number,
			Date:        item.Date.Format(dateLayout),
			ProjectID:   item.ProjectID,
			ProjectName: item.ProjectName,
			FieldID:     item.FieldID,
			FieldName:   item.FieldName,
			IsDigital:   item.IsDigital,
			Status:      string(item.Status),
			CreatedAt:   item.Base.CreatedAt.Format(time.RFC3339),
		}
	}

	return &WorkOrderDraftListResponse{
		PageInfo: pageInfo,
		Data:     resp,
	}
}

func NewBatchCreateResponse(items []domain.WorkOrderDraftBatchCreateResultItem) *WorkOrderDraftBatchCreateResponse {
	resp := make([]WorkOrderDraftBatchCreateResponseItem, len(items))
	for i, item := range items {
		resp[i] = WorkOrderDraftBatchCreateResponseItem{
			ID:            item.ID,
			Number:        item.Number,
			LotID:         item.LotID,
			EffectiveArea: item.EffectiveArea,
		}
	}

	return &WorkOrderDraftBatchCreateResponse{Items: resp}
}
