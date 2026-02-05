package category

import (
	"context"
	"net/http"

	dto "github.com/alphacodinggroup/ponti-backend/internal/category/handler/dto"
	domain "github.com/alphacodinggroup/ponti-backend/internal/category/usecases/domain"
	sharedhandlers "github.com/alphacodinggroup/ponti-backend/internal/shared/handlers"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/gin-gonic/gin"
)

// UseCasesPort expects domain.Category, not dto.Category
type UseCasesPort interface {
	ListCategories(context.Context) ([]domain.Category, error)
	CreateCategory(context.Context, *domain.Category) (int64, error)
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

// Handler encapsulates all dependencies for the Category HTTP handler.
type Handler struct {
	categoryUC UseCasesPort
	gsv        GinEnginePort
	acf        ConfigAPIPort
	mws        MiddlewaresEnginePort
}

// NewHandler creates a new Category handler.
func NewHandler(u UseCasesPort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *Handler {
	return &Handler{
		categoryUC: u,
		gsv:        s,
		acf:        c,
		mws:        m,
	}
}

// Routes registers all category routes.
func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/categories"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	group := r.Group(baseURL)
	{
		group.GET("", h.ListCategories)
		group.POST("", h.CreateCategory)
		group.PUT("/:category_id", h.UpdateCategory)
		group.DELETE("/:category_id", h.DeleteCategory)
	}
}

func (h *Handler) ListCategories(c *gin.Context) {
	categories, err := h.categoryUC.ListCategories(c.Request.Context())
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	out := make([]dto.Category, len(categories))
	for i := range categories {
		out[i] = *dto.FromDomain(&categories[i])
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) CreateCategory(c *gin.Context) {
	var req dto.Category
	if err := c.ShouldBindJSON(&req); err != nil {
		domErr := types.NewError(types.ErrBadRequest, "invalid request payload", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	newID, err := h.categoryUC.CreateCategory(c.Request.Context(), req.ToDomain())
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Category created successfully", "id": newID})
}

func (h *Handler) UpdateCategory(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("category_id"), "category_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req dto.Category
	if err := c.ShouldBindJSON(&req); err != nil {
		domErr := types.NewError(types.ErrBadRequest, "invalid request payload", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	dom := req.ToDomain()
	dom.ID = id
	if err := h.categoryUC.UpdateCategory(c.Request.Context(), dom); err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Category updated successfully"})
}

func (h *Handler) DeleteCategory(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("category_id"), "category_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.categoryUC.DeleteCategory(c.Request.Context(), id); err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Category deleted successfully"})
}
