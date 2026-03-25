package workorderdraft

import (
	"context"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/alphacodinggroup/ponti-backend/internal/work-order-draft/usecases/domain"
	workorderdomain "github.com/alphacodinggroup/ponti-backend/internal/work-order/usecases/domain"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

type RepositoryPort interface {
	CreateWorkOrderDraft(context.Context, *domain.WorkOrderDraft) (int64, error)
	GetWorkOrderDraftByID(context.Context, int64) (*domain.WorkOrderDraft, error)
	ListWorkOrderDrafts(context.Context, string) ([]domain.WorkOrderDraftListItem, error)
	UpdateWorkOrderDraftByID(context.Context, *domain.WorkOrderDraft) error
	MarkWorkOrderDraftAsPublished(context.Context, int64, int64) error
}

type PublisherPort interface {
	CreateWorkOrder(context.Context, *workorderdomain.WorkOrder) (int64, error)
}

type UseCases struct {
	repo      RepositoryPort
	publisher PublisherPort
}

func NewUseCases(r RepositoryPort, p PublisherPort) *UseCases {
	return &UseCases{
		repo:      r,
		publisher: p,
	}
}

func (u *UseCases) CreateWorkOrderDraft(ctx context.Context, d *domain.WorkOrderDraft) (int64, error) {
	if d == nil {
		return 0, types.NewError(types.ErrValidation, "work order draft is nil", nil)
	}

	if d.Status == "" {
		d.Status = domain.StatusDraft
	}

	if err := validateDraft(d); err != nil {
		return 0, err
	}

	return u.repo.CreateWorkOrderDraft(ctx, d)
}

func (u *UseCases) GetWorkOrderDraftByID(ctx context.Context, id int64) (*domain.WorkOrderDraft, error) {
	if id <= 0 {
		return nil, types.NewInvalidIDError("invalid work order draft id", nil)
	}
	return u.repo.GetWorkOrderDraftByID(ctx, id)
}

func (u *UseCases) ListWorkOrderDrafts(ctx context.Context, number string) ([]domain.WorkOrderDraftListItem, error) {
	return u.repo.ListWorkOrderDrafts(ctx, number)
}

func (u *UseCases) UpdateWorkOrderDraftByID(ctx context.Context, d *domain.WorkOrderDraft) error {
	if d == nil {
		return types.NewError(types.ErrValidation, "work order draft is nil", nil)
	}
	if d.ID <= 0 {
		return types.NewInvalidIDError("invalid work order draft id", nil)
	}
	if err := validateDraft(d); err != nil {
		return err
	}

	current, err := u.repo.GetWorkOrderDraftByID(ctx, d.ID)
	if err != nil {
		return err
	}
	if current.Status == domain.StatusPublished {
		return types.NewError(types.ErrConflict, "published work order drafts cannot be updated", nil)
	}

	if d.Status == "" {
		d.Status = current.Status
	}

	return u.repo.UpdateWorkOrderDraftByID(ctx, d)
}

func (u *UseCases) PublishWorkOrderDraft(ctx context.Context, id int64) (int64, error) {
	if id <= 0 {
		return 0, types.NewInvalidIDError("invalid work order draft id", nil)
	}

	draft, err := u.repo.GetWorkOrderDraftByID(ctx, id)
	if err != nil {
		return 0, err
	}

	if draft.Status == domain.StatusPublished {
		return 0, types.NewError(types.ErrConflict, "work order draft is already published", nil)
	}

	if err := validateDraft(draft); err != nil {
		return 0, err
	}

	workOrder := &workorderdomain.WorkOrder{
		Number:         draft.Number,
		ProjectID:      draft.ProjectID,
		FieldID:        draft.FieldID,
		LotID:          draft.LotID,
		CropID:         draft.CropID,
		LaborID:        draft.LaborID,
		Contractor:     draft.Contractor,
		Observations:   draft.Observations,
		Date:           draft.Date,
		InvestorID:     draft.InvestorID,
		EffectiveArea:  draft.EffectiveArea,
		Items:          make([]workorderdomain.WorkOrderItem, len(draft.Items)),
		InvestorSplits: make([]workorderdomain.WorkOrderInvestorSplit, len(draft.InvestorSplits)),
	}

	for i, item := range draft.Items {
		workOrder.Items[i] = workorderdomain.WorkOrderItem{
			SupplyID:  item.SupplyID,
			TotalUsed: item.TotalUsed,
			FinalDose: item.FinalDose,
		}
	}

	for i, split := range draft.InvestorSplits {
		workOrder.InvestorSplits[i] = workorderdomain.WorkOrderInvestorSplit{
			InvestorID: split.InvestorID,
			Percentage: split.Percentage,
		}
	}

	workOrderID, err := u.publisher.CreateWorkOrder(ctx, workOrder)
	if err != nil {
		return 0, err
	}

	if err := u.repo.MarkWorkOrderDraftAsPublished(ctx, draft.ID, workOrderID); err != nil {
		return 0, err
	}

	return workOrderID, nil
}

func validateDraft(d *domain.WorkOrderDraft) error {
	if strings.TrimSpace(d.Number) == "" {
		return types.NewError(types.ErrValidation, "number is required", nil)
	}
	if d.Date.IsZero() {
		return types.NewError(types.ErrValidation, "date is required", nil)
	}
	if d.CustomerID <= 0 {
		return types.NewError(types.ErrValidation, "customer_id must be greater than 0", nil)
	}
	if d.ProjectID <= 0 {
		return types.NewError(types.ErrValidation, "project_id must be greater than 0", nil)
	}
	if d.FieldID <= 0 {
		return types.NewError(types.ErrValidation, "field_id must be greater than 0", nil)
	}
	if d.LotID <= 0 {
		return types.NewError(types.ErrValidation, "lot_id must be greater than 0", nil)
	}
	if d.CropID <= 0 {
		return types.NewError(types.ErrValidation, "crop_id must be greater than 0", nil)
	}
	if d.LaborID <= 0 {
		return types.NewError(types.ErrValidation, "labor_id must be greater than 0", nil)
	}
	if strings.TrimSpace(d.Contractor) == "" {
		return types.NewError(types.ErrValidation, "contractor is required", nil)
	}
	if d.EffectiveArea.LessThanOrEqual(decimal.Zero) {
		return types.NewError(types.ErrValidation, "effective_area must be greater than 0", nil)
	}
	if len(d.Items) == 0 {
		return types.NewError(types.ErrValidation, "at least one item is required", nil)
	}

	for _, item := range d.Items {
		if item.SupplyID <= 0 {
			return types.NewError(types.ErrValidation, "item supply_id must be greater than 0", nil)
		}
		if item.TotalUsed.LessThanOrEqual(decimal.Zero) {
			return types.NewError(types.ErrValidation, "item total_used must be greater than 0", nil)
		}
		if item.FinalDose.LessThanOrEqual(decimal.Zero) {
			return types.NewError(types.ErrValidation, "item final_dose must be greater than 0", nil)
		}
	}

	if len(d.InvestorSplits) == 0 {
		if d.InvestorID <= 0 {
			return types.NewError(types.ErrValidation, "investor_id must be greater than 0", nil)
		}
		return nil
	}

	seen := make(map[int64]struct{})
	sum := decimal.Zero

	for _, split := range d.InvestorSplits {
		if split.InvestorID <= 0 {
			return types.NewError(types.ErrValidation, "investor_splits investor_id must be greater than 0", nil)
		}
		if split.Percentage.LessThanOrEqual(decimal.Zero) {
			return types.NewError(types.ErrValidation, "investor_splits percentage must be greater than 0", nil)
		}
		if _, exists := seen[split.InvestorID]; exists {
			return types.NewError(types.ErrValidation, "duplicate investor_id in investor_splits", nil)
		}
		seen[split.InvestorID] = struct{}{}
		sum = sum.Add(split.Percentage)
	}

	if sum.Sub(decimal.NewFromInt(100)).Abs().GreaterThan(decimal.NewFromFloat(0.001)) {
		return types.NewError(types.ErrValidation, "investor_splits percentage must sum to 100", nil)
	}

	if d.InvestorID <= 0 {
		d.InvestorID = d.InvestorSplits[0].InvestorID
	}

	return nil
}
