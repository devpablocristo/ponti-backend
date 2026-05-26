package dataintegrity

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dashboardDomain "github.com/devpablocristo/ponti-backend/internal/dashboard/usecases/domain"
	"github.com/devpablocristo/ponti-backend/internal/data-integrity/usecases/domain"
)

// dashboardWith arma un *DashboardData mínimo con los valores económicos que necesitan
// los 5 controles, evitando boilerplate en cada test.
func dashboardWith(directCosts, income, semillas, agroq, fert, structure, rent decimal.Decimal) *dashboardDomain.DashboardData {
	return &dashboardDomain.DashboardData{
		ManagementBalance: &dashboardDomain.DashboardManagementBalance{
			Summary: &dashboardDomain.DashboardBalanceSummary{
				DirectCostsExecutedUSD:     directCosts,
				IncomeUSD:                  income,
				SemillasInvertidosUSD:      semillas,
				AgroquimicosInvertidosUSD:  agroq,
				FertilizantesInvertidosUSD: fert,
				StructureExecutedUSD:       structure,
				RentExecutedUSD:            rent,
			},
		},
	}
}

func newUseCasesWithMocks(t *testing.T) (
	*UseCases,
	*MockDashboardRepositoryPort,
	*MockWorkOrderRepositoryPort,
	*MockReportRepositoryPort,
	*MockSupplyRepositoryPort,
	*MockProjectRepositoryPort,
	*MockLotRepositoryPort,
	*gomock.Controller,
) {
	t.Helper()
	ctrl := gomock.NewController(t)

	dash := NewMockDashboardRepositoryPort(ctrl)
	wo := NewMockWorkOrderRepositoryPort(ctrl)
	report := NewMockReportRepositoryPort(ctrl)
	supply := NewMockSupplyRepositoryPort(ctrl)
	project := NewMockProjectRepositoryPort(ctrl)
	lot := NewMockLotRepositoryPort(ctrl)

	uc := NewUseCases(dash, wo, report, supply, project, lot)
	return uc, dash, wo, report, supply, project, lot, ctrl
}

func dec(v float64) decimal.Decimal {
	return decimal.NewFromFloat(v)
}

// TestCheckCostsCoherence_AllOK verifica que con SSOT == RAW para los 5 controles,
// todos devuelven status OK con difference_a = 0.
func TestCheckCostsCoherence_AllOK(t *testing.T) {
	uc, dash, wo, report, supply, project, lot, ctrl := newUseCasesWithMocks(t)
	defer ctrl.Finish()

	projectID := int64(42)
	dashData := dashboardWith(
		dec(1000), // direct costs
		dec(5000), // income
		dec(200),  // semillas
		dec(150),  // agroquímicos
		dec(50),   // fertilizantes
		dec(800),  // structure (admin)
		dec(600),  // rent executed
	)

	dash.EXPECT().GetDashboard(gomock.Any(), gomock.Any()).Return(dashData, nil)
	wo.EXPECT().GetRawDirectCost(gomock.Any(), projectID).Return(dec(1000), nil)
	report.EXPECT().GetRawNetIncome(gomock.Any(), projectID).Return(dec(5000), nil)
	supply.EXPECT().GetRawSupplyInvestment(gomock.Any(), projectID).Return(dec(400), nil) // 200+150+50
	project.EXPECT().GetRawAdminCostTotal(gomock.Any(), projectID).Return(dec(800), nil)
	lot.EXPECT().GetRawLeaseExecuted(gomock.Any(), projectID).Return(dec(600), nil)

	report1, err := uc.CheckCostsCoherence(context.Background(), domain.CostsCheckFilter{ProjectID: &projectID})
	require.NoError(t, err)
	require.Len(t, report1.Checks, 5)
	for i, check := range report1.Checks {
		assert.Equal(t, checkStatusOK, check.Status, "control %d should be OK, got %s", i+1, check.Status)
		assert.Equal(t, "0.00", check.DifferenceA.StringFixed(2), "control %d diff should be 0", i+1)
		assert.Equal(t, checkTypeStrong, check.CheckType)
	}
}

