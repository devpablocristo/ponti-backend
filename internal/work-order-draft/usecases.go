package workorderdraft

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"

	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
	supplydomain "github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
	"github.com/devpablocristo/ponti-backend/internal/work-order-draft/usecases/domain"
	workorderdomain "github.com/devpablocristo/ponti-backend/internal/work-order/usecases/domain"
)

type RepositoryPort interface {
	CreateWorkOrderDraft(context.Context, *domain.WorkOrderDraft) (int64, error)
	CreateWorkOrderDraftBatch(context.Context, []*domain.WorkOrderDraft) ([]int64, error)
	GetWorkOrderDraftByID(context.Context, int64) (*domain.WorkOrderDraft, error)
	ListPendingSupplyNamesByIDs(context.Context, []int64) ([]string, error)
	ListRelatedDigitalWorkOrderDraftsByBaseNumber(context.Context, int64, string) ([]*domain.WorkOrderDraft, error)
	ListWorkOrderDrafts(context.Context, string, string, *bool, types.Input) ([]domain.WorkOrderDraftListItem, types.PageInfo, error)
	ListOccupiedWorkOrderNumbersByProject(context.Context, int64) ([]string, error)
	ListOccupiedWorkOrderNumbersByProjectExcludingDraft(context.Context, int64, int64) ([]string, error)
	ListPublishedWorkOrderNumbersByProject(context.Context, int64) ([]string, error)
	UpdateWorkOrderDraftByID(context.Context, *domain.WorkOrderDraft) error
	DeleteWorkOrderDraftByID(context.Context, int64) error
	MarkWorkOrderDraftAsPublished(context.Context, int64, int64) error
}

type PublisherPort interface {
	CreateWorkOrder(context.Context, *workorderdomain.WorkOrder) (int64, error)
}

type SupplyReaderPort interface {
	GetSupply(context.Context, int64) (*supplydomain.Supply, error)
}

type PDFExporterPort interface {
	ExportDraft(context.Context, *domain.WorkOrderDraft) ([]byte, error)
	ExportDraftGroup(context.Context, []*domain.WorkOrderDraft) ([]byte, error)
}

type UseCases struct {
	repo         RepositoryPort
	publisher    PublisherPort
	pdfExporter  PDFExporterPort
	supplyReader SupplyReaderPort
}

var (
	plainNumberRE        = regexp.MustCompile(`^\d+$`)
	digitalBaseNumberRE  = regexp.MustCompile(`^D-(\d+)$`)
	digitalSplitNumberRE = regexp.MustCompile(`^D-(\d+)\.(\d+)$`)
)

func NewUseCases(r RepositoryPort, p PublisherPort, pdf PDFExporterPort, sr SupplyReaderPort) *UseCases {
	return &UseCases{
		repo:         r,
		publisher:    p,
		pdfExporter:  pdf,
		supplyReader: sr,
	}
}

func (u *UseCases) CreateWorkOrderDraft(ctx context.Context, d *domain.WorkOrderDraft) (int64, error) {
	if d == nil {
		return 0, types.NewError(types.ErrValidation, "work order draft is nil", nil)
	}

	// Si el cliente no manda estado, una orden nueva queda abierta.
	if d.Status == "" {
		d.Status = domain.StatusDraft
	}

	if err := u.hydrateDraftSupplyNames(ctx, d); err != nil {
		return 0, err
	}

	if err := validateDraft(d); err != nil {
		return 0, err
	}

	return u.repo.CreateWorkOrderDraft(ctx, d)
}

func (u *UseCases) CreateDigitalWorkOrderDraft(ctx context.Context, d *domain.WorkOrderDraft) (int64, error) {
	if d == nil {
		return 0, types.NewError(types.ErrValidation, "work order draft is nil", nil)
	}

	d.IsDigital = true

	if d.Status == "" {
		d.Status = domain.StatusDraft
	}

	number, err := u.resolveDigitalDraftNumber(ctx, d.ProjectID, strings.TrimSpace(d.Number))
	if err != nil {
		return 0, err
	}
	d.Number = number

	if err := u.hydrateDraftSupplyNames(ctx, d); err != nil {
		return 0, err
	}

	if err := validateDraft(d); err != nil {
		return 0, err
	}

	return u.repo.CreateWorkOrderDraft(ctx, d)
}

