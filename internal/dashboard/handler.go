// Package dashboard expone endpoints del dashboard.
package dashboard

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	dto "github.com/devpablocristo/ponti-backend/internal/dashboard/handler/dto"
	domain "github.com/devpablocristo/ponti-backend/internal/dashboard/usecases/domain"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

type UseCasesPort interface {
	GetDashboard(context.Context, domain.DashboardFilter) (*domain.DashboardData, error)
}

type GinEnginePort interface {
	GetRouter() *gin.Engine
	RunServer(ctx context.Context) error
}

type ConfigAPIPort interface {
	APIVersion() string
	APIBaseURL() string
}

type MiddlewaresEnginePort interface {
	GetGlobal() []gin.HandlerFunc
	GetValidation() []gin.HandlerFunc
	GetProtected() []gin.HandlerFunc
}

// Handler encapsulates all dependencies for the Dashboard HTTP handler.
type Handler struct {
	ucs UseCasesPort
	gsv GinEnginePort
	acf ConfigAPIPort
	mws MiddlewaresEnginePort
}

// NewHandler creates a new Dashboard handler.
func NewHandler(u UseCasesPort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		acf: c,
		mws: m,
	}
}

// Routes registers all dashboard routes.
func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/dashboard"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	public := r.Group(baseURL)
	{
		public.GET("", h.GetDashboard)
	}
}

// GetDashboard retrieves dashboard data based on query parameters.
func (h *Handler) GetDashboard(c *gin.Context) {
	// Parse query parameters for filters
	workspaceFilter, err := sharedhandlers.ParseWorkspaceFilter(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	filter := domain.DashboardFilter{
		CustomerID: workspaceFilter.CustomerID,
		ProjectID:  workspaceFilter.ProjectID,
		CampaignID: workspaceFilter.CampaignID,
		FieldID:    workspaceFilter.FieldID,
	}

	// Get dashboard data
	dashboardData, err := h.ucs.GetDashboard(c.Request.Context(), filter)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	// Convert to DTO response
	response := dto.FromDashboardData(dashboardData)
	c.JSON(http.StatusOK, response)
}
