package project

import (
	"net/http"
	"strconv"

	mdw "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	gsv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	dto "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/handler/dto"
	"github.com/gin-gonic/gin"
)

// Handler encapsulates all dependencies for the Project HTTP handler.
type Handler struct {
	ucs UseCases
	gsv gsv.Server
	mws *mdw.Middlewares
}

// NewHandler creates a new Project handler.
func NewHandler(s gsv.Server, u UseCases, m *mdw.Middlewares) *Handler {
	return &Handler{ucs: u, gsv: s, mws: m}
}

// Routes registers all project routes.
func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	apiV := h.gsv.GetApiVersion()
	base := "/api/" + apiV + "/projects"

	public := r.Group(base)
	{
		public.POST("", h.CreateProject)                        // Create a project
		public.GET("", h.ListProjects)                          // List all projects
		public.GET("/customer/:id", h.ListProjectsByCustomerID) // List projects by customer ID
		public.GET("/:id", h.GetProject)                        // Get a project by ID
		public.PUT("/:id", h.UpdateProject)                     // Update a project
		public.DELETE("/:id", h.DeleteProject)                  // Delete a project
	}
}

// CreateProject handles project creation.
func (h *Handler) CreateProject(c *gin.Context) {
	var req dto.CreateProject
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	pID, err := h.ucs.CreateProject(c.Request.Context(), req.ToDomain())
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.CreateProjectResponse{Message: "created", ProjectID: pID})
}

// ListProjects maneja el endpoint GET /projects con paginación ligera
func (h *Handler) ListProjects(c *gin.Context) {
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

// GetProject returns a single project by ID.
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

// UpdateProject handles a full project update.
func (h *Handler) UpdateProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid project id"})
		return
	}
	var req dto.UpdateProject
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	dom := req.ToDomain()
	dom.ID = id
	if err := h.ucs.UpdateProject(c.Request.Context(), dom); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "updated"})
}

// DeleteProject removes a project by ID.
func (h *Handler) DeleteProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid project id"})
		return
	}
	if err := h.ucs.DeleteProject(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "deleted"})
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
		// Delegate to use case ListProjectsByName
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
		// existing ListProjects
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