func (u *UseCases) CreateDigitalWorkOrderDraftBatch(ctx context.Context, b *domain.WorkOrderDraftBatchCreate) ([]domain.WorkOrderDraftBatchCreateResultItem, error) {
	if b == nil {
		return nil, types.NewError(types.ErrValidation, "work order draft batch is nil", nil)
	}
	if len(b.Lots) == 0 {
		return nil, types.NewError(types.ErrValidation, "at least one lot is required", nil)
	}

	baseNumber, err := u.resolveDigitalDraftBatchBaseNumber(ctx, b.ProjectID, strings.TrimSpace(b.Number))
	if err != nil {
		return nil, err
	}

	seenLots := make(map[int64]struct{})
	drafts := make([]*domain.WorkOrderDraft, len(b.Lots))

	for i, lot := range b.Lots {
		if lot.LotID <= 0 {
			return nil, types.NewError(types.ErrValidation, "lot_id must be greater than 0", nil)
		}
		if lot.EffectiveArea.LessThanOrEqual(decimal.Zero) {
			return nil, types.NewError(types.ErrValidation, "effective_area must be greater than 0", nil)
		}
		if len(lot.Items) == 0 {
			return nil, types.NewError(types.ErrValidation, "at least one item is required for each lot", nil)
		}
		if _, exists := seenLots[lot.LotID]; exists {
			return nil, types.NewError(types.ErrValidation, "duplicate lot_id in lots", nil)
		}
		seenLots[lot.LotID] = struct{}{}

		number := baseNumber
		if len(b.Lots) > 1 {
			number = fmt.Sprintf("%s.%d", baseNumber, i+1)
		}

		items := make([]domain.WorkOrderDraftItem, len(lot.Items))
		for j, item := range lot.Items {
			if item.SupplyID <= 0 {
				return nil, types.NewError(types.ErrValidation, "item supply_id must be greater than 0", nil)
			}
			if item.TotalUsed.LessThanOrEqual(decimal.Zero) {
				return nil, types.NewError(types.ErrValidation, "item total_used must be greater than 0", nil)
			}

			finalDose := item.TotalUsed.Div(lot.EffectiveArea).Round(6)

			items[j] = domain.WorkOrderDraftItem{
				SupplyID:  item.SupplyID,
				TotalUsed: item.TotalUsed,
				FinalDose: finalDose,
			}
		}

		draft := &domain.WorkOrderDraft{
			Number:         number,
			Date:           b.Date,
			CustomerID:     b.CustomerID,
			ProjectID:      b.ProjectID,
			CampaignID:     b.CampaignID,
			FieldID:        b.FieldID,
			LotID:          lot.LotID,
			CropID:         b.CropID,
			LaborID:        b.LaborID,
			Contractor:     b.Contractor,
			EffectiveArea:  lot.EffectiveArea,
			Observations:   b.Observations,
			InvestorID:     b.InvestorID,
			IsDigital:      true,
			Status:         domain.StatusDraft,
			Items:          items,
			InvestorSplits: b.InvestorSplits,
		}

		if err := u.hydrateDraftSupplyNames(ctx, draft); err != nil {
			return nil, err
		}

		if err := validateDraft(draft); err != nil {
			return nil, err
		}

		drafts[i] = draft
	}

	ids, err := u.repo.CreateWorkOrderDraftBatch(ctx, drafts)
	if err != nil {
		return nil, err
	}

	result := make([]domain.WorkOrderDraftBatchCreateResultItem, len(drafts))
	for i := range drafts {
		result[i] = domain.WorkOrderDraftBatchCreateResultItem{
			ID:            ids[i],
			Number:        drafts[i].Number,
			LotID:         drafts[i].LotID,
			EffectiveArea: drafts[i].EffectiveArea,
		}
	}

	return result, nil
}

