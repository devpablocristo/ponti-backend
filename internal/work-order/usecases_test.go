package workorder

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
	"github.com/devpablocristo/ponti-backend/internal/work-order/usecases/domain"
)

type useCasesRepoStub struct {
	getHarvestAreaSnapshotFn  func(context.Context, int64, int64, int64) (bool, decimal.Decimal, decimal.Decimal, error)
	listWorkOrderFilterRowsFn func(context.Context, domain.WorkOrderFilter) ([]domain.WorkOrderListElement, error)
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
func (s useCasesRepoStub) ListWorkOrderFilterRows(ctx context.Context, filt domain.WorkOrderFilter) ([]domain.WorkOrderListElement, error) {
	if s.listWorkOrderFilterRowsFn != nil {
		return s.listWorkOrderFilterRowsFn(ctx, filt)
	}
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

func TestListWorkOrderFilterRowsDelegatesWithoutOwnValidation(t *testing.T) {
	t.Parallel()

	// El mínimo cliente+proyecto+campaña se exige en el handler (ValidateRequiredWorkspaceFilter);
	// el use case ya no aplica una validación propia más débil: debe delegar tal cual al repositorio,
	// incluso con un filtro vacío que antes habría rechazado.

	// 1) Delegación pura: el resultado del repo se devuelve sin modificar y el filtro llega intacto.
	sentinel := []domain.WorkOrderListElement{{ID: 7}, {ID: 9}}
	var gotFilter domain.WorkOrderFilter
	uc := NewUseCases(useCasesRepoStub{
		listWorkOrderFilterRowsFn: func(_ context.Context, filt domain.WorkOrderFilter) ([]domain.WorkOrderListElement, error) {
			gotFilter = filt
			return sentinel, nil
		},
	}, nil)

	rows, err := uc.ListWorkOrderFilterRows(context.Background(), domain.WorkOrderFilter{})
	if err != nil {
		t.Fatalf("use case should delegate without its own validation, got: %v", err)
	}
	if len(rows) != len(sentinel) || rows[0].ID != 7 || rows[1].ID != 9 {
		t.Fatalf("expected the repo result to be returned unchanged, got: %+v", rows)
	}
	if gotFilter != (domain.WorkOrderFilter{}) {
		t.Fatalf("expected the empty filter to reach the repo unchanged, got: %+v", gotFilter)
	}

	// 2) Propagación de error: lo que devuelve el repo se propaga sin envolver.
	repoErr := errors.New("repo boom")
	ucErr := NewUseCases(useCasesRepoStub{
		listWorkOrderFilterRowsFn: func(context.Context, domain.WorkOrderFilter) ([]domain.WorkOrderListElement, error) {
			return nil, repoErr
		},
	}, nil)
	if _, err := ucErr.ListWorkOrderFilterRows(context.Background(), domain.WorkOrderFilter{}); !errors.Is(err, repoErr) {
		t.Fatalf("expected the repo error to propagate, got: %v", err)
	}
}

func TestParseFiltersRequiresWorkspaceMinimum(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	// Sin cliente+proyecto+campaña -> error (regla compartida ValidateRequiredWorkspaceFilter).
	missing, _ := gin.CreateTestContext(httptest.NewRecorder())
	missing.Request = httptest.NewRequest(http.MethodGet, "/work-orders", nil)
	if _, err := parseFilters(missing); err == nil {
		t.Fatalf("expected error when customer_id/project_id/campaign_id are missing")
	}

	// Cada required ausente por separado (los otros dos presentes) -> error.
	partials := map[string]string{
		"falta customer_id": "/work-orders?project_id=2&campaign_id=3",
		"falta project_id":  "/work-orders?customer_id=1&campaign_id=3",
		"falta campaign_id": "/work-orders?customer_id=1&project_id=2",
	}
	for name, target := range partials {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, target, nil)
		if _, err := parseFilters(c); err == nil {
			t.Fatalf("%s: expected error when a required workspace id is missing", name)
		}
	}

	// ID mal formado -> error de parseo propagado (antes se tragaba y devolvía lista global).
	malformed, _ := gin.CreateTestContext(httptest.NewRecorder())
	malformed.Request = httptest.NewRequest(http.MethodGet, "/work-orders?customer_id=abc&project_id=2&campaign_id=3", nil)
	if _, err := parseFilters(malformed); err == nil {
		t.Fatalf("expected error for a malformed customer_id (must not fall back to an empty/global filter)")
	}

	// Con los tres -> ok; el campo es opcional (nil = todos los campos).
	full, _ := gin.CreateTestContext(httptest.NewRecorder())
	full.Request = httptest.NewRequest(http.MethodGet, "/work-orders?customer_id=1&project_id=2&campaign_id=3", nil)
	f, err := parseFilters(full)
	if err != nil {
		t.Fatalf("expected no error with full workspace, got: %v", err)
	}
	if f.CustomerID == nil || f.ProjectID == nil || f.CampaignID == nil {
		t.Fatalf("expected customer/project/campaign to be parsed")
	}
	if f.FieldID != nil {
		t.Fatalf("field_id should be optional (nil = all fields)")
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
