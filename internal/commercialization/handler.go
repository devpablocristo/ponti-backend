package commercialization

import (
	"context"
	"net/http"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/gin-gonic/gin"

	dto "github.com/alphacodinggroup/ponti-backend/internal/commercialization/handler/dto"
	domain "github.com/alphacodinggroup/ponti-backend/internal/commercialization/usecases/domain"
	sharedhandlers "github.com/alphacodinggroup/ponti-backend/internal/shared/handlers"
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

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	public := r.Group(baseURL)
	{
		public.GET("", h.ListByProject)
		public.POST("", h.CreateOrUpdateBulk)
	}
}

// Listar por proyecto
func (h *Handler) ListByProject(c *gin.Context) {
	projectID, err := sharedhandlers.ParseParamID(c.Param("project_id"), "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	items, err := h.ucs.ListByProject(c.Request.Context(), projectID)
	if err != nil {
		sharedhandlers.RespondErrorLegacyNotFound(c, err)
		return
	}

	resp := make([]dto.CommercializationResponse, len(items))
	for i, d := range items {
		resp[i] = dto.FromDomain(&d)
	}

	c.JSON(http.StatusOK, resp)
}

// Crear proyecto
func (h *Handler) CreateOrUpdateBulk(c *gin.Context) {
	projectID, err := sharedhandlers.ParseParamID(c.Param("project_id"), "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	userID, err := sharedhandlers.ParseUserID(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	var body dto.BulkCommercializationRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		domErr := types.NewError(types.ErrBadRequest, "invalid request payload", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	items := body.ToDomainSlice(projectID, userID)
	if err := h.ucs.CreateOrUpdateBulk(c.Request.Context(), items); err != nil {
		sharedhandlers.RespondErrorLegacyNotFound(c, err)
		return
	}
	c.JSON(http.StatusCreated, types.MessageResponse{Message: "Crop commercialization saved"})
}