func (u *UseCases) PreviewDigitalWorkOrderNumber(ctx context.Context, projectID int64, requested string) (string, error) {
	if projectID <= 0 {
		return "", types.NewError(types.ErrValidation, "project_id must be greater than 0", nil)
	}

	return u.resolveDigitalDraftNumber(ctx, projectID, requested)
}

func (u *UseCases) PreviewDigitalWorkOrderDraftBatchNumber(ctx context.Context, projectID int64, requested string) (string, error) {
	if projectID <= 0 {
		return "", types.NewError(types.ErrValidation, "project_id must be greater than 0", nil)
	}

	return u.resolveDigitalDraftBatchBaseNumber(ctx, projectID, requested)
}

func (u *UseCases) GetWorkOrderDraftByID(ctx context.Context, id int64) (*domain.WorkOrderDraft, error) {
	if id <= 0 {
		return nil, types.NewInvalidIDError("invalid work order draft id", nil)
	}
	return u.repo.GetWorkOrderDraftByID(ctx, id)
}

func (u *UseCases) ExportWorkOrderDraftPDF(ctx context.Context, id int64) ([]byte, error) {
	if id <= 0 {
		return nil, types.NewInvalidIDError("invalid work order draft id", nil)
	}

	draft, err := u.repo.GetWorkOrderDraftByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if u.pdfExporter == nil {
		return nil, types.NewError(types.ErrInternal, "pdf exporter not configured", nil)
	}

	return u.pdfExporter.ExportDraft(ctx, draft)
}

func (u *UseCases) ExportWorkOrderDraftGroupPDF(ctx context.Context, id int64) ([]byte, error) {
	if id <= 0 {
		return nil, types.NewInvalidIDError("invalid work order draft id", nil)
	}

	draft, err := u.repo.GetWorkOrderDraftByID(ctx, id)
	if err != nil {
		return nil, err
	}

	baseSequence, ok := extractBaseSequence(draft.Number)
	if !ok {
		return nil, types.NewError(types.ErrValidation, "invalid work order draft number", nil)
	}

	baseNumber := fmt.Sprintf("D-%d", baseSequence)

	related, err := u.repo.ListRelatedDigitalWorkOrderDraftsByBaseNumber(ctx, draft.ProjectID, baseNumber)
	if err != nil {
		return nil, err
	}
	if len(related) == 0 {
		return nil, types.NewError(types.ErrNotFound, "related work order drafts not found", nil)
	}

	sortDigitalDraftGroup(related)

	if u.pdfExporter == nil {
		return nil, types.NewError(types.ErrInternal, "pdf exporter not configured", nil)
	}

	return u.pdfExporter.ExportDraftGroup(ctx, related)
}

func (u *UseCases) ListWorkOrderDrafts(ctx context.Context, number string, status string, inp types.Input) ([]domain.WorkOrderDraftListItem, types.PageInfo, error) {
	return u.repo.ListWorkOrderDrafts(ctx, number, status, nil, inp)
}

func (u *UseCases) ListDigitalWorkOrderDrafts(ctx context.Context, number string, status string, inp types.Input) ([]domain.WorkOrderDraftListItem, types.PageInfo, error) {
	isDigital := true
	return u.repo.ListWorkOrderDrafts(ctx, number, status, &isDigital, inp)
}

