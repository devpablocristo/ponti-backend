package category

import (
	"context"

	"github.com/gin-gonic/gin"

	dto "github.com/devpablocristo/ponti-backend/internal/category/handler/dto"
	domain "github.com/devpablocristo/ponti-backend/internal/category/usecases/domain"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

type UseCasesPort interface {
	CreateCategory(context.Context, *domain.Category) (int64, error)
	ListCategories(context.Context, int, int) ([]domain.Category, int64, error)
	GetCategory(context.Context, int64) (*domain.Category, error)
	UpdateCategory(context.Context, *domain.Category) error
	DeleteCategory(context.Context, int64) error
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

// Handler encapsula las dependencias del handler HTTP de Category.
type Handler struct {
	ucs UseCasesPort
	gsv GinEnginePort
	acf ConfigAPIPort
	mws MiddlewaresEnginePort
}

// NewHandler crea un handler de Category.
func NewHandler(u UseCasesPort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		acf: c,
		mws: m,
	}
}

// Routes registra las rutas del módulo Category.
func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/categories"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	group := r.Group(baseURL)
	{
		group.POST("", h.CreateCategory)
		group.GET("", h.ListCategories)
		group.GET("/:category_id", h.GetCategory)
		group.PUT("/:category_id", h.UpdateCategory)
		group.DELETE("/:category_id", h.DeleteCategory)
	}
}

func (h *Handler) CreateCategory(c *gin.Context) {
	var req dto.CreateCategoryRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	id, err := h.ucs.CreateCategory(c.Request.Context(), req.ToDomain())
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondCreated(c, id)
}

func (h *Handler) ListCategories(c *gin.Context) {
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 1000)
	categories, total, err := h.ucs.ListCategories(c.Request.Context(), page, perPage)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.NewListCategoriesResponse(categories, page, perPage, total))
}

func (h *Handler) GetCategory(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("category_id"), "category_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	cat, err := h.ucs.GetCategory(c.Request.Context(), id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.CategoryFromDomain(cat))
}

func (h *Handler) UpdateCategory(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("category_id"), "category_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req dto.UpdateCategoryRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	if err := h.ucs.UpdateCategory(c.Request.Context(), req.ToDomain(id)); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) DeleteCategory(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("category_id"), "category_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.DeleteCategory(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}
