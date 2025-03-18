package group

import (
	"net/http"

	"github.com/gin-gonic/gin"

	mdw "github.com/alphacodinggroup/euxcel-backend/pkg/http/middlewares/gin"
	gsv "github.com/alphacodinggroup/euxcel-backend/pkg/http/servers/gin"
	types "github.com/alphacodinggroup/euxcel-backend/pkg/types"
	utils "github.com/alphacodinggroup/euxcel-backend/pkg/utils"

	dto "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/group/handler/dto"
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
	apiBase := "/api/" + apiVersion + "/groups"
	publicPrefix := apiBase + "/public"
	validatedPrefix := apiBase + "/validated"
	protectedPrefix := apiBase + "/protected"

	// Rutas públicas
	public := router.Group(publicPrefix)
	{
		public.POST("", h.CreateGroup)
		public.GET("", h.ListGroups)
		public.GET("/:id", h.GetGroupByID)
		public.PUT("/:id", h.UpdateGroup)
		public.DELETE("/:id", h.DeleteGroup)
	}

	validated := router.Group(validatedPrefix)
	{
		// Aplicar middleware de validación de credenciales
		validated.Use(h.mws.Validated...)
		// Puedes añadir rutas aquí si es necesario
	}

	// Rutas protegidas
	protected := router.Group(protectedPrefix)
	{
		protected.Use(h.mws.Protected...)

		protected.GET("/ping", h.ProtectedPing)

		protected.POST("", h.CreateGroup)
		protected.GET("", h.ListGroups)
		protected.GET("/:id", h.GetGroupByID)
		protected.PUT("/:id", h.UpdateGroup)
		protected.DELETE("/:id", h.DeleteGroup)
	}
}

func (h *Handler) ProtectedPing(c *gin.Context) {
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Protected Pong!",
	})
}

func (h *Handler) CreateGroup(c *gin.Context) {
	// cuil, err := mdw.ExtractClaim(c, "sub", "")
	// if err != nil {
	// 	c.JSON(http.StatusUnauthorized, sup.ErrorResponse{
	// 		Error: err.Error(),
	// 	})
	// 	return
	// }

	var req dto.CreateGroup
	if err := utils.ValidateRequest(c, &req); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	ctx := c.Request.Context()
	if err := h.ucs.CreateGroup(ctx, req.ToDomain()); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Group created successfully",
	})
}

func (h *Handler) ListGroups(c *gin.Context) {
	groups, err := h.ucs.ListGroups(c.Request.Context())
	if err != nil {
		apiErr, code := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(code)
		return
	}

	c.JSON(http.StatusOK, groups)
}

func (h *Handler) GetGroupByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := utils.ValidateStringID(idParam)
	if err != nil {
		apiErr, code := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(code)
		return
	}

	group, err := h.ucs.GetGroupByID(c.Request.Context(), id)
	if err != nil {
		apiErr, code := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(code)
		return
	}

	c.JSON(http.StatusOK, group)
}

func (h *Handler) UpdateGroup(c *gin.Context) {
	// idParam := c.Param("id")
	// id, err := utils.ValidateID(idParam)
	// if err != nil {
	// 	apiErr, code := types.NewAPIError(err)
	// 	c.Error(apiErr).SetMeta(code)
	// 	return
	// }

	id := c.Param("id")

	var req dto.UpdateGroup
	if err := utils.ValidateRequest(c, &req); err != nil {
		apiErr, code := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(code)
		return
	}

	group := req.ToDomain()
	group.ID = id

	if err := h.ucs.UpdateGroup(c.Request.Context(), group); err != nil {
		apiErr, code := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(code)
		return
	}

	c.JSON(http.StatusOK, types.MessageResponse{
		Message: "Group updated successfully",
	})
}

func (h *Handler) DeleteGroup(c *gin.Context) {
	idParam := c.Param("id")
	id, err := utils.ValidateStringID(idParam)
	if err != nil {
		apiErr, code := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(code)
		return
	}

	if err := h.ucs.DeleteGroup(c.Request.Context(), id); err != nil {
		apiErr, code := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(code)
		return
	}

	c.JSON(http.StatusOK, types.MessageResponse{
		Message: "Group deleted successfully",
	})
}