func (u *UseCases) UpdateWorkOrderDraftByID(ctx context.Context, d *domain.WorkOrderDraft) error {
	if d == nil {
		return types.NewError(types.ErrValidation, "work order draft is nil", nil)
	}
	if d.ID <= 0 {
		return types.NewInvalidIDError("invalid work order draft id", nil)
	}

	current, err := u.repo.GetWorkOrderDraftByID(ctx, d.ID)
	if err != nil {
		return err
	}
	if current.Status == domain.StatusPublished {
		return types.NewError(types.ErrConflict, "published work order drafts cannot be updated", nil)
	}

	// Si el cliente no manda status, conservamos el actual.
	if d.Status == "" {
		d.Status = current.Status
	}

	// Si el cliente no marca explícitamente el origen, conservamos el actual.
	if !d.IsDigital && current.IsDigital {
		d.IsDigital = current.IsDigital
	}

	// Si el cliente borra el número en edición, para digital recalculamos;
	// para no digital preservamos el actual.
	if strings.TrimSpace(d.Number) == "" && !current.IsDigital && !d.IsDigital {
		d.Number = current.Number
	}

	if current.IsDigital || d.IsDigital {
		d.IsDigital = true

		number, err := u.resolveDigitalDraftNumberForUpdate(ctx, d.ProjectID, d.ID, strings.TrimSpace(d.Number))
		if err != nil {
			return err
		}
		d.Number = number
	}

	if err := u.hydrateDraftSupplyNames(ctx, d); err != nil {
		return err
	}

	if err := validateDraft(d); err != nil {
		return err
	}

	return u.repo.UpdateWorkOrderDraftByID(ctx, d)
}

func (u *UseCases) DeleteWorkOrderDraftByID(ctx context.Context, id int64) error {
	if id <= 0 {
		return types.NewInvalidIDError("invalid work order draft id", nil)
	}

	current, err := u.repo.GetWorkOrderDraftByID(ctx, id)
	if err != nil {
		return err
	}

	if current.Status == domain.StatusPublished {
		return types.NewError(types.ErrConflict, "published work order drafts cannot be deleted", nil)
	}

	return u.repo.DeleteWorkOrderDraftByID(ctx, id)
}

