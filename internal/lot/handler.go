// Package lot expone los handlers HTTP para la entidad Lot.
package lot

import (
	// standard library
	"context"
	"net/http"
	"strconv"

	// third-party
	ginmw "github.com/devpablocristo/core/http/gin/go"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	// pkg
	"github.com/devpablocristo/core/errors/go/domainerr"
	types "github.com/devpablocristo/ponti-backend/internal/shared/types"

	// project
	dto "github.com/devpablocristo/ponti-backend/internal/lot/handler/dto"
	domain "github.com/devpablocristo/ponti-backend/internal/lot/usecases/domain"
	"github.com/devpablocristo/ponti-backend/internal/shared/csvexport"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
)

type UseCasesPort interface {
	CreateLot(context.Context, *domain.Lot) (int64, error)
	GetLot(context.Context, int64) (*domain.Lot, error)
	UpdateLot(context.Context, *domain.Lot) error
	UpdateLotTons(context.Context, int64, decimal.Decimal) error
	ArchiveLot(context.Context, int64) error
	RestoreLot(context.Context, int64) error
	HardDeleteLot(context.Context, int64) error
	ListArchivedLots(context.Context, int, int) ([]domain.LotTable, int64, error)
	ListLotsByField(context.Context, int64) ([]domain.Lot, error)
	ListLotsByProject(context.Context, int64) ([]domain.Lot, error)
	ListLotsByProjectAndField(context.Context, int64, int64) ([]domain.Lot, error)
	ListLotsByProjectFieldAndCrop(context.Context, int64, int64, int64, string) ([]domain.Lot, error)
	GetMetrics(context.Context, domain.LotListFilter) (*domain.LotMetrics, error)
	ListLots(context.Context, domain.LotListFilter, int, int) ([]domain.LotTable, int, decimal.Decimal, decimal.Decimal, error)
	ExportLots(context.Context, domain.LotListFilter, int, int) ([]byte, error)
}

type GinEnginePort interface {
	GetRouter() *gin.Engine
	RunServer(context.Context) error
}

type ConfigAPIPort interface {
	APIVersion() string
	APIBaseURL() string
}

type MiddlewaresEnginePort interface {
	GetGlobal() []gin.HandlerFunc
	GetValidation() []gin.HandlerFunc
}

const maxLotExportPageSize = 1000

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

func (h *Handler) runLotIDAction(c *gin.Context, action func(context.Context, int64) error) {
	id, err := ginmw.ParseParamID(c, "lot_id")
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
	baseURL := h.acf.APIBaseURL() + "/lots"

	public := r.Group(baseURL, h.mws.GetValidation()...)
	{
		public.POST("", ValidateLotRequest(), h.CreateLot)
		public.GET("", h.ListLots)
		public.GET("/archived", h.ListArchivedLots)
		public.GET("/metrics", h.GetMetrics)
		public.PUT("/:lot_id/tons", ValidateLotTonsUpdate(), h.UpdateLotTons)
		public.GET("/:lot_id", h.GetLot)
		public.PUT("/:lot_id", ValidateLotUpdate(), h.UpdateLot)
		public.POST("/:lot_id/archive", h.ArchiveLot)
		public.POST("/:lot_id/restore", h.RestoreLot)
		public.DELETE("/:lot_id/hard", h.HardDeleteLot)
		public.GET("/export", h.ExportLots)
	}
}

func (h *Handler) CreateLot(c *gin.Context) {
	// El lote ya fue validado por el middleware ValidateLotRequest
	req := c.MustGet("validated_lot").(*dto.Lot)

	lotDomain, err := req.ToDomain()
	if err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid domain conversion"))
		return
	}

	newID, err := h.ucs.CreateLot(c.Request.Context(), lotDomain)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondCreated(c, newID)
}

