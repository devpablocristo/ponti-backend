package event

import (
	"net/http"

	"github.com/gin-gonic/gin"

	mdw "github.com/devpablocristo/monorepo/pkg/http/middlewares/gin"
	gsv "github.com/devpablocristo/monorepo/pkg/http/servers/gin"
	types "github.com/devpablocristo/monorepo/pkg/types"
	utils "github.com/devpablocristo/monorepo/pkg/utils"

	dto "github.com/devpablocristo/monorepo/projects/qh/internal/event/handler/dto"
	sup "github.com/devpablocristo/monorepo/projects/qh/internal/event/handler/support"
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
	apiBase := "/api/" + apiVersion + "/events"
	publicPrefix := apiBase + "/public"
	validatedPrefix := apiBase + "/validated"
	protectedPrefix := apiBase + "/protected"

	// Rutas públicas
	public := router.Group(publicPrefix)
	{
		public.POST("", h.CreateEvent)
		public.GET("", h.ListEvents)
		// public.GET("/:id", h.GetEventByID)
		// public.PUT("/:id", h.UpdateEvent)
		// public.DELETE("/:id", h.DeleteEvent)
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
	}
}

func (h *Handler) ProtectedPing(c *gin.Context) {
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Protected Pong!",
	})
}

func (h *Handler) ListEvents(c *gin.Context) {
	groups, err := h.ucs.ListEvents(c.Request.Context())
	if err != nil {
		apiErr, code := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(code)
		return
	}

	var list dto.EventList
	c.JSON(http.StatusOK, sup.ListEventsResponse{
		List: list.FromDomain(groups),
	})
}

func (h *Handler) CreateEvent(c *gin.Context) {
	// cuil, err := mdw.ExtractClaim(c, "sub", "")
	// if err != nil {
	// 	c.JSON(http.StatusUnauthorized, sup.ErrorResponse{
	// 		Error: err.Error(),
	// 	})
	// 	return
	// }

	var req dto.CreateEvent
	if err := utils.ValidateRequest(c, &req); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	ctx := c.Request.Context()
	if err := h.ucs.CreateEvent(ctx, req.ToDomain()); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Event created successfully",
	})
}

// func (es *Handler) GetEvent(c *gin.Context) {
// 	eventID := c.Param("eventID")
// 	event, err := es.ucs.GetEvent(c, eventID)
// 	if err != nil {
// 		apiErr, errCode := types.NewAPIError(err)
// 		c.Error(apiErr).SetMeta(errCode)
// 		return
// 	}

// 	c.JSON(http.StatusOK, ctypes.NewAPIMessage("event founded", event))
// }

// func (es *Handler) DeleteEvent(c *gin.Context) {
// 	eventID := c.Param("eventID")
// 	_, err := es.useCase.DeleteEvent(c, eventID)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, ctypes.NewAPIError(http.StatusBadRequest, err.Error()))
// 		return
// 	}

// 	c.JSON(http.StatusOK, ctypes.NewAPIMessage("event successfully deleted"))
// }

// func (es *Handler) HardDeleteEvent(c *gin.Context) {
// 	eventID := c.Param("eventID")
// 	_, err := es.useCase.HardDeleteEvent(c, eventID)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, ctypes.NewAPIError(http.StatusBadRequest, err.Error()))
// 		return
// 	}

// 	c.JSON(http.StatusOK, ctypes.NewAPIMessage("event successfully deleted"))
// }

// func (es *Handler) UpdateEvent(c *gin.Context) {
// 	var dto *ctypes.Event
// 	if err := c.ShouldBindJSON(&dto); err != nil {
// 		c.JSON(http.StatusBadRequest, ctypes.NewAPIError(http.StatusBadRequest, err.Error()))
// 		return
// 	}
// 	eventID := c.Param("eventID")
// 	_, err := es.useCase.UpdateEvent(c, ctypes.EventDtoToDomain(dto), eventID)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, ctypes.NewAPIError(http.StatusBadRequest, err.Error()))
// 		return
// 	}

// 	c.JSON(http.StatusOK, ctypes.NewAPIMessage("event successfully updated"))
// }

// func (es *Handler) ReviveEvent(c *gin.Context) {
// 	eventID := c.Param("eventID")
// 	_, err := es.useCase.ReviveEvent(c, eventID)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, ctypes.NewAPIError(http.StatusBadRequest, err.Error()))
// 		return
// 	}

// 	c.JSON(http.StatusOK, ctypes.NewAPIMessage("event successfully undeleted"))
// }

// func (es *Handler) AddUserToEvent(c *gin.Context) {
// 	eventID := c.Param("eventID")

// 	var dto *ctypes.User
// 	if err := c.ShouldBindJSON(&dto); err != nil {
// 		c.JSON(http.StatusBadRequest, ctypes.NewAPIError(http.StatusBadRequest, err.Error()))
// 		return
// 	}

// 	_, err := es.useCase.AddUserToEvent(c, eventID, ctypes.UserDtoToDomain(dto))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, ctypes.NewAPIError(http.StatusBadRequest, err.Error()))
// 		return
// 	}

// 	c.JSON(http.StatusOK, ctypes.NewAPIMessage("person added to event"))
// }

// func (es *Handler) AddPersonsGroupToEvent(c *gin.Context) {
// 	events, err := es.useCase.ListEvents(c)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, ctypes.NewAPIError(http.StatusBadRequest, err.Error()))
// 		return
// 	}

// 	c.JSON(http.StatusOK, ctypes.NewAPIMessage("persons group to event"))

// }
