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

	"github.com/devpablocristo/core/security/go/contextkeys"
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
			"reporting": gin.H{
				"read_mode": deps.Config.Reporting.ReadMode,
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
	registerMeContextRoute(deps, apiBase)

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

func registerMeContextRoute(deps *wire.Dependencies, apiBase string) {
	group := deps.GinEngine.GetRouter().Group(apiBase, deps.Middlewares.GetValidation()...)
	group.GET("/me/context", func(c *gin.Context) {
		actor, _ := c.Request.Context().Value(ctxkeys.Actor).(string)
		if strings.TrimSpace(actor) == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "authentication context required"})
			return
		}

		type userRow struct {
			ID       uuid.UUID `json:"id"`
			IDPSub   string    `json:"idp_sub"`
			IDPEmail string    `json:"idp_email"`
			Email    string    `json:"email"`
		}
		var user userRow
		if err := deps.GormRepo.Client().
			WithContext(c.Request.Context()).
			Table("users").
			Select("id, idp_sub, idp_email, email").
			Where("idp_sub = ?", actor).
			Limit(1).
			Take(&user).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"message": "local user not found"})
			return
		}

		type tenantRow struct {
			TenantID uuid.UUID `json:"tenant_id"`
			Name     string    `json:"name"`
			RoleID   uuid.UUID `json:"-"`
			RoleName string    `json:"role"`
		}
		var tenants []tenantRow
		if err := deps.GormRepo.Client().
			WithContext(c.Request.Context()).
			Table("auth_memberships AS m").
			Select("m.tenant_id, t.name, m.role_id, r.name AS role_name").
			Joins("JOIN auth_tenants t ON t.id = m.tenant_id").
			Joins("JOIN auth_roles r ON r.id = m.role_id").
			Where("m.user_id = ? AND m.status = 'active'", user.ID).
			Order("t.name ASC").
			Find(&tenants).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "unable to load memberships"})
			return
		}

		type permissionRow struct {
			RoleID uuid.UUID
			Name   string
		}
		var roleIDs []uuid.UUID
		for _, tenant := range tenants {
			roleIDs = append(roleIDs, tenant.RoleID)
		}
		permissionsByRole := map[uuid.UUID][]string{}
		if len(roleIDs) > 0 {
			var perms []permissionRow
			if err := deps.GormRepo.Client().
				WithContext(c.Request.Context()).
				Table("auth_role_permissions rp").
				Select("rp.role_id, p.name").
				Joins("JOIN auth_permissions p ON p.id = rp.permission_id").
				Where("rp.role_id IN ?", roleIDs).
				Order("p.name ASC").
				Find(&perms).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "unable to load permissions"})
				return
			}
			for _, perm := range perms {
				permissionsByRole[perm.RoleID] = append(permissionsByRole[perm.RoleID], perm.Name)
			}
		}

		currentTenantID, _ := c.Request.Context().Value(ctxkeys.OrgID).(uuid.UUID)
		tenantPayload := make([]gin.H, 0, len(tenants))
		for _, tenant := range tenants {
			tenantPayload = append(tenantPayload, gin.H{
				"id":          tenant.TenantID,
				"name":        tenant.Name,
				"role":        tenant.RoleName,
				"permissions": permissionsByRole[tenant.RoleID],
				"is_current":  tenant.TenantID == currentTenantID,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"user": gin.H{
				"id":        user.ID,
				"idp_sub":   user.IDPSub,
				"idp_email": user.IDPEmail,
				"email":     user.Email,
			},
			"current_tenant_id": currentTenantID,
			"tenants":           tenantPayload,
		})
	})
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
	deps.ActorHandler.Routes()
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
	biHandler.Routes()
}
