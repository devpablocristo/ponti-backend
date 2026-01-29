package main

import (
	"context"
	"errors"
	"log"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	wire "github.com/alphacodinggroup/ponti-backend/wire"
)

// runHTTPServer registra rutas en Gin y levanta el servidor HTTP.
func runHTTPServer(ctx context.Context, deps *wire.Dependencies) error {
	if deps == nil {
		return errors.New("dependencies cannot be nil")
	}

	// Configurar Gin con middlewares globales.
	// Middlewares globales: ErrorHandling, RequestAndResponseLogger.
	deps.GinEngine.GetRouter().Use(deps.Middlewares.GetGlobal()...)

	// Registrar todas las rutas de la aplicación.
	// Cada handler aplica sus middlewares de validación específicos.
	registerHTTPRoutes(deps)

	log.Println("Starting HTTP Server on port: ", deps.Config.HTTPServer.Port)
	log.Println("Version: ", deps.Config.App.Version)
	log.Println("--------------------------------")
	log.Println("Database: ", deps.Config.DB.Name)
	log.Println("--------------------------------")

	// Iniciar el servidor HTTP (ej: puerto 8080).
	return deps.GinEngine.RunServer(ctx)
}

// registerHTTPRoutes registra todas las rutas en el router Gin.
func registerHTTPRoutes(deps *wire.Dependencies) {
	deps.LotHandler.Routes()
	deps.CustomerHandler.Routes()
	deps.CampaignHandler.Routes()
	deps.DashboardHandler.Routes()
	deps.DataIntegrityHandler.Routes()
	deps.InvestorHandler.Routes()
	deps.InvoiceHandler.Routes()
	deps.FieldHandler.Routes()
	deps.ProjectHandler.Routes()
	deps.ReportHandler.Routes()
	deps.CropHandler.Routes()
	deps.ManagerHandler.Routes()
	deps.LeaseTypeHandler.Routes()
	deps.SupplyHandler.Routes()
	deps.CategoryHandler.Routes()
	deps.ClassTypeHandler.Routes()
	deps.BusinessParametersHandler.Routes()
	deps.WorkOrderHandler.Routes()
	deps.DollarHandler.Routes()
	deps.LaborHandler.Routes()
	deps.StockHandler.Routes()
	deps.CommercializationHandler.Routes()
}
