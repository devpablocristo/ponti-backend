package investor

import (
	"context"

	ginmw "github.com/devpablocristo/core/http/gin/go"
	"github.com/gin-gonic/gin"

	dto "github.com/devpablocristo/ponti-backend/internal/investor/handler/dto"
	domain "github.com/devpablocristo/ponti-backend/internal/investor/usecases/domain"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

type UseCasesPort interface {
	CreateInvestor(context.Context, *domain.Investor) (int64, error)
	ListInvestors(context.Context, int, int) ([]domain.Investor, int64, error)
	GetInvestor(context.Context, int64) (*domain.Investor, error)
	UpdateInvestor(context.Context, *domain.Investor) error
	DeleteInvestor(context.Context, int64) error
	ArchiveInvestor(context.Context, int64) error
	RestoreInvestor(context.Context, int64) error
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
	baseURL := h.acf.APIBaseURL() + "/investors"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	public := r.Group(baseURL)
	{
		public.POST("", h.CreateInvestor)
		public.GET("", h.ListInvestors)
		public.GET("/:investor_id", h.GetInvestor)
		public.PUT("/:investor_id", h.UpdateInvestor)
		public.DELETE("/:investor_id", h.DeleteInvestor)
		public.POST("/:investor_id/archive", h.ArchiveInvestor)
		public.POST("/:investor_id/restore", h.RestoreInvestor)
	}
}

func (h *Handler) CreateInvestor(c *gin.Context) {
	var req dto.CreateInvestorRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	id, err := h.ucs.CreateInvestor(c.Request.Context(), req.ToDomain())
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondCreated(c, id)
}

func (h *Handler) ListInvestors(c *gin.Context) {
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 1000)
	investors, total, err := h.ucs.ListInvestors(c.Request.Context(), page, perPage)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.NewListInvestorsResponse(investors, page, perPage, total))
}

func (h *Handler) GetInvestor(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "investor_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	inv, err := h.ucs.GetInvestor(c.Request.Context(), id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.InvestorFromDomain(inv))
}

func (h *Handler) UpdateInvestor(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "investor_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req dto.UpdateInvestorRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	if err := h.ucs.UpdateInvestor(c.Request.Context(), req.ToDomain(id)); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) DeleteInvestor(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "investor_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.DeleteInvestor(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) ArchiveInvestor(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "investor_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.ArchiveInvestor(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) RestoreInvestor(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "investor_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.RestoreInvestor(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}
