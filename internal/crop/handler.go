package crop

import (
	"context"

	ginmw "github.com/devpablocristo/platform/http/gin/go"
	"github.com/gin-gonic/gin"

	dto "github.com/devpablocristo/ponti-backend/internal/crop/handler/dto"
	domain "github.com/devpablocristo/ponti-backend/internal/crop/usecases/domain"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

type UseCasesPort interface {
	CreateCrop(context.Context, *domain.Crop) (int64, error)
	ListCrops(context.Context, int, int) ([]domain.Crop, int64, error)
	ListArchivedCrops(context.Context, int, int) ([]domain.Crop, int64, error)
	GetCrop(context.Context, int64) (*domain.Crop, error)
	UpdateCrop(context.Context, *domain.Crop) error
	ArchiveCrop(context.Context, int64) error
	RestoreCrop(context.Context, int64) error
	HardDeleteCrop(context.Context, int64) error
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

	public := r.Group(baseURL, h.mws.GetValidation()...)
	{
		public.POST("", h.CreateCrop)
		public.GET("", h.ListCrops)
		public.GET("/archived", h.ListArchivedCrops)
		public.GET("/:crop_id", h.GetCrop)
		public.PUT("/:crop_id", h.UpdateCrop)
		public.POST("/:crop_id/archive", h.ArchiveCrop)
		public.POST("/:crop_id/restore", h.RestoreCrop)
		public.DELETE("/:crop_id/hard", h.HardDeleteCrop)
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

func (h *Handler) ListArchivedCrops(c *gin.Context) {
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 1000)
	crops, total, err := h.ucs.ListArchivedCrops(c.Request.Context(), page, perPage)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.NewListCropsResponse(crops, page, perPage, total))
}

func (h *Handler) GetCrop(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "crop_id")
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
	id, err := ginmw.ParseParamID(c, "crop_id")
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

func (h *Handler) ArchiveCrop(c *gin.Context) {
	h.runCropIDAction(c, h.ucs.ArchiveCrop)
}

func (h *Handler) RestoreCrop(c *gin.Context) {
	h.runCropIDAction(c, h.ucs.RestoreCrop)
}

func (h *Handler) HardDeleteCrop(c *gin.Context) {
	h.runCropIDAction(c, h.ucs.HardDeleteCrop)
}

func (h *Handler) runCropIDAction(c *gin.Context, action func(context.Context, int64) error) {
	id, err := ginmw.ParseParamID(c, "crop_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := action(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}
