package main

import (
	"context"
	"fmt"
	"log"
	"time"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	gormRepo "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"

	businessparametermodels "github.com/alphacodinggroup/ponti-backend/internal/business-parameters/repository/models"
	campaignmodels "github.com/alphacodinggroup/ponti-backend/internal/campaign/repository/models"
	categorymodels "github.com/alphacodinggroup/ponti-backend/internal/category/repository/models"
	classtypemodels "github.com/alphacodinggroup/ponti-backend/internal/class-type/repository/models"
	commercializationmodels "github.com/alphacodinggroup/ponti-backend/internal/commercialization/repository/models"
	cropmodels "github.com/alphacodinggroup/ponti-backend/internal/crop/repository/models"
	customermodels "github.com/alphacodinggroup/ponti-backend/internal/customer/repository/models"
	dollarmodels "github.com/alphacodinggroup/ponti-backend/internal/dollar/repository/models"
	fieldmodels "github.com/alphacodinggroup/ponti-backend/internal/field/repository/models"
	investormodels "github.com/alphacodinggroup/ponti-backend/internal/investor/repository/models"
	invoicemodels "github.com/alphacodinggroup/ponti-backend/internal/invoice/repository/models"
	leasetypemodels "github.com/alphacodinggroup/ponti-backend/internal/lease-type/repository/models"
	lotmodels "github.com/alphacodinggroup/ponti-backend/internal/lot/repository/models"
	managermodels "github.com/alphacodinggroup/ponti-backend/internal/manager/repository/models"
	projectmodels "github.com/alphacodinggroup/ponti-backend/internal/project/repository/models"
	stockmodels "github.com/alphacodinggroup/ponti-backend/internal/stock/repository/models"
	supplymodels "github.com/alphacodinggroup/ponti-backend/internal/supply/repository/models"
	workordermodels "github.com/alphacodinggroup/ponti-backend/internal/work-order/repository/models"
)

// runGormMigrations runs GORM AutoMigrate on all models and ensures
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
		&workordermodels.Workorder{},
		&workordermodels.WorkorderItem{},
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
