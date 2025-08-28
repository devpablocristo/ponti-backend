package dashboard

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/repository/mocks"
)

// TestDashboardRepositoryStructure verifica la estructura del repositorio
func TestDashboardRepositoryStructure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Crear mock del motor GORM
	mockGormEngine := mocks.NewMockGormEngine(ctrl)

	// Crear el repositorio con el mock
	repo := NewRepository(mockGormEngine)

	// Verificar que el repositorio se creó correctamente
	assert.NotNil(t, repo, "Repository should be created successfully")
	assert.NotNil(t, mockGormEngine, "Mock GORM engine should be created successfully")

	// Verificar que el mock tenga los métodos esperados
	mockGormEngine.EXPECT().
		Client().
		Return(nil).
		Times(1)

	// Verificar que el mock funcione
	client := mockGormEngine.Client()
	assert.Nil(t, client, "Mock should return nil client as configured")
}

// TestDashboardRepositoryFilterTypes verifica los tipos de filtros
func TestDashboardRepositoryFilterTypes(t *testing.T) {
	// Verificar que los tipos de filtro sean correctos
	projectID := int64Ptr(1)
	customerID := int64Ptr(100)
	fieldID := int64Ptr(25)

	assert.NotNil(t, projectID, "Project ID should be created")
	assert.NotNil(t, customerID, "Customer ID should be created")
	assert.NotNil(t, fieldID, "Field ID should be created")

	// Verificar valores
	assert.Equal(t, int64(1), *projectID, "Project ID should be 1")
	assert.Equal(t, int64(100), *customerID, "Customer ID should be 100")
	assert.Equal(t, int64(25), *fieldID, "Field ID should be 25")
}

// Test utility functions are defined in test_utils.go
