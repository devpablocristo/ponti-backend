package supply

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	categorymodels "github.com/devpablocristo/ponti-backend/internal/category/repository/models"
	classtypemodels "github.com/devpablocristo/ponti-backend/internal/class-type/repository/models"
	investormodels "github.com/devpablocristo/ponti-backend/internal/investor/repository/models"
	projectmodels "github.com/devpablocristo/ponti-backend/internal/project/repository/models"
	providermodels "github.com/devpablocristo/ponti-backend/internal/provider/repository/models"
	providerdomain "github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
	stockmodels "github.com/devpablocristo/ponti-backend/internal/stock/repository/models"
	models "github.com/devpablocristo/ponti-backend/internal/supply/repository/models"
	supplydomain "github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
)

func TestDeleteSupplyMovement_DeletesOnlyMatchingInvestorStock(t *testing.T) {
	repo, db := newSQLiteSupplyRepository(t)
	fixture := seedDeleteMovementFixture(t, db)

	err := repo.DeleteSupplyMovement(context.Background(), fixture.project.ID, fixture.movementA.ID)
	require.NoError(t, err)

	var activeMovementCount int64
	require.NoError(t, db.Model(&models.SupplyMovement{}).Where("id = ?", fixture.movementA.ID).Count(&activeMovementCount).Error)
	assert.Equal(t, int64(0), activeMovementCount)

	var activeStockACount int64
	require.NoError(t, db.Model(&stockmodels.Stock{}).Where("id = ?", fixture.stockA.ID).Count(&activeStockACount).Error)
	assert.Equal(t, int64(0), activeStockACount)

	var activeStockBCount int64
	require.NoError(t, db.Model(&stockmodels.Stock{}).Where("id = ?", fixture.stockB.ID).Count(&activeStockBCount).Error)
	assert.Equal(t, int64(1), activeStockBCount)
}

func TestDeleteSupplyMovement_ClosedStockCheckIsInvestorAware(t *testing.T) {
	repo, db := newSQLiteSupplyRepository(t)
	fixture := seedDeleteMovementFixture(t, db)

	closedAt := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
	closedStockOtherInvestor := &stockmodels.Stock{
		ProjectID:         fixture.project.ID,
		SupplyID:          fixture.supply.ID,
		InvestorID:        fixture.investorB.ID,
		CloseDate:         &closedAt,
		InitialStock:      decimal.NewFromInt(5),
		YearPeriod:        2026,
		MonthPeriod:       4,
		UnitsEntered:      decimal.NewFromInt(5),
		UnitsConsumed:     decimal.Zero,
		RealStockUnits:    decimal.NewFromInt(5),
		HasRealStockCount: true,
	}
	require.NoError(t, db.Create(closedStockOtherInvestor).Error)

	err := repo.DeleteSupplyMovement(context.Background(), fixture.project.ID, fixture.movementA.ID)
	require.NoError(t, err)

	var activeStockACount int64
	require.NoError(t, db.Model(&stockmodels.Stock{}).Where("id = ?", fixture.stockA.ID).Count(&activeStockACount).Error)
	assert.Equal(t, int64(0), activeStockACount)

	var closedStockCount int64
	require.NoError(t, db.Model(&stockmodels.Stock{}).Where("id = ?", closedStockOtherInvestor.ID).Count(&closedStockCount).Error)
	assert.Equal(t, int64(1), closedStockCount)
}

func TestDeleteSupplyMovement_AllowsActiveMovementWhenSameTripletHasClosedStock(t *testing.T) {
	repo, db := newSQLiteSupplyRepository(t)
	fixture := seedDeleteMovementFixture(t, db)

	closedAt := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
	closedStockSameTriplet := &stockmodels.Stock{
		ProjectID:         fixture.project.ID,
		SupplyID:          fixture.supply.ID,
		InvestorID:        fixture.investorA.ID,
		CloseDate:         &closedAt,
		InitialStock:      decimal.NewFromInt(5),
		YearPeriod:        2026,
		MonthPeriod:       4,
		UnitsEntered:      decimal.NewFromInt(5),
		UnitsConsumed:     decimal.Zero,
		RealStockUnits:    decimal.NewFromInt(5),
		HasRealStockCount: true,
	}
	require.NoError(t, db.Create(closedStockSameTriplet).Error)

	err := repo.DeleteSupplyMovement(context.Background(), fixture.project.ID, fixture.movementA.ID)
	require.NoError(t, err)

	var activeMovementCount int64
	require.NoError(t, db.Model(&models.SupplyMovement{}).Where("id = ?", fixture.movementA.ID).Count(&activeMovementCount).Error)
	assert.Equal(t, int64(0), activeMovementCount)

	var closedStockCount int64
	require.NoError(t, db.Model(&stockmodels.Stock{}).Where("id = ?", closedStockSameTriplet.ID).Count(&closedStockCount).Error)
	assert.Equal(t, int64(1), closedStockCount)
}

