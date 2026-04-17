package leasetype

import (
	"context"

	"github.com/gin-gonic/gin"

	dto "github.com/alphacodinggroup/ponti-backend/internal/lease-type/handler/dto"
	domain "github.com/alphacodinggroup/ponti-backend/internal/lease-type/usecases/domain"
	sharedhandlers "github.com/alphacodinggroup/ponti-backend/internal/shared/handlers"
)

type UseCasesPort interface {
	CreateLeaseType(context.Context, *domain.LeaseType) (int64, error)
	ListLeaseTypes(context.Context, int, int) ([]domain.LeaseType, int64, error)
	GetLeaseType(context.Context, int64) (*domain.LeaseType, error)
	UpdateLeaseType(context.Context, *domain.LeaseType) error
	DeleteLeaseType(context.Context, int64) error
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
	baseURL := h.acf.APIBaseURL() + "/lease-types"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	public := r.Group(baseURL)
	{
		public.POST("", h.CreateLeaseType)
		public.GET("", h.ListLeaseTypes)
		public.GET("/:lease_type_id", h.GetLeaseType)
		public.PUT("/:lease_type_id", h.UpdateLeaseType)
		public.DELETE("/:lease_type_id", h.DeleteLeaseType)
	}
}

func (h *Handler) CreateLeaseType(c *gin.Context) {
	var req dto.CreateLeaseTypeRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	id, err := h.ucs.CreateLeaseType(c.Request.Context(), req.ToDomain())
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondCreated(c, id)
}

func (h *Handler) ListLeaseTypes(c *gin.Context) {
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 1000)
	leaseTypes, total, err := h.ucs.ListLeaseTypes(c.Request.Context(), page, perPage)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.NewListLeaseTypesResponse(leaseTypes, page, perPage, total))
}

func (h *Handler) GetLeaseType(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("lease_type_id"), "lease_type_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	lt, err := h.ucs.GetLeaseType(c.Request.Context(), id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.LeaseTypeFromDomain(lt))
}

func (h *Handler) UpdateLeaseType(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("lease_type_id"), "lease_type_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req dto.UpdateLeaseTypeRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	if err := h.ucs.UpdateLeaseType(c.Request.Context(), req.ToDomain(id)); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) DeleteLeaseType(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("lease_type_id"), "lease_type_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.DeleteLeaseType(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}
