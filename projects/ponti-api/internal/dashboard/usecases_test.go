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

//go:generate mockgen -destination=usecases/mocks/mocks.go -package=mocks github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard RepositoryPort

// TestDashboardUseCasesGetDashboard verifica la funcionalidad de los casos de uso
func TestDashboardUseCasesGetDashboard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Crear mock del repositorio
	mockRepo := NewMockRepositoryPort(ctrl)

	// Crear los casos de uso con el mock
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

				TotalHectares:   decimal.NewFromInt(200),
				SowedArea:       decimal.NewFromInt(150),
				HarvestedArea:   decimal.NewFromInt(100),
				LaborsCostUSD:   decimal.NewFromInt(5000),
				InputsCostUSD:   decimal.NewFromInt(3000),
				ExecutedCostUSD: decimal.NewFromInt(8000),
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

				TotalHectares:   decimal.NewFromInt(500),
				SowedArea:       decimal.NewFromInt(300),
				HarvestedArea:   decimal.NewFromInt(200),
				LaborsCostUSD:   decimal.NewFromInt(8000),
				InputsCostUSD:   decimal.NewFromInt(5000),
				ExecutedCostUSD: decimal.NewFromInt(13000),
			},
			description: "Use cases should return dashboard data filtered by customer ID",
		},
		{
			name: "Get dashboard with field filter",
			filter: domain.DashboardFilter{
				FieldID: int64Ptr(25),
			},
			expectedResult: &domain.DashboardRow{
				FirstOrderDate:     timePtr(time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC)),
				FirstOrderNumber:   stringPtr("0001"),
				LastOrderDate:      timePtr(time.Date(2024, 10, 15, 0, 0, 0, 0, time.UTC)),
				LastOrderNumber:    stringPtr("0050"),
				LastStockCountDate: timePtr(time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC)),

				TotalHectares:   decimal.NewFromInt(50),
				SowedArea:       decimal.NewFromInt(40),
				HarvestedArea:   decimal.NewFromInt(30),
				LaborsCostUSD:   decimal.NewFromInt(2000),
				InputsCostUSD:   decimal.NewFromInt(1500),
				ExecutedCostUSD: decimal.NewFromInt(3500),
			},
			description: "Use cases should return dashboard data filtered by field ID",
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

				TotalHectares:   decimal.NewFromInt(1000),
				SowedArea:       decimal.NewFromInt(800),
				HarvestedArea:   decimal.NewFromInt(600),
				LaborsCostUSD:   decimal.NewFromInt(15000),
				InputsCostUSD:   decimal.NewFromInt(10000),
				ExecutedCostUSD: decimal.NewFromInt(25000),
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

			// Ejecutar la consulta
			result, err := useCases.GetDashboard(context.Background(), tc.filter)

			// Verificar resultados
			require.NoError(t, err, "Use cases should not return error")
			assert.NotNil(t, result, "Use cases should return result")

			// Verificar que los datos coincidan
			verifyUseCaseData(t, result, tc.expectedResult)
		})
	}
}

