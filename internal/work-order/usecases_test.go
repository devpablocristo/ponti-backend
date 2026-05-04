package workorder

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"

	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
	"github.com/devpablocristo/ponti-backend/internal/work-order/usecases/domain"
)

type useCasesRepoStub struct{}

func (useCasesRepoStub) CreateWorkOrder(context.Context, *domain.WorkOrder) (int64, error) {
	return 0, nil
}
func (useCasesRepoStub) GetWorkOrderByID(context.Context, int64) (*domain.WorkOrder, error) {
	return nil, nil
}
func (useCasesRepoStub) GetWorkOrderByNumberAndProjectID(context.Context, string, int64) (*domain.WorkOrder, error) {
	return nil, nil
}
func (useCasesRepoStub) UpdateWorkOrderByID(context.Context, *domain.WorkOrder) error {
	return nil
}
func (useCasesRepoStub) UpdateInvestorPaymentStatus(context.Context, int64, int64, string) error {
	return nil
}
func (useCasesRepoStub) DeleteWorkOrderByID(context.Context, int64) error {
	return nil
}
func (useCasesRepoStub) ArchiveWorkOrder(context.Context, int64) error {
	return nil
}
func (useCasesRepoStub) RestoreWorkOrder(context.Context, int64) error {
	return nil
}
func (useCasesRepoStub) ListWorkOrders(context.Context, domain.WorkOrderFilter, types.Input) ([]domain.WorkOrderListElement, types.PageInfo, error) {
	return nil, types.PageInfo{}, nil
}
func (useCasesRepoStub) ListWorkOrderFilterRows(context.Context, domain.WorkOrderFilter) ([]domain.WorkOrderListElement, error) {
	return nil, nil
}
func (useCasesRepoStub) GetMetrics(context.Context, domain.WorkOrderFilter) (*domain.WorkOrderMetrics, error) {
	return nil, nil
}
func (useCasesRepoStub) GetRawDirectCost(context.Context, int64) (decimal.Decimal, error) {
	return decimal.Zero, nil
}

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

func TestListWorkOrderFilterRowsRequiresProjectOrFieldScope(t *testing.T) {
	t.Parallel()

	uc := NewUseCases(useCasesRepoStub{}, nil)

	if _, err := uc.ListWorkOrderFilterRows(context.Background(), domain.WorkOrderFilter{}); err == nil {
		t.Fatalf("expected validation error for unscoped filter rows request")
	}
}
