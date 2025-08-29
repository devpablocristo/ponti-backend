package dashboard

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/handler/dto"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
)

type UseCasesPort interface {
	GetDashboard(context.Context, domain.DashboardFilter) (*domain.DashboardPayload, error)
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
	var f domain.DashboardFilter

	// Parse customer_ids array parameter
	if v := c.Query("customer_ids"); v != "" {
		ids, err := parseInt64Array(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid customer_ids format"})
			return
		}
		f.CustomerIDs = ids
	}

	// Parse project_ids array parameter
	if v := c.Query("project_ids"); v != "" {
		ids, err := parseInt64Array(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid project_ids format"})
			return
		}
		f.ProjectIDs = ids
	}

	// Parse campaign_ids array parameter
	if v := c.Query("campaign_ids"); v != "" {
		ids, err := parseInt64Array(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid campaign_ids format"})
			return
		}
		f.CampaignIDs = ids
	}

	// Parse field_ids array parameter
	if v := c.Query("field_ids"); v != "" {
		ids, err := parseInt64Array(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid field_ids format"})
			return
		}
		f.FieldIDs = ids
	}

	// Sin límites - mostrar todos los datos
	f.Limit = 0
	f.Offset = 0

	// Obtener el dashboard desde la vista SQL
	dashboard, err := h.ucs.GetDashboard(c.Request.Context(), f)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	// Convertir directamente desde el dominio a DTO (ya incluye redondeo a 3 decimales)
	response := dto.FromDashboardPayload(dashboard)

	c.JSON(http.StatusOK, response)
}

// parseInt64Array parses a comma-separated string of integers into a slice of int64
func parseInt64Array(s string) ([]int64, error) {
	if s == "" {
		return nil, nil
	}

	parts := strings.Split(s, ",")
	ids := make([]int64, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		id, err := strconv.ParseInt(part, 10, 64)
		if err != nil || id <= 0 {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}
