package assessment

import (
	"net/http"

	"github.com/gin-gonic/gin"

	mdw "github.com/alphacodinggroup/euxcel-backend/pkg/http/middlewares/gin"
	gsv "github.com/alphacodinggroup/euxcel-backend/pkg/http/servers/gin"
	types "github.com/alphacodinggroup/euxcel-backend/pkg/types"
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

		// Obtener detalles de un assessment validado
		//validated.GET("/:id", h.GetValidatedAssessment)
		//validated.POST("/:id/responses", h.SubmitAssessmentResponse) // Enviar respuestas a un assessment validado
	}

	// Rutas protegidas
	protected := router.Group(protectedPrefix)
	{
		protected.Use(h.mws.Protected...)

		protected.GET("/ping", h.ProtectedPing) // Endpoint de prueba protegido

		protected.POST("", h.CreateAssessment)       // Crear un assessment
		protected.GET("", h.ListAssessments)         // Listar todos los assessments
		protected.GET("/:id", h.GetAssessment)       // Obtener un assessment por ID
		protected.PUT("/:id", h.UpdateAssessment)    // Actualizar un assessment
		protected.DELETE("/:id", h.DeleteAssessment) // Eliminar un assessment
		protected.POST("/:id/link", h.GenerateLink)  // Generar link único para un assessment
		protected.GET("/:id/link", h.SendLink)       // Generar link único para un assessment
	}
}

func (h *Handler) ProtectedPing(c *gin.Context) {
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Protected Pong!",
	})
}
