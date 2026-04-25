package workorderdraft

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
	"github.com/devpablocristo/ponti-backend/internal/work-order-draft/handler/dto"
	"github.com/devpablocristo/ponti-backend/internal/work-order-draft/usecases/domain"
)

type UseCasesPort interface {
	CreateWorkOrderDraft(context.Context, *domain.WorkOrderDraft) (int64, error)
	CreateDigitalWorkOrderDraft(context.Context, *domain.WorkOrderDraft) (int64, error)
	CreateDigitalWorkOrderDraftBatch(context.Context, *domain.WorkOrderDraftBatchCreate) ([]domain.WorkOrderDraftBatchCreateResultItem, error)
	PreviewDigitalWorkOrderNumber(context.Context, int64, string) (string, error)
	PreviewDigitalWorkOrderDraftBatchNumber(context.Context, int64, string) (string, error)
	GetWorkOrderDraftByID(context.Context, int64) (*domain.WorkOrderDraft, error)
	ExportWorkOrderDraftPDF(context.Context, int64) ([]byte, error)
	ExportWorkOrderDraftGroupPDF(context.Context, int64) ([]byte, error)
	ListWorkOrderDrafts(context.Context, string, string, types.Input) ([]domain.WorkOrderDraftListItem, types.PageInfo, error)
	ListDigitalWorkOrderDrafts(context.Context, string, string, types.Input) ([]domain.WorkOrderDraftListItem, types.PageInfo, error)
	UpdateWorkOrderDraftByID(context.Context, *domain.WorkOrderDraft) error
	DeleteWorkOrderDraftByID(context.Context, int64) error
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

	grp := r.Group(base)
	for _, mw := range h.mws.GetValidation() {
		grp.Use(mw)
	}
	{
		grp.POST("", h.CreateWorkOrderDraft)
		grp.POST("/digital", h.CreateDigitalWorkOrderDraft)
		grp.POST("/digital/batch", h.CreateDigitalWorkOrderDraftBatch)
		grp.POST("/digital/preview-number", h.PreviewDigitalWorkOrderNumber)
		grp.POST("/digital/batch/preview-number", h.PreviewDigitalWorkOrderDraftBatchNumber)
		grp.GET("", h.ListWorkOrderDrafts)
		grp.GET("/digital", h.ListDigitalWorkOrderDrafts)
		grp.GET("/:work_order_draft_id", h.GetWorkOrderDraftByID)
		grp.GET("/:work_order_draft_id/pdf", h.ExportWorkOrderDraftPDF)
		grp.GET("/:work_order_draft_id/group-pdf", h.ExportWorkOrderDraftGroupPDF)
		grp.PUT("/:work_order_draft_id", h.UpdateWorkOrderDraftByID)
		grp.DELETE("/:work_order_draft_id", h.DeleteWorkOrderDraftByID)
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

func (h *Handler) CreateDigitalWorkOrderDraft(c *gin.Context) {
	var req dto.WorkOrderDraft
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}

	draft, err := req.ToDomain()
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	id, err := h.ucs.CreateDigitalWorkOrderDraft(c.Request.Context(), draft)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	sharedhandlers.RespondCreated(c, id)
}

func (h *Handler) PreviewDigitalWorkOrderNumber(c *gin.Context) {
	var req dto.WorkOrderDraftNumberPreviewRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}

	number, err := h.ucs.PreviewDigitalWorkOrderNumber(c.Request.Context(), req.ProjectID, req.Number)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	sharedhandlers.RespondOK(c, dto.WorkOrderDraftNumberPreviewResponse{
		Number: number,
	})
}

func (h *Handler) PreviewDigitalWorkOrderDraftBatchNumber(c *gin.Context) {
	var req dto.WorkOrderDraftNumberPreviewRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}

	number, err := h.ucs.PreviewDigitalWorkOrderDraftBatchNumber(c.Request.Context(), req.ProjectID, req.Number)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	sharedhandlers.RespondOK(c, dto.WorkOrderDraftNumberPreviewResponse{
		Number: number,
	})
}

func (h *Handler) CreateDigitalWorkOrderDraftBatch(c *gin.Context) {
	var req dto.WorkOrderDraftBatchCreateRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}

	batch, err := req.ToDomain()
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	items, err := h.ucs.CreateDigitalWorkOrderDraftBatch(c.Request.Context(), batch)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.NewBatchCreateResponse(items))
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

func (h *Handler) ExportWorkOrderDraftPDF(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("work_order_draft_id"), "work_order_draft_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	data, err := h.ucs.ExportWorkOrderDraftPDF(c.Request.Context(), id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	filename := fmt.Sprintf("orden-digital-%d.pdf", id)

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/pdf", data)
}

func (h *Handler) ExportWorkOrderDraftGroupPDF(c *gin.Context) {
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

	data, err := h.ucs.ExportWorkOrderDraftGroupPDF(c.Request.Context(), id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	baseNumber := draft.Number
	if base, ok := extractBaseSequence(draft.Number); ok {
		baseNumber = fmt.Sprintf("D-%d", base)
	}

	filename := fmt.Sprintf("orden-digital-%s.pdf", baseNumber)

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/pdf", data)
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

func (h *Handler) DeleteWorkOrderDraftByID(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("work_order_draft_id"), "work_order_draft_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	if err := h.ucs.DeleteWorkOrderDraftByID(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) ListWorkOrderDrafts(c *gin.Context) {
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 10)

	items, pageInfo, err := h.ucs.ListWorkOrderDrafts(
		c.Request.Context(),
		c.Query("number"),
		c.Query("status"),
		types.Input{
			Page:     uint(page),
			PageSize: uint(perPage),
		},
	)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	sharedhandlers.RespondOK(c, dto.NewListResponse(pageInfo, items))
}

func (h *Handler) ListDigitalWorkOrderDrafts(c *gin.Context) {
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 10)

	items, pageInfo, err := h.ucs.ListDigitalWorkOrderDrafts(
		c.Request.Context(),
		c.Query("number"),
		c.Query("status"),
		types.Input{
			Page:     uint(page),
			PageSize: uint(perPage),
		},
	)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	sharedhandlers.RespondOK(c, dto.NewListResponse(pageInfo, items))
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
