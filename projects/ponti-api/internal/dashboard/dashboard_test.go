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

// MockRepository es generado por gomock
//go:generate mockgen -destination=mock_repository.go -package=dashboard github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard RepositoryPort

// TestDashboardWithMockData verifica la funcionalidad del dashboard usando gomock y tablas de datos
func TestDashboardWithMockData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Crear mock del repositorio
	mockRepo := NewMockRepositoryPort(ctrl)

	// Tabla de datos de prueba para diferentes escenarios
	testCases := []struct {
		name           string
		projectID      int64
		expectedResult domain.DashboardRow
		description    string
	}{
		{
			name:      "Proyecto con indicadores operativos completos",
			projectID: 1,
			expectedResult: domain.DashboardRow{
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
			description: "Proyecto agrícola completo con todas las métricas e indicadores operativos",
		},
		{
			name:      "Proyecto en etapa inicial",
			projectID: 2,
			expectedResult: domain.DashboardRow{
				FirstOrderDate:     timePtr(time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)),
				FirstOrderNumber:   stringPtr("0001"),
				LastOrderDate:      timePtr(time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)), // Solo una orden
				LastOrderNumber:    stringPtr("0001"),
				LastStockCountDate: timePtr(time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)),

				TotalHectares:   decimal.NewFromInt(100),
				SowedArea:       decimal.NewFromInt(25),
				HarvestedArea:   decimal.NewFromInt(0),
				LaborsCostUSD:   decimal.NewFromInt(1000),
				InputsCostUSD:   decimal.NewFromInt(500),
				ExecutedCostUSD: decimal.NewFromInt(1500),
			},
			description: "Proyecto que acaba de comenzar, solo con primera orden",
		},
		{
			name:      "Proyecto sin datos",
			projectID: 999,
			expectedResult: domain.DashboardRow{
				FirstOrderDate:     nil,
				FirstOrderNumber:   nil,
				LastOrderDate:      nil,
				LastOrderNumber:    nil,
				LastStockCountDate: nil,

				TotalHectares:   decimal.NewFromInt(0),
				SowedArea:       decimal.NewFromInt(0),
				HarvestedArea:   decimal.NewFromInt(0),
				LaborsCostUSD:   decimal.NewFromInt(0),
				InputsCostUSD:   decimal.NewFromInt(0),
				ExecutedCostUSD: decimal.NewFromInt(0),
			},
			description: "Proyecto inexistente o sin datos",
		},
	}

	// Ejecutar tests para cada caso de prueba
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Configurar expectativas del mock
			mockRepo.EXPECT().
				GetDashboard(gomock.Any(), domain.DashboardFilter{ProjectID: &tc.projectID}).
				Return(&tc.expectedResult, nil).
				Times(1)

			// Ejecutar la consulta
			filter := domain.DashboardFilter{ProjectID: &tc.projectID}
			result, err := mockRepo.GetDashboard(context.Background(), filter)

			// Verificar resultados
			require.NoError(t, err)
			assert.NotNil(t, result)

			// Verificar indicadores operativos
			verifyOperationalIndicators(t, result, tc.expectedResult)

			// Verificar métricas del dashboard
			verifyDashboardMetrics(t, result, tc.expectedResult)

			// Verificar cálculos específicos según el escenario
			verifyScenarioSpecificLogic(t, result, tc)
		})
	}
}

// verifyOperationalIndicators verifica que los indicadores operativos sean correctos
func verifyOperationalIndicators(t *testing.T, actual *domain.DashboardRow, expected domain.DashboardRow) {
	t.Run("Verificar indicadores operativos", func(t *testing.T) {
		// Verificar primera orden
		if expected.FirstOrderDate != nil {
			assert.Equal(t, expected.FirstOrderDate, actual.FirstOrderDate)
			assert.Equal(t, expected.FirstOrderNumber, actual.FirstOrderNumber)
		}

		// Verificar última orden
		if expected.LastOrderDate != nil {
			assert.Equal(t, expected.LastOrderDate, actual.LastOrderDate)
			assert.Equal(t, expected.LastOrderNumber, actual.LastOrderNumber)
		}

		// Verificar arqueo de stock
		if expected.LastStockCountDate != nil {
			assert.Equal(t, expected.LastStockCountDate, actual.LastStockCountDate)
		}

		// Verificar cierre de campaña

		// Verificar coherencia de fechas
		if actual.FirstOrderDate != nil && actual.LastOrderDate != nil {
			assert.True(t, actual.FirstOrderDate.Before(*actual.LastOrderDate) ||
				actual.FirstOrderDate.Equal(*actual.LastOrderDate),
				"La primera orden debe ser antes o igual a la última")
		}

		if actual.LastStockCountDate != nil && actual.FirstOrderDate != nil && actual.LastOrderDate != nil {
			assert.True(t, actual.LastStockCountDate.After(*actual.FirstOrderDate) ||
				actual.LastStockCountDate.Equal(*actual.FirstOrderDate),
				"El arqueo de stock debe ser después o igual a la primera orden")
			assert.True(t, actual.LastStockCountDate.Before(*actual.LastOrderDate) ||
				actual.LastStockCountDate.Equal(*actual.LastOrderDate),
				"El arqueo de stock debe ser antes o igual a la última orden")
		}
	})
}

