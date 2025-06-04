package field

import (
	"context"
	"net/http"
	"strconv"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	dto "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/handler/dto"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	"github.com/gin-gonic/gin"
)

type UseCasesPort interface {
	CreateField(ctx context.Context, f *domain.Field) (int64, error)
	ListFields(ctx context.Context) ([]domain.Field, error)
	GetField(ctx context.Context, id int64) (*domain.Field, error)
	UpdateField(ctx context.Context, f *domain.Field) error
	DeleteField(ctx context.Context, id int64) error
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
	baseURL := h.acf.APIBaseURL() + "/fields"

	public := r.Group(baseURL)
	{
		public.POST("", h.CreateField)
		public.GET("", h.ListFields)
		public.GET("/:id", h.GetField)
		public.PUT("/:id", h.UpdateField)
		public.DELETE("/:id", h.DeleteField)
	}
}

func (h *Handler) CreateField(c *gin.Context) {
	var req dto.Field
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	id, err := h.ucs.CreateField(c.Request.Context(), req.ToDomain())
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.CreateFieldResponse{Message: "Field created", ID: id})
}

func (h *Handler) ListFields(c *gin.Context) {
	fields, err := h.ucs.ListFields(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	dtos := make([]dto.Field, len(fields))
	for i, f := range fields {
		dtos[i] = dto.FromDomain(f)
	}
	c.JSON(http.StatusOK, dtos)
}

func (h *Handler) GetField(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	f, err := h.ucs.GetField(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.FromDomain(*f))
}

func (h *Handler) UpdateField(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req dto.UpdateField
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	dom := req.ToDomain()
	dom.ID = id
	if err := h.ucs.UpdateField(c.Request.Context(), dom); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Field updated"})
}

func (h *Handler) DeleteField(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.ucs.DeleteField(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Field deleted"})
}
