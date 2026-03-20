// Package project expone endpoints HTTP para proyectos.
package project

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	types "github.com/devpablocristo/ponti-backend/pkg/types"

	domainField "github.com/devpablocristo/ponti-backend/internal/field/usecases/domain"
	dto "github.com/devpablocristo/ponti-backend/internal/project/handler/dto"
	domain "github.com/devpablocristo/ponti-backend/internal/project/usecases/domain"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
)

type UseCasesPort interface {
	CreateProject(context.Context, *domain.Project) (int64, error)
	GetProjects(context.Context, string, int64, int64, int, int) ([]domain.Project, decimal.Decimal, int64, error)
	ListArchivedProjects(context.Context, int, int) ([]domain.Project, decimal.Decimal, int64, error)
	ListProjects(context.Context, int, int) ([]domain.ListedProject, int64, error)
	ListProjectsByCustomerID(context.Context, int64, int, int) ([]domain.ListedProject, int64, error)
	ListProjectsByName(context.Context, string, int, int) ([]domain.ListedProject, int64, error)
	GetFieldsByProjectID(ctx context.Context, projectID int64) ([]domainField.Field, error)
	GetProject(context.Context, int64) (*domain.Project, error)
	UpdateProject(context.Context, *domain.Project) error
	ArchiveProject(context.Context, int64) error
	RestoreProject(context.Context, int64) error
	DeleteProject(context.Context, int64) error
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

// Handler encapsula dependencias del handler HTTP de Project.
type Handler struct {
	ucs UseCasesPort
	gsv GinEnginePort
	acf ConfigAPIPort
	mws MiddlewaresEnginePort
}

// NewHandler crea un handler de Project.
func NewHandler(u UseCasesPort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		acf: c,
		mws: m,
	}
}

// Routes registra las rutas del módulo Project.
func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/projects"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	public := r.Group(baseURL)
	{
		public.POST("", h.CreateProject)
		public.GET("", h.ListProjects)
		public.GET("/archived", h.ListArchivedProjects)
		public.GET("/:project_id/fields", h.GetFieldsByProjectID)
		public.GET("/dropdown", h.ListProjectsDropdown)
		// Compatibilidad con ruta legacy usada en remoto.
		public.GET("/customer/:customer_id", h.ListProjectsByCustomerID)
		public.GET("/customers/:customer_id", h.ListProjectsByCustomerID)
		public.GET("/:project_id", h.GetProject)
		public.PUT("/:project_id", h.UpdateProject)
		public.POST("/:project_id/archive", h.ArchiveProject)
		public.POST("/:project_id/restore", h.RestoreProject)
		public.DELETE("/:project_id", h.DeleteProject)
		public.GET("/search", h.ListProjectsByName)
	}
}

// CreateProject crea un proyecto.
func (h *Handler) CreateProject(c *gin.Context) {
	var req dto.Project
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}

	ctx := c.Request.Context()
	pID, err := h.ucs.CreateProject(ctx, req.ToDomain())
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondCreated(c, pID)
}

func (h *Handler) ListProjects(c *gin.Context) {
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 100)
	customerID, err := sharedhandlers.ParseOptionalInt64Query(c, "customer_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	name := c.Query("name")
	campaignID, err := sharedhandlers.ParseOptionalInt64Query(c, "campaign_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	var customerIDValue int64
	if customerID != nil {
		customerIDValue = *customerID
	}
	var campaignIDValue int64
	if campaignID != nil {
		campaignIDValue = *campaignID
	}
	items, totalHectares, total, err := h.ucs.GetProjects(c.Request.Context(), name, customerIDValue, campaignIDValue, page, perPage)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	resp := dto.NewProjectsResponse(items, page, perPage, total)
	resp.TotalHectares = totalHectares
	sharedhandlers.RespondOK(c, resp)
}

// ListArchivedProjects lista proyectos archivados con clientes activos.
func (h *Handler) ListArchivedProjects(c *gin.Context) {
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 100)
	items, totalHectares, total, err := h.ucs.ListArchivedProjects(c.Request.Context(), page, perPage)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	resp := dto.NewProjectsResponse(items, page, perPage, total)
	resp.TotalHectares = totalHectares
	sharedhandlers.RespondOK(c, resp)
}

