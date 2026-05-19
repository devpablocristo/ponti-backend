// Package provider expone endpoints HTTP para proveedores.
package provider

import (
	"context"

	ginmw "github.com/devpablocristo/core/http/gin/go"
	"github.com/gin-gonic/gin"

	"github.com/devpablocristo/ponti-backend/internal/provider/handler/dto"
	"github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

type UseCasesPort interface {
	GetProviders(context.Context) ([]domain.Provider, error)
	ListArchivedProviders(context.Context) ([]domain.Provider, error)
	GetProvider(context.Context, int64) (*domain.Provider, error)
	CreateProvider(context.Context, *domain.Provider) (int64, error)
	UpdateProvider(context.Context, *domain.Provider) error
	ArchiveProvider(context.Context, int64) error
	RestoreProvider(context.Context, int64) error
	HardDeleteProvider(context.Context, int64) error
	DeleteProvider(context.Context, int64) error
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
}

// Routes registra las rutas del módulo Provider.
func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL()

	publicGroup := r.Group(baseURL+"/providers", h.mws.GetValidation()...)
	{
		publicGroup.POST("", h.CreateProvider)
		publicGroup.GET("", h.GetProviders)
		publicGroup.GET("/archived", h.ListArchivedProviders)
		publicGroup.GET("/:provider_id", h.GetProvider)
		publicGroup.PUT("/:provider_id", h.UpdateProvider)
		publicGroup.POST("/:provider_id/archive", h.ArchiveProvider)
		publicGroup.POST("/:provider_id/restore", h.RestoreProvider)
		publicGroup.DELETE("/:provider_id/hard", h.HardDeleteProvider)
		publicGroup.DELETE("/:provider_id", h.DeleteProvider)
	}
}

type Handler struct {
	ucs UseCasesPort
	gsv GinEnginePort
	acf ConfigAPIPort
	mws MiddlewaresEnginePort
}

func NewHandler(
	ucs UseCasesPort,
	s GinEnginePort,
	c ConfigAPIPort,
	m MiddlewaresEnginePort,
) *Handler {
	return &Handler{
		ucs: ucs,
		gsv: s,
		acf: c,
		mws: m,
	}
}

func (h *Handler) GetProviders(c *gin.Context) {
	ctx := c.Request.Context()

	providers, err := h.ucs.GetProviders(ctx)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	sharedhandlers.RespondOK(c, dto.NewGetProvidersResponse(providers))
}

func (h *Handler) ListArchivedProviders(c *gin.Context) {
	providers, err := h.ucs.ListArchivedProviders(c.Request.Context())
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.NewGetProvidersResponse(providers))
}

func (h *Handler) GetProvider(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "provider_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	provider, err := h.ucs.GetProvider(c.Request.Context(), id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.NewGetProvidersResponse([]domain.Provider{*provider})[0])
}

func (h *Handler) CreateProvider(c *gin.Context) {
	var req dto.CreateProviderRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	id, err := h.ucs.CreateProvider(c.Request.Context(), req.ToDomain())
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondCreated(c, id)
}

func (h *Handler) UpdateProvider(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "provider_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req dto.UpdateProviderRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	if err := h.ucs.UpdateProvider(c.Request.Context(), req.ToDomain(id)); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) ArchiveProvider(c *gin.Context) {
	h.runProviderIDAction(c, h.ucs.ArchiveProvider)
}

func (h *Handler) RestoreProvider(c *gin.Context) {
	h.runProviderIDAction(c, h.ucs.RestoreProvider)
}

func (h *Handler) HardDeleteProvider(c *gin.Context) {
	h.runProviderIDAction(c, h.ucs.HardDeleteProvider)
}

func (h *Handler) DeleteProvider(c *gin.Context) {
	h.runProviderIDAction(c, h.ucs.DeleteProvider)
}

func (h *Handler) runProviderIDAction(c *gin.Context, action func(context.Context, int64) error) {
	id, err := ginmw.ParseParamID(c, "provider_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := action(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}
