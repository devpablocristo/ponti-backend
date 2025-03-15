package browserEvent

import (
	"net/http"

	"github.com/gin-gonic/gin"

	mdw "github.com/devpablocristo/monorepo/pkg/http/middlewares/gin"
	gsv "github.com/devpablocristo/monorepo/pkg/http/servers/gin"
	types "github.com/devpablocristo/monorepo/pkg/types"
)

// Handler gestiona los endpoints del recurso browserEvent.
type Handler struct {
	ucs UseCases
	gsv gsv.Server
	mws *mdw.Middlewares
	ws  WebSocket
}

// NewHandler crea una nueva instancia de Handler.
func NewHandler(s gsv.Server, u UseCases, m *mdw.Middlewares, w WebSocket) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		mws: m,
		ws:  w,
	}
}

// Routes registra las rutas REST y el endpoint de WebSocket.
// Se centraliza la configuración del WebSocket en este método.
func (h *Handler) Routes() {
	router := h.gsv.GetRouter()

	apiVersion := h.gsv.GetApiVersion()
	apiBase := "/api/" + apiVersion + "/browser-events"
	//publicPrefix := apiBase + "/public"
	validatedPrefix := apiBase + "/validated"
	protectedPrefix := apiBase + "/protected"

	publicWsPrefix := apiBase + "/public/ws"

	// Rutas públicas
	// public := router.Group(publicPrefix)
	// {
	// 	public.POST("", h.CreateCandidate)
	// 	public.GET("", h.ListCandidates)
	// 	public.GET("/:id", h.GetCandidate)
	// 	public.PUT("/:id", h.UpdateCandidate)
	// 	public.DELETE("/:id", h.DeleteCandidate)
	// }

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
	}

	wsGroup := router.Group(publicWsPrefix)
	{
		wsGroup.GET("", h.BrowserEvent)
		wsGroup.GET("/ping", h.WsPing)
	}
}

// ProtectedPing es un endpoint de ejemplo que responde "Protected Pong!".
func (h *Handler) ProtectedPing(c *gin.Context) {
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Protected Pong!",
	})
}

func (h *Handler) BrowserEvent(c *gin.Context) {
	h.ws.BrowserEvent(c.Writer, c.Request)
}

func (h *Handler) WsPing(c *gin.Context) {
	h.ws.Ping(c.Writer, c.Request)
}
