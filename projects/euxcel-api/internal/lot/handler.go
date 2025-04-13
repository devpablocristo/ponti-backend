package lot

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/euxcel-backend/pkg/types"
	utils "github.com/alphacodinggroup/euxcel-backend/pkg/utils"

	mdw "github.com/alphacodinggroup/euxcel-backend/pkg/http/middlewares/gin"
	gsv "github.com/alphacodinggroup/euxcel-backend/pkg/http/servers/gin"
	dto "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/lot/handler/dto"
)

// Handler encapsulates all dependencies for the Lot HTTP handler.
type Handler struct {
	ucs UseCases
	gsv gsv.Server
	mws *mdw.Middlewares
}

// NewHandler creates a new Lot handler.
func NewHandler(s gsv.Server, u UseCases, m *mdw.Middlewares) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		mws: m,
	}
}

// Routes registers all lot routes.
func (h *Handler) Routes() {
	router := h.gsv.GetRouter()

	apiVersion := h.gsv.GetApiVersion()
	apiBase := "/api/" + apiVersion + "/lots"
	publicPrefix := apiBase + "/public"
	protectedPrefix := apiBase + "/protected"

	public := router.Group(publicPrefix)
	{
		public.POST("", h.CreateLot)       // Create a lot
		public.GET("", h.ListLots)         // List all lots
		public.GET("/:id", h.GetLot)       // Get a lot by ID
		public.PUT("/:id", h.UpdateLot)    // Update a lot
		public.DELETE("/:id", h.DeleteLot) // Delete a lot
	}

	// Protected routes.
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

// CreateLot handles the creation of a new lot.
func (h *Handler) CreateLot(c *gin.Context) {
	var req dto.CreateLot
	if err := utils.ValidateRequest(c, &req); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	newID, err := h.ucs.CreateLot(ctx, req.ToDomain())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.CreateLotResponse{
		Message: "Lot created successfully",
		ID:      newID,
	})
}

// ListLots retrieves all lots.
func (h *Handler) ListLots(c *gin.Context) {
	lots, err := h.ucs.ListLots(c.Request.Context())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lots)
}

// GetLot retrieves a lot by its ID.
func (h *Handler) GetLot(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid lot id"})
		return
	}

	lot, err := h.ucs.GetLot(c.Request.Context(), id)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, lot)
}

// UpdateLot updates an existing lot.
func (h *Handler) UpdateLot(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid lot id"})
		return
	}
	var req dto.Lot
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid payload"})
		return
	}
	req.ID = id
	if err := h.ucs.UpdateLot(c.Request.Context(), req.ToDomain()); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Lot updated successfully"})
}

// DeleteLot deletes a lot by its ID.
func (h *Handler) DeleteLot(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid lot id"})
		return
	}
	if err := h.ucs.DeleteLot(c.Request.Context(), id); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Lot deleted successfully"})
}
