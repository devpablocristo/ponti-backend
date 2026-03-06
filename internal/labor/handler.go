package labor

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	labexcel "github.com/alphacodinggroup/ponti-backend/internal/labor/excel"
	"github.com/alphacodinggroup/ponti-backend/internal/labor/handler/dto"
	"github.com/alphacodinggroup/ponti-backend/internal/labor/usecases/domain"
	sharedhandlers "github.com/alphacodinggroup/ponti-backend/internal/shared/handlers"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/internal/shared/models"
)

type UseCasesPort interface {
	CreateLabor(context.Context, *domain.Labor) (int64, error)
	ListLabor(context.Context, int, int, int64) ([]domain.ListedLabor, int64, error)
	DeleteLabor(context.Context, int64) error
	UpdateLabor(context.Context, *domain.Labor) error
	CountWorkOrdersByLaborID(context.Context, int64) (int64, error)
	ListLaborCategoriesByTypeID(context.Context, int64) ([]domain.LaborCategory, error)
	ListLaborByWorkOrder(context.Context, int64) ([]domain.LaborRawItem, error)
	ListGroupLaborByWorkOrder(context.Context, types.Input, int64, int64) ([]domain.LaborListItem, types.PageInfo, error)
	ExportGroupLaborXLSX(context.Context, types.Input, int64, int64) ([]byte, error)
	ExportAllGroupLabors(context.Context, int64) ([]byte, error)
	GetMetrics(context.Context, domain.LaborFilter) (*domain.LaborMetrics, error)
	GetLabor(context.Context, int64) (*domain.Labor, error)
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

// Handler encapsulates all dependencies for the LeaseType HTTP handler.
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
	baseURL := h.acf.APIBaseURL()

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	// Endpoints de labores asociados a un proyecto específico
	projectLaborsGroup := r.Group(baseURL + "/projects/:project_id/labors")
	{
		projectLaborsGroup.POST("", h.CreateLabor)
		projectLaborsGroup.GET("", h.ListLabor)
		projectLaborsGroup.DELETE("/:labor_id", h.DeleteLabor)
		projectLaborsGroup.PUT("/:labor_id", h.UpdateLabor)
		projectLaborsGroup.GET("/:labor_id/workorders-count", h.CountWorkOrdersByLaborID)
		projectLaborsGroup.GET("/labor-categories/:type_id", h.ListLaborCategories)
		projectLaborsGroup.GET("/export", h.ExportProjectLabors)
	}

	// Endpoints de labores asociados a órdenes de trabajo y operaciones globales
	workOrderLaborsGroup := r.Group(baseURL + "/labors")
	{
		workOrderLaborsGroup.DELETE("/:labor_id", h.DeleteLaborByID)
		workOrderLaborsGroup.GET("/:work_order_id", h.ListLaborByWorkOrder)
		workOrderLaborsGroup.GET("/group/:project_id", h.ListGroupLaborByProject)
		workOrderLaborsGroup.GET("/export/:project_id", h.ExportGroupLaborXLSX)
		workOrderLaborsGroup.GET("/export/all", h.ExportAllGroupLabors)
		workOrderLaborsGroup.GET("/metrics", h.GetMetrics)
	}
}