func TestDeleteSupplyMovement_RejectsMovementFromClosedStock(t *testing.T) {
	repo, db := newSQLiteSupplyRepository(t)
	fixture := seedDeleteMovementFixture(t, db)
	_, closedMovement := seedClosedStockAndMovement(t, db, fixture)

	err := repo.DeleteSupplyMovement(context.Background(), fixture.project.ID, closedMovement.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "closed stock movement")

	var movementCount int64
	require.NoError(t, db.Model(&models.SupplyMovement{}).Where("id = ?", closedMovement.ID).Count(&movementCount).Error)
	assert.Equal(t, int64(1), movementCount)
}

func TestDeleteSupplyMovement_RejectsInternalMovementWhenRelatedStockIsClosed(t *testing.T) {
	repo, db := newSQLiteSupplyRepository(t)
	fixture := seedDeleteMovementFixture(t, db)
	closedStock, relatedMovement := seedClosedStockAndMovement(t, db, fixture)

	fixture.movementA.MovementType = "Movimiento interno"
	fixture.movementA.ReferenceNumber = relatedMovement.ReferenceNumber
	fixture.movementA.MovementDate = relatedMovement.MovementDate
	require.NoError(t, db.Save(fixture.movementA).Error)

	relatedMovement.MovementType = "Movimiento interno entrada"
	require.NoError(t, db.Save(relatedMovement).Error)

	err := repo.DeleteSupplyMovement(context.Background(), fixture.project.ID, fixture.movementA.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "closed stock movement")

	var closedStockCount int64
	require.NoError(t, db.Model(&stockmodels.Stock{}).Where("id = ?", closedStock.ID).Count(&closedStockCount).Error)
	assert.Equal(t, int64(1), closedStockCount)
}

func TestUpdateSupplyMovement_RejectsMovementFromClosedStock(t *testing.T) {
	repo, db := newSQLiteSupplyRepository(t)
	fixture := seedDeleteMovementFixture(t, db)
	_, closedMovement := seedClosedStockAndMovement(t, db, fixture)

	newDate := time.Date(2026, 4, 18, 0, 0, 0, 0, time.UTC)
	err := repo.UpdateSupplyMovement(context.Background(), &supplydomain.SupplyMovement{
		ID:                   closedMovement.ID,
		StockId:              fixture.stockA.ID,
		Quantity:             decimal.NewFromInt(9),
		MovementType:         "Remito oficial",
		MovementDate:         &newDate,
		ReferenceNumber:      "REF-UPDATED",
		ProjectId:            fixture.project.ID,
		ProjectDestinationId: 0,
		Supply:               fixture.supply.ToDomain(),
		Investor:             fixture.investorA.ToDomain(),
		Provider:             &providerdomain.Provider{ID: closedMovement.ProviderID},
		IsEntry:              true,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "closed stock movement")

	var persisted models.SupplyMovement
	require.NoError(t, db.First(&persisted, closedMovement.ID).Error)
	assert.Equal(t, closedMovement.StockId, persisted.StockId)
	assert.Equal(t, closedMovement.ReferenceNumber, persisted.ReferenceNumber)
}

type deleteMovementFixture struct {
	project   *projectmodels.Project
	supply    *models.Supply
	investorA *investormodels.Investor
	investorB *investormodels.Investor
	stockA    *stockmodels.Stock
	stockB    *stockmodels.Stock
	movementA *models.SupplyMovement
}

func newSQLiteSupplyRepository(t *testing.T) (*Repository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec("PRAGMA foreign_keys = ON").Error)
	require.NoError(t, db.AutoMigrate(
		&classtypemodels.ClassType{},
		&categorymodels.Category{},
		&projectmodels.Customer{},
		&projectmodels.Campaign{},
		&projectmodels.Project{},
		&investormodels.Investor{},
		&providermodels.Provider{},
		&models.Supply{},
		&stockmodels.Stock{},
		&models.SupplyMovement{},
	))

	return NewRepository(&gormEngineAdapter{client: db}), db
}

