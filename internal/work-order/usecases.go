// Package workorder contiene casos de uso para work orders.
package workorder

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/devpablocristo/core/errors/go/domainerr"
	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
	"github.com/devpablocristo/ponti-backend/internal/work-order/usecases/domain"
)

type RepositoryPort interface {
	CreateWorkOrder(context.Context, *domain.WorkOrder) (int64, error)
	GetWorkOrderByID(ctx context.Context, id int64) (*domain.WorkOrder, error)
	GetWorkOrderByNumberAndProjectID(ctx context.Context, number string, projectID int64) (*domain.WorkOrder, error)
	UpdateWorkOrderByID(context.Context, *domain.WorkOrder) error
	UpdateInvestorPaymentStatus(context.Context, int64, int64, string) error
	DeleteWorkOrderByID(context.Context, int64) error
	ArchiveWorkOrder(context.Context, int64) error
	RestoreWorkOrder(context.Context, int64) error
	ListWorkOrders(context.Context, domain.WorkOrderFilter, types.Input) ([]domain.WorkOrderListElement, types.PageInfo, error)
	GetMetrics(context.Context, domain.WorkOrderFilter) (*domain.WorkOrderMetrics, error)
	GetRawDirectCost(context.Context, int64) (decimal.Decimal, error)
}

type ExporterAdapterPort interface {
	Export(ctx context.Context, items []domain.WorkOrderListElement) ([]byte, error)
	Close() error
}

type UseCases struct {
	repo  RepositoryPort
	excel ExporterAdapterPort
}

var workOrderNumberPattern = regexp.MustCompile(`^\d+$`)

// NewUseCases crea una instancia de casos de uso para work orders.
func NewUseCases(r RepositoryPort, excel ExporterAdapterPort) *UseCases {
	return &UseCases{repo: r, excel: excel}
}

func (u *UseCases) CreateWorkOrder(ctx context.Context, o *domain.WorkOrder) (int64, error) {
	if o == nil {
		return 0, domainerr.Validation("work order is nil")
	}
	if err := validateDate(o); err != nil {
		return 0, err
	}
	if err := validateItems(o); err != nil {
		return 0, err
	}
	if err := validateWorkOrderNumberForCreate(o); err != nil {
		return 0, err
	}
	if err := ensureWorkOrderNumberIsUnique(ctx, u.repo, o.Number, o.ProjectID, 0); err != nil {
		return 0, err
	}
	if err := validateInvestorSplits(o); err != nil {
		return 0, err
	}
	if err := validateUniqueSupplyItems(o); err != nil {
		return 0, err
	}
	return u.repo.CreateWorkOrder(ctx, o)
}

func (u *UseCases) GetWorkOrderByID(ctx context.Context, id int64) (*domain.WorkOrder, error) {
	return u.repo.GetWorkOrderByID(ctx, id)
}

func (u *UseCases) DuplicateWorkOrder(ctx context.Context, number string) (string, error) {
	return "", nil
}

func (u *UseCases) UpdateWorkOrderByID(ctx context.Context, o *domain.WorkOrder) error {
	if o == nil {
		return domainerr.Validation("work order is nil")
	}
	if err := validateDate(o); err != nil {
		return err
	}
	if err := validateItems(o); err != nil {
		return err
	}
	existing, err := u.repo.GetWorkOrderByID(ctx, o.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return domainerr.NotFound("work order not found")
	}
	if err := validateWorkOrderNumberForUpdate(existing, o); err != nil {
		return err
	}
	if err := ensureWorkOrderNumberIsUnique(ctx, u.repo, o.Number, o.ProjectID, o.ID); err != nil {
		return err
	}
	if err := validateInvestorSplits(o); err != nil {
		return err
	}
	if err := validateUniqueSupplyItems(o); err != nil {
		return err
	}
	return u.repo.UpdateWorkOrderByID(ctx, o)
}

func ensureWorkOrderNumberIsUnique(
	ctx context.Context,
	repo RepositoryPort,
	number string,
	projectID int64,
	currentID int64,
) error {
	existing, err := repo.GetWorkOrderByNumberAndProjectID(ctx, number, projectID)
	if err != nil {
		return err
	}
	if existing != nil && existing.ID != currentID {
		return domainerr.Conflict("work order number already exists in this project")
	}
	return nil
}

func validateWorkOrderNumberForCreate(o *domain.WorkOrder) error {
	if o == nil {
		return domainerr.Validation("work order is nil")
	}

	o.Number = normalizeWorkOrderNumber(o.Number)
	if o.Number == "" {
		return domainerr.Validation("work order number is required")
	}
	if !isOfficialWorkOrderNumber(o.Number) {
		return domainerr.Validation("work order number must contain digits only")
	}

	return nil
}

func validateWorkOrderNumberForUpdate(existing, incoming *domain.WorkOrder) error {
	if existing == nil || incoming == nil {
		return domainerr.Validation("work order is nil")
	}

	incoming.Number = normalizeWorkOrderNumber(incoming.Number)
	if incoming.Number == "" {
		return domainerr.Validation("work order number is required")
	}

	currentOfficialNumber := normalizeWorkOrderNumber(existing.Number)
	if incoming.Number == currentOfficialNumber {
		return nil
	}

	if existing.LegacyNumber != nil && incoming.Number == normalizeWorkOrderNumber(*existing.LegacyNumber) {
		incoming.Number = currentOfficialNumber
		return nil
	}

	if !isOfficialWorkOrderNumber(incoming.Number) {
		return domainerr.Validation("work order number must contain digits only")
	}

	return nil
}

