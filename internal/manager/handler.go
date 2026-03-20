// Package manager expone endpoints HTTP para managers.
package manager

import (
	"context"

	"github.com/gin-gonic/gin"

	dto "github.com/devpablocristo/ponti-backend/internal/manager/handler/dto"
	domain "github.com/devpablocristo/ponti-backend/internal/manager/usecases/domain"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

type UseCasesPort interface {
	CreateManager(context.Context, *domain.Manager) (int64, error)
	ListManagers(context.Context, int, int) ([]domain.Manager, int64, error)
	GetManager(context.Context, int64) (*domain.Manager, error)
	UpdateManager(context.Context, *domain.Manager) error
	DeleteManager(context.Context, int64) error
	ArchiveManager(context.Context, int64) error
	RestoreManager(context.Context, int64) error
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

// Handler encapsula dependencias del handler HTTP de Manager.
type Handler struct {
	ucs UseCasesPort
	gsv GinEnginePort
	acf ConfigAPIPort
	mws MiddlewaresEnginePort
}

// NewHandler crea un handler de Manager.
func NewHandler(u UseCasesPort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		acf: c,
		mws: m,
	}
}

// Routes registra las rutas del módulo Manager.
func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/managers"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	group := r.Group(baseURL)
	{
		group.POST("", h.CreateManager)
		group.GET("", h.ListManagers)
		group.GET("/:manager_id", h.GetManager)
		group.PUT("/:manager_id", h.UpdateManager)
		group.DELETE("/:manager_id", h.DeleteManager)
		group.POST("/:manager_id/archive", h.ArchiveManager)
		group.POST("/:manager_id/restore", h.RestoreManager)
	}
}

func (h *Handler) CreateManager(c *gin.Context) {
	var req dto.CreateManagerRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	id, err := h.ucs.CreateManager(c.Request.Context(), req.ToDomain())
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondCreated(c, id)
}

func (h *Handler) ListManagers(c *gin.Context) {
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 1000)
	items, total, err := h.ucs.ListManagers(c.Request.Context(), page, perPage)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.NewListManagersResponse(items, page, perPage, total))
}

func (h *Handler) GetManager(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("manager_id"), "manager_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	m, err := h.ucs.GetManager(c.Request.Context(), id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.ManagerFromDomain(m))
}

func (h *Handler) UpdateManager(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("manager_id"), "manager_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req dto.UpdateManagerRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	if err := h.ucs.UpdateManager(c.Request.Context(), req.ToDomain(id)); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) DeleteManager(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("manager_id"), "manager_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.DeleteManager(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) ArchiveManager(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("manager_id"), "manager_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.ArchiveManager(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) RestoreManager(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("manager_id"), "manager_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.RestoreManager(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}