func seedDeleteMovementFixture(t *testing.T, db *gorm.DB) *deleteMovementFixture {
	t.Helper()

	classType := &classtypemodels.ClassType{Name: "test_type"}
	require.NoError(t, db.Create(classType).Error)

	category := &categorymodels.Category{Name: "test_category", TypeID: classType.ID}
	require.NoError(t, db.Create(category).Error)

	customer := &projectmodels.Customer{Name: "test_customer"}
	require.NoError(t, db.Create(customer).Error)

	campaign := &projectmodels.Campaign{Name: "test_campaign"}
	require.NoError(t, db.Create(campaign).Error)

	project := &projectmodels.Project{
		Name:        "test_project",
		CustomerID:  customer.ID,
		CampaignID:  campaign.ID,
		AdminCost:   decimal.Zero,
		PlannedCost: decimal.Zero,
	}
	require.NoError(t, db.Create(project).Error)

	investorA := &investormodels.Investor{Name: "investor_a"}
	require.NoError(t, db.Create(investorA).Error)

	investorB := &investormodels.Investor{Name: "investor_b"}
	require.NoError(t, db.Create(investorB).Error)

	provider := &providermodels.Provider{Name: "provider"}
	require.NoError(t, db.Create(provider).Error)

	supply := &models.Supply{
		ProjectID:      project.ID,
		Name:           "test_supply",
		Price:          decimal.NewFromInt(10),
		IsPartialPrice: false,
		UnitID:         1,
		CategoryID:     category.ID,
		TypeID:         classType.ID,
	}
	require.NoError(t, db.Create(supply).Error)

	stockA := &stockmodels.Stock{
		ProjectID:         project.ID,
		SupplyID:          supply.ID,
		InvestorID:        investorA.ID,
		InitialStock:      decimal.NewFromInt(5),
		YearPeriod:        2026,
		MonthPeriod:       4,
		UnitsEntered:      decimal.NewFromInt(5),
		UnitsConsumed:     decimal.Zero,
		RealStockUnits:    decimal.NewFromInt(5),
		HasRealStockCount: true,
	}
	require.NoError(t, db.Create(stockA).Error)

	stockB := &stockmodels.Stock{
		ProjectID:         project.ID,
		SupplyID:          supply.ID,
		InvestorID:        investorB.ID,
		InitialStock:      decimal.NewFromInt(7),
		YearPeriod:        2026,
		MonthPeriod:       4,
		UnitsEntered:      decimal.NewFromInt(7),
		UnitsConsumed:     decimal.Zero,
		RealStockUnits:    decimal.NewFromInt(7),
		HasRealStockCount: true,
	}
	require.NoError(t, db.Create(stockB).Error)

	movementDate := time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC)
	movementA := &models.SupplyMovement{
		StockId:              stockA.ID,
		Quantity:             decimal.NewFromInt(3),
		MovementType:         "Remito oficial",
		MovementDate:         &movementDate,
		ReferenceNumber:      "REF-1",
		ProjectId:            project.ID,
		ProjectDestinationId: 0,
		SupplyID:             supply.ID,
		InvestorID:           investorA.ID,
		ProviderID:           provider.ID,
		IsEntry:              true,
	}
	require.NoError(t, db.Create(movementA).Error)

	return &deleteMovementFixture{
		project:   project,
		supply:    supply,
		investorA: investorA,
		investorB: investorB,
		stockA:    stockA,
		stockB:    stockB,
		movementA: movementA,
	}
}

func seedClosedStockAndMovement(t *testing.T, db *gorm.DB, fixture *deleteMovementFixture) (*stockmodels.Stock, *models.SupplyMovement) {
	t.Helper()

	closedAt := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
	closedStock := &stockmodels.Stock{
		ProjectID:         fixture.project.ID,
		SupplyID:          fixture.supply.ID,
		InvestorID:        fixture.investorA.ID,
		CloseDate:         &closedAt,
		InitialStock:      decimal.NewFromInt(4),
		YearPeriod:        2026,
		MonthPeriod:       4,
		UnitsEntered:      decimal.NewFromInt(4),
		UnitsConsumed:     decimal.Zero,
		RealStockUnits:    decimal.NewFromInt(4),
		HasRealStockCount: true,
	}
	require.NoError(t, db.Create(closedStock).Error)

	movementDate := time.Date(2026, 4, 14, 0, 0, 0, 0, time.UTC)
	closedMovement := &models.SupplyMovement{
		StockId:              closedStock.ID,
		Quantity:             decimal.NewFromInt(2),
		MovementType:         "Remito oficial",
		MovementDate:         &movementDate,
		ReferenceNumber:      "REF-CLOSED",
		ProjectId:            fixture.project.ID,
		ProjectDestinationId: 0,
		SupplyID:             fixture.supply.ID,
		InvestorID:           fixture.investorA.ID,
		ProviderID:           fixture.movementA.ProviderID,
		IsEntry:              true,
	}
	require.NoError(t, db.Create(closedMovement).Error)

	return closedStock, closedMovement
}
