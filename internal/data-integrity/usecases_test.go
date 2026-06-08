package dataintegrity

import (
	"context"
	"testing"

	ctxkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dashboardDomain "github.com/devpablocristo/ponti-backend/internal/dashboard/usecases/domain"
	"github.com/devpablocristo/ponti-backend/internal/data-integrity/usecases/domain"
	lotDomain "github.com/devpablocristo/ponti-backend/internal/lot/usecases/domain"
	reportDomain "github.com/devpablocristo/ponti-backend/internal/report/usecases/domain"
	supplyDomain "github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
)

type fakeBusinessInsightsNotifier struct {
	notifyCalls                int
	resolveCalls               int
	tentativeNotifyCalls       int
	tentativeResolveCalls      int
	lastIssue                  DataIntegrityCriticalInput
	lastTentativePricesIssue   TentativePricesInput
	lastProject                string
	lastTentativePricesProject string
}

func (f *fakeBusinessInsightsNotifier) NotifyDataIntegrityCritical(_ context.Context, _ uuid.UUID, _ string, issue DataIntegrityCriticalInput) error {
	f.notifyCalls++
	f.lastIssue = issue
	return nil
}

func (f *fakeBusinessInsightsNotifier) MaybeResolveDataIntegrityCritical(_ context.Context, _ uuid.UUID, projectID string) error {
	f.resolveCalls++
	f.lastProject = projectID
	return nil
}

func (f *fakeBusinessInsightsNotifier) NotifyTentativePrices(_ context.Context, _ uuid.UUID, _ string, issue TentativePricesInput) error {
	f.tentativeNotifyCalls++
	f.lastTentativePricesIssue = issue
	return nil
}

func (f *fakeBusinessInsightsNotifier) MaybeResolveTentativePrices(_ context.Context, _ uuid.UUID, projectID string) error {
	f.tentativeResolveCalls++
	f.lastTentativePricesProject = projectID
	return nil
}

func TestEvaluateIntegrityInsight_NotifiesWhenChecksFail(t *testing.T) {
	notifier := &fakeBusinessInsightsNotifier{}
	useCases := NewUseCases(nil, nil, nil, nil, nil, nil)
	useCases.SetBusinessInsightsNotifier(notifier)
	projectID := int64(4)
	diffB := decimal.NewFromInt(9)
	recalcBSource := "summary_results"
	ctx := context.WithValue(context.Background(), ctxkeys.OrgID, uuid.New())
	ctx = context.WithValue(ctx, ctxkeys.Actor, "user-1")

	useCases.evaluateIntegrityInsight(ctx, &projectID, &domain.IntegrityReport{
		Checks: []domain.IntegrityCheck{
			{ControlNumber: 1, Status: "OK"},
			{
				ControlNumber: 2,
				Status:        "ERROR",
				DataToVerify:  "Costos directos",
				Description:   "Dashboard vs informes",
				DifferenceA:   decimal.NewFromInt(10),
				DifferenceB:   &diffB,
				Tolerance:     decimal.NewFromInt(1),
				SystemSource:  "dashboard",
				RecalcASource: "lot_list",
				RecalcBSource: &recalcBSource,
			},
		},
	})

	require.Equal(t, 1, notifier.notifyCalls)
	require.Equal(t, 0, notifier.resolveCalls)
	assert.Equal(t, "4", notifier.lastIssue.ProjectID)
	assert.Equal(t, 1, notifier.lastIssue.FailedChecks)
	assert.Equal(t, 2, notifier.lastIssue.TotalChecks)
	require.Len(t, notifier.lastIssue.Controls, 1)
	assert.Equal(t, 2, notifier.lastIssue.Controls[0].ControlNumber)
	assert.Equal(t, "summary_results", notifier.lastIssue.Controls[0].RecalcBSource)
}

func TestEvaluateIntegrityInsight_ResolvesWhenChecksRecover(t *testing.T) {
	notifier := &fakeBusinessInsightsNotifier{}
	useCases := NewUseCases(nil, nil, nil, nil, nil, nil)
	useCases.SetBusinessInsightsNotifier(notifier)
	projectID := int64(4)
	ctx := context.WithValue(context.Background(), ctxkeys.OrgID, uuid.New())

	useCases.evaluateIntegrityInsight(ctx, &projectID, &domain.IntegrityReport{
		Checks: []domain.IntegrityCheck{{ControlNumber: 1, Status: "OK"}},
	})

	require.Equal(t, 0, notifier.notifyCalls)
	require.Equal(t, 1, notifier.resolveCalls)
	assert.Equal(t, "4", notifier.lastProject)
}

