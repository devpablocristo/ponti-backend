package item

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/euxcel-backend/pkg/types"
	utils "github.com/alphacodinggroup/euxcel-backend/pkg/utils"

	dto "github.com/alphacodinggroup/euxcel-backend/internal/item/handler/dto"
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
	apiBase := "/api/" + apiVersion + "/items"
	publicPrefix := apiBase + "/public"
	// validatedPrefix := apiBase + "/validated"
	protectedPrefix := apiBase + "/protected"

	public := router.Group(publicPrefix)
	{
		public.POST("", h.CreateItem)       // Crear un ítem
		public.GET("", h.ListItems)         // Listar todos los ítems
		public.GET("/:id", h.GetItem)       // Obtener un ítem por ID
		public.PUT("/:id", h.UpdateItem)    // Actualizar un ítem
		public.DELETE("/:id", h.DeleteItem) // Eliminar un ítem
	}

	// validated := router.Group(validatedPrefix)
	// {
	// 	// Aplicar middleware de validación de credenciales
	// 	validated.Use(h.mws.Validated...)
	// }

	// Rutas protegidas
	protected := router.Group(protectedPrefix)
	{
		protected.Use(h.mws.Protected...)
		protected.GET("/ping", h.ProtectedPing) // Endpoint de prueba protegido
	}
}

func (h *Handler) ProtectedPing(c *gin.Context) {
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Protected Pong!",
	})
}

func (h *Handler) CreateItem(c *gin.Context) {
	var req dto.CreateItem
	if err := utils.ValidateRequest(c, &req); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	ctx := c.Request.Context()
	item, err := req.ToDomain()
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	newItemID, err := h.ucs.CreateItem(ctx, item)
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	c.JSON(http.StatusCreated, dto.CreateItemResponse{
		Message: "Item created successfully",
		ItemID:  newItemID,
	})
}

func (h *Handler) ListItems(c *gin.Context) {
	items, err := h.ucs.ListItems(c.Request.Context())
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler) GetItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "invalid item id",
		})
		return
	}

	item, err := h.ucs.GetItem(c.Request.Context(), id)
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) UpdateItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "invalid item id",
		})
		return
	}

	var req dto.Item
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	updatedItem, err := req.ToDomain()
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}
	// Aseguramos que el ID del item actualizado coincida con el parámetro de la URL.
	updatedItem.ID = id

	if err := h.ucs.UpdateItem(c.Request.Context(), updatedItem); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Item updated successfully",
	})
}

func (h *Handler) DeleteItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "invalid item id",
		})
		return
	}

	if err := h.ucs.DeleteItem(c.Request.Context(), id); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Item deleted successfully",
	})
}
