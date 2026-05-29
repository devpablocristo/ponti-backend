package main

import (
	"context"
	"fmt"
	"log"
	"time"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	gormRepo "github.com/devpablocristo/ponti-backend/internal/platform/persistence/gorm"

	businessparametermodels "github.com/devpablocristo/ponti-backend/internal/business-parameters/repository/models"
	campaignmodels "github.com/devpablocristo/ponti-backend/internal/campaign/repository/models"
	categorymodels "github.com/devpablocristo/ponti-backend/internal/category/repository/models"
	classtypemodels "github.com/devpablocristo/ponti-backend/internal/class-type/repository/models"
	commercializationmodels "github.com/devpablocristo/ponti-backend/internal/commercialization/repository/models"
	cropmodels "github.com/devpablocristo/ponti-backend/internal/crop/repository/models"
	customermodels "github.com/devpablocristo/ponti-backend/internal/customer/repository/models"
	dollarmodels "github.com/devpablocristo/ponti-backend/internal/dollar/repository/models"
	fieldmodels "github.com/devpablocristo/ponti-backend/internal/field/repository/models"
	investormodels "github.com/devpablocristo/ponti-backend/internal/investor/repository/models"
	invoicemodels "github.com/devpablocristo/ponti-backend/internal/invoice/repository/models"
	leasetypemodels "github.com/devpablocristo/ponti-backend/internal/lease-type/repository/models"
	lotmodels "github.com/devpablocristo/ponti-backend/internal/lot/repository/models"
	managermodels "github.com/devpablocristo/ponti-backend/internal/manager/repository/models"
	projectmodels "github.com/devpablocristo/ponti-backend/internal/project/repository/models"
	stockmodels "github.com/devpablocristo/ponti-backend/internal/stock/repository/models"
	supplymodels "github.com/devpablocristo/ponti-backend/internal/supply/repository/models"
	workOrderModels "github.com/devpablocristo/ponti-backend/internal/work-order/repository/models"
)

// runGormMigrations corre AutoMigrate de GORM sobre todos los modelos.
func runGormMigrations(ctx context.Context, repo *gormRepo.Repository) error {
	log.Println("Starting GORM migrations...")

	// Verify DB connection
	sqlDB, err := repo.Client().DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	modelsList := []any{
		&customermodels.Customer{},
		&campaignmodels.Campaign{},
		&leasetypemodels.LeaseType{},
		&managermodels.Manager{},
		&investormodels.Investor{},
		&cropmodels.Crop{},
		&projectmodels.Manager{},
		&commercializationmodels.CropCommercialization{},
		&fieldmodels.Field{},
		&fieldmodels.FieldInvestor{},
		&lotmodels.Lot{},
		&categorymodels.Category{},
		&businessparametermodels.BusinessParameter{},
		&classtypemodels.ClassType{},
		&supplymodels.Supply{},
		&dollarmodels.ProjectDollarValue{},
		&workOrderModels.WorkOrder{},
		&workOrderModels.WorkOrderItem{},
		&projectmodels.ProjectInvestor{},
		&projectmodels.AdminCostInvestor{},
		&invoicemodels.Invoice{},
		&projectmodels.Project{},
		&supplymodels.SupplyMovement{},
		&stockmodels.Stock{},
	}

	start := time.Now()
	for _, m := range modelsList {
		fmt.Printf("Migrating model: %T\n", m)
		if err := repo.AutoMigrate(m); err != nil {
			return fmt.Errorf("failed to migrate %T: %w", m, err)
		}
	}
	log.Printf("GORM migrations completed successfully in %s.", time.Since(start))
	return nil
}
