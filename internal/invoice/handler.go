package invoice

import (
	"context"
	"net/http"

	"github.com/alphacodinggroup/ponti-backend/internal/invoice/handler/dto"
	domain "github.com/alphacodinggroup/ponti-backend/internal/invoice/usecases/domain"
	sharedhandlers "github.com/alphacodinggroup/ponti-backend/internal/shared/handlers"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/gin-gonic/gin"
)

type UseCasePort interface {
	GetInvoiceByWorkOrder(context.Context, int64) (*domain.Invoice, error)
	CreateInvoice(context.Context, *domain.Invoice) (int64, error)
	UpdateInvoice(context.Context, *domain.Invoice) error
	DeleteInvoice(context.Context, int64) error
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
	ucs UseCasePort
	gsv GinEnginePort
	cfg ConfigAPIPort
	mws MiddlewaresEnginePort
}

func NewHandler(u UseCasePort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		cfg: c,
		mws: m,
	}
}

func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.cfg.APIBaseURL() + "/invoices"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	public := r.Group(baseURL)
	{
		public.GET("/:work_order_id", h.GetInvoiceByWorkOrder)
		public.POST("/:work_order_id", h.CreateInvoice)
		public.PUT("/:work_order_id", h.UpdateInvoice)
		public.DELETE("/:work_order_id", h.DeleteInvoice)

	}
}

func (h *Handler) GetInvoiceByWorkOrder(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("work_order_id"), "work_order_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	item, err := h.ucs.GetInvoiceByWorkOrder(c.Request.Context(), id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	resp := dto.FromDomain(item)

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateInvoice(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("work_order_id"), "work_order_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	userID, err := sharedhandlers.ParseUserID(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	var body dto.InvoiceRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		domErr := types.NewError(types.ErrBadRequest, "invalid request payload", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	item := body.ToDomain(id, userID)

	newID, err := h.ucs.CreateInvoice(c.Request.Context(), item)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	item.ID = newID
	c.JSON(http.StatusCreated, types.MessageResponse{Message: "Invoice saved"})
}

func (h *Handler) UpdateInvoice(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("work_order_id"), "work_order_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	userID, err := sharedhandlers.ParseUserID(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	var body dto.InvoiceRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		domErr := types.NewError(types.ErrBadRequest, "invalid request payload", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	item := body.ToDomain(id, userID)
	item.ID = id

	if err := h.ucs.UpdateInvoice(c.Request.Context(), item); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.MessageResponse{Message: "Invoice saved"})
}

func (h *Handler) DeleteInvoice(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("work_order_id"), "work_order_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	if err := h.ucs.DeleteInvoice(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

