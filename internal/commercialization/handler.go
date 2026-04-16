package commercialization

import (
	"context"

	"github.com/gin-gonic/gin"

	dto "github.com/devpablocristo/ponti-backend/internal/commercialization/handler/dto"
	domain "github.com/devpablocristo/ponti-backend/internal/commercialization/usecases/domain"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

type UseCasePort interface {
	CreateOrUpdateBulk(context.Context, []domain.CropCommercialization) error
	ListByProject(context.Context, int64) ([]domain.CropCommercialization, error)
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
	baseURL := h.cfg.APIBaseURL() + "/projects/:project_id/commercializations"

	public := r.Group(baseURL, h.mws.GetValidation()...)
	{
		public.GET("", h.ListByProject)
		public.POST("", h.CreateOrUpdateBulk)
	}
}

// Listar por proyecto
func (h *Handler) ListByProject(c *gin.Context) {
	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	items, err := h.ucs.ListByProject(c.Request.Context(), projectID)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	resp := make([]dto.CommercializationResponse, len(items))
	for i, d := range items {
		resp[i] = dto.FromDomain(&d)
	}

	sharedhandlers.RespondOK(c, resp)
}

// Crear proyecto
func (h *Handler) CreateOrUpdateBulk(c *gin.Context) {
	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	userID, err := sharedhandlers.ParseActor(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	var body dto.BulkCommercializationRequest
	if err := sharedhandlers.BindJSON(c, &body); err != nil {
		return
	}

	items := body.ToDomainSlice(projectID, userID)
	if err := h.ucs.CreateOrUpdateBulk(c.Request.Context(), items); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}