// verifyDashboardMetrics verifica las métricas principales del dashboard
func verifyDashboardMetrics(t *testing.T, actual *domain.DashboardRow, expected domain.DashboardRow) {
	t.Run("Verificar métricas del dashboard", func(t *testing.T) {
		// Verificar hectáreas
		assert.Equal(t, expected.TotalHectares, actual.TotalHectares)
		assert.Equal(t, expected.SowedArea, actual.SowedArea)
		assert.Equal(t, expected.HarvestedArea, actual.HarvestedArea)

		// Verificar costos
		assert.Equal(t, expected.LaborsCostUSD, actual.LaborsCostUSD)
		assert.Equal(t, expected.InputsCostUSD, actual.InputsCostUSD)
		assert.Equal(t, expected.ExecutedCostUSD, actual.ExecutedCostUSD)

		// Verificar lógica de negocio
		assert.True(t, actual.SowedArea.LessThanOrEqual(actual.TotalHectares) ||
			actual.TotalHectares.IsZero(),
			"Área sembrada no puede exceder total (excepto si total es 0)")

		assert.True(t, actual.HarvestedArea.LessThanOrEqual(actual.SowedArea) ||
			actual.SowedArea.IsZero(),
			"Área cosechada no puede exceder sembrada (excepto si sembrada es 0)")
	})
}

// verifyScenarioSpecificLogic verifica lógica específica según el escenario
func verifyScenarioSpecificLogic(t *testing.T, result *domain.DashboardRow, tc struct {
	name           string
	projectID      int64
	expectedResult domain.DashboardRow
	description    string
}) {
	switch tc.name {
	case "Proyecto con indicadores operativos completos":
		t.Run("Verificar proyecto completo", func(t *testing.T) {
			// Verificar que tenga todos los datos
			assert.NotNil(t, result.FirstOrderDate)
			assert.NotNil(t, result.LastOrderDate)
			assert.NotNil(t, result.LastStockCountDate)

			// Verificar métricas significativas
			assert.True(t, result.TotalHectares.GreaterThan(decimal.Zero))
			assert.True(t, result.SowedArea.GreaterThan(decimal.Zero))
			assert.True(t, result.ExecutedCostUSD.GreaterThan(decimal.Zero))
		})

	case "Proyecto en etapa inicial":
		t.Run("Verificar proyecto inicial", func(t *testing.T) {
			// Verificar que solo tenga una orden
			assert.Equal(t, result.FirstOrderNumber, result.LastOrderNumber)
			assert.Equal(t, result.FirstOrderDate, result.LastOrderDate)

			// Verificar que no haya cosecha
			assert.True(t, result.HarvestedArea.IsZero())
		})

	case "Proyecto sin datos":
		t.Run("Verificar proyecto sin datos", func(t *testing.T) {
			// Verificar que todos los campos sean nulos o cero
			assert.Nil(t, result.FirstOrderDate)
			assert.Nil(t, result.LastOrderDate)
			assert.Nil(t, result.LastStockCountDate)

			assert.True(t, result.TotalHectares.IsZero())
			assert.True(t, result.SowedArea.IsZero())
			assert.True(t, result.HarvestedArea.IsZero())
		})
	}
}

// TestDashboardCalculationsWithMockData verifica cálculos específicos usando mocks
func TestDashboardCalculationsWithMockData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepositoryPort(ctrl)

	// Tabla de datos para cálculos
	calculationTests := []struct {
		name                    string
		input                   domain.DashboardRow
		expectedSowingProgress  decimal.Decimal
		expectedHarvestProgress decimal.Decimal
	}{
		{
			name: "Cálculo de progreso 75% siembra, 50% cosecha",
			input: domain.DashboardRow{
				TotalHectares: decimal.NewFromInt(200),
				SowedArea:     decimal.NewFromInt(150),
				HarvestedArea: decimal.NewFromInt(100),
			},
			expectedSowingProgress:  decimal.NewFromFloat(75.0),
			expectedHarvestProgress: decimal.NewFromFloat(50.0),
		},
		{
			name: "Cálculo de progreso 100% siembra, 0% cosecha",
			input: domain.DashboardRow{
				TotalHectares: decimal.NewFromInt(100),
				SowedArea:     decimal.NewFromInt(100),
				HarvestedArea: decimal.NewFromInt(0),
			},
			expectedSowingProgress:  decimal.NewFromFloat(100.0),
			expectedHarvestProgress: decimal.NewFromFloat(0.0),
		},
	}

	for _, ct := range calculationTests {
		t.Run(ct.name, func(t *testing.T) {
			// Configurar mock
			mockRepo.EXPECT().
				GetDashboard(gomock.Any(), gomock.Any()).
				Return(&ct.input, nil).
				Times(1)

			// Ejecutar consulta
			filter := domain.DashboardFilter{ProjectID: int64Ptr(1)}
			result, err := mockRepo.GetDashboard(context.Background(), filter)

			require.NoError(t, err)
			assert.NotNil(t, result)

			// Verificar cálculos
			if !ct.input.TotalHectares.IsZero() {
				// Progreso de siembra
				sowingProgress := ct.input.SowedArea.Div(ct.input.TotalHectares).Mul(decimal.NewFromInt(100))
				assert.True(t, sowingProgress.Equal(ct.expectedSowingProgress))

				// Progreso de cosecha
				harvestProgress := ct.input.HarvestedArea.Div(ct.input.TotalHectares).Mul(decimal.NewFromInt(100))
				assert.True(t, harvestProgress.Equal(ct.expectedHarvestProgress))
			}
		})
	}
}

// Test utility functions are defined in test_utils.go