func (h *Handler) ListLots(c *gin.Context) {
	workspaceFilter, err := sharedhandlers.ParseWorkspaceFilter(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var cropID *int64
	if raw := c.Query("crop_id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			sharedhandlers.RespondError(c, domainerr.Validation("invalid crop_id"))
			return
		}
		cropID = &parsed
	}

	page, pageSize := sharedhandlers.ParsePaginationParams(c, 1, 1000)
	if pageSize > 1000 {
		pageSize = 1000
	}

	filter := domain.LotListFilter{
		CustomerID: workspaceFilter.CustomerID,
		ProjectID:  workspaceFilter.ProjectID,
		CampaignID: workspaceFilter.CampaignID,
		FieldID:    workspaceFilter.FieldID,
		CropID:     cropID,
	}
	rows, total, sumSowed, sumCost, err := h.ucs.ListLots(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	pageInfo := types.NewPageInfo(page, pageSize, int64(total))
	resp := dto.FromDomainList(pageInfo, rows, sumSowed, sumCost)
	sharedhandlers.RespondOK(c, resp)
}

func (h *Handler) GetLot(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "lot_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	lot, err := h.ucs.GetLot(c.Request.Context(), id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.FromDomain(lot))
}

func (h *Handler) UpdateLot(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "lot_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	// El lote ya fue validado por el middleware ValidateLotUpdate
	req := c.MustGet("validated_lot").(*dto.LotUpdate)

	dom, err := req.ToDomain()
	if err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid domain conversion"))
		return
	}
	dom.ID = id

	// Obtener user ID del contexto para campos de auditoría
	if userID, err := sharedmodels.ActorFromContext(c.Request.Context()); err == nil {
		dom.UpdatedBy = &userID
	}

	// Si el cliente no envía field_id, usamos el existente para evitar inconsistencias
	if dom.FieldID == 0 {
		if cur, getErr := h.ucs.GetLot(c.Request.Context(), id); getErr == nil {
			dom.FieldID = cur.FieldID
		} else {
			sharedhandlers.RespondError(c, domainerr.Validation("field_id is required"))
			return
		}
	}
	if err := h.ucs.UpdateLot(c.Request.Context(), dom); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) UpdateLotTons(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "lot_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	// Las toneladas ya fueron validadas por el middleware ValidateLotTonsUpdate
	tons := c.MustGet("validated_tons").(decimal.Decimal)

	if err := h.ucs.UpdateLotTons(c.Request.Context(), id, tons); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) ArchiveLot(c *gin.Context) {
	h.runLotIDAction(c, h.ucs.ArchiveLot)
}

func (h *Handler) RestoreLot(c *gin.Context) {
	h.runLotIDAction(c, h.ucs.RestoreLot)
}

func (h *Handler) HardDeleteLot(c *gin.Context) {
	h.runLotIDAction(c, h.ucs.HardDeleteLot)
}

func (h *Handler) ListArchivedLots(c *gin.Context) {
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 1000)
	lots, total, err := h.ucs.ListArchivedLots(c.Request.Context(), page, perPage)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	items := make([]dto.LotListElement, len(lots))
	for i := range lots {
		items[i] = dto.FromDomainListElement(lots[i])
	}
	sharedhandlers.RespondOK(c, gin.H{
		"data":      items,
		"page_info": types.NewPageInfo(page, perPage, total),
	})
}

func (h *Handler) GetMetrics(c *gin.Context) {
	workspaceFilter, err := sharedhandlers.ParseWorkspaceFilter(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var cropID *int64
	if raw := c.Query("crop_id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			sharedhandlers.RespondError(c, domainerr.Validation("invalid crop_id"))
			return
		}
		cropID = &parsed
	}

	m, err := h.ucs.GetMetrics(c.Request.Context(), domain.LotListFilter{
		CustomerID: workspaceFilter.CustomerID,
		ProjectID:  workspaceFilter.ProjectID,
		CampaignID: workspaceFilter.CampaignID,
		FieldID:    workspaceFilter.FieldID,
		CropID:     cropID,
	})
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.FromDomainMetrics(m))
}

func (h *Handler) ExportLots(c *gin.Context) {
	workspaceFilter, err := sharedhandlers.ParseWorkspaceFilter(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var cropID *int64
	if raw := c.Query("crop_id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			sharedhandlers.RespondError(c, domainerr.Validation("invalid crop_id"))
			return
		}
		cropID = &parsed
	}

	filter := domain.LotListFilter{
		CustomerID: workspaceFilter.CustomerID,
		ProjectID:  workspaceFilter.ProjectID,
		CampaignID: workspaceFilter.CampaignID,
		FieldID:    workspaceFilter.FieldID,
		CropID:     cropID,
	}
	data, err := h.ucs.ExportLots(c.Request.Context(), filter, 1, maxLotExportPageSize)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	c.Header("Content-Type", csvexport.ContentType)
	c.Header("Content-Disposition", `attachment; filename="`+CSVExportFilename+`"`)
	c.Data(http.StatusOK, csvexport.ContentType, data)
}