// TestCheckCostsCoherence_ErrorWhenDiffExceedsTolerance verifica que diferencias > 1 USD
// fuerzan status ERROR sin warning intermedio.
func TestCheckCostsCoherence_ErrorWhenDiffExceedsTolerance(t *testing.T) {
	uc, dash, wo, report, supply, project, lot, ctrl := newUseCasesWithMocks(t)
	defer ctrl.Finish()

	projectID := int64(7)
	dashData := dashboardWith(dec(1000), dec(5000), dec(200), dec(150), dec(50), dec(800), dec(600))

	// Forzamos divergencia >1 USD en direct cost y rent. Los demás coinciden.
	dash.EXPECT().GetDashboard(gomock.Any(), gomock.Any()).Return(dashData, nil)
	wo.EXPECT().GetRawDirectCost(gomock.Any(), projectID).Return(dec(950), nil)         // 50 USD off
	report.EXPECT().GetRawNetIncome(gomock.Any(), projectID).Return(dec(5000), nil)     // OK
	supply.EXPECT().GetRawSupplyInvestment(gomock.Any(), projectID).Return(dec(400), nil)
	project.EXPECT().GetRawAdminCostTotal(gomock.Any(), projectID).Return(dec(800), nil)
	lot.EXPECT().GetRawLeaseExecuted(gomock.Any(), projectID).Return(dec(700), nil) // -100 off

	report1, err := uc.CheckCostsCoherence(context.Background(), domain.CostsCheckFilter{ProjectID: &projectID})
	require.NoError(t, err)
	require.Len(t, report1.Checks, 5)

	// Mapeamos por ControlNumber porque los goroutines no garantizan orden de finalización.
	byCtl := make(map[int]domain.IntegrityCheck, len(report1.Checks))
	for _, c := range report1.Checks {
		byCtl[c.ControlNumber] = c
	}

	assert.Equal(t, checkStatusError, byCtl[1].Status, "control 1 (direct costs) must be ERROR")
	assert.Equal(t, "50.00", byCtl[1].DifferenceA.StringFixed(2))
	assert.Equal(t, checkSeverityError, byCtl[1].Severity)

	assert.Equal(t, checkStatusOK, byCtl[2].Status, "control 2 (net income) must be OK")
	assert.Equal(t, checkStatusOK, byCtl[3].Status, "control 3 (supplies) must be OK")
	assert.Equal(t, checkStatusOK, byCtl[4].Status, "control 4 (admin) must be OK")

	assert.Equal(t, checkStatusError, byCtl[5].Status, "control 5 (rent) must be ERROR")
	assert.Equal(t, "-100.00", byCtl[5].DifferenceA.StringFixed(2))
}

// TestCheckCostsCoherence_DiffWithinTolerance verifica que una diferencia ≤ 1 USD
// (caused by numeric rounding) no rompe el control.
func TestCheckCostsCoherence_DiffWithinTolerance(t *testing.T) {
	uc, dash, wo, report, supply, project, lot, ctrl := newUseCasesWithMocks(t)
	defer ctrl.Finish()

	projectID := int64(1)
	dashData := dashboardWith(dec(1000.50), dec(5000), dec(200), dec(150), dec(50), dec(800), dec(600))

	dash.EXPECT().GetDashboard(gomock.Any(), gomock.Any()).Return(dashData, nil)
	wo.EXPECT().GetRawDirectCost(gomock.Any(), projectID).Return(dec(1000), nil) // 0.50 off, dentro
	report.EXPECT().GetRawNetIncome(gomock.Any(), projectID).Return(dec(5000), nil)
	supply.EXPECT().GetRawSupplyInvestment(gomock.Any(), projectID).Return(dec(400), nil)
	project.EXPECT().GetRawAdminCostTotal(gomock.Any(), projectID).Return(dec(800), nil)
	lot.EXPECT().GetRawLeaseExecuted(gomock.Any(), projectID).Return(dec(600), nil)

	report1, err := uc.CheckCostsCoherence(context.Background(), domain.CostsCheckFilter{ProjectID: &projectID})
	require.NoError(t, err)
	require.Len(t, report1.Checks, 5)

	byCtl := make(map[int]domain.IntegrityCheck, len(report1.Checks))
	for _, c := range report1.Checks {
		byCtl[c.ControlNumber] = c
	}
	assert.Equal(t, checkStatusOK, byCtl[1].Status)
	assert.Equal(t, "0.50", byCtl[1].DifferenceA.StringFixed(2))
}

