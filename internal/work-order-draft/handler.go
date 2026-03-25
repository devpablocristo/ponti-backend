package workorderdraft

import (
	"context"

	"github.com/gin-gonic/gin"

	sharedhandlers "github.com/alphacodinggroup/ponti-backend/internal/shared/handlers"
	"github.com/alphacodinggroup/ponti-backend/internal/work-order-draft/handler/dto"
	"github.com/alphacodinggroup/ponti-backend/internal/work-order-draft/usecases/domain"
)

type UseCasesPort interface {
	CreateWorkOrderDraft(context.Context, *domain.WorkOrderDraft) (int64, error)
	GetWorkOrderDraftByID(context.Context, int64) (*domain.WorkOrderDraft, error)
	ListWorkOrderDrafts(context.Context, string) ([]domain.WorkOrderDraftListItem, error)
	UpdateWorkOrderDraftByID(context.Context, *domain.WorkOrderDraft) error
	PublishWorkOrderDraft(context.Context, int64) (int64, error)
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
	base := h.acf.APIBaseURL() + "/work-order-drafts"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	grp := r.Group(base)
	{
		grp.POST("", h.CreateWorkOrderDraft)
		grp.GET("", h.ListWorkOrderDrafts)
		grp.GET("/:work_order_draft_id", h.GetWorkOrderDraftByID)
		grp.PUT("/:work_order_draft_id", h.UpdateWorkOrderDraftByID)
		grp.POST("/:work_order_draft_id/publish", h.PublishWorkOrderDraft)
	}
}

func (h *Handler) CreateWorkOrderDraft(c *gin.Context) {
	var req dto.WorkOrderDraft
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}

	draft, err := req.ToDomain()
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	id, err := h.ucs.CreateWorkOrderDraft(c.Request.Context(), draft)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	sharedhandlers.RespondCreated(c, id)
}

func (h *Handler) GetWorkOrderDraftByID(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("work_order_draft_id"), "work_order_draft_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	draft, err := h.ucs.GetWorkOrderDraftByID(c.Request.Context(), id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	sharedhandlers.RespondOK(c, dto.FromDomain(draft))
}

func (h *Handler) UpdateWorkOrderDraftByID(c *gin.Context) {
	var req dto.WorkOrderDraft
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}

	id, err := sharedhandlers.ParseParamID(c.Param("work_order_draft_id"), "work_order_draft_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	draft, err := req.ToDomain()
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	draft.ID = id

	if err := h.ucs.UpdateWorkOrderDraftByID(c.Request.Context(), draft); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) ListWorkOrderDrafts(c *gin.Context) {
	items, err := h.ucs.ListWorkOrderDrafts(c.Request.Context(), c.Query("number"))
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	sharedhandlers.RespondOK(c, dto.NewListResponse(items))
}

func (h *Handler) PublishWorkOrderDraft(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("work_order_draft_id"), "work_order_draft_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	workOrderID, err := h.ucs.PublishWorkOrderDraft(c.Request.Context(), id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	sharedhandlers.RespondOK(c, gin.H{
		"draft_id":                id,
		"published_work_order_id": workOrderID,
		"status":                  "published",
	})
}
