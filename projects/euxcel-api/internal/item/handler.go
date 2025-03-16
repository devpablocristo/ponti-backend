package item

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

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
	apiBase := "/api/" + apiVersion + "/assessments"
	//publicPrefix := apiBase + "/public"
	validatedPrefix := apiBase + "/validated"
	protectedPrefix := apiBase + "/protected"

	// Rutas públicas
	// public := router.Group(publicPrefix){}

	validated := router.Group(validatedPrefix)
	{
		// Aplicar middleware de validación de credenciales
		validated.Use(h.mws.Validated...)

		// Obtener detalles de un item validado
		//validated.GET("/:id", h.GetValidatedAssessment)
		//validated.POST("/:id/responses", h.SubmitAssessmentResponse) // Enviar respuestas a un item validado
	}

	// Rutas protegidas
	protected := router.Group(protectedPrefix)
	{
		protected.Use(h.mws.Protected...)

		protected.GET("/ping", h.ProtectedPing) // Endpoint de prueba protegido

		protected.POST("", h.CreateAssessment)       // Crear un item
		protected.GET("", h.ListAssessments)         // Listar todos los assessments
		protected.GET("/:id", h.GetAssessment)       // Obtener un item por ID
		protected.PUT("/:id", h.UpdateAssessment)    // Actualizar un item
		protected.DELETE("/:id", h.DeleteAssessment) // Eliminar un item
	}
}

func (h *Handler) ProtectedPing(c *gin.Context) {
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Protected Pong!",
	})
}

func (h *Handler) CreateAssessment(c *gin.Context) {
	tokenInterface, exists := c.Get("token")
	if !exists {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error: "token not found in context",
		})
		return
	}
	token, ok := tokenInterface.(*jwt.Token)
	if !ok {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error: "invalid token type in context",
		})
		return
	}

	// Extraer el claim "sub" del token.
	userID, err := utils.ExtractClaim(token, "sub")
	if err != nil {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	var req dto.CreateAssessment
	if err := utils.ValidateRequest(c, &req); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}
	req.HRID = userID

	ctx := c.Request.Context()
	item, err := req.ToDomain()
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	newAssessmentID, err := h.ucs.CreateAssessment(ctx, item)
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	c.JSON(http.StatusCreated, dto.CreateAssessmentResponse{
		Message:      "Item created successfully",
		AssessmentID: newAssessmentID,
	})
}

func (h *Handler) ListAssessments(c *gin.Context) {
	users, err := h.ucs.ListAssessments(c.Request.Context())
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *Handler) GetAssessment(c *gin.Context) {
	id := c.Param("id")

	item, err := h.ucs.GetAssessment(c.Request.Context(), id)
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) UpdateAssessment(c *gin.Context) {
	var req dto.Item
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	updatedAssessment, err := req.ToDomain()
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}
	if err := h.ucs.UpdateAssessment(c.Request.Context(), updatedAssessment); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Item updated successfully",
	})
}

func (h *Handler) DeleteAssessment(c *gin.Context) {
	id := c.Param("id")
	if err := h.ucs.DeleteAssessment(c.Request.Context(), id); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Item deleted successfully",
	})
}
