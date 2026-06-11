// Package provider expone endpoints HTTP para proveedores.
package provider

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/devpablocristo/ponti-backend/internal/provider/handler/dto"
	"github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

type UseCasesPort interface {
	GetProviders(context.Context) ([]domain.Provider, error)
	GetArchivedProviders(context.Context) ([]domain.Provider, error)
	CreateProvider(context.Context, *domain.Provider) (int64, error)
	GetProvider(context.Context, int64) (*domain.Provider, error)
	UpdateProvider(context.Context, *domain.Provider) error
	DeleteProvider(context.Context, int64) error
	ArchiveProvider(context.Context, int64) error
	RestoreProvider(context.Context, int64) error
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

// Routes registra las rutas del módulo Provider.
func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL()

	publicGroup := r.Group(baseURL+"/providers", h.mws.GetValidation()...)
	{
		publicGroup.GET("", h.GetProviders)
		publicGroup.POST("", h.CreateProvider)
		publicGroup.GET("/archived", h.GetArchivedProviders)
		publicGroup.GET("/:provider_id", h.GetProvider)
		publicGroup.PUT("/:provider_id", h.UpdateProvider)
		publicGroup.DELETE("/:provider_id", h.DeleteProvider)
		publicGroup.POST("/:provider_id/archive", h.ArchiveProvider)
		publicGroup.POST("/:provider_id/restore", h.RestoreProvider)
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

	c.JSON(http.StatusOK, dto.NewGetProvidersResponse(providers))
}