func (u *UseCases) PublishWorkOrderDraft(ctx context.Context, id int64) (int64, error) {
	if id <= 0 {
		return 0, types.NewInvalidIDError("invalid work order draft id", nil)
	}

	draft, err := u.repo.GetWorkOrderDraftByID(ctx, id)
	if err != nil {
		return 0, err
	}
	// Publicar equivale a cerrar la orden digital y crear la workorder real.
	if draft.Status == domain.StatusPublished {
		return 0, types.NewError(types.ErrConflict, "work order draft is already published", nil)
	}

	if err := validateDraft(draft); err != nil {
		return 0, err
	}

	if draft.IsDigital {
		if err := u.validateDigitalNumberForPublish(ctx, draft.ProjectID, draft.Number); err != nil {
			return 0, err
		}
	}

	if err := u.validateDraftSuppliesReadyForPublish(ctx, draft); err != nil {
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
			SupplyID:   item.SupplyID,
			SupplyName: item.SupplyName,
			TotalUsed:  item.TotalUsed,
			FinalDose:  item.FinalDose,
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

func (u *UseCases) validateDraftSuppliesReadyForPublish(ctx context.Context, draft *domain.WorkOrderDraft) error {
	supplyIDs := make([]int64, 0, len(draft.Items))
	seen := make(map[int64]struct{}, len(draft.Items))
	for _, item := range draft.Items {
		if item.SupplyID <= 0 {
			continue
		}
		if _, exists := seen[item.SupplyID]; exists {
			continue
		}
		seen[item.SupplyID] = struct{}{}
		supplyIDs = append(supplyIDs, item.SupplyID)
	}

	if len(supplyIDs) == 0 {
		return nil
	}

	pendingNames, err := u.repo.ListPendingSupplyNamesByIDs(ctx, supplyIDs)
	if err != nil {
		return err
	}
	if len(pendingNames) == 0 {
		return nil
	}

	return types.NewError(
		types.ErrConflict,
		fmt.Sprintf("cannot publish work order draft with pending supplies: %s", strings.Join(pendingNames, ", ")),
		nil,
	)
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

	seenSupplyIDs := make(map[int64]struct{})

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
		if _, exists := seenSupplyIDs[item.SupplyID]; exists {
			return types.NewError(types.ErrValidation, "duplicate supply_id in items", nil)
		}
		seenSupplyIDs[item.SupplyID] = struct{}{}
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

func (u *UseCases) resolveDigitalDraftNumber(ctx context.Context, projectID int64, requested string) (string, error) {
	occupied, err := u.repo.ListOccupiedWorkOrderNumbersByProject(ctx, projectID)
	if err != nil {
		return "", err
	}

	return resolveDigitalDraftNumberWithOccupied(projectID, requested, occupied)
}

func (u *UseCases) resolveDigitalDraftNumberForUpdate(ctx context.Context, projectID int64, draftID int64, requested string) (string, error) {
	occupied, err := u.repo.ListOccupiedWorkOrderNumbersByProjectExcludingDraft(ctx, projectID, draftID)
	if err != nil {
		return "", err
	}

	return resolveDigitalDraftNumberWithOccupied(projectID, requested, occupied)
}

func (u *UseCases) resolveDigitalDraftBatchBaseNumber(ctx context.Context, projectID int64, requested string) (string, error) {
	occupied, err := u.repo.ListOccupiedWorkOrderNumbersByProject(ctx, projectID)
	if err != nil {
		return "", err
	}

	requested = strings.TrimSpace(requested)
	if requested == "" {
		return nextAvailableDigitalBaseNumber(occupied), nil
	}

	if !digitalBaseNumberRE.MatchString(requested) {
		return "", types.NewError(types.ErrValidation, "batch digital work order number must have format D-<number>", nil)
	}

	base, ok := extractBaseSequence(requested)
	if !ok {
		return "", types.NewError(types.ErrValidation, "batch digital work order number must have format D-<number>", nil)
	}

	if exactNumberExists(occupied, requested) || baseSequenceUsedByDifferentNumber(base, requested, occupied) {
		return "", newWorkOrderNumberConflictError(requested, projectID)
	}

	return requested, nil
}

func (u *UseCases) validateDigitalNumberForPublish(ctx context.Context, projectID int64, number string) error {
	occupied, err := u.repo.ListPublishedWorkOrderNumbersByProject(ctx, projectID)
	if err != nil {
		return err
	}

	number = strings.TrimSpace(number)
	if number == "" {
		return types.NewError(types.ErrValidation, "number is required", nil)
	}

	if exactNumberExists(occupied, number) {
		return newWorkOrderNumberConflictError(number, projectID)
	}

	base, ok := extractBaseSequence(number)
	if !ok {
		return types.NewError(types.ErrValidation, "digital work order number must have format D-<number> or D-<number>.<suffix>", nil)
	}

	if baseSequenceUsedByDifferentNumber(base, number, occupied) {
		return newWorkOrderNumberConflictError(number, projectID)
	}

	return nil
}

func (u *UseCases) hydrateDraftSupplyNames(ctx context.Context, draft *domain.WorkOrderDraft) error {
	if draft == nil {
		return types.NewError(types.ErrValidation, "work order draft is nil", nil)
	}
	if u.supplyReader == nil {
		return types.NewError(types.ErrInternal, "supply reader not configured", nil)
	}

	for i := range draft.Items {
		if draft.Items[i].SupplyID <= 0 {
			return types.NewError(types.ErrValidation, "item supply_id must be greater than 0", nil)
		}

		supply, err := u.supplyReader.GetSupply(ctx, draft.Items[i].SupplyID)
		if err != nil {
			return err
		}

		draft.Items[i].SupplyName = supply.Name
	}

	return nil
}

func resolveDigitalDraftNumberWithOccupied(projectID int64, requested string, occupied []string) (string, error) {
	requested = strings.TrimSpace(requested)
	if requested == "" {
		return nextAvailableDigitalBaseNumber(occupied), nil
	}

	if exactNumberExists(occupied, requested) {
		if digitalBaseNumberRE.MatchString(requested) {
			return nextAvailableDigitalSplitNumber(requested, occupied), nil
		}
		return "", newWorkOrderNumberConflictError(requested, projectID)
	}

	base, ok := extractBaseSequence(requested)
	if !ok {
		return "", types.NewError(types.ErrValidation, "digital work order number must have format D-<number> or D-<number>.<suffix>", nil)
	}

	if baseSequenceUsedByDifferentNumber(base, requested, occupied) {
		return "", newWorkOrderNumberConflictError(requested, projectID)
	}

	return requested, nil
}

func nextAvailableDigitalBaseNumber(occupied []string) string {
	maxBase := 0

	for _, number := range occupied {
		base, ok := extractBaseSequence(number)
		if !ok {
			continue
		}
		if base > maxBase {
			maxBase = base
		}
	}

	return fmt.Sprintf("D-%d", maxBase+1)
}

func nextAvailableDigitalSplitNumber(baseNumber string, occupied []string) string {
	base, ok := extractBaseSequence(baseNumber)
	if !ok {
		return baseNumber
	}

	maxSuffix := 0
	for _, number := range occupied {
		if strings.TrimSpace(number) == baseNumber {
			continue
		}
		currentBase, suffix, ok := extractDigitalSplitSequence(number)
		if !ok {
			continue
		}
		if currentBase == base && suffix > maxSuffix {
			maxSuffix = suffix
		}
	}

	return fmt.Sprintf("%s.%d", baseNumber, maxSuffix+1)
}

func exactNumberExists(numbers []string, target string) bool {
	target = strings.TrimSpace(target)
	for _, number := range numbers {
		if strings.TrimSpace(number) == target {
			return true
		}
	}
	return false
}

func baseSequenceUsedByDifferentNumber(base int, requested string, numbers []string) bool {
	requested = strings.TrimSpace(requested)
	for _, number := range numbers {
		number = strings.TrimSpace(number)
		if number == requested {
			continue
		}
		currentBase, ok := extractBaseSequence(number)
		if !ok {
			continue
		}
		if currentBase == base {
			return true
		}
	}
	return false
}

func extractBaseSequence(number string) (int, bool) {
	number = strings.TrimSpace(number)

	if plainNumberRE.MatchString(number) {
		value, err := strconv.Atoi(number)
		if err != nil {
			return 0, false
		}
		return value, true
	}

	if matches := digitalBaseNumberRE.FindStringSubmatch(number); len(matches) == 2 {
		value, err := strconv.Atoi(matches[1])
		if err != nil {
			return 0, false
		}
		return value, true
	}

	if matches := digitalSplitNumberRE.FindStringSubmatch(number); len(matches) == 3 {
		value, err := strconv.Atoi(matches[1])
		if err != nil {
			return 0, false
		}
		return value, true
	}

	return 0, false
}

func extractDigitalSplitSequence(number string) (int, int, bool) {
	number = strings.TrimSpace(number)

	matches := digitalSplitNumberRE.FindStringSubmatch(number)
	if len(matches) != 3 {
		return 0, 0, false
	}

	base, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, 0, false
	}

	suffix, err := strconv.Atoi(matches[2])
	if err != nil {
		return 0, 0, false
	}

	return base, suffix, true
}

func sortDigitalDraftGroup(drafts []*domain.WorkOrderDraft) {
	sort.Slice(drafts, func(i, j int) bool {
		left := strings.TrimSpace(drafts[i].Number)
		right := strings.TrimSpace(drafts[j].Number)

		leftBase, leftSuffix, leftIsSplit := extractDigitalSplitSequence(left)
		rightBase, rightSuffix, rightIsSplit := extractDigitalSplitSequence(right)

		if leftIsSplit && rightIsSplit {
			if leftBase != rightBase {
				return leftBase < rightBase
			}
			return leftSuffix < rightSuffix
		}

		leftBaseOnly := digitalBaseNumberRE.MatchString(left)
		rightBaseOnly := digitalBaseNumberRE.MatchString(right)

		if leftBaseOnly && rightIsSplit {
			return true
		}
		if leftIsSplit && rightBaseOnly {
			return false
		}

		return left < right
	})
}

func newWorkOrderNumberConflictError(number string, projectID int64) error {
	return types.NewError(
		types.ErrConflict,
		fmt.Sprintf("work order already exists for number %s and project %d", number, projectID),
		nil,
	)
}
