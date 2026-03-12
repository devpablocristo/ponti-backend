package crop

import (
	"context"

	"github.com/gin-gonic/gin"

	dto "github.com/alphacodinggroup/ponti-backend/internal/crop/handler/dto"
	domain "github.com/alphacodinggroup/ponti-backend/internal/crop/usecases/domain"
	sharedhandlers "github.com/alphacodinggroup/ponti-backend/internal/shared/handlers"
)

type UseCasesPort interface {
	CreateCrop(context.Context, *domain.Crop) (int64, error)
	ListCrops(context.Context, int, int) ([]domain.Crop, int64, error)
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

// Handler encapsula las dependencias del handler HTTP de Crop.
type Handler struct {
	ucs UseCasesPort
	gsv GinEnginePort
	acf ConfigAPIPort
	mws MiddlewaresEnginePort
}

// NewHandler crea un handler de Crop.
func NewHandler(u UseCasesPort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		acf: c,
		mws: m,
	}
}

// Routes registra las rutas del módulo Crop.
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
	var req dto.CreateCropRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	id, err := h.ucs.CreateCrop(c.Request.Context(), req.ToDomain())
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondCreated(c, id)
}

func (h *Handler) ListCrops(c *gin.Context) {
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 1000)
	crops, total, err := h.ucs.ListCrops(c.Request.Context(), page, perPage)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.NewListCropsResponse(crops, page, perPage, total))
}

func (h *Handler) GetCrop(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("crop_id"), "crop_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	crop, err := h.ucs.GetCrop(c.Request.Context(), id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.CropFromDomain(crop))
}

func (h *Handler) UpdateCrop(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("crop_id"), "crop_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req dto.UpdateCropRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	if err := h.ucs.UpdateCrop(c.Request.Context(), req.ToDomain(id)); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) DeleteCrop(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("crop_id"), "crop_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.DeleteCrop(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}
