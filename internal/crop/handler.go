package crop

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	dto "github.com/alphacodinggroup/ponti-backend/internal/crop/handler/dto"
	domain "github.com/alphacodinggroup/ponti-backend/internal/crop/usecases/domain"
)

type UseCasesPort interface {
	CreateCrop(context.Context, *domain.Crop) (int64, error)
	ListCrops(context.Context) ([]domain.Crop, error)
	GetCrop(context.Context, int64) (*domain.Crop, error)
	UpdateCrop(context.Context, *domain.Crop) error
	DeleteCrop(context.Context, int64) error
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

// Handler encapsulates all dependencies for the Project HTTP handler.
type Handler struct {
	ucs UseCasesPort
	gsv GinEnginePort
	acf ConfigAPIPort
	mws MiddlewaresEnginePort
}

// NewHandler creates a new Project handler.
func NewHandler(u UseCasesPort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		acf: c,
		mws: m,
	}
}

// Routes registers all project routes.
func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/crops"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	public := r.Group(baseURL)
	{
		public.POST("", h.CreateCrop)
		public.GET("", h.ListCrops)
		public.GET("/:crop_id", h.GetCrop)
		public.PUT("/:crop_id", h.UpdateCrop)
		public.DELETE("/:crop_id", h.DeleteCrop)
	}
}

func (h *Handler) CreateCrop(c *gin.Context) {
	var req dto.CreateCrop
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	newID, err := h.ucs.CreateCrop(ctx, req.ToDomain())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.CreateCropResponse{
		Message: "Crop created successfully",
		ID:      newID,
	})
}

// ListCrops retrieves all crops.
func (h *Handler) ListCrops(c *gin.Context) {
	crops, err := h.ucs.ListCrops(c.Request.Context())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	resp := dto.NewListCropsResponse(crops)
	c.JSON(http.StatusOK, resp)
}

// GetCrop retrieves a crop by its ID.
func (h *Handler) GetCrop(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("crop_id"), 10, 64)
	if err != nil {
		domErr := types.NewError(types.ErrInvalidID, "invalid crop id", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	crop, err := h.ucs.GetCrop(c.Request.Context(), id)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	c.JSON(http.StatusOK, crop)
}

// UpdateCrop updates an existing crop.
func (h *Handler) UpdateCrop(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("crop_id"), 10, 64)
	if err != nil {
		domErr := types.NewError(types.ErrInvalidID, "invalid crop id", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	var req dto.Crop
	if err := c.ShouldBindJSON(&req); err != nil {
		domErr := types.NewError(types.ErrBadRequest, "invalid request payload", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	req.ID = id
	if err := h.ucs.UpdateCrop(c.Request.Context(), req.ToDomain()); err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Crop updated successfully"})
}

// DeleteCrop deletes a crop by its ID.
func (h *Handler) DeleteCrop(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("crop_id"), 10, 64)
	if err != nil {
		domErr := types.NewError(types.ErrInvalidID, "invalid crop id", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	if err := h.ucs.DeleteCrop(c.Request.Context(), id); err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Crop deleted successfully"})
}
