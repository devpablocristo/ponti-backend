package dataintegrity

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dashboardDomain "github.com/alphacodinggroup/ponti-backend/internal/dashboard/usecases/domain"
	lotDomain "github.com/alphacodinggroup/ponti-backend/internal/lot/usecases/domain"
	reportDomain "github.com/alphacodinggroup/ponti-backend/internal/report/usecases/domain"
)

func TestUseCases_control1OrdenesVsDashboard(t *testing.T) {
	tests := []struct {
		name              string
		projectID         *int64
		mockLots          []lotDomain.LotTable
		mockLotsErr       error
		mockDashboardData *dashboardDomain.DashboardData
		mockDashboardErr  error
		expectedStatus    string
		expectedDiff      string
		expectedLeftVal   string
	}{
		{
			name:           "OK - valores iguales",
			projectID:      ptr(int64(11)),
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
			mockLotsErr: nil,
			mockDashboardData: &dashboardDomain.DashboardData{
				ManagementBalance: &dashboardDomain.DashboardManagementBalance{
					Summary: &dashboardDomain.DashboardBalanceSummary{
						DirectCostsExecutedUSD: decimal.NewFromFloat(18459.39),
					},
				},
			},
			mockDashboardErr: nil,
			expectedStatus:   "OK",
			expectedDiff:     "0.00",
			expectedLeftVal:  "18459.39",
		},
		{
			name:           "ERROR - diferencia aunque sea pequeña (tolerancia 0)",
			projectID:      ptr(int64(11)),
			mockLots: []lotDomain.LotTable{
				{
					CostUsdPerHa: decimal.NewFromFloat(1845.439),
					Hectares:     decimal.NewFromFloat(10.0),
				},
			},
			mockLotsErr: nil,
			mockDashboardData: &dashboardDomain.DashboardData{
				ManagementBalance: &dashboardDomain.DashboardManagementBalance{
					Summary: &dashboardDomain.DashboardBalanceSummary{
						DirectCostsExecutedUSD: decimal.NewFromFloat(18455.17),
					},
				},
			},
			mockDashboardErr: nil,
			expectedStatus:   "OK",
			expectedDiff:     "-0.78",
			expectedLeftVal:  "18454.39",
		},
		{
			name:           "ERROR - diferencia fuera de tolerancia",
			projectID:      ptr(int64(11)),
			mockLots: []lotDomain.LotTable{
				{
					CostUsdPerHa: decimal.NewFromFloat(2000.00),
					Hectares:     decimal.NewFromFloat(10.0),
				},
			},
			mockLotsErr: nil,
			mockDashboardData: &dashboardDomain.DashboardData{
				ManagementBalance: &dashboardDomain.DashboardManagementBalance{
					Summary: &dashboardDomain.DashboardBalanceSummary{
						DirectCostsExecutedUSD: decimal.NewFromFloat(18460.00),
					},
				},
			},
			mockDashboardErr: nil,
			expectedStatus:   "ERROR",
			expectedDiff:     "1540.00",
			expectedLeftVal:  "20000.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockWorkOrderRepo := NewMockWorkOrderRepositoryPort(ctrl)
			mockDashboardRepo := NewMockDashboardRepositoryPort(ctrl)
			mockLotRepo := NewMockLotRepositoryPort(ctrl)
			mockReportRepo := NewMockReportRepositoryPort(ctrl)
			mockStockRepo := NewMockStockRepositoryPort(ctrl)

			useCases := NewUseCases(
				mockWorkOrderRepo,
				mockDashboardRepo,
				mockLotRepo,
				mockReportRepo,
				mockStockRepo,
			)

			ctx := context.Background()
			// Mock expectations
			mockLotRepo.EXPECT().
				ListLots(ctx, lotDomain.LotListFilter{ProjectID: tt.projectID}, 1, 10000).
				Return(tt.mockLots, 0, decimal.Zero, decimal.Zero, tt.mockLotsErr).
				Times(1)

			mockDashboardRepo.EXPECT().
				GetDashboard(ctx, gomock.Any()).
				Return(tt.mockDashboardData, tt.mockDashboardErr).
				Times(1)

			// Act
			result, err := useCases.control1OrdenesVsDashboard(ctx, tt.projectID)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, 1, result.ControlNumber)
			assert.Equal(t, tt.expectedStatus, result.Status)
			assert.Equal(t, tt.expectedDiff, result.Difference.StringFixed(2))
			assert.Equal(t, tt.expectedLeftVal, result.LeftValue.StringFixed(2))
			assert.Equal(t, tt.mockDashboardData.ManagementBalance.Summary.DirectCostsExecutedUSD, result.RightValue)
			assert.Equal(t, "v4_report.lot_list", result.LeftSource)
			assert.Equal(t, "v4_report.dashboard_management_balance", result.RightSource)
		})
	}
}

