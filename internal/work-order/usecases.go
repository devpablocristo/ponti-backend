// Package workorder contiene casos de uso para work orders.
package workorder

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/devpablocristo/core/backend/go/domainerr"
	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
	"github.com/devpablocristo/ponti-backend/internal/work-order/usecases/domain"
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
		return 0, domainerr.Validation("work order is nil")
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
	if err := validateInvestorSplits(o); err != nil {
		return err
	}
	return u.repo.UpdateWorkOrderByID(ctx, o)
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
		seen[s.InvestorID] = struct{}{}
		sum = sum.Add(s.Percentage)
	}

	// Permitir un margen mínimo por decimales.
	if sum.Sub(decimal.NewFromInt(100)).Abs().GreaterThan(decimal.NewFromFloat(0.001)) {
		return domainerr.Validation("investor_splits percentage must sum to 100")
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