func TestGetTentativePrices_NotifiesWhenTentativePricesExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	supplyRepo := NewMockSupplyRepositoryPort(ctrl)
	projectID := int64(4)
	customerID := int64(1)
	campaignID := int64(2)
	supplyRepo.EXPECT().
		ListTentativePrices(gomock.Any(), supplyDomain.SupplyFilter{
			CustomerID: &customerID,
			ProjectID:  &projectID,
			CampaignID: &campaignID,
		}, 10).
		Return([]supplyDomain.TentativePriceItem{
			{SupplyID: 10, Name: "Insumo", CategoryName: "Fertilizante", Price: decimal.RequireFromString("123.45")},
		}, int64(1), nil)
	notifier := &fakeBusinessInsightsNotifier{}
	useCases := NewUseCases(nil, nil, nil, nil, nil, supplyRepo)
	useCases.SetBusinessInsightsNotifier(notifier)
	ctx := context.WithValue(context.Background(), ctxkeys.OrgID, uuid.New())
	ctx = context.WithValue(ctx, ctxkeys.Actor, "user-1")

	report, err := useCases.GetTentativePrices(ctx, domain.TentativePricesFilter{
		CustomerID: &customerID,
		ProjectID:  &projectID,
		CampaignID: &campaignID,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), report.Count)
	require.Equal(t, 1, notifier.tentativeNotifyCalls)
	require.Equal(t, 0, notifier.tentativeResolveCalls)
	assert.Equal(t, "4", notifier.lastTentativePricesIssue.ProjectID)
	require.Len(t, notifier.lastTentativePricesIssue.SampleItems, 1)
	assert.Equal(t, "Insumo", notifier.lastTentativePricesIssue.SampleItems[0].Name)
}

func TestGetTentativePrices_ResolvesWhenNoTentativePricesRemain(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	supplyRepo := NewMockSupplyRepositoryPort(ctrl)
	projectID := int64(4)
	supplyRepo.EXPECT().
		ListTentativePrices(gomock.Any(), supplyDomain.SupplyFilter{ProjectID: &projectID}, 10).
		Return([]supplyDomain.TentativePriceItem{}, int64(0), nil)
	notifier := &fakeBusinessInsightsNotifier{}
	useCases := NewUseCases(nil, nil, nil, nil, nil, supplyRepo)
	useCases.SetBusinessInsightsNotifier(notifier)
	ctx := context.WithValue(context.Background(), ctxkeys.OrgID, uuid.New())

	report, err := useCases.GetTentativePrices(ctx, domain.TentativePricesFilter{ProjectID: &projectID})

	require.NoError(t, err)
	require.Equal(t, int64(0), report.Count)
	require.Equal(t, 0, notifier.tentativeNotifyCalls)
	require.Equal(t, 1, notifier.tentativeResolveCalls)
	assert.Equal(t, "4", notifier.lastTentativePricesProject)
}

