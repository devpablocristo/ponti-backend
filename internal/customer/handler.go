// Package customer expone endpoints HTTP para clientes.
package customer

import (
	"context"

	ginmw "github.com/devpablocristo/core/http/gin/go"
	"github.com/gin-gonic/gin"

	dto "github.com/devpablocristo/ponti-backend/internal/customer/handler/dto"
	domain "github.com/devpablocristo/ponti-backend/internal/customer/usecases/domain"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

type UseCasesPort interface {
	CreateCustomer(context.Context, *domain.Customer) (int64, error)
	ListCustomers(context.Context, int, int) ([]domain.ListedCustomer, int64, error)
	ListArchivedCustomers(context.Context, int, int) ([]domain.ListedCustomer, int64, error)
	GetCustomer(context.Context, int64) (*domain.Customer, error)
	UpdateCustomer(context.Context, *domain.Customer) error
	DeleteCustomer(context.Context, int64) error
	ArchiveCustomer(context.Context, int64) error
	RestoreCustomer(context.Context, int64) error
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
	baseURL := h.acf.APIBaseURL() + "/customers"

	public := r.Group(baseURL, h.mws.GetValidation()...)
	{
		public.POST("", h.CreateCustomer)
		public.GET("", h.ListCustomers)
		public.GET("/archived", h.ListArchivedCustomers)
		public.GET("/:customer_id", h.GetCustomer)
		public.PUT("/:customer_id", h.UpdateCustomer)
		public.DELETE("/:customer_id", h.DeleteCustomer)
		public.POST("/:customer_id/archive", h.ArchiveCustomer)
		public.POST("/:customer_id/restore", h.RestoreCustomer)
	}
}

func (h *Handler) CreateCustomer(c *gin.Context) {
	var req dto.CreateCustomerRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	id, err := h.ucs.CreateCustomer(c.Request.Context(), req.ToDomain())
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondCreated(c, id)
}

func (h *Handler) ListCustomers(c *gin.Context) {
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 100)

	status := c.DefaultQuery("status", "active")
	switch status {
	case "archived":
		h.listByStatus(c, page, perPage, true)
	case "all":
		h.listAll(c, page, perPage)
	default:
		h.listByStatus(c, page, perPage, false)
	}
}

// ListArchivedCustomers mantiene compatibilidad con GET /archived.
func (h *Handler) ListArchivedCustomers(c *gin.Context) {
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 100)
	h.listByStatus(c, page, perPage, true)
}

func (h *Handler) listByStatus(c *gin.Context, page, perPage int, archived bool) {
	var (
		items []domain.ListedCustomer
		total int64
		err   error
	)
	if archived {
		items, total, err = h.ucs.ListArchivedCustomers(c.Request.Context(), page, perPage)
	} else {
		items, total, err = h.ucs.ListCustomers(c.Request.Context(), page, perPage)
	}
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.NewListCustomersResponse(items, page, perPage, total))
}

func (h *Handler) listAll(c *gin.Context, page, perPage int) {
	active, totalA, err := h.ucs.ListCustomers(c.Request.Context(), page, perPage)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	archived, totalAr, err := h.ucs.ListArchivedCustomers(c.Request.Context(), page, perPage)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	all := append(active, archived...)
	sharedhandlers.RespondOK(c, dto.NewListCustomersResponse(all, page, perPage, totalA+totalAr))
}

func (h *Handler) GetCustomer(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "customer_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	cust, err := h.ucs.GetCustomer(c.Request.Context(), id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.CustomerFromDomain(cust))
}

func (h *Handler) UpdateCustomer(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "customer_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req dto.UpdateCustomerRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	if err := h.ucs.UpdateCustomer(c.Request.Context(), req.ToDomain(id)); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

// DeleteCustomer ejecuta hard delete del customer.
func (h *Handler) DeleteCustomer(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "customer_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.DeleteCustomer(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

// ArchiveCustomer ejecuta soft delete (archivado) del customer.
func (h *Handler) ArchiveCustomer(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "customer_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.ArchiveCustomer(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) RestoreCustomer(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "customer_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.RestoreCustomer(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}
