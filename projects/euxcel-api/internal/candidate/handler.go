package candidate

import (
	"net/http"

	"github.com/gin-gonic/gin"

	mdw "github.com/devpablocristo/monorepo/pkg/http/middlewares/gin"
	gsv "github.com/devpablocristo/monorepo/pkg/http/servers/gin"
	types "github.com/devpablocristo/monorepo/pkg/types"
	utils "github.com/devpablocristo/monorepo/pkg/utils"

	dto "github.com/devpablocristo/monorepo/projects/qh/internal/candidate/handler/dto"
)

// Handler gestiona los endpoints del recurso candidate.
type Handler struct {
	ucs UseCases
	gsv gsv.Server
	mws *mdw.Middlewares
}

// NewHandler crea una nueva instancia de Handler.
func NewHandler(s gsv.Server, u UseCases, m *mdw.Middlewares) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		mws: m,
	}
}

// Routes registra las rutas REST y el endpoint de WebSocket.
// Se centraliza la configuración del WebSocket en este método.
func (h *Handler) Routes() {
	router := h.gsv.GetRouter()

	apiVersion := h.gsv.GetApiVersion()
	apiBase := "/api/" + apiVersion + "/candidates"
	//publicPrefix := apiBase + "/public"
	validatedPrefix := apiBase + "/validated"
	protectedPrefix := apiBase + "/protected"

	// Rutas públicas
	// public := router.Group(publicPrefix)
	// {
	// 	public.POST("", h.CreateCandidate)
	// 	public.GET("", h.ListCandidates)
	// 	public.GET("/:id", h.GetCandidate)
	// 	public.PUT("/:id", h.UpdateCandidate)
	// 	public.DELETE("/:id", h.DeleteCandidate)
	// }

	validated := router.Group(validatedPrefix)
	{
		// Aplicar middleware de validación de credenciales
		validated.Use(h.mws.Validated...)
		// Puedes añadir rutas aquí si es necesario
	}

	// Rutas protegidas
	protected := router.Group(protectedPrefix)
	{
		protected.Use(h.mws.Protected...)

		protected.GET("/ping", h.ProtectedPing)

		protected.POST("", h.CreateCandidate)
		protected.GET("", h.ListCandidates)
		protected.GET("/:id", h.GetCandidate)
		protected.PUT("/:id", h.UpdateCandidate)
		protected.DELETE("/:id", h.DeleteCandidate)
	}
}

// ProtectedPing es un endpoint de ejemplo que responde "Protected Pong!".
func (h *Handler) ProtectedPing(c *gin.Context) {
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Protected Pong!",
	})
}

// CreateCandidate procesa la creación de un candidate.
func (h *Handler) CreateCandidate(c *gin.Context) {
	var req dto.CreateCandidate
	if err := utils.ValidateRequest(c, &req); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	ctx := c.Request.Context()
	newCandidateID, err := h.ucs.CreateCandidate(ctx, req.ToDomain())
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	c.JSON(http.StatusCreated, dto.CreateCandidateResponse{
		Message:     "Candidate created successfully",
		CandidateID: newCandidateID,
	})
}

// ListCandidates retorna la lista de candidates.
func (h *Handler) ListCandidates(c *gin.Context) {
	candidates, err := h.ucs.ListCandidates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching candidates"})
		return
	}
	c.JSON(http.StatusOK, candidates)
}

// GetCandidate retorna la información de un candidate en función del ID.
func (h *Handler) GetCandidate(c *gin.Context) {
	id := c.Param("id")

	person, err := h.ucs.GetCandidate(c.Request.Context(), id)
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	c.JSON(http.StatusOK, person)
}

// UpdateCandidate procesa la actualización de un candidate.
func (h *Handler) UpdateCandidate(c *gin.Context) {
	var updatedCandidate dto.Candidate
	if err := c.ShouldBindJSON(&updatedCandidate); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	if err := h.ucs.UpdateCandidate(c.Request.Context(), updatedCandidate.ToDomain()); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Candidate updated successfully",
	})
}

// DeleteCandidate procesa la eliminación de un candidate.
func (h *Handler) DeleteCandidate(c *gin.Context) {
	id := c.Param("id")
	if err := h.ucs.DeleteCandidate(c.Request.Context(), id); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Candidate deleted successfully",
	})
}
