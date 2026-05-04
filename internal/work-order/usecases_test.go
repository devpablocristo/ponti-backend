package workorder

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"

	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
	"github.com/devpablocristo/ponti-backend/internal/work-order/usecases/domain"
)

type useCasesRepoStub struct {
	getHarvestAreaSnapshotFn func(context.Context, int64, int64, int64) (bool, decimal.Decimal, decimal.Decimal, error)
}

func (s useCasesRepoStub) GetHarvestAreaSnapshot(
	ctx context.Context,
	lotID int64,
	laborID int64,
	excludeWorkOrderID int64,
) (bool, decimal.Decimal, decimal.Decimal, error) {
	if s.getHarvestAreaSnapshotFn != nil {
		return s.getHarvestAreaSnapshotFn(ctx, lotID, laborID, excludeWorkOrderID)
	}
	return false, decimal.Zero, decimal.Zero, nil
}

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

func TestCreateWorkOrderAllowsPartialHarvestWithinLotArea(t *testing.T) {
	t.Parallel()

	uc := NewUseCases(useCasesRepoStub{
		getHarvestAreaSnapshotFn: func(context.Context, int64, int64, int64) (bool, decimal.Decimal, decimal.Decimal, error) {
			return true, decimal.NewFromInt(100), decimal.NewFromInt(40), nil
		},
	}, nil)

	workOrder := validHarvestWorkOrder()
	workOrder.EffectiveArea = decimal.NewFromInt(30)

	if _, err := uc.CreateWorkOrder(context.Background(), workOrder); err != nil {
		t.Fatalf("expected partial harvest to be allowed, got %v", err)
	}
}

func TestCreateWorkOrderRejectsHarvestAreaOverLotArea(t *testing.T) {
	t.Parallel()

	uc := NewUseCases(useCasesRepoStub{
		getHarvestAreaSnapshotFn: func(context.Context, int64, int64, int64) (bool, decimal.Decimal, decimal.Decimal, error) {
			return true, decimal.NewFromInt(100), decimal.NewFromInt(80), nil
		},
	}, nil)

	workOrder := validHarvestWorkOrder()
	workOrder.EffectiveArea = decimal.NewFromInt(25)

	if _, err := uc.CreateWorkOrder(context.Background(), workOrder); err == nil {
		t.Fatal("expected validation error for harvest area over lot area")
	}
}

func TestUpdateWorkOrderExcludesCurrentWorkOrderFromHarvestArea(t *testing.T) {
	t.Parallel()

	var gotExcludeID int64
	uc := NewUseCases(useCasesRepoStub{
		getHarvestAreaSnapshotFn: func(_ context.Context, _ int64, _ int64, excludeWorkOrderID int64) (bool, decimal.Decimal, decimal.Decimal, error) {
			gotExcludeID = excludeWorkOrderID
			return true, decimal.NewFromInt(100), decimal.NewFromInt(40), nil
		},
	}, nil)

	workOrder := validHarvestWorkOrder()
	workOrder.ID = 55
	workOrder.EffectiveArea = decimal.NewFromInt(50)

	if err := uc.UpdateWorkOrderByID(context.Background(), workOrder); err != nil {
		t.Fatalf("expected update to be allowed, got %v", err)
	}
	if gotExcludeID != 55 {
		t.Fatalf("expected excludeWorkOrderID 55, got %d", gotExcludeID)
	}
}

func TestCreateWorkOrderSkipsHarvestLimitForNonHarvestLabor(t *testing.T) {
	t.Parallel()

	uc := NewUseCases(useCasesRepoStub{
		getHarvestAreaSnapshotFn: func(context.Context, int64, int64, int64) (bool, decimal.Decimal, decimal.Decimal, error) {
			return false, decimal.NewFromInt(100), decimal.NewFromInt(100), nil
		},
	}, nil)

	workOrder := validHarvestWorkOrder()
	workOrder.EffectiveArea = decimal.NewFromInt(500)

	if _, err := uc.CreateWorkOrder(context.Background(), workOrder); err != nil {
		t.Fatalf("expected non-harvest labor to skip harvest area limit, got %v", err)
	}
}

func validHarvestWorkOrder() *domain.WorkOrder {
	return &domain.WorkOrder{
		Number:        "OT-1",
		ProjectID:     1,
		FieldID:       2,
		LotID:         3,
		CropID:        4,
		LaborID:       5,
		InvestorID:    6,
		EffectiveArea: decimal.NewFromInt(10),
		Items: []domain.WorkOrderItem{
			{
				SupplyID:  7,
				TotalUsed: decimal.NewFromInt(1),
				FinalDose: decimal.NewFromInt(1),
			},
		},
	}
}