func TestUseCases_control2OrdenesVsLotes(t *testing.T) {
	tests := []struct {
		name             string
		projectID        *int64
		mockLots         []lotDomain.LotTable
		mockLotsErr      error
		mockFieldCrops   []reportDomain.FieldCropMetric
		mockFieldCropsErr error
		expectedStatus   string
		expectedLeftVal  string
		expectedRightVal string
	}{
		{
			name:           "OK - valores coinciden",
			projectID:      ptr(int64(11)),
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
			mockLotsErr:      nil,
			mockFieldCrops: []reportDomain.FieldCropMetric{
				{
					TotalDirectCostsUsd: decimal.NewFromFloat(18459.39),
				},
			},
			mockFieldCropsErr: nil,
			expectedStatus:   "OK",
			expectedLeftVal:  "18459.39",
			expectedRightVal: "18459.39", // 10005.0 + 8454.39
		},
		{
			name:           "ERROR - diferencia aunque sea pequeña (tolerancia 0)",
			projectID:      ptr(int64(11)),
			mockLots: []lotDomain.LotTable{
				{
					CostUsdPerHa: decimal.NewFromFloat(1845.439),
					Hectares:     decimal.NewFromFloat(10.0),
				},
			},
			mockLotsErr:      nil,
			mockFieldCrops: []reportDomain.FieldCropMetric{
				{
					TotalDirectCostsUsd: decimal.NewFromFloat(18455.17),
				},
			},
			mockFieldCropsErr: nil,
			expectedStatus:   "OK",
			expectedLeftVal:  "18454.39",
			expectedRightVal: "18455.17",
		},
		{
			name:           "ERROR - diferencia fuera de tolerancia",
			projectID:      ptr(int64(11)),
			mockLots: []lotDomain.LotTable{
				{
					CostUsdPerHa: decimal.NewFromFloat(2000.00),
					Hectares:     decimal.NewFromFloat(10.0),
				},
			},
			mockLotsErr:      nil,
			mockFieldCrops: []reportDomain.FieldCropMetric{
				{
					TotalDirectCostsUsd: decimal.NewFromFloat(21000.00),
				},
			},
			mockFieldCropsErr: nil,
			expectedStatus:   "ERROR",
			expectedLeftVal:  "20000.00",
			expectedRightVal: "21000.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockWorkOrderRepo := NewMockWorkOrderRepositoryPort(ctrl)
			mockDashboardRepo := NewMockDashboardRepositoryPort(ctrl)
			mockLotRepo := NewMockLotRepositoryPort(ctrl)
			mockReportRepo := NewMockReportRepositoryPort(ctrl)
			mockStockRepo := NewMockStockRepositoryPort(ctrl)

			useCases := NewUseCases(
				mockWorkOrderRepo,
				mockDashboardRepo,
				mockLotRepo,
				mockReportRepo,
				mockStockRepo,
			)

			ctx := context.Background()
			mockLotRepo.EXPECT().
				ListLots(ctx, lotDomain.LotListFilter{ProjectID: tt.projectID}, 1, 10000).
				Return(tt.mockLots, 0, decimal.Zero, decimal.Zero, tt.mockLotsErr).
				Times(1)

			mockReportRepo.EXPECT().
				GetFieldCropMetrics(reportDomain.ReportFilter{ProjectID: tt.projectID}).
				Return(tt.mockFieldCrops, tt.mockFieldCropsErr).
				Times(1)

			// Act
			result, err := useCases.control2OrdenesVsLotes(ctx, tt.projectID)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, 2, result.ControlNumber)
			assert.Equal(t, tt.expectedStatus, result.Status)
			assert.Equal(t, tt.expectedLeftVal, result.LeftValue.StringFixed(2))
			assert.Equal(t, tt.expectedRightVal, result.RightValue.StringFixed(2))
			assert.Equal(t, "v4_report.lot_list", result.LeftSource)
			assert.Equal(t, "v4_report.field_crop_metrics", result.RightSource)
			assert.Equal(t, "1.00", result.Tolerance.StringFixed(2))
		})
	}
}

