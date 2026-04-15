package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"

	"github.com/devpablocristo/ponti-backend/internal/businessinsights"
	"github.com/devpablocristo/ponti-backend/internal/reviewproxy"
	stockmod "github.com/devpablocristo/ponti-backend/internal/stock"
	wire "github.com/devpablocristo/ponti-backend/wire"
)

// runHTTPServer registra rutas en Gin y levanta el servidor HTTP.
func runHTTPServer(ctx context.Context, deps *wire.Dependencies) error {
	if deps == nil {
		return errors.New("dependencies cannot be nil")
	}

	// Configurar Gin con middlewares globales.
	// Middlewares globales: ErrorHandling, RequestAndResponseLogger.
	deps.GinEngine.GetRouter().Use(deps.Middlewares.GetGlobal()...)

	// Service de notificaciones reactivas (stock bajo, etc). El Service es
	// opcional: si REVIEW_URL no esta seteado, el service degrada gracioso
	// y los use cases que lo llamen quedan no-op. Wireado manualmente
	// porque todavia no tiene consumidores en google/wire.
	biRepo := businessinsights.NewRepository(deps.GormRepo.Client())
	biService := buildBusinessInsightsService(deps, biRepo)
	biHandler := businessinsights.NewHandler(biRepo, biService, deps.GinEngine, &deps.Config.API, deps.Middlewares)
	deps.StockUseCases.SetBusinessInsightsNotifier(&stockNegativeAdapter{svc: biService})

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
	registerHTTPRoutes(deps, biHandler)

	log.Println("Starting HTTP Server on port: ", deps.Config.HTTPServer.Port)
	log.Println("Version: ", deps.Config.Service.Version)
	log.Println("--------------------------------")
	log.Println("Database: ", deps.Config.DB.Name)
	log.Println("--------------------------------")

	// Iniciar el servidor HTTP (ej: puerto 8080).
	return deps.GinEngine.RunServer(ctx)
}

// stockNegativeAdapter traduce el StockNegativeInput del stock.Handler al
// StockLevel del businessinsights.Service, manteniendo desacoplados los
// tipos de los paquetes (stock no importa businessinsights y viceversa).
type stockNegativeAdapter struct {
	svc *businessinsights.Service
}

func (a *stockNegativeAdapter) NotifyStockNegative(ctx context.Context, tenantID uuid.UUID, actor string, in stockmod.StockNegativeInput) error {
	if a == nil || a.svc == nil {
		return nil
	}
	return a.svc.NotifyStockNegative(ctx, tenantID, actor, businessinsights.StockLevel{
		ProductID:   in.ProductID,
		ProductName: in.ProductName,
		Quantity:    in.Quantity,
	})
}

func (a *stockNegativeAdapter) MaybeResolveStockNegative(ctx context.Context, tenantID uuid.UUID, productID string) error {
	if a == nil || a.svc == nil {
		return nil
	}
	return a.svc.MaybeResolveStockNegative(ctx, tenantID, productID)
}

// buildBusinessInsightsService arma el Service para notificaciones reactivas.
// Si REVIEW_URL esta vacio devuelve un Service con review=nil (no-op gracioso).
func buildBusinessInsightsService(deps *wire.Dependencies, repo *businessinsights.Repository) *businessinsights.Service {
	reviewURL := strings.TrimSpace(deps.Config.Review.URL)
	var client businessinsights.ReviewClient
	if reviewURL != "" {
		client = reviewproxy.NewClient(reviewURL, strings.TrimSpace(deps.Config.Review.APIKey))
	}
	return businessinsights.NewService(repo, repo, repo, client, businessinsights.Config{})
}

// registerHTTPRoutes registra todas las rutas en el router Gin.
func registerHTTPRoutes(deps *wire.Dependencies, biHandler *businessinsights.Handler) {
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
	deps.DollarHandler.Routes()
	deps.LaborHandler.Routes()
	deps.StockHandler.Routes()
	deps.CommercializationHandler.Routes()
	deps.AIHandler.Routes()
	deps.AdminHandler.Routes()
	biHandler.Routes()
}
