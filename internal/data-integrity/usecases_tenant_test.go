package dataintegrity

import (
	"context"
	"testing"
	"time"

	contextkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	dashboardDomain "github.com/devpablocristo/ponti-backend/internal/dashboard/usecases/domain"
	lotDomain "github.com/devpablocristo/ponti-backend/internal/lot/usecases/domain"
	reportDomain "github.com/devpablocristo/ponti-backend/internal/report/usecases/domain"
	stockDomain "github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
	workOrderDomain "github.com/devpablocristo/ponti-backend/internal/work-order/usecases/domain"
)

func dataIntegrityTenantContext(tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "tenant-user@example.com")
	ctx = context.WithValue(ctx, contextkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, contextkeys.Role, "tenant_viewer")
	ctx = context.WithValue(ctx, contextkeys.Scopes, []string{"data-integrity.read"})
	return ctx
}

type dataIntegrityCapture struct {
	t         *testing.T
	tenantID  uuid.UUID
	projectID int64
	calls     []string
}

func (c *dataIntegrityCapture) assertTenantContext(ctx context.Context, call string) {
	c.t.Helper()
	got, _ := ctx.Value(contextkeys.OrgID).(uuid.UUID)
	if got != c.tenantID {
		c.t.Fatalf("%s received wrong tenant context: got %s want %s", call, got, c.tenantID)
	}
	c.calls = append(c.calls, call)
}

func (c *dataIntegrityCapture) assertProject(call string, got *int64) {
	c.t.Helper()
	if got == nil || *got != c.projectID {
		c.t.Fatalf("%s received wrong project filter: got %#v want %d", call, got, c.projectID)
	}
}

type captureLotRepo struct{ capture *dataIntegrityCapture }

func (r captureLotRepo) ListLots(ctx context.Context, filter lotDomain.LotListFilter, _ int, _ int) ([]lotDomain.LotTable, int, decimal.Decimal, decimal.Decimal, error) {
	r.capture.assertTenantContext(ctx, "lots")
	r.capture.assertProject("lots", filter.ProjectID)
	return []lotDomain.LotTable{}, 0, decimal.Zero, decimal.Zero, nil
}

type captureDashboardRepo struct{ capture *dataIntegrityCapture }

func (r captureDashboardRepo) GetDashboard(ctx context.Context, filter dashboardDomain.DashboardFilter) (*dashboardDomain.DashboardData, error) {
	r.capture.assertTenantContext(ctx, "dashboard")
	r.capture.assertProject("dashboard", filter.ProjectID)
	return &dashboardDomain.DashboardData{}, nil
}

type captureReportRepo struct{ capture *dataIntegrityCapture }

func (r captureReportRepo) GetSummaryResults(ctx context.Context, filter reportDomain.SummaryResultsFilter) ([]reportDomain.SummaryResults, error) {
	r.capture.assertTenantContext(ctx, "summary")
	r.capture.assertProject("summary", filter.ProjectID)
	return []reportDomain.SummaryResults{}, nil
}

func (r captureReportRepo) GetFieldCropMetrics(ctx context.Context, filter reportDomain.ReportFilter) ([]reportDomain.FieldCropMetric, error) {
	r.capture.assertTenantContext(ctx, "field-crop")
	r.capture.assertProject("field-crop", filter.ProjectID)
	return []reportDomain.FieldCropMetric{}, nil
}

func (r captureReportRepo) GetInvestorContributionReport(ctx context.Context, filter reportDomain.ReportFilter) (*reportDomain.InvestorContributionReport, error) {
	r.capture.assertTenantContext(ctx, "investor-contribution")
	r.capture.assertProject("investor-contribution", filter.ProjectID)
	return &reportDomain.InvestorContributionReport{}, nil
}

type captureWorkOrderRepo struct{ capture *dataIntegrityCapture }

func (r captureWorkOrderRepo) GetMetrics(ctx context.Context, filter workOrderDomain.WorkOrderFilter) (*workOrderDomain.WorkOrderMetrics, error) {
	r.capture.assertTenantContext(ctx, "workorder-metrics")
	r.capture.assertProject("workorder-metrics", filter.ProjectID)
	return &workOrderDomain.WorkOrderMetrics{}, nil
}

func (r captureWorkOrderRepo) GetRawDirectCost(ctx context.Context, projectID int64) (decimal.Decimal, error) {
	r.capture.assertTenantContext(ctx, "workorder-raw")
	if projectID != r.capture.projectID {
		r.capture.t.Fatalf("workorder-raw received wrong project id: got %d want %d", projectID, r.capture.projectID)
	}
	return decimal.Zero, nil
}

type unusedStockRepo struct{}

func (unusedStockRepo) GetStocks(context.Context, int64, time.Time) ([]*stockDomain.Stock, error) {
	panic("stock repo should not be called while fetching shared data")
}

func TestFetchSharedDataPropagatesTenantContextAndProjectFilter(t *testing.T) {
	tenantID := uuid.New()
	projectID := int64(42)
	capture := &dataIntegrityCapture{t: t, tenantID: tenantID, projectID: projectID}

	ucs := NewUseCases(
		captureWorkOrderRepo{capture: capture},
		captureDashboardRepo{capture: capture},
		captureLotRepo{capture: capture},
		captureReportRepo{capture: capture},
		unusedStockRepo{},
	)

	if _, err := ucs.fetchSharedData(dataIntegrityTenantContext(tenantID), &projectID); err != nil {
		t.Fatalf("fetch shared data: %v", err)
	}

	expectedCalls := map[string]bool{
		"lots":                  false,
		"dashboard":             false,
		"field-crop":            false,
		"summary":               false,
		"investor-contribution": false,
		"workorder-metrics":     false,
		"workorder-raw":         false,
	}
	for _, call := range capture.calls {
		expectedCalls[call] = true
	}
	for call, seen := range expectedCalls {
		if !seen {
			t.Fatalf("expected %s to be called, calls=%#v", call, capture.calls)
		}
	}
}