func TestUseCases_control1LotesVsDashboard(t *testing.T) {
	tests := []struct {
		name               string
		mockLots           []lotDomain.LotTable
		mockDashboardData  *dashboardDomain.DashboardData
		mockSummaryResults []reportDomain.SummaryResults
		expectedStatus     string
		expectedDiffA      string
		expectedSystemVal  string
		hasDiffB           bool
		expectedDiffB      string
	}{
		{
			name: "OK - todos los valores iguales",
			mockLots: []lotDomain.LotTable{
				{
					CostUsdPerHa: decimal.NewFromFloat(1000.50),
					Hectares:     decimal.NewFromFloat(10.0),
				},
				{
					CostUsdPerHa: decimal.NewFromFloat(845.439),
					Hectares:     decimal.NewFromFloat(10.0),
				},
			},
			mockDashboardData: &dashboardDomain.DashboardData{
				ManagementBalance: &dashboardDomain.DashboardManagementBalance{
					Summary: &dashboardDomain.DashboardBalanceSummary{
						DirectCostsExecutedUSD: decimal.NewFromFloat(18459.39),
					},
				},
			},
			mockSummaryResults: []reportDomain.SummaryResults{
				{TotalDirectCostsUsd: decimal.NewFromFloat(18459.39)},
			},
			expectedStatus:    "OK",
			expectedDiffA:     "0.00",
			expectedSystemVal: "18459.39",
			hasDiffB:          true,
			expectedDiffB:     "0.00",
		},
		{
			name: "OK - diferencia dentro de tolerancia",
			mockLots: []lotDomain.LotTable{
				{
					CostUsdPerHa: decimal.NewFromFloat(1845.439),
					Hectares:     decimal.NewFromFloat(10.0),
				},
			},
			mockDashboardData: &dashboardDomain.DashboardData{
				ManagementBalance: &dashboardDomain.DashboardManagementBalance{
					Summary: &dashboardDomain.DashboardBalanceSummary{
						DirectCostsExecutedUSD: decimal.NewFromFloat(18455.17),
					},
				},
			},
			mockSummaryResults: []reportDomain.SummaryResults{
				{TotalDirectCostsUsd: decimal.NewFromFloat(18455.17)},
			},
			expectedStatus:    "OK",
			expectedDiffA:     "0.78",
			expectedSystemVal: "18455.17",
			hasDiffB:          true,
			expectedDiffB:     "0.00",
		},
		{
			name: "ERROR - diferencia fuera de tolerancia",
			mockLots: []lotDomain.LotTable{
				{
					CostUsdPerHa: decimal.NewFromFloat(2000.00),
					Hectares:     decimal.NewFromFloat(10.0),
				},
			},
			mockDashboardData: &dashboardDomain.DashboardData{
				ManagementBalance: &dashboardDomain.DashboardManagementBalance{
					Summary: &dashboardDomain.DashboardBalanceSummary{
						DirectCostsExecutedUSD: decimal.NewFromFloat(18460.00),
					},
				},
			},
			mockSummaryResults: []reportDomain.SummaryResults{
				{TotalDirectCostsUsd: decimal.NewFromFloat(18460.00)},
			},
			expectedStatus:    "ERROR",
			expectedDiffA:     "-1540.00",
			expectedSystemVal: "18460.00",
			hasDiffB:          true,
			expectedDiffB:     "0.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			useCases := NewUseCases(
				NewMockWorkOrderRepositoryPort(ctrl),
				NewMockDashboardRepositoryPort(ctrl),
				NewMockLotRepositoryPort(ctrl),
				NewMockReportRepositoryPort(ctrl),
				NewMockStockRepositoryPort(ctrl),
				NewMockSupplyRepositoryPort(ctrl),
			)

			sd := &sharedData{
				lots:           tt.mockLots,
				dashboardData:  tt.mockDashboardData,
				summaryResults: tt.mockSummaryResults,
			}

			ctx := context.Background()
			result, err := useCases.control1LotesVsDashboard(ctx, sd)

			require.NoError(t, err)
			assert.Equal(t, 1, result.ControlNumber)
			assert.Equal(t, tt.expectedStatus, result.Status)
			assert.Equal(t, tt.expectedDiffA, result.DifferenceA.StringFixed(2))
			assert.Equal(t, tt.expectedSystemVal, result.SystemValue.StringFixed(2))
			assert.Equal(t, "v4_report.dashboard_management_balance", result.SystemSource)
			assert.Equal(t, "v4_report.lot_list", result.RecalcASource)

			if tt.hasDiffB {
				require.NotNil(t, result.DifferenceB)
				assert.Equal(t, tt.expectedDiffB, result.DifferenceB.StringFixed(2))
			}
		})
	}
}

