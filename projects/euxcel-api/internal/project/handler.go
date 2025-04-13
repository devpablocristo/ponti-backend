package project

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/euxcel-backend/pkg/types"
	utils "github.com/alphacodinggroup/euxcel-backend/pkg/utils"

	mdw "github.com/alphacodinggroup/euxcel-backend/pkg/http/middlewares/gin"
	gsv "github.com/alphacodinggroup/euxcel-backend/pkg/http/servers/gin"
	dto "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/project/handler/dto"
)

// Handler encapsulates all dependencies for the Project HTTP handler.
type Handler struct {
	ucs UseCases
	gsv gsv.Server
	mws *mdw.Middlewares
}

// NewHandler creates a new Project handler.
func NewHandler(s gsv.Server, u UseCases, m *mdw.Middlewares) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		mws: m,
	}
}

// Routes registers all project routes.
func (h *Handler) Routes() {
	router := h.gsv.GetRouter()

	apiVersion := h.gsv.GetApiVersion()
	apiBase := "/api/" + apiVersion + "/projects"
	publicPrefix := apiBase + "/public"
	protectedPrefix := apiBase + "/protected"

	public := router.Group(publicPrefix)
	{
		public.POST("", h.CreateProject)       // Create a project
		public.GET("", h.ListProjects)         // List all projects
		public.GET("/:id", h.GetProject)       // Get a project by ID
		public.PUT("/:id", h.UpdateProject)    // Update a project
		public.DELETE("/:id", h.DeleteProject) // Delete a project
	}

	// Protected routes.
	protected := router.Group(protectedPrefix)
	{
		protected.Use(h.mws.Protected...)
		protected.GET("/ping", h.ProtectedPing) // Protected test endpoint
	}
}

func (h *Handler) ProtectedPing(c *gin.Context) {
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Protected Pong!",
	})
}

// CreateProject handles the creation of a new project.
func (h *Handler) CreateProject(c *gin.Context) {
	var req dto.CreateProject
	if err := utils.ValidateRequest(c, &req); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	newID, err := h.ucs.CreateProject(ctx, req.ToDomain())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.CreateProjectResponse{
		Message:   "Project created successfully",
		ProjectID: newID,
	})
}

// ListProjects retrieves all projects.
func (h *Handler) ListProjects(c *gin.Context) {
	projects, err := h.ucs.ListProjects(c.Request.Context())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, projects)
}

// GetProject retrieves a project by its ID.
func (h *Handler) GetProject(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid project id"})
		return
	}

	project, err := h.ucs.GetProject(c.Request.Context(), id)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, project)
}

// UpdateProject updates an existing project.
func (h *Handler) UpdateProject(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid project id"})
		return
	}
	var req dto.Project
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid payload"})
		return
	}
	req.ID = id
	if err := h.ucs.UpdateProject(c.Request.Context(), req.ToDomain()); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Project updated successfully"})
}

// DeleteProject deletes a project by its ID.
func (h *Handler) DeleteProject(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid project id"})
		return
	}
	if err := h.ucs.DeleteProject(c.Request.Context(), id); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Project deleted successfully"})
}