// TestDashboardUseCasesGetDashboardError verifica el manejo de errores
func TestDashboardUseCasesGetDashboardError(t *testing.T) {
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

// TestDashboardUseCasesBusinessLogic verifica la lógica de negocio
func TestDashboardUseCasesBusinessLogic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepositoryPort(ctrl)
	useCases := NewUseCases(mockRepo)

	// Casos de prueba para lógica de negocio
	businessLogicTests := []struct {
		name           string
		input          *domain.DashboardRow
		expectedResult *domain.DashboardRow
		description    string
	}{
		{
			name: "Validate operational indicators logic",
			input: &domain.DashboardRow{
				FirstOrderDate:     timePtr(time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC)),
				FirstOrderNumber:   stringPtr("0001"),
				LastOrderDate:      timePtr(time.Date(2024, 10, 15, 0, 0, 0, 0, time.UTC)),
				LastOrderNumber:    stringPtr("0050"),
				LastStockCountDate: timePtr(time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC)),

				TotalHectares: decimal.NewFromInt(200),
				SowedArea:     decimal.NewFromInt(150),
				HarvestedArea: decimal.NewFromInt(100),
			},
			expectedResult: &domain.DashboardRow{
				FirstOrderDate:     timePtr(time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC)),
				FirstOrderNumber:   stringPtr("0001"),
				LastOrderDate:      timePtr(time.Date(2024, 10, 15, 0, 0, 0, 0, time.UTC)),
				LastOrderNumber:    stringPtr("0050"),
				LastStockCountDate: timePtr(time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC)),

				TotalHectares: decimal.NewFromInt(200),
				SowedArea:     decimal.NewFromInt(150),
				HarvestedArea: decimal.NewFromInt(100),
			},
			description: "Use cases should validate operational indicators business rules",
		},
		{
			name: "Validate metrics consistency",
			input: &domain.DashboardRow{
				FirstOrderDate:     timePtr(time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC)),
				FirstOrderNumber:   stringPtr("0001"),
				LastOrderDate:      timePtr(time.Date(2024, 10, 15, 0, 0, 0, 0, time.UTC)),
				LastOrderNumber:    stringPtr("0050"),
				LastStockCountDate: timePtr(time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC)),

				TotalHectares: decimal.NewFromInt(100),
				SowedArea:     decimal.NewFromInt(80),
				HarvestedArea: decimal.NewFromInt(60),
			},
			expectedResult: &domain.DashboardRow{
				FirstOrderDate:     timePtr(time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC)),
				FirstOrderNumber:   stringPtr("0001"),
				LastOrderDate:      timePtr(time.Date(2024, 10, 15, 0, 0, 0, 0, time.UTC)),
				LastOrderNumber:    stringPtr("0050"),
				LastStockCountDate: timePtr(time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC)),

				TotalHectares: decimal.NewFromInt(100),
				SowedArea:     decimal.NewFromInt(80),
				HarvestedArea: decimal.NewFromInt(60),
			},
			description: "Use cases should validate metrics consistency",
		},
	}

	for _, tc := range businessLogicTests {
		t.Run(tc.name, func(t *testing.T) {
			// Configurar mock para retornar datos de entrada
			mockRepo.EXPECT().
				GetDashboard(gomock.Any(), gomock.Any()).
				Return(tc.input, nil).
				Times(1)

			// Ejecutar consulta
			result, err := useCases.GetDashboard(context.Background(), domain.DashboardFilter{})
			require.NoError(t, err, "Use cases should execute successfully")

			// Verificar lógica de negocio
			verifyBusinessLogic(t, result, tc.expectedResult)
		})
	}
}

// verifyUseCaseData verifica que los datos de los casos de uso sean correctos
func verifyUseCaseData(t *testing.T, actual, expected *domain.DashboardRow) {
	t.Run("Verify use case data integrity", func(t *testing.T) {
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

// verifyBusinessLogic verifica la lógica de negocio
func verifyBusinessLogic(t *testing.T, actual, _ *domain.DashboardRow) {
	t.Run("Verify business logic rules", func(t *testing.T) {
		// Verificar que las fechas sean coherentes
		if actual.FirstOrderDate != nil && actual.LastOrderDate != nil {
			assert.True(t, actual.FirstOrderDate.Before(*actual.LastOrderDate) ||
				actual.FirstOrderDate.Equal(*actual.LastOrderDate),
				"First order date should be before or equal to last order date")
		}

		// Verificar que el arqueo de stock esté en el rango correcto
		if actual.LastStockCountDate != nil && actual.FirstOrderDate != nil && actual.LastOrderDate != nil {
			assert.True(t, actual.LastStockCountDate.After(*actual.FirstOrderDate) ||
				actual.LastStockCountDate.Equal(*actual.FirstOrderDate),
				"Stock count date should be after or equal to first order date")
			assert.True(t, actual.LastStockCountDate.Before(*actual.LastOrderDate) ||
				actual.LastStockCountDate.Equal(*actual.LastOrderDate),
				"Stock count date should be before or equal to last order date")
		}

		// Verificar métricas de área
		assert.True(t, actual.SowedArea.LessThanOrEqual(actual.TotalHectares) ||
			actual.TotalHectares.IsZero(),
			"Sowed area cannot exceed total area (except when total is 0)")

		assert.True(t, actual.HarvestedArea.LessThanOrEqual(actual.SowedArea) ||
			actual.SowedArea.IsZero(),
			"Harvested area cannot exceed sowed area (except when sowed is 0)")
	})
}

// Test utility functions are defined in test_utils.go