func normalizeWorkOrderNumber(number string) string {
	return strings.TrimSpace(number)
}

func isOfficialWorkOrderNumber(number string) bool {
	return workOrderNumberPattern.MatchString(number)
}

func validateDate(o *domain.WorkOrder) error {
	if o == nil {
		return nil
	}
	if o.Date.IsZero() {
		return nil
	}
	today := time.Now().Truncate(24 * time.Hour)
	if o.Date.After(today) {
		return domainerr.Validation("la fecha de la orden de trabajo no puede ser futura")
	}
	return nil
}

func (u *UseCases) UpdateInvestorPaymentStatus(
	ctx context.Context,
	workOrderID int64,
	investorID int64,
	paymentStatus string,
) error {
	normalized, err := normalizeInvestorPaymentStatus(paymentStatus, false)
	if err != nil {
		return err
	}
	return u.repo.UpdateInvestorPaymentStatus(ctx, workOrderID, investorID, normalized)
}

func validateInvestorSplits(o *domain.WorkOrder) error {
	if o == nil {
		return domainerr.Validation("work order is nil")
	}
	if len(o.InvestorSplits) == 0 {
		return nil
	}

	seen := map[int64]struct{}{}
	sum := decimal.Zero
	for _, s := range o.InvestorSplits {
		if s.InvestorID <= 0 {
			return domainerr.Validation("invalid investor_id in investor_splits")
		}
		if s.Percentage.LessThanOrEqual(decimal.Zero) {
			return domainerr.Validation("invalid percentage in investor_splits")
		}
		if _, ok := seen[s.InvestorID]; ok {
			return domainerr.Validation("duplicate investor_id in investor_splits")
		}
		if _, err := normalizeInvestorPaymentStatus(s.PaymentStatus, true); err != nil {
			return err
		}
		seen[s.InvestorID] = struct{}{}
		sum = sum.Add(s.Percentage)
	}

	// Permitir un margen mínimo por decimales.
	if sum.Sub(decimal.NewFromInt(100)).Abs().GreaterThan(decimal.NewFromFloat(0.001)) {
		return domainerr.Validation("investor_splits percentage must sum to 100")
	}
	return nil
}

func validateItems(o *domain.WorkOrder) error {
	if o == nil {
		return types.NewError(types.ErrValidation, "work order is nil", nil)
	}

	seenSupplyIDs := make(map[int64]struct{})

	for _, item := range o.Items {
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

	return nil
}

func validateUniqueSupplyItems(o *domain.WorkOrder) error {
	if o == nil {
		return domainerr.Validation("work order is nil")
	}

	if len(o.Items) == 0 {
		return nil
	}

	seen := map[int64]struct{}{}
	for _, item := range o.Items {
		if item.SupplyID <= 0 {
			return domainerr.Validation("invalid supply_id in items")
		}

		if _, ok := seen[item.SupplyID]; ok {
			return domainerr.Validation("duplicate supply_id in items")
		}

		seen[item.SupplyID] = struct{}{}
	}

	return nil
}

func normalizeInvestorPaymentStatus(status string, allowEmpty bool) (string, error) {
	normalized := strings.TrimSpace(status)
	if normalized == "" && allowEmpty {
		return "", nil
	}

	switch normalized {
	case "", domain.InvestorPaymentStatusPending:
		return domain.InvestorPaymentStatusPending, nil
	case domain.InvestorPaymentStatusPaid:
		return domain.InvestorPaymentStatusPaid, nil
	default:
		return "", domainerr.Validation("invalid investor payment status")
	}
}

func (u *UseCases) DeleteWorkOrderByID(ctx context.Context, id int64) error {
	return u.repo.DeleteWorkOrderByID(ctx, id)
}

func (u *UseCases) ArchiveWorkOrder(ctx context.Context, id int64) error {
	return u.repo.ArchiveWorkOrder(ctx, id)
}

func (u *UseCases) RestoreWorkOrder(ctx context.Context, id int64) error {
	return u.repo.RestoreWorkOrder(ctx, id)
}

// ListWorkOrders delega al repositorio.
func (u *UseCases) ListWorkOrders(
	ctx context.Context,
	filt domain.WorkOrderFilter,
	inp types.Input,
) ([]domain.WorkOrderListElement, types.PageInfo, error) {
	return u.repo.ListWorkOrders(ctx, filt, inp)
}

// GetMetrics delega al repositorio.
func (u *UseCases) GetMetrics(ctx context.Context, f domain.WorkOrderFilter) (*domain.WorkOrderMetrics, error) {
	return u.repo.GetMetrics(ctx, f)
}

func (u *UseCases) ExportWorkOrders(ctx context.Context, filt domain.WorkOrderFilter, inp types.Input) ([]byte, error) {
	if u.excel == nil {
		return nil, domainerr.Internal("exporter not configured")
	}

	items, _, err := u.ListWorkOrders(ctx, filt, inp)
	if err != nil {
		return nil, domainerr.Internal("list work orders")
	}

	if len(items) == 0 {
		return nil, domainerr.NotFound("there is no data to export")
	}

	return u.excel.Export(ctx, items)
}
