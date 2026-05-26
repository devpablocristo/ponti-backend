package dataintegrity

import (
	"context"
	"testing"

	contextkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	dashboardDomain "github.com/devpablocristo/ponti-backend/internal/dashboard/usecases/domain"
	"github.com/devpablocristo/ponti-backend/internal/data-integrity/usecases/domain"
)

func dataIntegrityTenantContext(tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "tenant-user@example.com")
	ctx = context.WithValue(ctx, contextkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, contextkeys.Role, "tenant_viewer")
	ctx = context.WithValue(ctx, contextkeys.Scopes, []string{"data-integrity.read"})
	return ctx
}

type tenantCapture struct {
	t        *testing.T
	tenantID uuid.UUID
	calls    []string
}

func (c *tenantCapture) assertTenant(ctx context.Context, call string) {
	c.t.Helper()
	got, _ := ctx.Value(contextkeys.OrgID).(uuid.UUID)
	if got != c.tenantID {
		c.t.Fatalf("%s received wrong tenant context: got %s want %s", call, got, c.tenantID)
	}
	c.calls = append(c.calls, call)
}

// Mocks que validan tenant context además de devolver valores fijos.

type captureDashboardRepo struct {
	capture *tenantCapture
}

func (r captureDashboardRepo) GetDashboard(ctx context.Context, filter dashboardDomain.DashboardFilter) (*dashboardDomain.DashboardData, error) {
	r.capture.assertTenant(ctx, "dashboard")
	return &dashboardDomain.DashboardData{
		ManagementBalance: &dashboardDomain.DashboardManagementBalance{
			Summary: &dashboardDomain.DashboardBalanceSummary{},
		},
	}, nil
}

type captureWorkOrderRepo struct{ capture *tenantCapture }

func (r captureWorkOrderRepo) GetRawDirectCost(ctx context.Context, _ int64) (decimal.Decimal, error) {
	r.capture.assertTenant(ctx, "workorder-raw")
	return decimal.Zero, nil
}

type captureReportRepo struct{ capture *tenantCapture }

func (r captureReportRepo) GetRawNetIncome(ctx context.Context, _ int64) (decimal.Decimal, error) {
	r.capture.assertTenant(ctx, "report-raw")
	return decimal.Zero, nil
}

type captureSupplyRepo struct{ capture *tenantCapture }

func (r captureSupplyRepo) GetRawSupplyInvestment(ctx context.Context, _ int64) (decimal.Decimal, error) {
	r.capture.assertTenant(ctx, "supply-raw")
	return decimal.Zero, nil
}

type captureProjectRepo struct{ capture *tenantCapture }

func (r captureProjectRepo) GetRawAdminCostTotal(ctx context.Context, _ int64) (decimal.Decimal, error) {
	r.capture.assertTenant(ctx, "project-raw")
	return decimal.Zero, nil
}

type captureLotRepo struct{ capture *tenantCapture }

func (r captureLotRepo) GetRawLeaseExecuted(ctx context.Context, _ int64) (decimal.Decimal, error) {
	r.capture.assertTenant(ctx, "lot-raw")
	return decimal.Zero, nil
}

// TestCheckCostsCoherence_PropagatesTenantContext verifica que el contexto con tenant_id
// llega a cada repo invocado (dashboard SSOT + 5 RAW), respetando aislamiento multitenant.
func TestCheckCostsCoherence_PropagatesTenantContext(t *testing.T) {
	tenantID := uuid.New()
	projectID := int64(42)
	capture := &tenantCapture{t: t, tenantID: tenantID}

	uc := NewUseCases(
		captureDashboardRepo{capture: capture},
		captureWorkOrderRepo{capture: capture},
		captureReportRepo{capture: capture},
		captureSupplyRepo{capture: capture},
		captureProjectRepo{capture: capture},
		captureLotRepo{capture: capture},
	)

	if _, err := uc.CheckCostsCoherence(dataIntegrityTenantContext(tenantID), domain.CostsCheckFilter{ProjectID: &projectID}); err != nil {
		t.Fatalf("CheckCostsCoherence: %v", err)
	}

	expectedCalls := map[string]bool{
		"dashboard":     false,
		"workorder-raw": false,
		"report-raw":    false,
		"supply-raw":    false,
		"project-raw":   false,
		"lot-raw":       false,
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
