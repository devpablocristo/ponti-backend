package main

import (
	"context"
	"fmt"
	"log"
	"time"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"

	campaignmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign/repository/models"
	categorymodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/category/repository/models"
	classtypemodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/classtype/repository/models"
	cropmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/repository/models"
	customermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/repository/models"
	fieldmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/repository/models"
	investormodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/repository/models"
	leasetypemodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/leasetype/repository/models"
	lotmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/repository/models"
	managermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/repository/models"
	projectmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/repository/models"
	supplymodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/repository/models"
	unitmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/unit/repository/models"
)

// RunGormMigrations runs SQL migrations using GORM.
func runGormMigrations(ctx context.Context, repo *gorm.Repository) error {
	log.Println("Starting GORM migrations...")

	sqlDB, err := repo.Client().DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	models := []any{
		&campaignmodels.Campaign{}, // primero4
		&leasetypemodels.LeaseType{},
		&managermodels.Manager{},
		&investormodels.Investor{},
		&cropmodels.Crop{},
		&fieldmodels.Field{},
		&lotmodels.Lot{},
		&customermodels.Customer{},
		&supplymodels.Supply{},
		&categorymodels.Category{},
		&classtypemodels.ClassType{},
		&unitmodels.Unit{},
		&projectmodels.ProjectInvestor{},
		&projectmodels.Project{}, // último
	}

	start := time.Now()
	for _, model := range models {
		fmt.Printf("Migrating model: %T\n", model)
		if err := repo.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate %T: %w", model, err)
		}
	}
	duration := time.Since(start)
	log.Printf("GORM migrations completed successfully in %s.", duration)

	return nil
}
