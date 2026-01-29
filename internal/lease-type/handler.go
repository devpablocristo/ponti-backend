// Package leasetype expone endpoints HTTP para tipos de arrendamiento.
package leasetype

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	dto "github.com/alphacodinggroup/ponti-backend/internal/lease-type/handler/dto"
	domain "github.com/alphacodinggroup/ponti-backend/internal/lease-type/usecases/domain"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

type UseCasesPort interface {
	CreateLeaseType(context.Context, *domain.LeaseType) (int64, error)
	ListLeaseTypes(context.Context) ([]domain.LeaseType, error)
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
	var req dto.LeaseType
	if err := c.ShouldBindJSON(&req); err != nil {
		domErr := types.NewError(types.ErrBadRequest, "invalid request payload", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	dom := &domain.LeaseType{Name: req.Name}
	id, err := h.ucs.CreateLeaseType(c.Request.Context(), dom)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusCreated, dto.CreateLeaseTypeResponse{Message: "Lease type created", ID: id})
}

func (h *Handler) ListLeaseTypes(c *gin.Context) {
	list, err := h.ucs.ListLeaseTypes(c.Request.Context())
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	resp := make([]dto.LeaseType, len(list))
	for i, lt := range list {
		resp[i] = *dto.FromDomain(lt)
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetLeaseType(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("lease_type_id"), 10, 64)
	if err != nil {
		domErr := types.NewError(types.ErrInvalidID, "invalid lease type id", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	lt, err := h.ucs.GetLeaseType(c.Request.Context(), id)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, dto.FromDomain(*lt))
}

func (h *Handler) UpdateLeaseType(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("lease_type_id"), 10, 64)
	if err != nil {
		domErr := types.NewError(types.ErrInvalidID, "invalid lease type id", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	var req dto.LeaseType
	if err := c.ShouldBindJSON(&req); err != nil {
		domErr := types.NewError(types.ErrBadRequest, "invalid request payload", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	req.ID = id
	if err := h.ucs.UpdateLeaseType(c.Request.Context(), req.ToDomain()); err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Lease type updated"})
}

func (h *Handler) DeleteLeaseType(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("lease_type_id"), 10, 64)
	if err != nil {
		domErr := types.NewError(types.ErrInvalidID, "invalid lease type id", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	if err := h.ucs.DeleteLeaseType(c.Request.Context(), id); err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Lease type deleted"})
}
