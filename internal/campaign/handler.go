package campaign

import (
	"context"
	"net/http"

	ginmw "github.com/devpablocristo/platform/http/gin/go"
	"github.com/gin-gonic/gin"

	dto "github.com/devpablocristo/ponti-backend/internal/campaign/handler/dto"
	domain "github.com/devpablocristo/ponti-backend/internal/campaign/usecases/domain"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
)

type UseCasesPort interface {
	CreateCampaign(context.Context, *domain.Campaign) (int64, error)
	ListCampaigns(context.Context, int64, string) ([]domain.Campaign, error)
	ListArchivedCampaigns(context.Context, int, int) ([]domain.Campaign, int64, error)
	GetCampaign(context.Context, int64) (*domain.Campaign, error)
	UpdateCampaign(context.Context, *domain.Campaign) error
	ArchiveCampaign(context.Context, int64) error
	RestoreCampaign(context.Context, int64) error
	HardDeleteCampaign(context.Context, int64) error
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

func (h *Handler) runCampaignIDAction(c *gin.Context, action func(context.Context, int64) error) {
	id, err := ginmw.ParseParamID(c, "campaign_id")
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

func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/campaigns"

	public := r.Group(baseURL, h.mws.GetValidation()...)
	{
		public.POST("", h.CreateCampaign)
		public.GET("", h.ListCampaigns)
		public.GET("/archived", h.ListArchivedCampaigns)
		public.GET("/:campaign_id", h.GetCampaign)
		public.PUT("/:campaign_id", h.UpdateCampaign)
		public.POST("/:campaign_id/archive", h.ArchiveCampaign)
		public.POST("/:campaign_id/restore", h.RestoreCampaign)
		public.DELETE("/:campaign_id/hard", h.HardDeleteCampaign)
	}
}

func (h *Handler) CreateCampaign(c *gin.Context) {
	var req dto.CreateCampaignRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	id, err := h.ucs.CreateCampaign(c.Request.Context(), req.ToDomain())
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondCreated(c, id)
}

func (h *Handler) GetCampaign(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "campaign_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	camp, err := h.ucs.GetCampaign(c.Request.Context(), id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.FromDomain(*camp))
}

func (h *Handler) UpdateCampaign(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "campaign_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req dto.UpdateCampaignRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	if err := h.ucs.UpdateCampaign(c.Request.Context(), req.ToDomain(id)); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
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
	out := make([]dto.Campaign, 0, len(campaigns))
	for _, d := range campaigns {
		out = append(out, *dto.FromDomain(d))
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) ListArchivedCampaigns(c *gin.Context) {
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 1000)
	campaigns, total, err := h.ucs.ListArchivedCampaigns(c.Request.Context(), page, perPage)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	out := make([]dto.Campaign, 0, len(campaigns))
	for _, d := range campaigns {
		out = append(out, *dto.FromDomain(d))
	}
	sharedhandlers.RespondOK(c, gin.H{
		"data":      out,
		"page_info": types.NewPageInfo(page, perPage, total),
	})
}

func (h *Handler) ArchiveCampaign(c *gin.Context) {
	h.runCampaignIDAction(c, h.ucs.ArchiveCampaign)
}

func (h *Handler) RestoreCampaign(c *gin.Context) {
	h.runCampaignIDAction(c, h.ucs.RestoreCampaign)
}

func (h *Handler) HardDeleteCampaign(c *gin.Context) {
	h.runCampaignIDAction(c, h.ucs.HardDeleteCampaign)
}
