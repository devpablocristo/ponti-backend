package capabilities

import (
	"net/http"
	"strings"

	"github.com/devpablocristo/core/errors/go/domainerr"
	"github.com/gin-gonic/gin"

	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

// GinEnginePort expone el router HTTP.
type GinEnginePort interface {
	GetRouter() *gin.Engine
}

// ConfigAPIPort expone la base URL del API (ej: "/api/v1").
type ConfigAPIPort interface {
	APIBaseURL() string
}

// MiddlewaresEnginePort expone middlewares estándar (auth, validación).
type MiddlewaresEnginePort interface {
	GetValidation() []gin.HandlerFunc
}

// Handler expone el catálogo de capabilities publicadas por Ponti hacia
// Companion. Endpoints son autenticados como cualquier otro: el caller debe
// presentar JWT válido y el filtro por roles/modules ocurre en el cliente
// (Companion) usando la metadata del manifest. Acá sólo se valida tenant
// scope en el sentido de que el usuario está autenticado en algún tenant.
type Handler struct {
	eng GinEnginePort
	cfg ConfigAPIPort
	mws MiddlewaresEnginePort
}

func NewHandler(eng GinEnginePort, cfg ConfigAPIPort, mws MiddlewaresEnginePort) *Handler {
	return &Handler{eng: eng, cfg: cfg, mws: mws}
}

// Routes registra GET /api/v1/capabilities y GET /api/v1/capabilities/:id.
func (h *Handler) Routes() {
	base := h.cfg.APIBaseURL() + "/capabilities"
	group := h.eng.GetRouter().Group(base, h.mws.GetValidation()...)
	group.GET("", h.List)
	group.GET("/:id", h.Get)
}

// List devuelve todos los manifests publicados. La validez tenant-scoped y de
// roles la aplica Companion (el cliente) en su Registry vía MatchesFilter.
// Ponti expone TODOS los manifests porque las capabilities son metadata, no
// datos sensibles del tenant — el control de acceso real ocurre cuando el
// caller invoca el endpoint subyacente (e.g. /insights/summary), no acá.
func (h *Handler) List(c *gin.Context) {
	if _, err := sharedhandlers.ParseOrgID(c); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": All()})
}

// Get devuelve un manifest puntual por ID.
func (h *Handler) Get(c *gin.Context) {
	if _, err := sharedhandlers.ParseOrgID(c); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		sharedhandlers.RespondError(c, domainerr.Validation("capability id is required"))
		return
	}
	m, ok := FindByID(id)
	if !ok {
		sharedhandlers.RespondError(c, domainerr.NotFound("capability not found"))
		return
	}
	c.JSON(http.StatusOK, m)
}
