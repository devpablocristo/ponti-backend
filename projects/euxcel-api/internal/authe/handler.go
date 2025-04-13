package authe

import (
	"net/http"

	"github.com/gin-gonic/gin"

	mdw "github.com/alphacodinggroup/euxcel-backend/pkg/http/middlewares/gin"
	gsv "github.com/alphacodinggroup/euxcel-backend/pkg/http/servers/gin"
	types "github.com/alphacodinggroup/euxcel-backend/pkg/types"

	dto "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/authe/handler/dto"
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

	router.Use(h.mws.Global...)

	apiVersion := h.gsv.GetApiVersion()
	apiBase := "/api/" + apiVersion + "/authe"
	publicPrefix := apiBase + "/public"
	validatedPrefix := apiBase + "/validated"
	protectedPrefix := apiBase + "/protected"

	public := router.Group(publicPrefix)
	{
		public.POST("", h.Login)
	}

	validated := router.Group(validatedPrefix)
	{
		// Aplicar middleware de validación de credenciales
		validated.Use(h.mws.Validated...)

		validated.POST("", h.Login)
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

func (h *Handler) Login(c *gin.Context) {
	// Recuperar las credenciales validadas por el middleware
	credentialsRaw, exists := c.Get("credentials")
	if !exists {
		apiErr, errCode := types.NewAPIError(
			types.NewError(
				types.ErrorType(types.APIErrUnauthorized),
				"credentials not found in context",
				nil,
			),
		)
		c.Error(apiErr).SetMeta(errCode)
	}

	// Asegurarte de que las credenciales tengan el tipo esperado
	credentials, ok := credentialsRaw.(types.LoginCredentials)
	if !ok {
		apiErr, errCode := types.NewAPIError(
			types.NewError(
				types.ErrorType(types.APIErrUnauthorized),
				"invalid credentials type",
				nil,
			),
		)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	loginType := c.Query("type")

	switch loginType {
	case "jwt":
		h.jwtLogin(c, credentials)
	case "auth0":
		h.auth0Login(c, credentials)
	default:
		apiErr, errCode := types.NewAPIError(
			types.NewError(
				types.ErrorType(types.APIErrUnauthorized),
				"Invalid login type",
				nil,
			),
		)
		c.Error(apiErr).SetMeta(errCode)
	}

}

func (h *Handler) jwtLogin(c *gin.Context, credentials types.LoginCredentials) {
	token, err := h.ucs.JwtLogin(c.Request.Context(), credentials.Username, credentials.Email, credentials.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, dto.LoginResponse{
		AccessToken: token.AccessToken,
	})
}

func (h *Handler) auth0Login(c *gin.Context, credentials types.LoginCredentials) {
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Auth0 coming soon!",
	})
}
