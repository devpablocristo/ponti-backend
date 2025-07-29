package stock

import (
	"context"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/handler/dto"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/usecases/domain"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

type UseCasesPort interface {
	GetStocksSummary(context.Context, int64, int64, time.Time) ([]*domain.Stock, error)
	CreateStock(context.Context, *domain.Stock) (int64, error)
	UpdateCloseDateByProjectAndField(context.Context, int64, int64, *domain.Stock) error
	UpdateRealStockUnits(context.Context, int64, *domain.Stock) error
	GetStockById(context.Context, int64) (*domain.Stock, error)
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
	ucs  UseCasesPort
	gsv  GinEnginePort
	acf  ConfigAPIPort
	mws  MiddlewaresEnginePort
	ucps project.UseCasesPort
}

func NewHandler(
	u UseCasesPort,
	s GinEnginePort,
	c ConfigAPIPort,
	m MiddlewaresEnginePort,
	ucps project.UseCasesPort) *Handler {
	return &Handler{
		ucs:  u,
		gsv:  s,
		acf:  c,
		mws:  m,
		ucps: ucps,
	}
}

func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/projects/:id/fields/:idField/stock"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}
	public := r.Group(baseURL)
	public.GET("/summary", h.getStocks)
	public.POST("", h.CreateStock)
	public.PUT("/close-date", h.UpdateStocksCloseDate)
	public.PUT("/real-stock/:stockId", h.UpdateRealStock)
	public.GET("/:stockId", h.GetStockById)
}

func (h *Handler) getStocks(c *gin.Context) {
	ctx := c.Request.Context()
	projectIdStr := c.Param("id")
	fieldIdStr := c.Param("idField")

	projectId, err := strconv.ParseInt(projectIdStr, 10, 64)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	_, err = h.ucps.GetProject(ctx, projectId)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	fieldId, err := strconv.ParseInt(fieldIdStr, 10, 64)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	cutoffDateStr := c.Query("cutoff_date")

	var cutoffDate time.Time

	if cutoffDateStr != "" {
		cutoffDate, err := time.Parse("2006-01-02", cutoffDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cutoff_date format, expected YYYY-MM-DD"})
			return
		}
		cutoffDate.UTC()
	}

	stocks, err := h.ucs.GetStocksSummary(ctx, projectId, fieldId, cutoffDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := dto.NewGetStocksListed(stocks)
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateStock(c *gin.Context) {
	var req dto.CreateStocksRequest
	ctx := c.Request.Context()
	projectIdStr := c.Param("id")
	fieldIdStr := c.Param("idField")

	projectId, err := strconv.ParseInt(projectIdStr, 10, 64)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	_, err = h.ucps.GetProject(ctx, projectId)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	fieldId, err := strconv.ParseInt(fieldIdStr, 10, 64)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	userID, err := sharedmodels.ConvertStringToID(c)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	_, err = h.ucps.GetProject(ctx, projectId)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	var responses []dto.CreateStocksResponse
	for _, stockReq := range req.Stocks {
		stockReq.ProjectID = projectId
		stockReq.FieldID = fieldId
		stockReq.CreatedBy = &userID
		stockReq.UpdatedBy = &userID
		stockId, err := h.ucs.CreateStock(ctx, stockReq.ToDomain())
		resp := dto.CreateStocksResponse{
			StockID:  stockId,
			SupplyID: stockReq.SupplyID,
			IsSaved:  err == nil,
		}
		if err != nil {
			resp.ErrorDetail = err.Error()
		}
		responses = append(responses, resp)
	}
	c.JSON(http.StatusMultiStatus, gin.H{"stocks": responses, "message": "stocks created"})
}

// UpdateStocksCloseDate actualiza el close_date de los stocks por proyecto y field
func (h *Handler) UpdateStocksCloseDate(c *gin.Context) {
	ctx := c.Request.Context()
	projectIdStr := c.Param("id")
	fieldIdStr := c.Param("idField")
	var req dto.UpdateCloseDateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	userID, err := sharedmodels.ConvertStringToID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid user_id in header"})
		return
	}
	req.UpdatedBy = &userID
	projectId, err := strconv.ParseInt(projectIdStr, 10, 64)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	_, err = h.ucps.GetProject(ctx, projectId)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	fieldId, err := strconv.ParseInt(fieldIdStr, 10, 64)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	err = h.ucs.UpdateCloseDateByProjectAndField(ctx, projectId, fieldId, req.ToDomain())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.UpdateCloseDateResponse{Message: "close_date updated successfully"})
}

func (h *Handler) UpdateRealStock(c *gin.Context) {
	ctx := c.Request.Context()
	stockIdStr := c.Param("stockId")
	var req dto.UpdateRealStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	userID, err := sharedmodels.ConvertStringToID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid user_id in header"})
		return
	}
	req.UpdatedBy = &userID
	stockId, err := strconv.ParseInt(stockIdStr, 10, 64)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	stockDomain, err := h.ucs.GetStockById(ctx, stockId)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	stockDomain.RealStockUnits = req.RealStockUnits
	err = h.ucs.UpdateRealStockUnits(ctx, stockId, stockDomain)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.UpdateRealStockResponse{Message: "real_stock_units updated successfully"})
}

func (h *Handler) GetStockById(c *gin.Context) {
	ctx := c.Request.Context()
	stockIdStr := c.Param("stockId")
	stockId, err := strconv.ParseInt(stockIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid stock id"})
		return
	}
	stock, err := h.ucs.GetStockById(ctx, stockId)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	resp := dto.NewGetStockByIdResponse(stock)
	c.JSON(http.StatusOK, resp)
}
