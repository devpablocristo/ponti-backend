package classtype

import (
	"context"

	ginmw "github.com/devpablocristo/core/http/gin/go"
	"github.com/gin-gonic/gin"

	dto "github.com/devpablocristo/ponti-backend/internal/class-type/handler/dto"
	domain "github.com/devpablocristo/ponti-backend/internal/class-type/usecases/domain"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

type UseCasesPort interface {
	CreateClassType(context.Context, *domain.ClassType) (int64, error)
	ListClassTypes(context.Context, int, int) ([]domain.ClassType, int64, error)
	ListArchivedClassTypes(context.Context, int, int) ([]domain.ClassType, int64, error)
	GetClassType(context.Context, int64) (*domain.ClassType, error)
	UpdateClassType(context.Context, *domain.ClassType) error
	ArchiveClassType(context.Context, int64) error
	RestoreClassType(context.Context, int64) error
	HardDeleteClassType(context.Context, int64) error
	DeleteClassType(context.Context, int64) error
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

func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/types"

	group := r.Group(baseURL, h.mws.GetValidation()...)
	{
		group.POST("", h.CreateClassType)
		group.GET("", h.ListClassTypes)
		group.GET("/archived", h.ListArchivedClassTypes)
		group.GET("/:class_type_id", h.GetClassType)
		group.PUT("/:class_type_id", h.UpdateClassType)
		group.POST("/:class_type_id/archive", h.ArchiveClassType)
		group.POST("/:class_type_id/restore", h.RestoreClassType)
		group.DELETE("/:class_type_id/hard", h.HardDeleteClassType)
		group.DELETE("/:class_type_id", h.DeleteClassType)
	}
}

func (h *Handler) CreateClassType(c *gin.Context) {
	var req dto.CreateClassTypeRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	id, err := h.ucs.CreateClassType(c.Request.Context(), req.ToDomain())
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondCreated(c, id)
}

func (h *Handler) ListClassTypes(c *gin.Context) {
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 1000)
	items, total, err := h.ucs.ListClassTypes(c.Request.Context(), page, perPage)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.NewListClassTypesResponse(items, page, perPage, total))
}

func (h *Handler) ListArchivedClassTypes(c *gin.Context) {
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 1000)
	items, total, err := h.ucs.ListArchivedClassTypes(c.Request.Context(), page, perPage)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.NewListClassTypesResponse(items, page, perPage, total))
}

func (h *Handler) GetClassType(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "class_type_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	ct, err := h.ucs.GetClassType(c.Request.Context(), id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.ClassTypeFromDomain(ct))
}

func (h *Handler) UpdateClassType(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "class_type_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req dto.UpdateClassTypeRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	if err := h.ucs.UpdateClassType(c.Request.Context(), req.ToDomain(id)); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) DeleteClassType(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "class_type_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.DeleteClassType(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) ArchiveClassType(c *gin.Context) {
	h.runClassTypeIDAction(c, h.ucs.ArchiveClassType)
}

func (h *Handler) RestoreClassType(c *gin.Context) {
	h.runClassTypeIDAction(c, h.ucs.RestoreClassType)
}

func (h *Handler) HardDeleteClassType(c *gin.Context) {
	h.runClassTypeIDAction(c, h.ucs.HardDeleteClassType)
}

func (h *Handler) runClassTypeIDAction(c *gin.Context, action func(context.Context, int64) error) {
	id, err := ginmw.ParseParamID(c, "class_type_id")
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
