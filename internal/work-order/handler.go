// Package workorder expone endpoints HTTP para work orders.
package workorder

import (
	"context"
	"net/http"

	ginmw "github.com/devpablocristo/core/http/gin/go"
	"github.com/gin-gonic/gin"

	types "github.com/devpablocristo/ponti-backend/internal/shared/types"

	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
	workOrderExcel "github.com/devpablocristo/ponti-backend/internal/work-order/excel"
	"github.com/devpablocristo/ponti-backend/internal/work-order/handler/dto"
	"github.com/devpablocristo/ponti-backend/internal/work-order/usecases/domain"
)

type UseCasesPort interface {
	CreateWorkOrder(context.Context, *domain.WorkOrder) (int64, error)
	GetWorkOrderByID(context.Context, int64) (*domain.WorkOrder, error)
	DuplicateWorkOrder(context.Context, string) (string, error)
	UpdateWorkOrderByID(context.Context, *domain.WorkOrder) error
	DeleteWorkOrderByID(context.Context, int64) error
	ArchiveWorkOrder(context.Context, int64) error
	RestoreWorkOrder(context.Context, int64) error
	ListWorkOrders(context.Context, domain.WorkOrderFilter, types.Input) ([]domain.WorkOrderListElement, types.PageInfo, error)
	GetMetrics(context.Context, domain.WorkOrderFilter) (*domain.WorkOrderMetrics, error)
	ExportWorkOrders(context.Context, domain.WorkOrderFilter, types.Input) ([]byte, error)
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

// NewHandler crea un handler de work orders.
func NewHandler(u UseCasesPort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *Handler {
	return &Handler{ucs: u, gsv: s, acf: c, mws: m}
}

// Routes registra las rutas del módulo work orders.
func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	base := h.acf.APIBaseURL() + "/work-orders"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	grp := r.Group(base)
	{
		grp.POST("", h.CreateWorkOrder)
		grp.GET("/:work_order_id", h.GetWorkOrderByID)
		grp.PUT("/:work_order_id", h.UpdateWorkOrderByID)
		grp.DELETE("/:work_order_id", h.DeleteWorkOrderByID)
		grp.POST("/:work_order_id/archive", h.ArchiveWorkOrder)
		grp.POST("/:work_order_id/restore", h.RestoreWorkOrder)
		grp.POST("/:work_order_id/duplicate", h.DuplicateWorkOrder)
		grp.GET("", h.ListWorkOrders)
		grp.GET("/metrics", h.GetMetrics)
		grp.GET("/export", h.ExportWorkOrders)
	}
}

// CreateWorkOrder crea una orden de trabajo.
func (h *Handler) CreateWorkOrder(c *gin.Context) {
	var req dto.WorkOrder
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}

	id, err := h.ucs.CreateWorkOrder(c.Request.Context(), req.ToDomain())
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondCreated(c, id)
}

// GetWorkOrderByID obtiene una orden por ID.
func (h *Handler) GetWorkOrderByID(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "work_order_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	wo, err := h.ucs.GetWorkOrderByID(c.Request.Context(), id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.FromDomain(wo))
}

// DuplicateWorkOrder duplica una orden de trabajo.
func (h *Handler) DuplicateWorkOrder(c *gin.Context) {
	// orig := c.Param("work_order_number")
	// newNum, err := h.ucs.DuplicateWorkOrder(c.Request.Context(), orig)
	// if err != nil {
	// 	status, apiErr := httperr.Normalize(err)
	// 	c.JSON(status, apiErr.ToResponse())
	// 	return
	// }
	c.JSON(http.StatusCreated, dto.WorkOrderResponse{
		Message: "Work order duplicated",
		Number:  0,
	})
}

// UpdateWorkOrderByID actualiza una orden de trabajo.
func (h *Handler) UpdateWorkOrderByID(c *gin.Context) {
	var req dto.WorkOrder
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	id, err := ginmw.ParseParamID(c, "work_order_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	req.ID = id
	if err := h.ucs.UpdateWorkOrderByID(c.Request.Context(), req.ToDomain()); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

// DeleteWorkOrderByID elimina una orden de trabajo.
func (h *Handler) DeleteWorkOrderByID(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "work_order_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	if err := h.ucs.DeleteWorkOrderByID(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

// ListWorkOrders lista órdenes de trabajo con filtros.
func (h *Handler) ListWorkOrders(c *gin.Context) {
	filt := parseFilters(c)
	input := types.NewInput(c.Request)

	// Devuelve ([]domain.WorkOrderListElement, types.PageInfo, error)
	list, pageInfo, err := h.ucs.ListWorkOrders(c.Request.Context(), filt, input)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	sharedhandlers.RespondOK(c, dto.FromDomainList(pageInfo, list))
}

// parseFilters extrae project_id, field_id, customer_id y campaign_id.
func parseFilters(c *gin.Context) domain.WorkOrderFilter {
	f := domain.WorkOrderFilter{}
	workspaceFilter, err := sharedhandlers.ParseWorkspaceFilter(c)
	if err != nil {
		return f
	}
	f.ProjectID = workspaceFilter.ProjectID
	f.FieldID = workspaceFilter.FieldID
	f.CustomerID = workspaceFilter.CustomerID
	f.CampaignID = workspaceFilter.CampaignID
	return f
}

func (h *Handler) GetMetrics(c *gin.Context) {
	var filt domain.WorkOrderFilter
	workspaceFilter, err := sharedhandlers.ParseWorkspaceFilter(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	filt.ProjectID = workspaceFilter.ProjectID
	filt.FieldID = workspaceFilter.FieldID
	filt.CustomerID = workspaceFilter.CustomerID
	filt.CampaignID = workspaceFilter.CampaignID
	m, err := h.ucs.GetMetrics(c.Request.Context(), filt)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.FromDomainMetrics(m))
}

// ExportWorkOrders exporta órdenes de trabajo.
func (h *Handler) ExportWorkOrders(c *gin.Context) {
	filt := parseFilters(c)
	// Para exportación, usar un page_size muy grande para obtener todos los registros
	input := types.Input{
		Page:     1,
		PageSize: 100000, // Límite suficientemente grande para exportar todos
	}

	data, err := h.ucs.ExportWorkOrders(c.Request.Context(), filt, input)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	filename := workOrderExcel.DefaultFilename

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// ArchiveWorkOrder ejecuta soft delete (archivado) de la work order.
func (h *Handler) ArchiveWorkOrder(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "work_order_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.ArchiveWorkOrder(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

// RestoreWorkOrder restaura una work order archivada.
func (h *Handler) RestoreWorkOrder(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "work_order_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.RestoreWorkOrder(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}
