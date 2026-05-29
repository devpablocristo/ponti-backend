package customer

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	models "github.com/devpablocristo/ponti-backend/internal/customer/repository/models"
	prjmodels "github.com/devpablocristo/ponti-backend/internal/project/repository/models"
)

// TestHardDeleteCustomer_SucceedsWithCascadeWhenCustomerHasProjects verifica que DeleteCustomer (hard)
// completa correctamente eliminando primero los proyectos del customer en cascada.
// Requiere TEST_DB_HOST o DB_HOST para ejecutarse.
func TestHardDeleteCustomer_SucceedsWithCascadeWhenCustomerHasProjects(t *testing.T) {
	host := getEnvOrDefault("TEST_DB_HOST", os.Getenv("DB_HOST"))
	if host == "" {
		t.Skip("Skipping integration test: TEST_DB_HOST or DB_HOST not set")
	}

	port := 5432
	if p := getEnvOrDefault("TEST_DB_PORT", os.Getenv("DB_PORT")); p != "" {
		port, _ = strconv.Atoi(p)
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		host,
		getEnvOrDefault("TEST_DB_USER", os.Getenv("DB_USER")),
		getEnvOrDefault("TEST_DB_PASSWORD", os.Getenv("DB_PASSWORD")),
		getEnvOrDefault("TEST_DB_NAME", os.Getenv("DB_NAME")),
		port,
		getEnvOrDefault("TEST_DB_SSL_MODE", os.Getenv("DB_SSL_MODE")),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	// GormEnginePort minimal: solo necesita Client()
	engine := &gormEngineAdapter{client: db}
	repo := NewRepository(engine)

	ctx := context.Background()

	// Crear customer y project para provocar FK al intentar hard delete
	tx := db.Begin()
	defer tx.Rollback()

	// Customer
	cust := &models.Customer{
		Name: fmt.Sprintf("test_harddelete_%d", time.Now().UnixNano()),
	}
	require.NoError(t, tx.Create(cust).Error)
	customerID := cust.ID

	// Campaign (projects requieren campaign_id)
	var campaign prjmodels.Campaign
	if err := tx.Table("campaigns").Where("deleted_at IS NULL").First(&campaign).Error; err != nil {
		cn := fmt.Sprintf("test_campaign_%d", time.Now().UnixNano())
		require.NoError(t, tx.Exec("INSERT INTO campaigns (name, created_at, updated_at) VALUES ($1, now(), now())", cn).Error)
		require.NoError(t, tx.Table("campaigns").Where("name = ?", cn).First(&campaign).Error)
	}

	// Project que referencia al customer
	proj := &prjmodels.Project{
		Name:        fmt.Sprintf("test_proj_%d", time.Now().UnixNano()),
		CustomerID:  customerID,
		CampaignID:  campaign.ID,
		AdminCost:   decimal.Zero,
		PlannedCost: decimal.Zero,
	}
	require.NoError(t, tx.Create(proj).Error)

	require.NoError(t, tx.Commit().Error)

	// Act: DeleteCustomer (hard) debe completar (cascade delete de proyectos primero)
	err = repo.DeleteCustomer(ctx, customerID)

	// Assert: debe completar sin error
	require.NoError(t, err)

	// Verificar que el customer fue eliminado
	var count int64
	db.Unscoped().Model(&models.Customer{}).Where("id = ?", customerID).Count(&count)
	assert.Equal(t, int64(0), count, "customer debe estar eliminado")
}

type gormEngineAdapter struct {
	client *gorm.DB
}

func (a *gormEngineAdapter) Client() *gorm.DB {
	return a.client
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
