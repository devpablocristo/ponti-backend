package dashboard

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
)

// TestDashboardHandlerGetDashboard verifica el endpoint GET /dashboard
func TestDashboardHandlerGetDashboard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Crear mock del repositorio (que es lo que realmente necesitamos)
	mockRepo := NewMockRepositoryPort(ctrl)

	// Crear los casos de uso con el mock del repositorio
	useCases := NewUseCases(mockRepo)

	// Tabla de datos de prueba para diferentes escenarios
	testCases := []struct {
		name           string
		filter         domain.DashboardFilter
		expectedResult *domain.DashboardRow
		description    string
	}{
		{
			name: "Get dashboard with project filter",
			filter: domain.DashboardFilter{
				ProjectID: int64Ptr(1),
			},
			expectedResult: &domain.DashboardRow{
				FirstOrderDate:     timePtr(time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC)),
				FirstOrderNumber:   stringPtr("0001"),
				LastOrderDate:      timePtr(time.Date(2024, 10, 15, 0, 0, 0, 0, time.UTC)),
				LastOrderNumber:    stringPtr("0050"),
				LastStockCountDate: timePtr(time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC)),
				TotalHectares:      decimal.NewFromInt(200),
				SowedArea:          decimal.NewFromInt(150),
				HarvestedArea:      decimal.NewFromInt(100),
				LaborsCostUSD:      decimal.NewFromInt(5000),
				InputsCostUSD:      decimal.NewFromInt(3000),
				ExecutedCostUSD:    decimal.NewFromInt(8000),
			},
			description: "Use cases should return dashboard data filtered by project ID",
		},
		{
			name: "Get dashboard with customer filter",
			filter: domain.DashboardFilter{
				CustomerID: int64Ptr(100),
			},
			expectedResult: &domain.DashboardRow{
				FirstOrderDate:     timePtr(time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC)),
				FirstOrderNumber:   stringPtr("0001"),
				LastOrderDate:      timePtr(time.Date(2024, 10, 15, 0, 0, 0, 0, time.UTC)),
				LastOrderNumber:    stringPtr("0050"),
				LastStockCountDate: timePtr(time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC)),
				TotalHectares:      decimal.NewFromInt(500),
				SowedArea:          decimal.NewFromInt(300),
				HarvestedArea:      decimal.NewFromInt(200),
				LaborsCostUSD:      decimal.NewFromInt(8000),
				InputsCostUSD:      decimal.NewFromInt(5000),
				ExecutedCostUSD:    decimal.NewFromInt(13000),
			},
			description: "Use cases should return dashboard data filtered by customer ID",
		},
		{
			name:   "Get dashboard without filters",
			filter: domain.DashboardFilter{},
			expectedResult: &domain.DashboardRow{
				FirstOrderDate:     timePtr(time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC)),
				FirstOrderNumber:   stringPtr("0001"),
				LastOrderDate:      timePtr(time.Date(2024, 10, 15, 0, 0, 0, 0, time.UTC)),
				LastOrderNumber:    stringPtr("0050"),
				LastStockCountDate: timePtr(time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC)),
				TotalHectares:      decimal.NewFromInt(1000),
				SowedArea:          decimal.NewFromInt(800),
				HarvestedArea:      decimal.NewFromInt(600),
				LaborsCostUSD:      decimal.NewFromInt(15000),
				InputsCostUSD:      decimal.NewFromInt(10000),
				ExecutedCostUSD:    decimal.NewFromInt(25000),
			},
			description: "Use cases should return general dashboard data without filters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Configurar expectativas del mock
			mockRepo.EXPECT().
				GetDashboard(gomock.Any(), tc.filter).
				Return(tc.expectedResult, nil).
				Times(1)

			// Ejecutar la consulta a través de los casos de uso
			result, err := useCases.GetDashboard(context.Background(), tc.filter)

			// Verificar resultados
			require.NoError(t, err, "Use cases should not return error")
			assert.NotNil(t, result, "Use cases should return result")

			// Verificar que los datos coincidan
			verifyHandlerData(t, result, tc.expectedResult)
		})
	}
}

// TestDashboardHandlerGetDashboardError verifica el manejo de errores
func TestDashboardHandlerGetDashboardError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepositoryPort(ctrl)
	useCases := NewUseCases(mockRepo)

	// Casos de error
	errorCases := []struct {
		name          string
		filter        domain.DashboardFilter
		expectedError string
		description   string
	}{
		{
			name: "Repository error",
			filter: domain.DashboardFilter{
				ProjectID: int64Ptr(1),
			},
			expectedError: "database error",
			description:   "Use cases should handle repository errors",
		},
		{
			name: "Invalid filter parameters",
			filter: domain.DashboardFilter{
				CustomerID: int64Ptr(-1), // Invalid ID
			},
			expectedError: "invalid parameters",
			description:   "Use cases should handle invalid filter parameters",
		},
	}

	for _, tc := range errorCases {
		t.Run(tc.name, func(t *testing.T) {
			// Configurar mock para retornar error
			mockRepo.EXPECT().
				GetDashboard(gomock.Any(), tc.filter).
				Return(nil, assert.AnError).
				Times(1)

			// Ejecutar la consulta
			result, err := useCases.GetDashboard(context.Background(), tc.filter)

			// Verificar que se retorne error
			assert.Error(t, err, "Use cases should return error")
			assert.Nil(t, result, "Use cases should not return result when error occurs")
		})
	}
}

// verifyHandlerData verifica que los datos del handler sean correctos
func verifyHandlerData(t *testing.T, actual, expected *domain.DashboardRow) {
	t.Run("Verify handler data integrity", func(t *testing.T) {
		// Verificar indicadores operativos
		if expected.FirstOrderDate != nil {
			assert.Equal(t, expected.FirstOrderDate, actual.FirstOrderDate)
			assert.Equal(t, expected.FirstOrderNumber, actual.FirstOrderNumber)
		}

		if expected.LastOrderDate != nil {
			assert.Equal(t, expected.LastOrderDate, actual.LastOrderDate)
			assert.Equal(t, expected.LastOrderNumber, actual.LastOrderNumber)
		}

		if expected.LastStockCountDate != nil {
			assert.Equal(t, expected.LastStockCountDate, actual.LastStockCountDate)
		}

		// Verificar métricas
		assert.Equal(t, expected.TotalHectares, actual.TotalHectares)
		assert.Equal(t, expected.SowedArea, actual.SowedArea)
		assert.Equal(t, expected.HarvestedArea, actual.HarvestedArea)
		assert.Equal(t, expected.LaborsCostUSD, actual.LaborsCostUSD)
		assert.Equal(t, expected.InputsCostUSD, actual.InputsCostUSD)
		assert.Equal(t, expected.ExecutedCostUSD, actual.ExecutedCostUSD)
	})
}

// Test utility functions are defined in test_utils.go
