// Package stock expone endpoints HTTP para stock.
package stock

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/devpablocristo/saas-core/shared/domainerr"

	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
	stockExcel "github.com/devpablocristo/ponti-backend/internal/stock/excel"
	"github.com/devpablocristo/ponti-backend/internal/stock/handler/dto"
	"github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
)

type UseCasesPort interface {
	GetStocksSummary(context.Context, int64, time.Time) ([]*domain.Stock, error)
	GetStocksPeriods(context.Context, int64) ([]string, error)
	CreateStock(context.Context, *domain.Stock) (int64, error)
	UpdateCloseDateByProject(context.Context, int64, int64, int64, *domain.Stock) error
	UpdateRealStockUnits(context.Context, int64, *domain.Stock) error
	GetStockByID(context.Context, int64) (*domain.Stock, error)
	GetLastStockByProjectID(context.Context, int64, int64) (*domain.Stock, bool, error)
	ExportStocksByProject(ctx context.Context, projectID int64) ([]byte, error)
	UpdateUnitsConsumed(context.Context, domain.Stock, decimal.Decimal) error
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

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	public := r.Group(baseURL + "/projects/:project_id/stocks")
	{
		public.GET("/summary", h.getStocksSummary)
		public.GET("/periods", h.getStocksPeriods)
		public.GET("/export", h.ExportStocksByProject)
		public.PUT("/close-date", h.UpdateStocksCloseDate)
		public.PUT("/real-stock/:stock_id", h.UpdateRealStock)
	}
}

func (h *Handler) getStocksSummary(c *gin.Context) {
	ctx := c.Request.Context()
	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	cutoffDateStr := c.Query("cutoff_date")
	var cutoffDate time.Time
	if cutoffDateStr != "" {
		cutoffDate, err = time.Parse("2006-01-02", cutoffDateStr)
		if err != nil {
			sharedhandlers.RespondError(c, err)
			return
		}
	}

	stocks, err := h.ucs.GetStocksSummary(ctx, projectID, cutoffDate)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	resp := dto.NewGetStocksListed(stocks)
	sharedhandlers.RespondOK(c, resp)
}

func (h *Handler) getStocksPeriods(c *gin.Context) {
	ctx := c.Request.Context()
	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	periods, err := h.ucs.GetStocksPeriods(ctx, projectID)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	sharedhandlers.RespondOK(c, periods)
}

// UpdateStocksCloseDate actualiza el close_date de los stocks por proyecto y field
func (h *Handler) UpdateStocksCloseDate(c *gin.Context) {
	ctx := c.Request.Context()

	monthPeriod, err := getMonthPeriod(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	yearPeriod, err := getYearPeriod(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	var req dto.UpdateCloseDateRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	userID, err := sharedmodels.ActorFromContext(ctx)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	if err := req.Validate(); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	err = h.ucs.UpdateCloseDateByProject(
		ctx,
		projectID,
		monthPeriod,
		yearPeriod,
		req.ToDomain(&userID),
	)

	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	sharedhandlers.RespondOK(c, dto.UpdateCloseDateResponse{Message: "close_date updated successfully"})
}

func (h *Handler) UpdateRealStock(c *gin.Context) {
	ctx := c.Request.Context()
	stockIDStr := c.Param("stock_id")

	var req dto.UpdateRealStockRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}

	userID, err := sharedmodels.ActorFromContext(ctx)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	stockID, err := sharedhandlers.ParseParamID(stockIDStr, "stock_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	stockDomain, err := h.ucs.GetStockByID(ctx, stockID)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	stockDomain.RealStockUnits = req.RealStockUnits
	stockDomain.UpdatedBy = &userID

	err = h.ucs.UpdateRealStockUnits(ctx, stockID, stockDomain)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.NewUpdateRealStockResponse("real stock updated successfully"))
}

func getMonthPeriodOrDefault(c *gin.Context) (int64, error) {
	monthPeriodStr := c.Query("month_period")
	if monthPeriodStr == "" {
		return int64(time.Now().Month()), nil
	}
	monthPeriod, err := strconv.ParseInt(monthPeriodStr, 10, 64)
	if err != nil {
		return 0, domainerr.Validation("Month period is in invalid format")
	}

	if monthPeriod < 1 || monthPeriod > 12 {
		return 0, domainerr.Validation("Month period must be between 1 and 12")
	}
	return monthPeriod, nil
}

func getYearPeriodOrDefault(c *gin.Context) (int64, error) {
	yearPeriodStr := c.Query("year_period")
	if yearPeriodStr == "" {
		return int64(time.Now().Year()), nil
	}
	yearPeriod, err := strconv.ParseInt(yearPeriodStr, 10, 64)
	if err != nil {
		return 0, err
	}

	if yearPeriod < 0 {
		return 0, domainerr.Validation("Year period must be greater than 0")
	}

	return yearPeriod, nil
}

func getMonthPeriod(c *gin.Context) (int64, error) {
	monthPeriodStr := c.Query("month_period")

	if monthPeriodStr == "" {
		return 0, domainerr.Validation("The field 'month_period' is required")
	}
	return getMonthPeriodOrDefault(c)
}

func getYearPeriod(c *gin.Context) (int64, error) {
	yearPeriodStr := c.Query("year_period")

	if yearPeriodStr == "" {
		return 0, domainerr.Validation("The field 'year_period' is required")
	}
	return getYearPeriodOrDefault(c)
}

// ExportStocksByProject exporta stocks filtrados por proyecto
// Ruta nueva: /api/v1/projects/:project_id/stocks/export
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

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
