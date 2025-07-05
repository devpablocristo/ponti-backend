package supply

import (
	"context"
	"net/http"
	"strconv"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	dto "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/handler/dto"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/usecases/domain"
	"github.com/gin-gonic/gin"
)

type UseCasesPort interface {
	CreateSupply(context.Context, *domain.Supply) (int64, error)
	GetSupply(context.Context, int64) (*domain.Supply, error)
	UpdateSupply(context.Context, *domain.Supply) error
	DeleteSupply(context.Context, int64) error
	ListSuppliesByProject(context.Context, int64) ([]domain.Supply, error)
	ListSuppliesByProjectAndCampaign(context.Context, int64, int64) ([]domain.Supply, error)
}

type Handler struct {
	ucs UseCasesPort
}

func NewHandler(u UseCasesPort) *Handler {
	return &Handler{ucs: u}
}

func (h *Handler) Routes(r *gin.Engine, baseURL string) {
	group := r.Group(baseURL + "/supplies")
	group.POST("", h.CreateSupply)
	group.GET("", h.ListSupplies)
	group.GET("/:id", h.GetSupply)
	group.PUT("/:id", h.UpdateSupply)
	group.DELETE("/:id", h.DeleteSupply)
}

func (h *Handler) CreateSupply(c *gin.Context) {
	var req dto.Supply
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	newID, err := h.ucs.CreateSupply(c.Request.Context(), req.ToDomain())
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Supply created successfully", "id": newID})
}

func (h *Handler) ListSupplies(c *gin.Context) {
	projectID, _ := strconv.ParseInt(c.Query("project_id"), 10, 64)
	campaignID, _ := strconv.ParseInt(c.Query("campaign_id"), 10, 64)

	var (
		supplies []domain.Supply
		err      error
	)
	if projectID > 0 && campaignID > 0 {
		supplies, err = h.ucs.ListSuppliesByProjectAndCampaign(c.Request.Context(), projectID, campaignID)
	} else if projectID > 0 {
		supplies, err = h.ucs.ListSuppliesByProject(c.Request.Context(), projectID)
	} else {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Missing required parameters"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	out := make([]dto.Supply, len(supplies))
	for i := range supplies {
		out[i] = *dto.FromDomain(&supplies[i])
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) GetSupply(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid supply id"})
		return
	}
	supply, err := h.ucs.GetSupply(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.FromDomain(supply))
}

func (h *Handler) UpdateSupply(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid supply id"})
		return
	}
	var req dto.Supply
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	dom := req.ToDomain()
	dom.ID = id
	if err := h.ucs.UpdateSupply(c.Request.Context(), dom); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Supply updated successfully"})
}

func (h *Handler) DeleteSupply(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid supply id"})
		return
	}
	if err := h.ucs.DeleteSupply(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Supply deleted successfully"})
}
