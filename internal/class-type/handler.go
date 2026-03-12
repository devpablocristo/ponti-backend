package classtype

import (
	"context"

	"github.com/gin-gonic/gin"

	dto "github.com/alphacodinggroup/ponti-backend/internal/class-type/handler/dto"
	domain "github.com/alphacodinggroup/ponti-backend/internal/class-type/usecases/domain"
	sharedhandlers "github.com/alphacodinggroup/ponti-backend/internal/shared/handlers"
)

type UseCasesPort interface {
	CreateClassType(context.Context, *domain.ClassType) (int64, error)
	ListClassTypes(context.Context, int, int) ([]domain.ClassType, int64, error)
	GetClassType(context.Context, int64) (*domain.ClassType, error)
	UpdateClassType(context.Context, *domain.ClassType) error
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
	baseURL := h.acf.APIBaseURL() + "/types"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	group := r.Group(baseURL)
	{
		group.POST("", h.CreateClassType)
		group.GET("", h.ListClassTypes)
		group.GET("/:class_type_id", h.GetClassType)
		group.PUT("/:class_type_id", h.UpdateClassType)
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

func (h *Handler) GetClassType(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("class_type_id"), "class_type_id")
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
	id, err := sharedhandlers.ParseParamID(c.Param("class_type_id"), "class_type_id")
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
	id, err := sharedhandlers.ParseParamID(c.Param("class_type_id"), "class_type_id")
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
