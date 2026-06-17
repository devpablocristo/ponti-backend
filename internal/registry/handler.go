package registry

import (
	"context"
	"strconv"

	ginmw "github.com/devpablocristo/platform/http/gin/go"
	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/gin-gonic/gin"

	dto "github.com/devpablocristo/ponti-backend/internal/registry/handler/dto"
	domain "github.com/devpablocristo/ponti-backend/internal/registry/usecases/domain"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

type UseCasesPort interface {
	SearchRegistry(ctx context.Context, q, typ, status string, page, perPage int) (domain.RegistryResult, error)
	SetAliases(ctx context.Context, actorID int64, aliases []string) error
	GetUsages(ctx context.Context, entityType string, id int64) (domain.UsageResult, error)
}

type GinEnginePort interface {
	GetRouter() *gin.Engine
	RunServer(ctx context.Context) error
}

type ConfigAPIPort interface {
	APIVersion() string
	APIBaseURL() string
}

type MiddlewaresEnginePort interface {
	GetGlobal() []gin.HandlerFunc
	GetValidation() []gin.HandlerFunc
	GetProtected() []gin.HandlerFunc
}

// Handler HTTP del registry (búsqueda unificada + edición de alias).
type Handler struct {
	ucs UseCasesPort
	gsv GinEnginePort
	acf ConfigAPIPort
	mws MiddlewaresEnginePort
}

func NewHandler(u UseCasesPort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *Handler {
	return &Handler{ucs: u, gsv: s, acf: c, mws: m}
}

// Routes registra las rutas del módulo registry.
func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/registry"

	public := r.Group(baseURL, h.mws.GetValidation()...)
	{
		public.GET("", h.Search)
		public.GET("/usages", h.GetUsages)
		public.PUT("/actors/:actor_id/aliases", h.SetActorAliases)
	}
}

// Search (GET /registry?q=&type=&status=&page=&per_page=) — búsqueda unificada paginada.
func (h *Handler) Search(c *gin.Context) {
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 100)
	typ := c.Query("type")
	if typ == "" {
		typ = "all"
	}
	res, err := h.ucs.SearchRegistry(c.Request.Context(), c.Query("q"), typ, c.Query("status"), page, perPage)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.NewSearchResponse(res, page, perPage))
}

// GetUsages (GET /registry/usages?entity_type=&id=) — proyectos que usan una entidad del catálogo.
// Soporta entity_type: crops | lease-types | types | lot | field | project | actor (arrendatario,
// vía field_lessees; el BFF resuelve customer/investor/manager por su cuenta).
func (h *Handler) GetUsages(c *gin.Context) {
	entityType := c.Query("entity_type")
	idStr := c.Query("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		sharedhandlers.RespondError(c, domainerr.Validation("id must be a positive integer"))
		return
	}
	res, err := h.ucs.GetUsages(c.Request.Context(), entityType, id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.NewUsageResponse(res))
}

// SetActorAliases (PUT /registry/actors/:actor_id/aliases) — reemplaza el set de alias del actor.
func (h *Handler) SetActorAliases(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "actor_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req dto.SetAliasesRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	if err := h.ucs.SetAliases(c.Request.Context(), id, req.Aliases); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}