// TestCheckCostsCoherence_RequiresProjectID verifica que sin project_id devuelve validation error
// sin tocar la base.
func TestCheckCostsCoherence_RequiresProjectID(t *testing.T) {
	uc, _, _, _, _, _, _, ctrl := newUseCasesWithMocks(t)
	defer ctrl.Finish()

	_, err := uc.CheckCostsCoherence(context.Background(), domain.CostsCheckFilter{})
	require.Error(t, err)

	zero := int64(0)
	_, err = uc.CheckCostsCoherence(context.Background(), domain.CostsCheckFilter{ProjectID: &zero})
	require.Error(t, err)
}

// TestCheckCostsCoherence_PropagatesRepoError verifica que un error de cualquier repo RAW
// se propaga arriba (cancela los demás controles y aborta).
func TestCheckCostsCoherence_PropagatesRepoError(t *testing.T) {
	uc, dash, wo, report, supply, project, lot, ctrl := newUseCasesWithMocks(t)
	defer ctrl.Finish()

	projectID := int64(99)
	dashData := dashboardWith(dec(0), dec(0), dec(0), dec(0), dec(0), dec(0), dec(0))

	dash.EXPECT().GetDashboard(gomock.Any(), gomock.Any()).Return(dashData, nil)
	wo.EXPECT().GetRawDirectCost(gomock.Any(), projectID).Return(decimal.Zero, assert.AnError).AnyTimes()
	report.EXPECT().GetRawNetIncome(gomock.Any(), projectID).Return(decimal.Zero, nil).AnyTimes()
	supply.EXPECT().GetRawSupplyInvestment(gomock.Any(), projectID).Return(decimal.Zero, nil).AnyTimes()
	project.EXPECT().GetRawAdminCostTotal(gomock.Any(), projectID).Return(decimal.Zero, nil).AnyTimes()
	lot.EXPECT().GetRawLeaseExecuted(gomock.Any(), projectID).Return(decimal.Zero, nil).AnyTimes()

	_, err := uc.CheckCostsCoherence(context.Background(), domain.CostsCheckFilter{ProjectID: &projectID})
	require.Error(t, err)
}

// TestBuildCheck cubre las dos ramas de la decisión de status (OK vs ERROR) y los bordes
// numéricos de la tolerancia.
func TestBuildCheck(t *testing.T) {
	tests := []struct {
		name            string
		systemValue     decimal.Decimal
		recalcAValue    decimal.Decimal
		expectedStatus  string
		expectedDiffA   string
		expectedSeverty string
	}{
		{"OK - diferencia cero", dec(100), dec(100), checkStatusOK, "0.00", checkSeverityInfo},
		{"OK - diferencia positiva dentro de tolerancia", dec(100.50), dec(100), checkStatusOK, "0.50", checkSeverityInfo},
		{"OK - diferencia negativa dentro de tolerancia", dec(100), dec(100.80), checkStatusOK, "-0.80", checkSeverityInfo},
		{"OK - diferencia exacta en límite", dec(101), dec(100), checkStatusOK, "1.00", checkSeverityInfo},
		{"ERROR - excede tolerancia positiva", dec(105), dec(100), checkStatusError, "5.00", checkSeverityError},
		{"ERROR - excede tolerancia negativa", dec(100), dec(105), checkStatusError, "-5.00", checkSeverityError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCheck(
				1,
				"Data", "Description", "Rule",
				"System calc", tt.systemValue, "System source", "System meaning",
				"RecalcA calc", tt.recalcAValue, "RecalcA source", "RecalcA meaning",
			)
			assert.Equal(t, tt.expectedStatus, result.Status)
			assert.Equal(t, tt.expectedSeverty, result.Severity)
			assert.Equal(t, tt.expectedDiffA, result.DifferenceA.StringFixed(2))
			assert.Equal(t, checkTypeStrong, result.CheckType)
			assert.Equal(t, defaultTolerance, result.Tolerance)
		})
	}
}
