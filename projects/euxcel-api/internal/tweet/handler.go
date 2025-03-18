package tweet

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	mdw "github.com/alphacodinggroup/euxcel-backend/pkg/http/middlewares/gin"
	gsv "github.com/alphacodinggroup/euxcel-backend/pkg/http/servers/gin"
	types "github.com/alphacodinggroup/euxcel-backend/pkg/types"
	utils "github.com/alphacodinggroup/euxcel-backend/pkg/utils"

	dto "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/tweet/handler/dto"
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
	apiBase := "/api/" + apiVersion + "/tweets"
	publicPrefix := apiBase + "/public"
	validatedPrefix := apiBase + "/validated"
	protectedPrefix := apiBase + "/protected"

	//Rutas públicas
	public := router.Group(publicPrefix)
	{
		public.POST("", h.CreateTweet)
		public.GET("/:id/timeline", h.GetTimeline)
	}

	validated := router.Group(validatedPrefix)
	{
		// Aplicar middleware de validación (ej. de credenciales)
		validated.Use(h.mws.Validated...)
	}

	// Rutas protegidas
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

func (h *Handler) CreateTweet(c *gin.Context) {
	var req dto.Tweet
	if err := utils.ValidateRequest(c, &req); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	// Convertir el DTO de entrada al objeto del dominio.
	domainTweet := req.ToDomain()

	// Crear el tweet a través del caso de uso.
	ctx := c.Request.Context()
	newTweetID, err := h.ucs.CreateTweet(ctx, domainTweet)
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	// Responder con el ID del tweet creado.
	c.JSON(http.StatusCreated, dto.CreateTweetResponse{
		Message: "Tweet created successfully",
		TweetID: newTweetID,
	})
}

// GetTimeline obtiene el timeline de tweets de un usuario y lo mapea a DTOs.
func (h *Handler) GetTimeline(c *gin.Context) {
	userID := c.Param("id")

	// Obtener el timeline (slice de domain.Tweet) desde la capa de usecases.
	timeline, err := h.ucs.GetTimeline(c.Request.Context(), userID)
	if err != nil {
		apiErr, code := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(code)
		c.JSON(code, apiErr.ToResponse())
		return
	}

	// Mapear cada tweet del dominio al DTO GetTimeline.
	dtoTimeline := make([]dto.GetTimeline, 0, len(timeline))
	for _, dTweet := range timeline {
		mapped, err := dto.FromDomainToGetTimeline(&dTweet)
		if err != nil {
			log.Printf("Error mapping domain tweet to DTO: %v", err)
			continue
		}
		dtoTimeline = append(dtoTimeline, *mapped)
	}

	// Enviar la respuesta con el timeline mapeado.
	c.JSON(http.StatusOK, dtoTimeline)
}
