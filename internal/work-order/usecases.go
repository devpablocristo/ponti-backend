// Package workorder contiene casos de uso para work orders.
package workorder

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/alphacodinggroup/ponti-backend/internal/work-order/usecases/domain"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

type RepositoryPort interface {
	CreateWorkOrder(context.Context, *domain.WorkOrder) (int64, error)
	GetWorkOrderByID(ctx context.Context, id int64) (*domain.WorkOrder, error)
	GetWorkOrderByNumberAndProjectID(ctx context.Context, number string, projectID int64) (*domain.WorkOrder, error)
	UpdateWorkOrderByID(context.Context, *domain.WorkOrder) error
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

// NewUseCases crea una instancia de casos de uso para work orders.
func NewUseCases(r RepositoryPort, excel ExporterAdapterPort) *UseCases {
	return &UseCases{repo: r, excel: excel}
}

func (u *UseCases) CreateWorkOrder(ctx context.Context, o *domain.WorkOrder) (int64, error) {
	if o == nil {
		return 0, types.NewError(types.ErrValidation, "work order is nil", nil)
	}
	if err := validateItems(o); err != nil {
		return 0, err
	}
	if err := validateInvestorSplits(o); err != nil {
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
	if err := validateItems(o); err != nil {
		return err
	}
	if err := validateInvestorSplits(o); err != nil {
		return err
	}
	return u.repo.UpdateWorkOrderByID(ctx, o)
}

func validateInvestorSplits(o *domain.WorkOrder) error {
	if o == nil {
		return types.NewError(types.ErrValidation, "work order is nil", nil)
	}
	if len(o.InvestorSplits) == 0 {
		return nil
	}

	seen := map[int64]struct{}{}
	sum := decimal.Zero
	for _, s := range o.InvestorSplits {
		if s.InvestorID <= 0 {
			return types.NewError(types.ErrValidation, "invalid investor_id in investor_splits", nil)
		}
		if s.Percentage.LessThanOrEqual(decimal.Zero) {
			return types.NewError(types.ErrValidation, "invalid percentage in investor_splits", nil)
		}
		if _, ok := seen[s.InvestorID]; ok {
			return types.NewError(types.ErrValidation, "duplicate investor_id in investor_splits", nil)
		}
		seen[s.InvestorID] = struct{}{}
		sum = sum.Add(s.Percentage)
	}

	// Permitir un margen mínimo por decimales.
	if sum.Sub(decimal.NewFromInt(100)).Abs().GreaterThan(decimal.NewFromFloat(0.001)) {
		return types.NewError(types.ErrValidation, "investor_splits percentage must sum to 100", nil)
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
		return nil, types.NewError(types.ErrInternal, "exporter not configured", nil)
	}

	items, _, err := u.ListWorkOrders(ctx, filt, inp)
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "list work orders", err)
	}

	if len(items) == 0 {
		return nil, types.NewError(types.ErrNotFound, "there is no data to export", nil)
	}

	return u.excel.Export(ctx, items)
}
