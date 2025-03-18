package supplier

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/euxcel-backend/pkg/types"
	utils "github.com/alphacodinggroup/euxcel-backend/pkg/utils"

	mdw "github.com/alphacodinggroup/euxcel-backend/pkg/http/middlewares/gin"
	gsv "github.com/alphacodinggroup/euxcel-backend/pkg/http/servers/gin"
	dto "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/supplier/handler/dto"
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
	apiBase := "/api/" + apiVersion + "/supplier"
	publicPrefix := apiBase + "/public"
	protectedPrefix := apiBase + "/protected"

	public := router.Group(publicPrefix)
	{
		public.POST("", h.CreateSupplier)
		public.GET("", h.ListSuppliers)
		public.GET("/:id", h.GetSupplier)
		public.PUT("/:id", h.UpdateSupplier)
		public.DELETE("/:id", h.DeleteSupplier)
	}

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

func (h *Handler) CreateSupplier(c *gin.Context) {
	var req dto.CreateSupplier
	if err := utils.ValidateRequest(c, &req); err != nil {
		apiErr, _ := types.NewAPIError(err)
		// Include error detail in meta.
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	newSupplierID, err := h.ucs.CreateSupplier(ctx, req.ToDomain())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.CreateSupplierResponse{
		Message:    "Supplier created successfully",
		SupplierID: newSupplierID,
	})
}

func (h *Handler) ListSuppliers(c *gin.Context) {
	suppliers, err := h.ucs.ListSuppliers(c.Request.Context())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, suppliers)
}

func (h *Handler) GetSupplier(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid supplier id"})
		return
	}
	s, err := h.ucs.GetSupplier(c.Request.Context(), id)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *Handler) UpdateSupplier(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid supplier id"})
		return
	}
	var req dto.Supplier
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	updatedSupplier := req.ToDomain()
	// Ensure that the updated supplier's ID matches the URL parameter.
	updatedSupplier.ID = id
	if err := h.ucs.UpdateSupplier(c.Request.Context(), updatedSupplier); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, types.MessageResponse{Message: "Supplier updated successfully"})
}

func (h *Handler) DeleteSupplier(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid supplier id"})
		return
	}
	if err := h.ucs.DeleteSupplier(c.Request.Context(), id); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, types.MessageResponse{Message: "Supplier deleted successfully"})
}
