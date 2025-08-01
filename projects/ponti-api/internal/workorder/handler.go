package workorder

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/handler/dto"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/usecases/domain"
)

type UseCasesPort interface {
	CreateWorkorder(context.Context, *domain.Workorder) (string, error)
	GetWorkorderByNumber(context.Context, string) (*domain.Workorder, error)
	DuplicateWorkorder(context.Context, string) (string, error)
	UpdateWorkorder(context.Context, *domain.Workorder) error
	DeleteWorkorder(context.Context, string) error
	ListWorkorders(context.Context, domain.WorkorderFilter, types.Input) ([]domain.WorkorderListElement, types.PageInfo, error)
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
	base := h.acf.APIBaseURL() + "/workorders"
	grp := r.Group(base)
	{
		grp.POST("", h.CreateWorkorder)
		grp.GET("/number/:number", h.GetWorkorderByNumber)
		grp.POST("/:number/duplicate", h.DuplicateWorkorder)
		grp.PUT("/:number", h.UpdateWorkorder)
		grp.DELETE("/:number", h.DeleteWorkorder)
		grp.GET("", h.ListWorkorders)

	}
}

func (h *Handler) CreateWorkorder(c *gin.Context) {
	var req dto.Workorder
	if err := c.ShouldBindJSON(&req); err != nil {
		domErr := types.NewError(types.ErrBadRequest, "invalid request payload", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	num, err := h.ucs.CreateWorkorder(c.Request.Context(), req.ToDomain())
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusCreated, dto.WorkorderResponse{
		Message: "Workorder created",
		Number:  num,
	})
}

func (h *Handler) GetWorkorderByNumber(c *gin.Context) {
	number := c.Param("number")
	wo, err := h.ucs.GetWorkorderByNumber(c.Request.Context(), number)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, dto.FromDomain(wo))
}

func (h *Handler) DuplicateWorkorder(c *gin.Context) {
	orig := c.Param("number")
	newNum, err := h.ucs.DuplicateWorkorder(c.Request.Context(), orig)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusCreated, dto.WorkorderResponse{
		Message: "Workorder duplicated",
		Number:  newNum,
	})
}

func (h *Handler) UpdateWorkorder(c *gin.Context) {
	var req dto.Workorder
	if err := c.ShouldBindJSON(&req); err != nil {
		domErr := types.NewError(types.ErrBadRequest, "invalid request payload", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	req.Number = c.Param("number")
	if err := h.ucs.UpdateWorkorder(c.Request.Context(), req.ToDomain()); err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) DeleteWorkorder(c *gin.Context) {
	number := c.Param("number")
	if err := h.ucs.DeleteWorkorder(c.Request.Context(), number); err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) ListWorkorders(c *gin.Context) {
	filt := parseFilters(c)
	input := types.NewInput(c.Request)

	// Devuelve ([]domain.WorkorderListElement, types.PageInfo, error)
	list, pageInfo, err := h.ucs.ListWorkorders(c.Request.Context(), filt, input)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	// Usamos el helper del DTO para mapear y construir la respuesta
	resp := dto.FromDomainList(pageInfo, list)

	c.JSON(http.StatusOK, resp)
}

// ParseFilters extrae project_id, field_id y state
func parseFilters(c *gin.Context) domain.WorkorderFilter {
	var f domain.WorkorderFilter
	if v := c.Query("project_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.ProjectID = &id
		}
	}
	if v := c.Query("field_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.FieldID = &id
		}
	}
	return f
}
