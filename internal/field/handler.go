package field

import (
	"context"

	ginmw "github.com/devpablocristo/platform/http/gin/go"
	"github.com/gin-gonic/gin"

	dto "github.com/devpablocristo/ponti-backend/internal/field/handler/dto"
	domain "github.com/devpablocristo/ponti-backend/internal/field/usecases/domain"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

type UseCasesPort interface {
	CreateField(ctx context.Context, f *domain.Field) (int64, error)
	ListFields(ctx context.Context, page, perPage int) ([]domain.Field, int64, error)
	ListArchivedFields(ctx context.Context, page, perPage int) ([]domain.Field, int64, error)
	GetField(ctx context.Context, id int64) (*domain.Field, error)
	UpdateField(ctx context.Context, f *domain.Field) error
	ArchiveField(ctx context.Context, id int64) error
	RestoreField(ctx context.Context, id int64) error
	HardDeleteField(ctx context.Context, id int64) error
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

func (h *Handler) runFieldIDAction(c *gin.Context, action func(context.Context, int64) error) {
	id, err := ginmw.ParseParamID(c, "field_id")
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

func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/fields"

	public := r.Group(baseURL, h.mws.GetValidation()...)
	{
		public.POST("", h.CreateField)
		public.GET("", h.ListFields)
		public.GET("/archived", h.ListArchivedFields)
		public.GET("/:field_id", h.GetField)
		public.PUT("/:field_id", h.UpdateField)
		public.POST("/:field_id/archive", h.ArchiveField)
		public.POST("/:field_id/restore", h.RestoreField)
		public.DELETE("/:field_id/hard", h.HardDeleteField)
	}
}

func (h *Handler) CreateField(c *gin.Context) {
	var req dto.CreateFieldRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	id, err := h.ucs.CreateField(c.Request.Context(), req.ToDomain())
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondCreated(c, id)
}

func (h *Handler) ListFields(c *gin.Context) {
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 1000)
	fields, total, err := h.ucs.ListFields(c.Request.Context(), page, perPage)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.NewListFieldsResponse(fields, page, perPage, total))
}

func (h *Handler) GetField(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "field_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	f, err := h.ucs.GetField(c.Request.Context(), id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.FieldFromDomain(f))
}

func (h *Handler) UpdateField(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "field_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req dto.UpdateFieldRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	if err := h.ucs.UpdateField(c.Request.Context(), req.ToDomain(id)); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) ListArchivedFields(c *gin.Context) {
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 1000)
	fields, total, err := h.ucs.ListArchivedFields(c.Request.Context(), page, perPage)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.NewListFieldsResponse(fields, page, perPage, total))
}

func (h *Handler) HardDeleteField(c *gin.Context) {
	h.runFieldIDAction(c, h.ucs.HardDeleteField)
}

func (h *Handler) ArchiveField(c *gin.Context) {
	h.runFieldIDAction(c, h.ucs.ArchiveField)
}

func (h *Handler) RestoreField(c *gin.Context) {
	h.runFieldIDAction(c, h.ucs.RestoreField)
}
