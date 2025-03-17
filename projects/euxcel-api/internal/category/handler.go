package category

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/euxcel-backend/pkg/types"
	utils "github.com/alphacodinggroup/euxcel-backend/pkg/utils"

	dto "github.com/alphacodinggroup/euxcel-backend/internal/category/handler/dto"
	mdw "github.com/alphacodinggroup/euxcel-backend/pkg/http/middlewares/gin"
	gsv "github.com/alphacodinggroup/euxcel-backend/pkg/http/servers/gin"
)

type Handler struct {
	ucs UseCases
	gsv gsv.Server
	mws *mdw.Middlewares
}

func NewHandler(s gsv.Server, u UseCases, m *mdw.Middlewares) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		mws: m,
	}
}

func (h *Handler) Routes() {
	router := h.gsv.GetRouter()

	apiVersion := h.gsv.GetApiVersion()
	apiBase := "/api/" + apiVersion + "/categories"
	publicPrefix := apiBase + "/public"
	// validatedPrefix := apiBase + "/validated"
	protectedPrefix := apiBase + "/protected"

	public := router.Group(publicPrefix)
	{
		public.POST("", h.CreateCategory)
		public.GET("", h.ListCategories)
		public.GET("/:id", h.GetCategory)
		public.PUT("/:id", h.UpdateCategory)
		public.DELETE("/:id", h.DeleteCategory)
	}

	// Protected routes
	protected := router.Group(protectedPrefix)
	{
		protected.Use(h.mws.Protected...)
		protected.GET("/ping", h.ProtectedPing) // Protected test endpoint
	}
}

func (h *Handler) ProtectedPing(c *gin.Context) {
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Protected Pong!",
	})
}

func (h *Handler) CreateCategory(c *gin.Context) {
	var req dto.CreateCategory
	if err := utils.ValidateRequest(c, &req); err != nil {
		apiErr, _ := types.NewAPIError(err)
		// Include error detail in meta.
		c.Error(apiErr).SetMeta(map[string]any{
			"details": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	newCategoryID, err := h.ucs.CreateCategory(ctx, req.ToDomain())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.CreateCategoryResponse{
		Message:    "Category created successfully",
		CategoryID: newCategoryID,
	})
}

func (h *Handler) ListCategories(c *gin.Context) {
	cats, err := h.ucs.ListCategory(c.Request.Context())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, cats)
}

func (h *Handler) GetCategory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid category id"})
		return
	}

	cat, err := h.ucs.GetCategory(c.Request.Context(), id)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, cat)
}

func (h *Handler) UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid category id"})
		return
	}
	var req dto.Category
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid payload"})
		return
	}
	req.ID = id
	if err := h.ucs.UpdateCategory(c.Request.Context(), req.ToDomain()); err != nil {
		apiErr, _ := types.NewAPIError(err)
		// Include details such as "category with id X does not exist" in the response meta.
		c.Error(apiErr).SetMeta(map[string]any{
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{
		Message: "Category updated successfully",
	})
}

func (h *Handler) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid category id"})
		return
	}
	if err := h.ucs.DeleteCategory(c.Request.Context(), id); err != nil {
		apiErr, _ := types.NewAPIError(err)
		// Include details such as "category with id X does not exist" in the response meta.
		c.Error(apiErr).SetMeta(map[string]any{
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{
		Message: "Category deleted successfully",
	})
}
