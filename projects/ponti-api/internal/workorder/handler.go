package workorder

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	pkgtypes "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/handler/dto"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/usecases/domain"
)

type UseCasesPort interface {
	CreateWorkOrder(context.Context, *domain.WorkOrder) (string, error)
	GetWorkOrder(context.Context, string) (*domain.WorkOrder, error)
	DuplicateWorkOrder(context.Context, string) (string, error)
	UpdateWorkOrder(context.Context, *domain.WorkOrder) error
	DeleteWorkOrder(context.Context, string) error
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
		grp.POST("", h.CreateWorkOrder)
		grp.GET("/:number", h.GetWorkOrder)
		grp.POST("/:number/duplicate", h.DuplicateWorkOrder)
		grp.PUT("/:number", h.UpdateWorkOrder)
		grp.DELETE("/:number", h.DeleteWorkOrder)
	}
}

func (h *Handler) CreateWorkOrder(c *gin.Context) {
	var req dto.WorkOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domErr := pkgtypes.NewError(pkgtypes.ErrBadRequest, "invalid request payload", err)
		apiErr, status := pkgtypes.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	num, err := h.ucs.CreateWorkOrder(c.Request.Context(), req.ToDomain())
	if err != nil {
		apiErr, status := pkgtypes.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusCreated, dto.WorkOrderResponse{
		Message: "WorkOrder created",
		Number:  num,
	})
}

func (h *Handler) GetWorkOrder(c *gin.Context) {
	number := c.Param("number")
	ord, err := h.ucs.GetWorkOrder(c.Request.Context(), number)
	if err != nil {
		apiErr, status := pkgtypes.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, dto.FromDomain(ord))
}

func (h *Handler) DuplicateWorkOrder(c *gin.Context) {
	orig := c.Param("number")
	newNum, err := h.ucs.DuplicateWorkOrder(c.Request.Context(), orig)
	if err != nil {
		apiErr, status := pkgtypes.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusCreated, dto.WorkOrderResponse{
		Message: "WorkOrder duplicated",
		Number:  newNum,
	})
}

func (h *Handler) UpdateWorkOrder(c *gin.Context) {
	var req dto.WorkOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domErr := pkgtypes.NewError(pkgtypes.ErrBadRequest, "invalid request payload", err)
		apiErr, status := pkgtypes.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	req.Number = c.Param("number")
	if err := h.ucs.UpdateWorkOrder(c.Request.Context(), req.ToDomain()); err != nil {
		apiErr, status := pkgtypes.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) DeleteWorkOrder(c *gin.Context) {
	number := c.Param("number")
	if err := h.ucs.DeleteWorkOrder(c.Request.Context(), number); err != nil {
		apiErr, status := pkgtypes.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.Status(http.StatusNoContent)
}
