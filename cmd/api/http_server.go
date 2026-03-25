package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"runtime"
	"time"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gin-gonic/gin"

	ai "github.com/alphacodinggroup/ponti-backend/internal/ai"
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

	// Middleware: auto-trigger de insights después de mutaciones exitosas.
	// Dispara cómputo async y throttleado en ponti-ai cuando una ruta con
	// :project_id en el path responde 2xx a POST/PUT/PATCH/DELETE.
	trigger := ai.NewInsightTrigger(
		deps.Config.AI.ServiceURL,
		deps.Config.AI.ServiceKey,
		deps.Config.AI.TimeoutMS,
		deps.Config.AI.ComputeThrottleSec,
	)
	deps.GinEngine.GetRouter().Use(ai.InsightTriggerMiddleware(trigger))

	// Meta endpoints (version + health) bajo /api/v1 (o el APIBaseURL configurado).
	apiBase := deps.Config.API.APIBaseURL()
	deps.GinEngine.GetRouter().GET(apiBase+"/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": gin.H{
				"name":       deps.Config.Service.Name,
				"version":    deps.Config.Service.Version,
				"git_sha":    deps.Config.Service.GitSHA,
				"build_time": deps.Config.Service.BuildTime,
			},
			"api": gin.H{
				"base_url": apiBase,
				"version":  deps.Config.API.APIVersion(),
			},
			"runtime": gin.H{
				"go": runtime.Version(),
			},
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})
	deps.GinEngine.GetRouter().GET(apiBase+"/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})
	deps.GinEngine.GetRouter().GET(apiBase+"/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// Registrar todas las rutas de la aplicación.
	// Cada handler aplica sus middlewares de validación específicos.
	registerHTTPRoutes(deps)

	log.Println("Starting HTTP Server on port: ", deps.Config.HTTPServer.Port)
	log.Println("Version: ", deps.Config.Service.Version)
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
	deps.ProviderHandler.Routes()
	deps.ReportHandler.Routes()
	deps.CropHandler.Routes()
	deps.ManagerHandler.Routes()
	deps.LeaseTypeHandler.Routes()
	deps.SupplyHandler.Routes()
	deps.CategoryHandler.Routes()
	deps.ClassTypeHandler.Routes()
	deps.BusinessParametersHandler.Routes()
	deps.WorkOrderHandler.Routes()
	deps.WorkOrderDraftHandler.Routes()
	deps.DollarHandler.Routes()
	deps.LaborHandler.Routes()
	deps.StockHandler.Routes()
	deps.CommercializationHandler.Routes()
	deps.AIHandler.Routes()
	deps.AdminHandler.Routes()
}
