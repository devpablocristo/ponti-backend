package dashboard

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	dto "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/handler/dto"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
)

type UseCasesPort interface {
	CreateDashboard(context.Context, *domain.Dashboard) (int64, error)
	ListDashboards(context.Context) ([]domain.Dashboard, error)
	GetDashboard(context.Context, int64) (*domain.Dashboard, error)
	UpdateDashboard(context.Context, *domain.Dashboard) error
	DeleteDashboard(context.Context, int64) error
	GetDashboardData(context.Context, domain.DashboardFilter) (*domain.DashboardData, error)
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

func (h *Handler) CreateDashboard(c *gin.Context) {
	var req dto.CreateDashboard
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	newID, err := h.ucs.CreateDashboard(ctx, req.ToDomain())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.CreateDashboardResponse{
		Message: "Dashboard created successfully",
		ID:      newID,
	})
}

// ListDashboards retrieves all dashboards.
func (h *Handler) ListDashboards(c *gin.Context) {
	dashboards, err := h.ucs.ListDashboards(c.Request.Context())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	resp := dto.NewListDashboardsResponse(dashboards)
	c.JSON(http.StatusOK, resp)
}

// GetDashboard retrieves dashboard data based on query parameters.
func (h *Handler) GetDashboard(c *gin.Context) {
	// Parse query parameters for filters
	var filter domain.DashboardFilter

	if customerIDs := c.QueryArray("customer_ids"); len(customerIDs) > 0 {
		for _, idStr := range customerIDs {
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				filter.CustomerIDs = append(filter.CustomerIDs, id)
			}
		}
	}

	if projectIDs := c.QueryArray("project_ids"); len(projectIDs) > 0 {
		for _, idStr := range projectIDs {
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				filter.ProjectIDs = append(filter.ProjectIDs, id)
			}
		}
	}

	if campaignIDs := c.QueryArray("campaign_ids"); len(campaignIDs) > 0 {
		for _, idStr := range campaignIDs {
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				filter.CampaignIDs = append(filter.CampaignIDs, id)
			}
		}
	}

	if fieldIDs := c.QueryArray("field_ids"); len(fieldIDs) > 0 {
		for _, idStr := range fieldIDs {
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				filter.FieldIDs = append(filter.FieldIDs, id)
			}
		}
	}

	// Get dashboard data
	dashboardData, err := h.ucs.GetDashboardData(c.Request.Context(), filter)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	// Convert to DTO response
	response := dto.FromDashboardData(dashboardData)
	c.JSON(http.StatusOK, response)
}

// UpdateDashboard updates an existing dashboard.
func (h *Handler) UpdateDashboard(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid dashboard id"})
		return
	}
	var req dto.Dashboard
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid payload"})
		return
	}
	req.ID = id
	if err := h.ucs.UpdateDashboard(c.Request.Context(), req.ToDomain()); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Dashboard updated successfully"})
}

// DeleteDashboard deletes a dashboard by its ID.
func (h *Handler) DeleteDashboard(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid dashboard id"})
		return
	}
	if err := h.ucs.DeleteDashboard(c.Request.Context(), id); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Dashboard updated successfully"})
}
