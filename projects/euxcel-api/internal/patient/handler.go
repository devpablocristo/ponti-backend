package patient

import (
	"net/http"

	"github.com/gin-gonic/gin"

	mdw "github.com/alphacodinggroup/euxcel-backend/pkg/http/middlewares/gin"
	gsv "github.com/alphacodinggroup/euxcel-backend/pkg/http/servers/gin"
	types "github.com/alphacodinggroup/euxcel-backend/pkg/types"
	utils "github.com/alphacodinggroup/euxcel-backend/pkg/utils"

	dto "github.com/alphacodinggroup/euxcel-backend/internal/patient/handler/dto"
)

// Handler gestiona las peticiones relacionadas a Patient.
type Handler struct {
	ucs UseCases   // UseCases es la interfaz de casos de uso para Patient.
	gsv gsv.Server // gsv provee las funcionalidades del servidor Gin.
	mws *mdw.Middlewares
}

// NewHandler crea una nueva instancia del handler de Patient.
func NewHandler(s gsv.Server, u UseCases, m *mdw.Middlewares) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		mws: m,
	}
}

// Routes define los endpoints para Patient en el router de Gin.
func (h *Handler) Routes() {
	router := h.gsv.GetRouter()

	// Middleware global

	apiVersion := h.gsv.GetApiVersion()
	apiBase := "/api/" + apiVersion + "/patients"
	publicPrefix := apiBase + "/public"
	validatedPrefix := apiBase + "/validated"
	protectedPrefix := apiBase + "/protected"

	// Rutas públicas para Patient
	public := router.Group(publicPrefix)
	{
		public.POST("", h.CreatePatient)
		//public.GET("", h.ListPatients)
		// public.GET("/:id", h.GetPatientByID)
		// public.PUT("/:id", h.UpdatePatient)
		// public.DELETE("/:id", h.DeletePatient)
	}

	// Rutas validadas (se pueden aplicar middlewares de validación específicos)
	validated := router.Group(validatedPrefix)
	{
		validated.Use(h.mws.Validated...)
		// Agregar rutas adicionales para usuarios validados, si es necesario.
	}

	// Rutas protegidas (requieren de autenticación)
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

func (h *Handler) CreatePatient(c *gin.Context) {
	// Declaramos el DTO para el paciente.
	var req dto.Patient

	// Validamos la request utilizando alguna función de validación, similar a la de CreateEvent.
	if err := utils.ValidateRequest(c, &req); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	// Convertimos el DTO a la entidad de dominio usando el método ToDomain del DTO.
	patientDomain, err := req.ToDomain()
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	// Extraemos el contexto de la request.
	ctx := c.Request.Context()

	// Llamamos al caso de uso para crear el paciente. Se asume que h.ucs.CreatePatient está implementado.
	if err := h.ucs.CreatePatient(ctx, patientDomain); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	// Enviamos la respuesta exitosa.
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Patient created successfully",
	})
}