func TestUseCases_control13LotesResultadoVsDashboard_withRecalcB(t *testing.T) {
	tests := []struct {
		name               string
		mockLots           []lotDomain.LotTable
		mockDashboardData  *dashboardDomain.DashboardData
		mockSummaryResults []reportDomain.SummaryResults
		expectedStatus     string
		expectedDiffA      string
		hasDiffB           bool
		expectedDiffB      string
	}{
		{
			name: "OK - 3 valores coinciden",
			mockLots: []lotDomain.LotTable{
				{
					OperatingResultPerHa: decimal.NewFromFloat(500.00),
					Hectares:             decimal.NewFromFloat(10.0),
				},
			},
			mockDashboardData: &dashboardDomain.DashboardData{
				Metrics: &dashboardDomain.DashboardMetrics{
					OperatingResult: &dashboardDomain.DashboardOperatingResult{
						ResultUSD: decimal.NewFromFloat(5000.00),
					},
				},
			},
			mockSummaryResults: []reportDomain.SummaryResults{
				{TotalOperatingResultUsd: decimal.NewFromFloat(5000.00)},
			},
			expectedStatus: "OK",
			expectedDiffA:  "0.00",
			hasDiffB:       true,
			expectedDiffB:  "0.00",
		},
		{
			name: "ERROR - RecalcB fuera de tolerancia",
			mockLots: []lotDomain.LotTable{
				{
					OperatingResultPerHa: decimal.NewFromFloat(500.00),
					Hectares:             decimal.NewFromFloat(10.0),
				},
			},
			mockDashboardData: &dashboardDomain.DashboardData{
				Metrics: &dashboardDomain.DashboardMetrics{
					OperatingResult: &dashboardDomain.DashboardOperatingResult{
						ResultUSD: decimal.NewFromFloat(5000.00),
					},
				},
			},
			mockSummaryResults: []reportDomain.SummaryResults{
				{TotalOperatingResultUsd: decimal.NewFromFloat(5005.00)},
			},
			expectedStatus: "ERROR",
			expectedDiffA:  "0.00",
			hasDiffB:       true,
			expectedDiffB:  "-5.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			useCases := NewUseCases(
				NewMockWorkOrderRepositoryPort(ctrl),
				NewMockDashboardRepositoryPort(ctrl),
				NewMockLotRepositoryPort(ctrl),
				NewMockReportRepositoryPort(ctrl),
				NewMockStockRepositoryPort(ctrl),
				NewMockSupplyRepositoryPort(ctrl),
			)

			sd := &sharedData{
				lots:           tt.mockLots,
				dashboardData:  tt.mockDashboardData,
				summaryResults: tt.mockSummaryResults,
			}

			ctx := context.Background()
			result, err := useCases.control13LotesResultadoVsDashboard(ctx, sd)

			require.NoError(t, err)
			assert.Equal(t, 13, result.ControlNumber)
			assert.Equal(t, tt.expectedStatus, result.Status)
			assert.Equal(t, tt.expectedDiffA, result.DifferenceA.StringFixed(2))

			if tt.hasDiffB {
				require.NotNil(t, result.DifferenceB)
				assert.Equal(t, tt.expectedDiffB, result.DifferenceB.StringFixed(2))
			}
		})
	}
}

