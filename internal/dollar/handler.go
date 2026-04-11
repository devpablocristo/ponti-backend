package dollar

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/devpablocristo/ponti-backend/internal/dollar/handler/dto"
	domain "github.com/devpablocristo/ponti-backend/internal/dollar/usecases/domain"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

type UseCasePort interface {
	ListByProject(context.Context, int64) ([]domain.DollarAverage, error)
	CreateOrUpdateBulk(context.Context, []domain.DollarAverage) error
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

type Handler struct {
	ucs UseCasePort
	gsv GinEnginePort
	cfg ConfigAPIPort
	mws MiddlewaresEnginePort
}

func NewHandler(u UseCasePort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		cfg: c,
		mws: m,
	}
}

func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.cfg.APIBaseURL() + "/projects/:project_id/dollar-values"

	public := r.Group(baseURL, h.mws.GetValidation()...)
	{
		public.GET("", h.ListByProject)
		public.PUT("", h.CreateorUpdateBulk)
	}
}

func (h *Handler) ListByProject(c *gin.Context) {
	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	// Llamo al caso de uso
	items, err := h.ucs.ListByProject(c.Request.Context(), projectID)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	// mapeo los domains a DTOs
	resp := make([]dto.MonthResponse, len(items))
	for i, d := range items {
		resp[i] = dto.FromDomainMonth(&d)
	}

	sharedhandlers.RespondOK(c, resp)
}

func (h *Handler) CreateorUpdateBulk(c *gin.Context) {
	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	// Parseo el body JSON a DTO
	var in dto.BulkDollarAverageRequest
	if err := sharedhandlers.BindJSON(c, &in); err != nil {
		return
	}

	// Convierto el DTO a Slice de domain
	items := in.ToDomainSlice(projectID)
	if err := h.ucs.CreateOrUpdateBulk(c.Request.Context(), items); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	sharedhandlers.RespondNoContent(c)
}