func (h *Handler) CreateLabor(c *gin.Context) {
	var req dto.LaborList
	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		_ = c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	userID, err := sharedmodels.ConvertStringToID(c.Request.Context())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		_ = c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	ctx := c.Request.Context()

	if err = c.ShouldBindJSON(&req); err != nil {
		apiErr, _ := types.NewAPIError(err)
		_ = c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	var labors []dto.CreateLabor

	for _, labor := range req.Labors {
		var laborResponse dto.CreateLabor

		laborID, err := h.ucs.CreateLabor(ctx, labor.ToDomain(projectID, userID))
		if err != nil {
			laborResponse = dto.CreateLabor{
				LaborID:     0,
				LaborName:   labor.Name,
				IsSaved:     false,
				ErrorDetail: err.Error(),
			}
		} else {
			laborResponse = dto.CreateLabor{
				LaborID:   laborID,
				LaborName: labor.Name,
				IsSaved:   true,
			}
		}

		labors = append(labors, laborResponse)
	}

	c.JSON(http.StatusMultiStatus, dto.CreateLaborsResponse{
		Message: "labors created",
		Labors:  labors,
	})
}

func (h *Handler) ListLabor(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "100"))

	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		_ = c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	items, total, err := h.ucs.ListLabor(c.Request.Context(), page, perPage, projectID)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		_ = c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	resp := dto.NewListLaborsResponse(items, page, perPage, total)
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) UpdateLabor(c *gin.Context) {
	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		_ = c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	userID, err := sharedmodels.ConvertStringToID(c.Request.Context())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		_ = c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	id, err := sharedhandlers.ParseParamID(c.Param("labor_id"), "labor_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req dto.Labor
	if err := c.ShouldBindJSON(&req); err != nil {
		domErr := types.NewError(types.ErrBadRequest, "invalid request payload", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	req.ID = id
	dom := req.ToDomain(projectID, userID)

	if req.IsPartialPrice == nil {
		currentLabor, err := h.ucs.GetLabor(c.Request.Context(), id)
		if err != nil {
			apiErr, status := types.NewAPIError(err)
			c.JSON(status, apiErr.ToResponse())
			return
		}
		dom.IsPartialPrice = currentLabor.IsPartialPrice
	}

	if err := h.ucs.UpdateLabor(c.Request.Context(), dom); err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Labor updated successfully"})
}

func (h *Handler) DeleteLabor(c *gin.Context) {
	if _, err := sharedhandlers.ParseProjectIDParam(c, "project_id"); err != nil {
		domErr := types.NewError(types.ErrInvalidID, "invalid project_id", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	id, err := sharedhandlers.ParseParamID(c.Param("labor_id"), "labor_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.DeleteLabor(c.Request.Context(), id); err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Labor deleted successfully"})
}

func (h *Handler) CountWorkOrdersByLaborID(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("labor_id"), "labor_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	count, err := h.ucs.CountWorkOrdersByLaborID(c.Request.Context(), id)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (h *Handler) DeleteLaborByID(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("labor_id"), "labor_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.DeleteLabor(c.Request.Context(), id); err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Labor deleted successfully"})
}

func (h *Handler) ListLaborCategories(c *gin.Context) {
	if _, err := sharedhandlers.ParseProjectIDParam(c, "project_id"); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	id, err := sharedhandlers.ParseParamID(c.Param("type_id"), "type_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	laborCategories, err := h.ucs.ListLaborCategoriesByTypeID(c, id)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	resp := dto.NewLaborCategoriesListResponse(laborCategories)
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ListLaborByWorkOrder(c *gin.Context) {
	workOrderID, ok := parseParamID(c, "work_order_id")
	if !ok {
		return
	}

	items, err := h.ucs.ListLaborByWorkOrder(c.Request.Context(), workOrderID)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	resp := dto.ToLaborListResponse(items)
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ListGroupLaborByProject(c *gin.Context) {
	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	fieldIDParam := c.Query("field_id")
	if fieldIDParam == "" && projectID == 0 {
		domErr := types.NewError(types.ErrBadRequest, "field_id or project_id is required", nil)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	var fieldID int64
	if fieldIDParam != "" {
		var err error
		fieldID, err = strconv.ParseInt(fieldIDParam, 10, 64)
		if err != nil {
			domErr := types.NewError(types.ErrInvalidID, "invalid field_id", err)
			apiErr, status := types.NewAPIError(domErr)
			c.JSON(status, apiErr.ToResponse())
			return
		}
	}

	input := types.NewInput(c.Request)

	list, pageInfo, err := h.ucs.ListGroupLaborByWorkOrder(c.Request.Context(), input, projectID, fieldID)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		_ = c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	resp := dto.FromDomainList(pageInfo, list)

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ExportGroupLaborXLSX(c *gin.Context) {
	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	fieldIDParam := c.Query("field_id")
	if fieldIDParam == "" && projectID == 0 {
		types.NewErrorResponseHelper().BadRequest(c, "fieldID or projectID requires a value", nil)
		return
	}

	var fieldID int64
	if fieldIDParam != "" {
		var err error
		fieldID, err = strconv.ParseInt(fieldIDParam, 10, 64)
		if err != nil {
			types.NewErrorResponseHelper().BadRequest(c, "fieldID is not a valid integer", nil)
			return
		}
	}

	// Para exportación, usar un page_size muy grande para obtener todos los registros
	input := types.Input{
		Page:     1,
		PageSize: 100000, // Límite suficientemente grande para exportar todos
	}

	data, err := h.ucs.ExportGroupLaborXLSX(c.Request.Context(), input, projectID, fieldID)
	if err != nil {
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}

	filename := labexcel.DefaultFilename

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

func (h *Handler) GetMetrics(c *gin.Context) {
	var filt domain.LaborFilter
	workspaceFilter, err := sharedhandlers.ParseWorkspaceFilter(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	filt.CustomerID = workspaceFilter.CustomerID
	filt.ProjectID = workspaceFilter.ProjectID
	filt.CampaignID = workspaceFilter.CampaignID
	filt.FieldID = workspaceFilter.FieldID
	m, err := h.ucs.GetMetrics(c.Request.Context(), filt)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, dto.FromDomainMetrics(m))
}

func (h *Handler) ExportProjectLabors(c *gin.Context) {
	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	data, err := h.ucs.ExportAllGroupLabors(c.Request.Context(), projectID)
	if err != nil {
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}

	filename := "labores_base_datos.xlsx"

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

func (h *Handler) ExportAllGroupLabors(c *gin.Context) {
	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		types.NewErrorResponseHelper().BadRequest(c, "project_id is required", nil)
		return
	}
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil || projectID <= 0 {
		types.NewErrorResponseHelper().BadRequest(c, "project_id must be a valid positive integer", nil)
		return
	}

	data, err := h.ucs.ExportAllGroupLabors(c.Request.Context(), projectID)
	if err != nil {
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}

	filename := "labores_base_datos.xlsx"

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// ----- HELPER -----

func parseParamID(c *gin.Context, param string) (int64, bool) {
	raw := strings.TrimSpace(c.Param(param))
	if raw == "" {
		apiErr := types.NewError(types.ErrInvalidID, param+" is required", nil)
		_ = c.Error(apiErr).SetMeta(map[string]any{"param": param})
		return 0, false
	}

	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		apiErr := types.NewError(types.ErrInvalidID, param+" must be a positive integer", err)
		_ = c.Error(apiErr).SetMeta(map[string]any{"param": param, "value": raw})
		return 0, false
	}

	return id, true
}
