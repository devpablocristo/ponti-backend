// Package project expone endpoints HTTP para proyectos.
package project

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	domainField "github.com/alphacodinggroup/ponti-backend/internal/field/usecases/domain"
	dto "github.com/alphacodinggroup/ponti-backend/internal/project/handler/dto"
	domain "github.com/alphacodinggroup/ponti-backend/internal/project/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/internal/shared/domain"
)

type UseCasesPort interface {
	CreateProject(context.Context, *domain.Project) (int64, error)
	GetProjects(context.Context, string, int64, int64, int, int) ([]domain.Project, decimal.Decimal, int64, error)
	ListProjects(context.Context, int, int) ([]domain.ListedProject, int64, error)
	ListProjectsByCustomerID(context.Context, int64, int, int) ([]domain.ListedProject, int64, error)
	ListProjectsByName(context.Context, string, int, int) ([]domain.ListedProject, int64, error)
	GetFieldsByProjectID(ctx context.Context, projectID int64) ([]domainField.Field, error)
	GetProject(context.Context, int64) (*domain.Project, error)
	UpdateProject(context.Context, *domain.Project) error
	DeleteProject(context.Context, int64) error
	RestoreProject(context.Context, int64) error
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
		public.GET("/:id/fields", h.GetFieldsByProjectID)
		public.GET("/dropdown", h.ListProjectsDropdown)
		public.GET("/customer/:id", h.ListProjectsByCustomerID)
		public.GET("/:id", h.GetProject)
		public.PUT("/:id", h.UpdateProject)
		public.PUT("/:id/restore", h.RestoreProject)
		public.DELETE("/:id", h.DeleteProject)
		public.GET("/search", h.ListProjectsByName)
	}
}

// CreateProject crea un proyecto.
func (h *Handler) CreateProject(c *gin.Context) {
	var req dto.Project
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}

	ctx := c.Request.Context()
	pID, err := h.ucs.CreateProject(ctx, req.ToDomain())
	if err != nil {
		switch {
		case types.IsConflict(err):
			c.JSON(http.StatusConflict, types.ErrorResponse{Error: err.Error()})
			return
		default:
			c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
			return
		}
	}
	c.JSON(http.StatusCreated, dto.CreateProjectResponse{Message: "created", ProjectID: pID})
}

func (h *Handler) ListProjects(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "100"))
	customerID, _ := strconv.ParseInt(c.Query("customer_id"), 10, 64)
	name := c.Query("name")
	campaignID, _ := strconv.ParseInt(c.Query("campaign_id"), 10, 64)

	// Obtener los proyectos ligeros y total
	items, totalHectares, total, err := h.ucs.GetProjects(c.Request.Context(), name, customerID, campaignID, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}

	// Construir y devolver la respuesta paginada
	resp := dto.NewProjectsResponse(items, page, perPage, total)
	resp.TotalHectares = totalHectares
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetFieldsByProjectID(c *gin.Context) {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	fields, err := h.ucs.GetFieldsByProjectID(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	dtos := make([]dto.Field, len(fields))
	for i, f := range fields {
		dtos[i] = dto.FieldsFromDomain(f)
	}
	c.JSON(http.StatusOK, dtos)
}

// ListProjectsDropdown maneja el endpoint GET /projects con paginación ligera.
func (h *Handler) ListProjectsDropdown(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "100"))

	// Obtener los proyectos ligeros y total
	items, total, err := h.ucs.ListProjects(c.Request.Context(), page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}

	// Construir y devolver la respuesta paginada
	resp := dto.NewListProjectsResponse(items, page, perPage, total)
	c.JSON(http.StatusOK, resp)
}

// ListProjectsByCustomerID maneja GET /projects/customer/:id con paginación ligera
func (h *Handler) ListProjectsByCustomerID(c *gin.Context) {
	customerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid customer id"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "100"))

	items, total, err := h.ucs.ListProjectsByCustomerID(c.Request.Context(), customerID, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}

	resp := dto.NewListProjectsResponse(items, page, perPage, total)
	c.JSON(http.StatusOK, resp)
}

// GetProject devuelve un proyecto por ID.
func (h *Handler) GetProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid project id"})
		return
	}
	proj, err := h.ucs.GetProject(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.FromDomain(proj))
}

// UpdateProject actualiza un proyecto.
func (h *Handler) UpdateProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid project id"})
		return
	}
	var req dto.Project
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	if req.UpdatedAt == nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "updated_at is required"})
		return
	}
	dom := req.ToDomain()
	dom.ID = id
	dom.Base = shareddomain.Base{
		UpdatedAt: *req.UpdatedAt,
	}
	if err := h.ucs.UpdateProject(c.Request.Context(), dom); err != nil {
		switch {
		case types.IsNotFound(err):
			c.JSON(http.StatusNotFound, types.ErrorResponse{Error: err.Error()})
			return
		default:
			c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "updated"})
}

// DeleteProject elimina un proyecto por ID.
func (h *Handler) DeleteProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid project id"})
		return
	}
	if err := h.ucs.DeleteProject(c.Request.Context(), id); err != nil {
		switch {
		case types.IsNotFound(err):
			c.JSON(http.StatusNotFound, types.ErrorResponse{Error: err.Error()})
			return
		default:
			c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "deleted"})
}

// RestoreProject restaura un proyecto eliminado junto con todas sus entidades relacionadas
func (h *Handler) RestoreProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid project id"})
		return
	}
	if err := h.ucs.RestoreProject(c.Request.Context(), id); err != nil {
		switch {
		case types.IsNotFound(err):
			c.JSON(http.StatusNotFound, types.ErrorResponse{Error: err.Error()})
			return
		case types.IsValidationError(err):
			c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
			return
		default:
			c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "project restored successfully"})
}

func (h *Handler) ListProjectsByName(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "100"))
	name := c.Query("name")
	var (
		items []dto.ListedProject
		total int64
		err   error
	)
	if name != "" {
		// Delegar al caso de uso ListProjectsByName
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
		// Usar ListProjects como fallback
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
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	resp := dto.ListProjectsResponse{
		Data: items,
		PageInfo: dto.PageInfo{
			PerPage: perPage,
			Page:    page,
			MaxPage: int((total + int64(perPage) - 1) / int64(perPage)),
			Total:   total,
		},
	}
	c.JSON(http.StatusOK, resp)
}
