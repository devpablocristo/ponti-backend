package workorder

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/handler/dto"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/usecases/domain"
)

type UseCasesPort interface {
	CreateWorkOrder(context.Context, *domain.WorkOrder) (string, error)
	GetOrder(context.Context, string) (*domain.WorkOrder, error)
	DuplicateOrder(context.Context, string) (string, error)
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
	return &Handler{
		ucs: u,
		gsv: s,
		acf: c,
		mws: m,
	}
}

func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/workorders"

	group := r.Group(baseURL)
	{
		group.POST("", h.CreateWorkOrder)
		group.GET(":number", h.GetOrder)
		group.POST(":number/duplicar", h.DuplicateOrder)
	}
}

func (h *Handler) CreateWorkOrder(c *gin.Context) {
	var req dto.WorkOrder
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	number, err := h.ucs.CreateWorkOrder(c.Request.Context(), req.ToDomain())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.WorkOrderResponse{
		Message: "WorkOrder created",
		Number:  number,
	})
}

func (h *Handler) GetOrder(c *gin.Context) {
	number := c.Param("number")
	ord, err := h.ucs.GetOrder(c.Request.Context(), number)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ord)
}

func (h *Handler) DuplicateOrder(c *gin.Context) {
	orig := c.Param("number")
	newNum, err := h.ucs.DuplicateOrder(c.Request.Context(), orig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.WorkOrderResponse{
		Message: "WorkOrder duplicated",
		Number:  newNum,
	})
}
