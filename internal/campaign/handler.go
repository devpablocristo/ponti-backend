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
	GetArchivedCampaigns(context.Context) ([]domain.Campaign, error)
	UpdateCampaign(context.Context, *domain.Campaign) error
	DeleteCampaign(context.Context, int64) error
	ArchiveCampaign(context.Context, int64) error
	RestoreCampaign(context.Context, int64) error
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

	public := r.Group(baseURL, h.mws.GetValidation()...)
	{
		public.GET("", h.ListCampaigns)
		public.POST("", h.CreateCampaign)
		public.GET("/archived", h.GetArchivedCampaigns)
		public.GET("/:campaign_id", h.GetCampaign)
		public.PUT("/:campaign_id", h.UpdateCampaign)
		public.DELETE("/:campaign_id", h.DeleteCampaign)
		public.POST("/:campaign_id/archive", h.ArchiveCampaign)
		public.POST("/:campaign_id/restore", h.RestoreCampaign)
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
