package campaign

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	domain "github.com/devpablocristo/ponti-backend/internal/campaign/usecases/domain"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

type UseCasesPort interface {
	CreateCampaign(context.Context, *domain.Campaign) (int64, error)
	ListCampaigns(context.Context, int64, string) ([]domain.Campaign, error)
	GetCampaign(context.Context, int64) (*domain.Campaign, error)
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
	ucs UseCasesPort
	gsv GinEnginePort
	acf ConfigAPIPort
	mws MiddlewaresEnginePort
}

func NewHandler(u UseCasesPort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		acf: c,
		mws: m,
	}
}

func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/campaigns"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	public := r.Group(baseURL)
	{
		public.GET("", h.ListCampaigns)
	}
}

func (h *Handler) ListCampaigns(c *gin.Context) {
	var customerID int64
	customerIDValue, err := sharedhandlers.ParseOptionalInt64Query(c, "customer_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if customerIDValue != nil {
		customerID = *customerIDValue
	}

	campaigns, err := h.ucs.ListCampaigns(c.Request.Context(), customerID, c.Query("project_name"))
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, campaigns)
}