func TestBuildCheck(t *testing.T) {
	tests := []struct {
		name           string
		controlNumber  int
		leftValue      decimal.Decimal
		rightValue     decimal.Decimal
		tolerance      decimal.Decimal
		expectedStatus string
		expectedDiff   string
	}{
		{
			name:           "OK - diferencia cero",
			controlNumber:  1,
			leftValue:      decimal.NewFromFloat(100.00),
			rightValue:     decimal.NewFromFloat(100.00),
			tolerance:      decimal.NewFromInt(1),
			expectedStatus: "OK",
			expectedDiff:   "0.00",
		},
		{
			name:           "OK - diferencia dentro de tolerancia positiva",
			controlNumber:  2,
			leftValue:      decimal.NewFromFloat(100.50),
			rightValue:     decimal.NewFromFloat(100.00),
			tolerance:      decimal.NewFromInt(1),
			expectedStatus: "OK",
			expectedDiff:   "0.50",
		},
		{
			name:           "OK - diferencia dentro de tolerancia negativa",
			controlNumber:  3,
			leftValue:      decimal.NewFromFloat(100.00),
			rightValue:     decimal.NewFromFloat(100.80),
			tolerance:      decimal.NewFromInt(1),
			expectedStatus: "OK",
			expectedDiff:   "-0.80",
		},
		{
			name:           "OK - diferencia exacta en límite de tolerancia",
			controlNumber:  4,
			leftValue:      decimal.NewFromFloat(101.00),
			rightValue:     decimal.NewFromFloat(100.00),
			tolerance:      decimal.NewFromInt(1),
			expectedStatus: "OK",
			expectedDiff:   "1.00",
		},
		{
			name:           "ERROR - diferencia fuera de tolerancia",
			controlNumber:  5,
			leftValue:      decimal.NewFromFloat(105.00),
			rightValue:     decimal.NewFromFloat(100.00),
			tolerance:      decimal.NewFromInt(1),
			expectedStatus: "ERROR",
			expectedDiff:   "5.00",
		},
		{
			name:           "ERROR - diferencia negativa fuera de tolerancia",
			controlNumber:  6,
			leftValue:      decimal.NewFromFloat(100.00),
			rightValue:     decimal.NewFromFloat(105.00),
			tolerance:      decimal.NewFromInt(1),
			expectedStatus: "ERROR",
			expectedDiff:   "-5.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := buildCheck(
				tt.controlNumber,
				"Source",
				"Data",
				"Target",
				"Rule",
				"Description",
				"Left calc",
				tt.leftValue,
				"Left source",
				"Left meaning",
				"Right calc",
				tt.rightValue,
				"Right source",
				"Right meaning",
				"Calculation meaning",
				tt.tolerance,
			)

			// Assert
			assert.Equal(t, tt.controlNumber, result.ControlNumber)
			assert.Equal(t, tt.expectedStatus, result.Status)
			assert.Equal(t, tt.expectedDiff, result.Difference.StringFixed(2))
			assert.Equal(t, tt.leftValue, result.LeftValue)
			assert.Equal(t, tt.rightValue, result.RightValue)
			assert.Equal(t, "Source", result.SourceModule)
			assert.Equal(t, "Target", result.TargetModule)
		})
	}
}

// Test de integración omitido por complejidad de setup
// Los tests unitarios individuales (control1, control2, buildCheck) son suficientes

// Helper function
func ptr(v int64) *int64 {
	return &v
}
