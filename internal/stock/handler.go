// Package stock expone endpoints HTTP para stock continuo por proyecto.
package stock

import (
	"context"
	"net/http"
	"time"

	ginmw "github.com/devpablocristo/core/http/gin/go"
	"github.com/gin-gonic/gin"

	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
	stockExcel "github.com/devpablocristo/ponti-backend/internal/stock/excel"
	"github.com/devpablocristo/ponti-backend/internal/stock/handler/dto"
	"github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
)

type UseCasesPort interface {
	GetStocksSummary(context.Context, int64, time.Time) ([]*domain.Stock, error)
	GetStockBySupplyID(context.Context, int64, int64, time.Time) (*domain.Stock, error)
	CreateStockCount(context.Context, int64, int64, *domain.StockCount) (int64, error)
	ExportStocksByProject(ctx context.Context, projectID int64) ([]byte, error)
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

func NewHandler(
	u UseCasesPort,
	s GinEnginePort,
	c ConfigAPIPort,
	m MiddlewaresEnginePort,
) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		acf: c,
		mws: m,
	}
}

func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL()

	stocks := r.Group(baseURL+"/projects/:project_id/stocks", h.mws.GetValidation()...)
	{
		stocks.GET("/summary", h.getStocksSummary)
		stocks.GET("/export", h.ExportStocksByProject)
	}

	counts := r.Group(baseURL+"/projects/:project_id/supplies/:supply_id/stock-counts", h.mws.GetValidation()...)
	{
		counts.POST("", h.CreateStockCount)
	}
}

func (h *Handler) getStocksSummary(c *gin.Context) {
	ctx := c.Request.Context()
	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	cutoffDate, err := parseCutoffDate(c.Query("cutoff_date"))
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	stocks, err := h.ucs.GetStocksSummary(ctx, projectID, cutoffDate)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	resp := dto.NewGetStocksListed(stocks)
	sharedhandlers.RespondOK(c, resp)
}

func (h *Handler) CreateStockCount(c *gin.Context) {
	ctx := c.Request.Context()

	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	supplyID, err := ginmw.ParseParamID(c, "supply_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	var req dto.CreateStockCountRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}

	if err := req.Validate(); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	actor, err := sharedmodels.ActorFromContext(ctx)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	id, err := h.ucs.CreateStockCount(ctx, projectID, supplyID, req.ToDomain(&actor))
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.NewCreateStockCountResponse(id, "stock count created successfully"))
}

func parseCutoffDate(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	return time.Parse("2006-01-02", value)
}

func (h *Handler) ExportStocksByProject(c *gin.Context) {
	ctx := c.Request.Context()
	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	data, err := h.ucs.ExportStocksByProject(ctx, projectID)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	filename := stockExcel.DefaultFilename
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
