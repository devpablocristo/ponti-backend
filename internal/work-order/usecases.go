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
	workOrder, err := u.repo.GetWorkOrderByNumberAndProjectID(ctx, o.Number, o.ProjectID)
	if err != nil {
		return 0, err
	}
	if workOrder != nil {
		return 0, types.NewError(types.ErrConflict, "work order already exists", nil)
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
	return u.repo.UpdateWorkOrderByID(ctx, o)
}

func (u *UseCases) DeleteWorkOrderByID(ctx context.Context, id int64) error {
	return u.repo.DeleteWorkOrderByID(ctx, id)
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
