package supply_movement

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/handler/dto"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/usecases/domain"
	"github.com/gin-gonic/gin"
)

// Handler for supply_movement operations

type UseCasesPort interface {
	GetSupplyMovements(context.Context, int64, int64, time.Time, time.Time) ([]*domain.SupplyMovement, error)
	CreateSupplyMovement(context.Context, *domain.SupplyMovement) (int64, error)
	GetSupplyMovementById(context.Context, int64) (*domain.SupplyMovement, error)
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
	ucs  UseCasesPort
	gsv  GinEnginePort
	acf  ConfigAPIPort
	mws  MiddlewaresEnginePort
	ucpp project.UseCasesPort
	ucps stock.UseCasesPort
}

func NewHandler(
	u UseCasesPort,
	s GinEnginePort,
	c ConfigAPIPort,
	m MiddlewaresEnginePort,
	ucpp project.UseCasesPort,
) *Handler {
	return &Handler{
		ucs:  u,
		gsv:  s,
		acf:  c,
		mws:  m,
		ucpp: ucpp,
	}
}

// --- HANDLER ENDPOINTS ---

// GetSupplyMovements handles GET /supply-movements
func (h *Handler) GetSupplyMovements(c *gin.Context) {
	projectIdStr := c.Query("project_id")
	supplyIdStr := c.Query("supply_id")
	fromDateStr := c.Query("from_date")
	toDateStr := c.Query("to_date")

	projectId, _ := strconv.ParseInt(projectIdStr, 10, 64)
	supplyId, _ := strconv.ParseInt(supplyIdStr, 10, 64)

	var fromDate, toDate time.Time
	var err error
	if fromDateStr != "" {
		fromDate, err = time.Parse(time.RFC3339, fromDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from_date"})
			return
		}
	}
	if toDateStr != "" {
		toDate, err = time.Parse(time.RFC3339, toDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to_date"})
			return
		}
	}

	movements, err := h.ucs.GetSupplyMovements(c.Request.Context(), projectId, supplyId, fromDate, toDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.FromDomainList(movements))
}

// CreateSupplyMovement handles POST /supply-movements
func (h *Handler) CreateSupplyMovement(c *gin.Context) {
	var req dto.CreateSupplyMovementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	movement := req.ToDomain()
	id, err := h.ucs.CreateSupplyMovement(c.Request.Context(), movement)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

// GetSupplyMovementById handles GET /supply-movements/:id
func (h *Handler) GetSupplyMovementById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	movement, err := h.ucs.GetSupplyMovementById(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.FromDomain(movement))
}
