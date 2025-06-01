package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"

	campaignmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign/repository/models"
	cropmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/repository/models"
	customermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/repository/models"
	fieldmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/repository/models"
	investormodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/repository/models"
	lotmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/repository/models"
	managermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/repository/models"
	projectmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/repository/models"

	wire "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/wire"
)

// RunHttpServer registers routes in the Gin router and starts the HTTP server.
func RunHttpServer(ctx context.Context, deps *wire.Dependencies) error {
	if deps == nil {
		return errors.New("dependencies cannot be nil")
	}

	log.Println("Registering HTTP routes...")

	// Set up the Gin router with global middlewares
	deps.GinEngine.GetRouter().Use(deps.Middlewares.GetGlobal()...)

	// Register all application routes.
	log.Println("Starting HTTP Server...")
	registerHttpRoutes(deps)

	// Start the HTTP server (e.g., on port 8080).
	return deps.GinEngine.RunServer(ctx)
}

// registerHttpRoutes registers all application routes in the Gin router.
func registerHttpRoutes(deps *wire.Dependencies) {
	deps.LotHandler.Routes()
	deps.CustomerHandler.Routes()
	deps.CampaignHandler.Routes()
	deps.InvestorHandler.Routes()
	deps.FieldHandler.Routes()
	deps.ProjectHandler.Routes()
	deps.CropHandler.Routes()
	deps.ManagerHandler.Routes()
}

// RunGormMigrations runs SQL migrations using GORM.
func RunGormMigrations(ctx context.Context, repo *gorm.Repository) error {
	log.Println("Starting GORM migrations...")

	sqlDB, err := repo.Client().DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	models := []any{
		&cropmodels.Crop{},
		&fieldmodels.Field{},
		&customermodels.Customer{},
		&investormodels.Investor{},
		&lotmodels.Lot{},
		&projectmodels.Project{},
		&projectmodels.ProjectInvestor{},
		&managermodels.Manager{},
		&campaignmodels.Campaign{},
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
