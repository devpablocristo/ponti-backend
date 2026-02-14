// Package lot expone los handlers HTTP para la entidad Lot.
package lot

import (
	// standard library
	"context"
	"net/http"
	"strconv"

	// third-party
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	// pkg
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	// excel
	lotExcel "github.com/alphacodinggroup/ponti-backend/internal/lot/excel"

	// project
	dto "github.com/alphacodinggroup/ponti-backend/internal/lot/handler/dto"
	domain "github.com/alphacodinggroup/ponti-backend/internal/lot/usecases/domain"
	sharedhandlers "github.com/alphacodinggroup/ponti-backend/internal/shared/handlers"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/internal/shared/models"
)

type UseCasesPort interface {
	CreateLot(context.Context, *domain.Lot) (int64, error)
	GetLot(context.Context, int64) (*domain.Lot, error)
	UpdateLot(context.Context, *domain.Lot) error
	UpdateLotTons(context.Context, int64, decimal.Decimal) error
	DeleteLot(context.Context, int64) error
	ListLotsByField(context.Context, int64) ([]domain.Lot, error)
	ListLotsByProject(context.Context, int64) ([]domain.Lot, error)
	ListLotsByProjectAndField(context.Context, int64, int64) ([]domain.Lot, error)
	ListLotsByProjectFieldAndCrop(context.Context, int64, int64, int64, string) ([]domain.Lot, error)
	GetMetrics(context.Context, int64, int64, int64) (*domain.LotMetrics, error)
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
	baseURL := h.acf.APIBaseURL() + "/lots"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	public := r.Group(baseURL)
	{
		public.POST("", ValidateLotRequest(), h.CreateLot)
		public.GET("", h.ListLots)
		public.GET("/metrics", h.GetMetrics)
		public.PUT("/:lot_id/tons", ValidateLotTonsUpdate(), h.UpdateLotTons)
		public.GET("/:lot_id", h.GetLot)
		public.PUT("/:lot_id", ValidateLotUpdate(), h.UpdateLot)
		public.DELETE("/:lot_id", h.DeleteLot)
		public.GET("/export", h.ExportLots)
	}
}

func (h *Handler) CreateLot(c *gin.Context) {
	// El lote ya fue validado por el middleware ValidateLotRequest
	req := c.MustGet("validated_lot").(*dto.Lot)

	lotDomain, err := req.ToDomain()
	if err != nil {
		types.NewErrorResponseHelper().InvalidPayload(c, types.NewError(types.ErrValidation, "invalid domain conversion", err))
		return
	}

	newID, err := h.ucs.CreateLot(c.Request.Context(), lotDomain)
	if err != nil {
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.CreateLotResponse{Message: "Lot created successfully", ID: newID})
}

func (h *Handler) ListLots(c *gin.Context) {
	workspaceFilter, err := sharedhandlers.ParseWorkspaceFilter(c)
	if err != nil {
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}
	var cropID *int64
	if raw := c.Query("crop_id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			types.NewErrorResponseHelper().InvalidPayload(c, types.NewError(types.ErrInvalidID, "invalid crop_id", err))
			return
		}
		cropID = &parsed
	}

	page, pageSize := sharedhandlers.ParsePaginationParams(c, 1, 1000)
	// Cap de paginación
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
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}

	pageInfo := types.NewPageInfo(page, pageSize, int64(total))
	resp := dto.FromDomainList(pageInfo, rows, sumSowed, sumCost)
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetLot(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("lot_id"), "lot_id")
	if err != nil {
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}
	lot, err := h.ucs.GetLot(c.Request.Context(), id)
	if err != nil {
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.FromDomain(lot))
}

func (h *Handler) UpdateLot(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("lot_id"), "lot_id")
	if err != nil {
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}

	// El lote ya fue validado por el middleware ValidateLotUpdate
	req := c.MustGet("validated_lot").(*dto.LotUpdate)

	dom, err := req.ToDomain()
	if err != nil {
		types.NewErrorResponseHelper().InvalidPayload(c, types.NewError(types.ErrValidation, "invalid domain conversion", err))
		return
	}
	dom.ID = id

	// Obtener user ID del contexto para campos de auditoría
	if userID, err := sharedmodels.ConvertStringToID(c.Request.Context()); err == nil {
		dom.Base.UpdatedBy = &userID
	}

	// Si el cliente no envía field_id, usamos el existente para evitar inconsistencias
	if dom.FieldID == 0 {
		if cur, getErr := h.ucs.GetLot(c.Request.Context(), id); getErr == nil {
			dom.FieldID = cur.FieldID
		} else {
			types.NewErrorResponseHelper().InvalidPayload(
				c,
				types.NewError(types.ErrInvalidID, "field_id is required", nil),
			)
			return
		}
	}
	if err := h.ucs.UpdateLot(c.Request.Context(), dom); err != nil {
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) UpdateLotTons(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("lot_id"), "lot_id")
	if err != nil {
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}

	// Las toneladas ya fueron validadas por el middleware ValidateLotTonsUpdate
	tons := c.MustGet("validated_tons").(decimal.Decimal)

	if err := h.ucs.UpdateLotTons(c.Request.Context(), id, tons); err != nil {
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) DeleteLot(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("lot_id"), "lot_id")
	if err != nil {
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}
	if err := h.ucs.DeleteLot(c.Request.Context(), id); err != nil {
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) GetMetrics(c *gin.Context) {
	projectID := int64(0)
	if raw := c.Query("project_id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			types.NewErrorResponseHelper().InvalidPayload(c, types.NewError(types.ErrInvalidID, "invalid project_id", err))
			return
		}
		projectID = parsed
	}
	fieldID := int64(0)
	if raw := c.Query("field_id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			types.NewErrorResponseHelper().InvalidPayload(c, types.NewError(types.ErrInvalidID, "invalid field_id", err))
			return
		}
		fieldID = parsed
	}
	cropID := int64(0)
	if raw := c.Query("crop_id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			types.NewErrorResponseHelper().InvalidPayload(c, types.NewError(types.ErrInvalidID, "invalid crop_id", err))
			return
		}
		cropID = parsed
	}

	// Los filtros por ID son opcionales para permitir búsquedas globales
	// Si no se proporcionan filtros, se retornan métricas de todos los lotes

	m, err := h.ucs.GetMetrics(c.Request.Context(), projectID, fieldID, cropID)
	if err != nil {
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.FromDomainMetrics(m))
}

func (h *Handler) ExportLots(c *gin.Context) {
	workspaceFilter, err := sharedhandlers.ParseWorkspaceFilter(c)
	if err != nil {
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}
	var cropID *int64
	if raw := c.Query("crop_id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			types.NewErrorResponseHelper().InvalidPayload(c, types.NewError(types.ErrInvalidID, "invalid crop_id", err))
			return
		}
		cropID = &parsed
	}

	page, pageSize := sharedhandlers.ParsePaginationParams(c, 1, 1000)
	// Cap de paginación
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
	data, err := h.ucs.ExportLots(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}

	filename := lotExcel.DefaultFilename

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
