package dashboard

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/handler/dto"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
)

type UseCasesPort interface {
	GetDashboard(context.Context, domain.DashboardFilter) (*domain.Dashboard, error)
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

type Handler struct {
	ucs UseCasesPort
	gsv GinEnginePort
	acf ConfigAPIPort
	mws MiddlewaresEnginePort
}

func NewHandler(u UseCasesPort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *Handler {
	return &Handler{ucs: u, gsv: s, acf: c, mws: m}
}

func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	base := h.acf.APIBaseURL() + "/dashboard"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	grp := r.Group(base)
	{
		grp.GET("", h.GetDashboard)
	}
}

func (h *Handler) GetDashboard(c *gin.Context) {
	// Crear DTO de filtro desde los query parameters
	filterDTO := dto.DashboardFilterRequest{}

	// Parse customer_id parameter (solo 1 ID)
	if v := c.Query("customer_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid customer_id format"})
			return
		}
		filterDTO.CustomerIDs = []int64{id}
	}

	// Parse project_id parameter (solo 1 ID)
	if v := c.Query("project_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid project_id format"})
			return
		}
		filterDTO.ProjectIDs = []int64{id}
	}

	// Parse campaign_id parameter (solo 1 ID)
	if v := c.Query("campaign_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid campaign_id format"})
			return
		}
		filterDTO.CampaignIDs = []int64{id}
	}

	// Parse field_id parameter (solo 1 ID)
	if v := c.Query("field_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid field_id format"})
			return
		}
		filterDTO.FieldIDs = []int64{id}
	}

	// Usar el mapper para convertir DTO a entidad de dominio
	f := dto.ToDashboardFilter(filterDTO)

	// Obtener el dashboard desde la vista SQL
	dashboard, err := h.ucs.GetDashboard(c.Request.Context(), f)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	// Convertir directamente desde el dominio a DTO (ya incluye redondeo a 3 decimales)
	response := dto.FromDashboard(dashboard)

	c.JSON(http.StatusOK, response)
}