func (h *Handler) GetFieldsByProjectID(c *gin.Context) {
	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	fields, err := h.ucs.GetFieldsByProjectID(c.Request.Context(), projectID)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	dtos := make([]dto.Field, len(fields))
	for i, f := range fields {
		dtos[i] = dto.FieldsFromDomain(f)
	}
	sharedhandlers.RespondOK(c, dtos)
}

// ListProjectsDropdown maneja el endpoint GET /projects con paginación ligera.
func (h *Handler) ListProjectsDropdown(c *gin.Context) {
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 100)

	items, total, err := h.ucs.ListProjects(c.Request.Context(), page, perPage)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	resp := dto.NewListProjectsResponse(items, page, perPage, total)
	sharedhandlers.RespondOK(c, resp)
}

// ListProjectsByCustomerID maneja GET /projects/customers/:customer_id con paginación ligera
func (h *Handler) ListProjectsByCustomerID(c *gin.Context) {
	customerID, err := sharedhandlers.ParseParamID(c.Param("customer_id"), "customer_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 100)

	items, total, err := h.ucs.ListProjectsByCustomerID(c.Request.Context(), customerID, page, perPage)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	resp := dto.NewListProjectsResponse(items, page, perPage, total)
	sharedhandlers.RespondOK(c, resp)
}

// GetProject devuelve un proyecto por ID.
func (h *Handler) GetProject(c *gin.Context) {
	id, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	proj, err := h.ucs.GetProject(c.Request.Context(), id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.FromDomain(proj))
}

// UpdateProject actualiza un proyecto.
func (h *Handler) UpdateProject(c *gin.Context) {
	id, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req dto.Project
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	if req.UpdatedAt == nil {
		sharedhandlers.RespondError(c, types.NewError(types.ErrBadRequest, "updated_at is required", nil))
		return
	}
	dom := req.ToDomain()
	dom.ID = id
	dom.Base = shareddomain.Base{
		UpdatedAt: *req.UpdatedAt,
	}
	if err := h.ucs.UpdateProject(c.Request.Context(), dom); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

// ArchiveProject archiva un proyecto por ID.
func (h *Handler) ArchiveProject(c *gin.Context) {
	id, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.ArchiveProject(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

// RestoreProject restaura un proyecto eliminado junto con todas sus entidades relacionadas
func (h *Handler) RestoreProject(c *gin.Context) {
	id, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.RestoreProject(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

// DeleteProject elimina físicamente un proyecto por ID.
func (h *Handler) DeleteProject(c *gin.Context) {
	id, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.DeleteProject(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) ListProjectsByName(c *gin.Context) {
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 100)
	name := c.Query("name")
	var (
		items []dto.ListedProject
		total int64
		err   error
	)
	if name != "" {
		list, t, errName := h.ucs.ListProjectsByName(c.Request.Context(), name, page, perPage)
		if errName != nil {
			err = errName
		} else {
			total = t
			items = make([]dto.ListedProject, len(list))
			for i, p := range list {
				items[i] = dto.ListedProject{ID: p.ID, Name: p.Name}
			}
		}
	} else {
		list, t, errList := h.ucs.ListProjects(c.Request.Context(), page, perPage)
		if errList != nil {
			err = errList
		} else {
			total = t
			items = make([]dto.ListedProject, len(list))
			for i, p := range list {
				items[i] = dto.ListedProject{ID: p.ID, Name: p.Name}
			}
		}
	}
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	resp := dto.ListProjectsResponse{
		Items:    items,
		PageInfo: types.NewPageInfo(page, perPage, total),
	}
	sharedhandlers.RespondOK(c, resp)
}
