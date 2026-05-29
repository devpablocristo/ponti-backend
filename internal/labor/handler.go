package labor

import (
	"context"
	"net/http"
	"strconv"

	ginmw "github.com/devpablocristo/core/http/gin/go"
	"github.com/gin-gonic/gin"

	"github.com/devpablocristo/core/errors/go/domainerr"
	types "github.com/devpablocristo/ponti-backend/internal/shared/types"

	labexcel "github.com/devpablocristo/ponti-backend/internal/labor/excel"
	"github.com/devpablocristo/ponti-backend/internal/labor/handler/dto"
	"github.com/devpablocristo/ponti-backend/internal/labor/usecases/domain"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
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

	// Endpoints de labores asociados a un proyecto específico
	projectLaborsGroup := r.Group(baseURL+"/projects/:project_id/labors", h.mws.GetValidation()...)
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
	workOrderLaborsGroup := r.Group(baseURL+"/labors", h.mws.GetValidation()...)
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
		sharedhandlers.RespondError(c, err)
		return
	}

	userID, err := sharedmodels.ActorFromContext(c.Request.Context())
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	ctx := c.Request.Context()

	if err := sharedhandlers.BindJSON(c, &req); err != nil {
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
		sharedhandlers.RespondError(c, err)
		return
	}

	items, total, err := h.ucs.ListLabor(c.Request.Context(), page, perPage, projectID)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	resp := dto.NewListLaborsResponse(items, page, perPage, total)
	sharedhandlers.RespondOK(c, resp)
}

func (h *Handler) UpdateLabor(c *gin.Context) {
	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	userID, err := sharedmodels.ActorFromContext(c.Request.Context())
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	id, err := ginmw.ParseParamID(c, "labor_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req dto.Labor
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	req.ID = id
	dom := req.ToDomain(projectID, userID)

	if req.IsPartialPrice == nil {
		currentLabor, err := h.ucs.GetLabor(c.Request.Context(), id)
		if err != nil {
			sharedhandlers.RespondError(c, err)
			return
		}
		dom.IsPartialPrice = currentLabor.IsPartialPrice
	}

	if err := h.ucs.UpdateLabor(c.Request.Context(), dom); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) DeleteLabor(c *gin.Context) {
	if _, err := sharedhandlers.ParseProjectIDParam(c, "project_id"); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	id, err := ginmw.ParseParamID(c, "labor_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.DeleteLabor(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) CountWorkOrdersByLaborID(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "labor_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	count, err := h.ucs.CountWorkOrdersByLaborID(c.Request.Context(), id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, gin.H{"count": count})
}

func (h *Handler) DeleteLaborByID(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "labor_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.DeleteLabor(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) ListLaborCategories(c *gin.Context) {
	if _, err := sharedhandlers.ParseProjectIDParam(c, "project_id"); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	id, err := ginmw.ParseParamID(c, "type_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	laborCategories, err := h.ucs.ListLaborCategoriesByTypeID(c.Request.Context(), id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	resp := dto.NewLaborCategoriesListResponse(laborCategories)
	sharedhandlers.RespondOK(c, resp)
}

func (h *Handler) ListLaborByWorkOrder(c *gin.Context) {
	workOrderID, err := ginmw.ParseParamID(c, "work_order_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	items, err := h.ucs.ListLaborByWorkOrder(c.Request.Context(), workOrderID)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	resp := dto.ToLaborListResponse(items)
	sharedhandlers.RespondOK(c, resp)
}

func (h *Handler) ListGroupLaborByProject(c *gin.Context) {
	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	fieldIDParam := c.Query("field_id")
	if fieldIDParam == "" && projectID == 0 {
		domErr := domainerr.Validation("field_id or project_id is required")
		sharedhandlers.RespondError(c, domErr)
		return
	}

	var fieldID int64
	if fieldIDParam != "" {
		parsedFieldID, err := ginmw.ParseOptionalInt64Query(c, "field_id")
		if err != nil {
			sharedhandlers.RespondError(c, err)
			return
		}
		fieldID = *parsedFieldID
	}

	input := types.NewInput(c.Request)

	list, pageInfo, err := h.ucs.ListGroupLaborByWorkOrder(c.Request.Context(), input, projectID, fieldID)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	resp := dto.FromDomainList(pageInfo, list)

	sharedhandlers.RespondOK(c, resp)
}

func (h *Handler) ExportGroupLaborXLSX(c *gin.Context) {
	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	fieldIDParam := c.Query("field_id")
	if fieldIDParam == "" && projectID == 0 {
		domErr := domainerr.Validation("field_id or project_id is required")
		sharedhandlers.RespondError(c, domErr)
		return
	}

	var fieldID int64
	if fieldIDParam != "" {
		parsedFieldID, err := ginmw.ParseOptionalInt64Query(c, "field_id")
		if err != nil {
			sharedhandlers.RespondError(c, err)
			return
		}
		fieldID = *parsedFieldID
	}

	// Para exportación, usar un page_size muy grande para obtener todos los registros
	input := types.Input{
		Page:     1,
		PageSize: 100000, // Límite suficientemente grande para exportar todos
	}

	data, err := h.ucs.ExportGroupLaborXLSX(c.Request.Context(), input, projectID, fieldID)
	if err != nil {
		sharedhandlers.RespondError(c, err)
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
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.FromDomainMetrics(m))
}

func (h *Handler) ExportProjectLabors(c *gin.Context) {
	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	data, err := h.ucs.ExportAllGroupLabors(c.Request.Context(), projectID)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	filename := "labores_base_datos.xlsx"

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

func (h *Handler) ExportAllGroupLabors(c *gin.Context) {
	projectIDPtr, err := ginmw.ParseOptionalInt64Query(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if projectIDPtr == nil {
		sharedhandlers.RespondError(c, domainerr.Validation("project_id is required"))
		return
	}
	projectID := *projectIDPtr

	data, err := h.ucs.ExportAllGroupLabors(c.Request.Context(), projectID)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	filename := "labores_base_datos.xlsx"

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