func TestBuildCheck(t *testing.T) {
	tests := []struct {
		name           string
		controlNumber  int
		systemValue    decimal.Decimal
		recalcAValue   decimal.Decimal
		recalcBValue   *decimal.Decimal
		tolerance      decimal.Decimal
		expectedStatus string
		expectedDiffA  string
		hasDiffB       bool
		expectedDiffB  string
	}{
		{
			name:           "OK - diferencia cero, sin RecalcB",
			controlNumber:  1,
			systemValue:    decimal.NewFromFloat(100.00),
			recalcAValue:   decimal.NewFromFloat(100.00),
			recalcBValue:   nil,
			tolerance:      decimal.NewFromInt(1),
			expectedStatus: "OK",
			expectedDiffA:  "0.00",
			hasDiffB:       false,
		},
		{
			name:           "OK - diferencia dentro de tolerancia positiva",
			controlNumber:  2,
			systemValue:    decimal.NewFromFloat(100.50),
			recalcAValue:   decimal.NewFromFloat(100.00),
			recalcBValue:   nil,
			tolerance:      decimal.NewFromInt(1),
			expectedStatus: "OK",
			expectedDiffA:  "0.50",
			hasDiffB:       false,
		},
		{
			name:           "OK - diferencia dentro de tolerancia negativa",
			controlNumber:  3,
			systemValue:    decimal.NewFromFloat(100.00),
			recalcAValue:   decimal.NewFromFloat(100.80),
			recalcBValue:   nil,
			tolerance:      decimal.NewFromInt(1),
			expectedStatus: "OK",
			expectedDiffA:  "-0.80",
			hasDiffB:       false,
		},
		{
			name:           "OK - diferencia exacta en límite de tolerancia",
			controlNumber:  4,
			systemValue:    decimal.NewFromFloat(101.00),
			recalcAValue:   decimal.NewFromFloat(100.00),
			recalcBValue:   nil,
			tolerance:      decimal.NewFromInt(1),
			expectedStatus: "OK",
			expectedDiffA:  "1.00",
			hasDiffB:       false,
		},
		{
			name:           "ERROR - diferencia fuera de tolerancia",
			controlNumber:  5,
			systemValue:    decimal.NewFromFloat(105.00),
			recalcAValue:   decimal.NewFromFloat(100.00),
			recalcBValue:   nil,
			tolerance:      decimal.NewFromInt(1),
			expectedStatus: "ERROR",
			expectedDiffA:  "5.00",
			hasDiffB:       false,
		},
		{
			name:           "ERROR - diferencia negativa fuera de tolerancia",
			controlNumber:  6,
			systemValue:    decimal.NewFromFloat(100.00),
			recalcAValue:   decimal.NewFromFloat(105.00),
			recalcBValue:   nil,
			tolerance:      decimal.NewFromInt(1),
			expectedStatus: "ERROR",
			expectedDiffA:  "-5.00",
			hasDiffB:       false,
		},
		{
			name:           "OK - 3 valores coinciden",
			controlNumber:  1,
			systemValue:    decimal.NewFromFloat(100.00),
			recalcAValue:   decimal.NewFromFloat(100.00),
			recalcBValue:   decPtr(decimal.NewFromFloat(100.00)),
			tolerance:      decimal.NewFromInt(1),
			expectedStatus: "OK",
			expectedDiffA:  "0.00",
			hasDiffB:       true,
			expectedDiffB:  "0.00",
		},
		{
			name:           "ERROR - RecalcB fuera de tolerancia",
			controlNumber:  1,
			systemValue:    decimal.NewFromFloat(100.00),
			recalcAValue:   decimal.NewFromFloat(100.00),
			recalcBValue:   decPtr(decimal.NewFromFloat(105.00)),
			tolerance:      decimal.NewFromInt(1),
			expectedStatus: "ERROR",
			expectedDiffA:  "0.00",
			hasDiffB:       true,
			expectedDiffB:  "-5.00",
		},
		{
			name:           "ERROR - RecalcA OK pero RecalcB falla",
			controlNumber:  1,
			systemValue:    decimal.NewFromFloat(100.00),
			recalcAValue:   decimal.NewFromFloat(100.50),
			recalcBValue:   decPtr(decimal.NewFromFloat(103.00)),
			tolerance:      decimal.NewFromInt(1),
			expectedStatus: "ERROR",
			expectedDiffA:  "-0.50",
			hasDiffB:       true,
			expectedDiffB:  "-3.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var recalcBCalc *string
			var recalcBSrc *string
			var recalcBMeaning *string
			if tt.recalcBValue != nil {
				recalcBCalc = strPtr("RecalcB calc")
				recalcBSrc = strPtr("RecalcB source")
				recalcBMeaning = strPtr("RecalcB meaning")
			}

			result := buildCheck(
				tt.controlNumber,
				"Data",
				"Description",
				"Rule",
				"System calc",
				tt.systemValue,
				"System source",
				"System meaning",
				"RecalcA calc",
				tt.recalcAValue,
				"RecalcA source",
				"RecalcA meaning",
				recalcBCalc,
				tt.recalcBValue,
				recalcBSrc,
				recalcBMeaning,
				tt.tolerance,
			)

			assert.Equal(t, tt.controlNumber, result.ControlNumber)
			assert.Equal(t, tt.expectedStatus, result.Status)
			assert.Equal(t, tt.expectedDiffA, result.DifferenceA.StringFixed(2))
			assert.Equal(t, tt.systemValue, result.SystemValue)
			assert.Equal(t, tt.recalcAValue, result.RecalcAValue)

			if tt.hasDiffB {
				require.NotNil(t, result.DifferenceB)
				assert.Equal(t, tt.expectedDiffB, result.DifferenceB.StringFixed(2))
				require.NotNil(t, result.RecalcBValue)
				assert.Equal(t, *tt.recalcBValue, *result.RecalcBValue)
			} else {
				assert.Nil(t, result.DifferenceB)
				assert.Nil(t, result.RecalcBValue)
			}
		})
	}
}

func decPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}
