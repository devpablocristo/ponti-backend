package supply

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	categorymodels "github.com/devpablocristo/ponti-backend/internal/category/repository/models"
	classtypemodels "github.com/devpablocristo/ponti-backend/internal/class-type/repository/models"
	investormodels "github.com/devpablocristo/ponti-backend/internal/investor/repository/models"
	projectmodels "github.com/devpablocristo/ponti-backend/internal/project/repository/models"
	providermodels "github.com/devpablocristo/ponti-backend/internal/provider/repository/models"
	stockmodels "github.com/devpablocristo/ponti-backend/internal/stock/repository/models"
	models "github.com/devpablocristo/ponti-backend/internal/supply/repository/models"
)

func TestIsForeignKeyViolation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "postgres foreign key violation",
			err:  &pgconn.PgError{Code: "23503"},
			want: true,
		},
		{
			name: "wrapped postgres foreign key violation",
			err:  fmt.Errorf("wrap: %w", &pgconn.PgError{Code: "23503"}),
			want: true,
		},
		{
			name: "different postgres error",
			err:  &pgconn.PgError{Code: "23505"},
			want: false,
		},
		{
			name: "generic error",
			err:  errors.New("boom"),
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isForeignKeyViolation(tt.err); got != tt.want {
				t.Fatalf("isForeignKeyViolation() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDeleteSupply_ReturnsBusinessRuleWhenActiveReferencesExist: un insumo con
// referencias ACTIVAS (stock + movimiento vivos) no puede eliminarse. El gate
// countActiveSupplyReferences cuenta esas filas vivas y DeleteSupply devuelve
// BusinessRule (-> 422), que el FE traduce a "in_use" y bloquea.
func TestDeleteSupply_ReturnsBusinessRuleWhenActiveReferencesExist(t *testing.T) {
	fx := setupDeleteSupplyFixture(t)

	err := fx.repo.DeleteSupply(fx.ctx, fx.supplyID)
	require.Error(t, err)
	assert.True(t, domainerr.IsKind(err, domainerr.KindBusinessRule),
		"se espera BusinessRule (422) cuando hay referencias activas, got: %v", err)
	assert.Contains(t, err.Error(), "in use by active records")

	var count int64
	require.NoError(t, fx.db.Unscoped().Model(&models.Supply{}).Where("id = ?", fx.supplyID).Count(&count).Error)
	assert.Equal(t, int64(1), count, "supply debe seguir existiendo cuando hay referencias activas")
}

// TestDeleteSupply_ReturnsConflictWhenOnlyHistoricalReferences: cuando el stock y
// el movimiento estan SOFT-DELETED, countActiveSupplyReferences cuenta 0 (GORM
// excluye deleted_at) pero las filas siguen fisicamente presentes y las FK
// RESTRICT impiden el hard-delete. DeleteSupply devuelve Conflict (-> 409), que el
// FE traduce a "conflict" y archiva en lugar de eliminar.
func TestDeleteSupply_ReturnsConflictWhenOnlyHistoricalReferences(t *testing.T) {
	fx := setupDeleteSupplyFixture(t)

	// Soft-delete de las referencias: deja deleted_at seteado pero las filas
	// fisicas siguen, manteniendo vivas las FK RESTRICT hacia el supply.
	require.NoError(t, fx.db.Delete(&models.SupplyMovement{}, fx.movementID).Error)
	require.NoError(t, fx.db.Delete(&stockmodels.Stock{}, fx.stockID).Error)

	err := fx.repo.DeleteSupply(fx.ctx, fx.supplyID)
	require.Error(t, err)
	assert.True(t, domainerr.IsConflict(err),
		"se espera Conflict (409) cuando solo hay referencias historicas, got: %v", err)
	assert.Contains(t, err.Error(), "historical references")

	var count int64
	require.NoError(t, fx.db.Unscoped().Model(&models.Supply{}).Where("id = ?", fx.supplyID).Count(&count).Error)
	assert.Equal(t, int64(1), count, "supply debe seguir existiendo cuando hay referencias historicas")
}

// deleteSupplyFixture agrupa lo necesario para ejercitar DeleteSupply contra una DB real.
type deleteSupplyFixture struct {
	db         *gorm.DB
	repo       *Repository
	ctx        context.Context
	supplyID   int64
	stockID    int64
	movementID int64
}

// setupDeleteSupplyFixture crea un insumo con un stock y un movimiento VIVOS
// (referencias activas) y registra la limpieza. Hace Skip si no hay DB de test.
func setupDeleteSupplyFixture(t *testing.T) deleteSupplyFixture {
	t.Helper()

	host := getEnvOrDefault("TEST_DB_HOST", os.Getenv("DB_HOST"))
	if host == "" {
		t.Skip("Skipping integration test: TEST_DB_HOST or DB_HOST not set")
	}

	port := 5432
	if p := getEnvOrDefault("TEST_DB_PORT", os.Getenv("DB_PORT")); p != "" {
		parsedPort, err := strconv.Atoi(p)
		require.NoError(t, err)
		port = parsedPort
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

	now := time.Now().UTC()
	suffix := now.UnixNano()

	tx := db.Begin()
	require.NoError(t, tx.Error)
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback().Error
		}
	}()

	classType := &classtypemodels.ClassType{Name: fmt.Sprintf("test_type_%d", suffix)}
	require.NoError(t, tx.Create(classType).Error)

	category := &categorymodels.Category{
		Name:   fmt.Sprintf("test_category_%d", suffix),
		TypeID: classType.ID,
	}
	require.NoError(t, tx.Create(category).Error)

	customer := &projectmodels.Customer{Name: fmt.Sprintf("test_customer_%d", suffix)}
	require.NoError(t, tx.Create(customer).Error)

	campaign := &projectmodels.Campaign{Name: fmt.Sprintf("test_campaign_%d", suffix)}
	require.NoError(t, tx.Create(campaign).Error)

	project := &projectmodels.Project{
		Name:        fmt.Sprintf("test_project_%d", suffix),
		CustomerID:  customer.ID,
		CampaignID:  campaign.ID,
		AdminCost:   decimal.Zero,
		PlannedCost: decimal.Zero,
	}
	require.NoError(t, tx.Create(project).Error)

	investor := &investormodels.Investor{Name: fmt.Sprintf("test_investor_%d", suffix)}
	require.NoError(t, tx.Create(investor).Error)

	provider := &providermodels.Provider{Name: fmt.Sprintf("test_provider_%d", suffix)}
	require.NoError(t, tx.Create(provider).Error)

	supply := &models.Supply{
		ProjectID:      project.ID,
		Name:           fmt.Sprintf("test_supply_%d", suffix),
		Price:          decimal.NewFromInt(10),
		IsPartialPrice: false,
		UnitID:         1,
		CategoryID:     category.ID,
		TypeID:         classType.ID,
	}
	require.NoError(t, tx.Create(supply).Error)

	stock := &stockmodels.Stock{
		ProjectID:      project.ID,
		SupplyID:       supply.ID,
		InvestorID:     investor.ID,
		InitialStock:   decimal.NewFromInt(5),
		YearPeriod:     int64(now.Year()),
		MonthPeriod:    int64(now.Month()),
		UnitsEntered:   decimal.NewFromInt(5),
		UnitsConsumed:  decimal.Zero,
		RealStockUnits: decimal.NewFromInt(5),
	}
	require.NoError(t, tx.Create(stock).Error)

	movementDate := now
	movement := &models.SupplyMovement{
		StockId:              stock.ID,
		Quantity:             decimal.NewFromInt(1),
		MovementType:         "Stock",
		MovementDate:         &movementDate,
		ReferenceNumber:      fmt.Sprintf("TEST-REF-%d", suffix),
		ProjectId:            project.ID,
		ProjectDestinationId: project.ID,
		SupplyID:             supply.ID,
		InvestorID:           investor.ID,
		ProviderID:           provider.ID,
		IsEntry:              true,
	}
	require.NoError(t, tx.Create(movement).Error)

	require.NoError(t, tx.Commit().Error)
	committed = true

	t.Cleanup(func() {
		cleanupDeleteSupplyTestData(t, db, movement.ID, stock.ID, supply.ID, provider.ID, investor.ID, project.ID, campaign.ID, customer.ID, category.ID, classType.ID)
	})

	return deleteSupplyFixture{
		db:         db,
		repo:       NewRepository(&gormEngineAdapter{client: db}),
		ctx:        context.Background(),
		supplyID:   supply.ID,
		stockID:    stock.ID,
		movementID: movement.ID,
	}
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

func cleanupDeleteSupplyTestData(
	t *testing.T,
	db *gorm.DB,
	movementID, stockID, supplyID, providerID, investorID, projectID, campaignID, customerID, categoryID, classTypeID int64,
) {
	t.Helper()

	require.NoError(t, db.Unscoped().Delete(&models.SupplyMovement{}, movementID).Error)
	require.NoError(t, db.Unscoped().Delete(&stockmodels.Stock{}, stockID).Error)
	require.NoError(t, db.Unscoped().Delete(&models.Supply{}, supplyID).Error)
	require.NoError(t, db.Unscoped().Delete(&providermodels.Provider{}, providerID).Error)
	require.NoError(t, db.Unscoped().Delete(&investormodels.Investor{}, investorID).Error)
	require.NoError(t, db.Unscoped().Delete(&projectmodels.Project{}, projectID).Error)
	require.NoError(t, db.Unscoped().Delete(&projectmodels.Campaign{}, campaignID).Error)
	require.NoError(t, db.Unscoped().Delete(&projectmodels.Customer{}, customerID).Error)
	require.NoError(t, db.Unscoped().Delete(&categorymodels.Category{}, categoryID).Error)
	require.NoError(t, db.Unscoped().Delete(&classtypemodels.ClassType{}, classTypeID).Error)
}
