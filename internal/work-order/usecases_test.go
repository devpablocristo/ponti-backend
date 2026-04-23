package workorder

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"

	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
	"github.com/devpablocristo/ponti-backend/internal/work-order/usecases/domain"
)

type workOrderRepoStub struct {
	getByIDFn               func(context.Context, int64) (*domain.WorkOrder, error)
	getByNumberAndProjectFn func(context.Context, string, int64) (*domain.WorkOrder, error)
	createFn                func(context.Context, *domain.WorkOrder) (int64, error)
	updateFn                func(context.Context, *domain.WorkOrder) error
}

func (s *workOrderRepoStub) CreateWorkOrder(ctx context.Context, wo *domain.WorkOrder) (int64, error) {
	if s.createFn != nil {
		return s.createFn(ctx, wo)
	}
	return 1, nil
}

func (s *workOrderRepoStub) GetWorkOrderByID(ctx context.Context, id int64) (*domain.WorkOrder, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (s *workOrderRepoStub) GetWorkOrderByNumberAndProjectID(
	ctx context.Context,
	number string,
	projectID int64,
) (*domain.WorkOrder, error) {
	if s.getByNumberAndProjectFn != nil {
		return s.getByNumberAndProjectFn(ctx, number, projectID)
	}
	return nil, nil
}

func (s *workOrderRepoStub) UpdateWorkOrderByID(ctx context.Context, wo *domain.WorkOrder) error {
	if s.updateFn != nil {
		return s.updateFn(ctx, wo)
	}
	return nil
}

func (s *workOrderRepoStub) UpdateInvestorPaymentStatus(context.Context, int64, int64, string) error {
	return nil
}
func (s *workOrderRepoStub) DeleteWorkOrderByID(context.Context, int64) error { return nil }
func (s *workOrderRepoStub) ArchiveWorkOrder(context.Context, int64) error    { return nil }
func (s *workOrderRepoStub) RestoreWorkOrder(context.Context, int64) error    { return nil }
func (s *workOrderRepoStub) ListWorkOrders(context.Context, domain.WorkOrderFilter, types.Input) ([]domain.WorkOrderListElement, types.PageInfo, error) {
	return nil, types.PageInfo{}, nil
}
func (s *workOrderRepoStub) GetMetrics(context.Context, domain.WorkOrderFilter) (*domain.WorkOrderMetrics, error) {
	return nil, nil
}
func (s *workOrderRepoStub) GetRawDirectCost(context.Context, int64) (decimal.Decimal, error) {
	return decimal.Zero, nil
}

type workOrderExporterStub struct{}

func (workOrderExporterStub) Export(context.Context, []domain.WorkOrderListElement) ([]byte, error) {
	return nil, nil
}

func (workOrderExporterStub) Close() error { return nil }

func TestNormalizeInvestorPaymentStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     string
		allowEmpty bool
		want       string
		wantErr    bool
	}{
		{
			name:       "default pending on empty when allowed",
			status:     "",
			allowEmpty: false,
			want:       domain.InvestorPaymentStatusPending,
		},
		{
			name:       "allow explicit empty on split payload",
			status:     "",
			allowEmpty: true,
			want:       "",
		},
		{
			name:       "accept paid",
			status:     domain.InvestorPaymentStatusPaid,
			allowEmpty: false,
			want:       domain.InvestorPaymentStatusPaid,
		},
		{
			name:       "reject unknown status",
			status:     "Facturada",
			allowEmpty: false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := normalizeInvestorPaymentStatus(tt.status, tt.allowEmpty)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestValidateInvestorSplitsRejectsInvalidPaymentStatus(t *testing.T) {
	t.Parallel()

	workOrder := &domain.WorkOrder{
		InvestorSplits: []domain.WorkOrderInvestorSplit{
			{
				InvestorID:    10,
				Percentage:    decimal.NewFromInt(100),
				PaymentStatus: "Facturada",
			},
		},
	}

	if err := validateInvestorSplits(workOrder); err == nil {
		t.Fatalf("expected validation error for invalid payment status")
	}
}

func TestCreateWorkOrderRejectsInvalidNumber(t *testing.T) {
	t.Parallel()

	ucs := NewUseCases(&workOrderRepoStub{}, workOrderExporterStub{})

	_, err := ucs.CreateWorkOrder(context.Background(), &domain.WorkOrder{
		Number:    "1861.1",
		ProjectID: 30,
	})
	if err == nil {
		t.Fatalf("expected validation error for invalid work order number")
	}
}

func TestCreateWorkOrderTrimsAndPersistsOfficialNumber(t *testing.T) {
	t.Parallel()

	var created *domain.WorkOrder
	ucs := NewUseCases(&workOrderRepoStub{
		createFn: func(_ context.Context, wo *domain.WorkOrder) (int64, error) {
			created = wo
			return 1, nil
		},
	}, workOrderExporterStub{})

	if _, err := ucs.CreateWorkOrder(context.Background(), &domain.WorkOrder{
		Number:    " 1901 ",
		ProjectID: 30,
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if created == nil || created.Number != "1901" {
		t.Fatalf("expected trimmed number 1901, got %+v", created)
	}
}

func TestUpdateWorkOrderAllowsUnchangedOfficialNumber(t *testing.T) {
	t.Parallel()

	var updated *domain.WorkOrder
	ucs := NewUseCases(&workOrderRepoStub{
		getByIDFn: func(context.Context, int64) (*domain.WorkOrder, error) {
			return &domain.WorkOrder{ID: 10, Number: "1901", ProjectID: 30}, nil
		},
		updateFn: func(_ context.Context, wo *domain.WorkOrder) error {
			updated = wo
			return nil
		},
	}, workOrderExporterStub{})

	err := ucs.UpdateWorkOrderByID(context.Background(), &domain.WorkOrder{
		ID:        10,
		Number:    "1901",
		ProjectID: 30,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated == nil || updated.Number != "1901" {
		t.Fatalf("expected unchanged official number, got %+v", updated)
	}
}

func TestUpdateWorkOrderRejectsChangedInvalidNumber(t *testing.T) {
	t.Parallel()

	ucs := NewUseCases(&workOrderRepoStub{
		getByIDFn: func(context.Context, int64) (*domain.WorkOrder, error) {
			return &domain.WorkOrder{ID: 10, Number: "1901", ProjectID: 30}, nil
		},
	}, workOrderExporterStub{})

	err := ucs.UpdateWorkOrderByID(context.Background(), &domain.WorkOrder{
		ID:        10,
		Number:    "1901.1",
		ProjectID: 30,
	})
	if err == nil {
		t.Fatalf("expected validation error for changed invalid number")
	}
}
