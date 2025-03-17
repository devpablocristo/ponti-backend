package macrocategory

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/euxcel-backend/pkg/types"

	dto "github.com/alphacodinggroup/euxcel-backend/internal/macrocategory/handler/dto"
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
	apiBase := "/api/" + apiVersion + "/macrocategories"
	publicPrefix := apiBase + "/public"
	// validatedPrefix := apiBase + "/validated"
	protectedPrefix := apiBase + "/protected"

	public := router.Group(publicPrefix)
	{
		public.POST("", h.CreateMacroCategory)
		public.GET("", h.ListMacroCategories)
		public.GET("/:id", h.GetMacroCategory)
		public.PUT("/:id", h.UpdateMacroCategory)
		public.DELETE("/:id", h.DeleteMacroCategory)
	}

	// Protected routes
	protected := router.Group(protectedPrefix)
	{
		protected.Use(h.mws.Protected...)
		protected.GET("/ping", h.ProtectedPing)
	}
}

func (h *Handler) ProtectedPing(c *gin.Context) {
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Protected Pong!",
	})
}

func (h *Handler) CreateMacroCategory(c *gin.Context) {
	var req dto.CreateMacroCategory
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr := types.NewError(types.ErrValidation, "invalid payload", err)
		c.JSON(http.StatusBadRequest, apiErr.ToJSON())
		return
	}

	ctx := c.Request.Context()
	newID, err := h.ucs.CreateMacroCategory(ctx, req.ToDomain())
	if err != nil {
		apiErr, code := types.NewAPIError(err)
		c.JSON(code, apiErr.ToResponse())
		return
	}

	c.JSON(http.StatusCreated, dto.CreateMacroCategoryResponse{
		Message:         "Macro category created successfully",
		MacroCategoryID: newID,
	})
}

func (h *Handler) ListMacroCategories(c *gin.Context) {
	list, err := h.ucs.ListMacroCategory(c.Request.Context())
	if err != nil {
		apiErr, code := types.NewAPIError(err)
		c.JSON(code, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) GetMacroCategory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid macro category id"})
		return
	}
	mc, err := h.ucs.GetMacroCategory(c.Request.Context(), id)
	if err != nil {
		apiErr, code := types.NewAPIError(err)
		c.JSON(code, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, mc)
}

func (h *Handler) UpdateMacroCategory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid macro category id"})
		return
	}
	var req dto.MacroCategory
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid payload"})
		return
	}
	req.ID = id
	if err := h.ucs.UpdateMacroCategory(c.Request.Context(), req.ToDomain()); err != nil {
		apiErr, code := types.NewAPIError(err)
		c.JSON(code, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{
		Message: "Macro category updated successfully",
	})
}

func (h *Handler) DeleteMacroCategory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid macro category id"})
		return
	}
	if err := h.ucs.DeleteMacroCategory(c.Request.Context(), id); err != nil {
		apiErr, code := types.NewAPIError(err)
		c.JSON(code, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{
		Message: "Macro category deleted successfully",
	})
}
