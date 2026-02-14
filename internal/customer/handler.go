// Package customer expone endpoints HTTP para clientes.
package customer

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	dto "github.com/alphacodinggroup/ponti-backend/internal/customer/handler/dto"
	domain "github.com/alphacodinggroup/ponti-backend/internal/customer/usecases/domain"
	sharedhandlers "github.com/alphacodinggroup/ponti-backend/internal/shared/handlers"
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
	HardDeleteCustomer(context.Context, int64) error
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

// Handler encapsulates all dependencies for the Project HTTP handler.
type Handler struct {
	ucs UseCasesPort
	gsv GinEnginePort
	acf ConfigAPIPort
	mws MiddlewaresEnginePort
}

// NewHandler creates a new Project handler.
func NewHandler(u UseCasesPort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		acf: c,
		mws: m,
	}
}

// Routes registers all project routes.
func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/customers"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	public := r.Group(baseURL)
	{
		public.POST("", h.CreateCustomer)                // Crear un customer
		public.GET("", h.ListCustomers)                  // Listar todos los customers
		public.GET("/archived", h.ListArchivedCustomers) // Listar customers archivados
		public.GET("/:customer_id", h.GetCustomer)       // Obtener un customer por ID
		public.PUT("/:customer_id", h.UpdateCustomer)    // Actualizar un customer
		public.PUT("/:customer_id/archive", h.ArchiveCustomer)
		public.PUT("/:customer_id/restore", h.RestoreCustomer)
		public.DELETE("/:customer_id/hard", h.HardDeleteCustomer)
		public.DELETE("/:customer_id", h.DeleteCustomer) // Eliminar (soft) un customer
	}
}

// ListArchivedCustomers lista customers archivados.
func (h *Handler) ListArchivedCustomers(c *gin.Context) {
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 100)
	items, total, err := h.ucs.ListArchivedCustomers(c.Request.Context(), page, perPage)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	resp := dto.NewListCustomersResponse(items, page, perPage, total)
	c.JSON(http.StatusOK, resp)
}

// CreateCustomer maneja la creación de un nuevo customer.
func (h *Handler) CreateCustomer(c *gin.Context) {
	var req dto.CreateCustomer
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr, _ := types.NewAPIError(err)
		_ = c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	newID, err := h.ucs.CreateCustomer(ctx, req.ToDomain())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		_ = c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.CreateCustomerResponse{
		Message: "Customer created successfully",
		ID:      newID,
	})
}

// ListCustomers recupera todos los customers.
func (h *Handler) ListCustomers(c *gin.Context) {
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 100)

	items, total, err := h.ucs.ListCustomers(c.Request.Context(), page, perPage)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	resp := dto.NewListCustomersResponse(items, page, perPage, total)
	c.JSON(http.StatusOK, resp)
}

// GetCustomer recupera un customer por su ID.
func (h *Handler) GetCustomer(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("customer_id"), "customer_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	customer, err := h.ucs.GetCustomer(c.Request.Context(), id)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	c.JSON(http.StatusOK, customer)
}

// UpdateCustomer actualiza un customer existente.
func (h *Handler) UpdateCustomer(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("customer_id"), "customer_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req dto.Customer
	if err := c.ShouldBindJSON(&req); err != nil {
		domErr := types.NewError(types.ErrBadRequest, "invalid request payload", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	req.ID = id
	if err := h.ucs.UpdateCustomer(c.Request.Context(), req.ToDomain()); err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteCustomer elimina un customer por su ID.
func (h *Handler) DeleteCustomer(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("customer_id"), "customer_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.DeleteCustomer(c.Request.Context(), id); err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.Status(http.StatusNoContent)
}

// ArchiveCustomer archiva un customer por su ID.
func (h *Handler) ArchiveCustomer(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("customer_id"), "customer_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.ArchiveCustomer(c.Request.Context(), id); err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.Status(http.StatusNoContent)
}

// RestoreCustomer restaura un customer archivado.
func (h *Handler) RestoreCustomer(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("customer_id"), "customer_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.RestoreCustomer(c.Request.Context(), id); err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.Status(http.StatusNoContent)
}

// HardDeleteCustomer elimina físicamente un customer por ID.
func (h *Handler) HardDeleteCustomer(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("customer_id"), "customer_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.HardDeleteCustomer(c.Request.Context(), id); err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.Status(http.StatusNoContent)
}
