package main

import (
	"context"
	"errors"
	"log"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	wire "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/wire"
)

// runHttpServer registers routes in the Gin router and starts the HTTP server.
func runHttpServer(ctx context.Context, deps *wire.Dependencies) error {
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
	deps.LeaseTypeHandler.Routes()
	deps.SupplyHandler.Routes()
	deps.CategoryHandler.Routes()
	deps.ClassTypeHandler.Routes()
	deps.UnitHandler.Routes()
	deps.DollarHandler.Routes()
	deps.LaborHandler.Routes()

	log.Println("HTTP routes registered successfully.")
}
