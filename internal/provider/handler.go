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

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	publicGroup := r.Group(baseURL + "/providers")
	{
		publicGroup.GET("", h.GetProviders)
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
