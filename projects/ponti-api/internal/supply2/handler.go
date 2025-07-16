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
	ListSupplies(context.Context, SupplyFilters) ([]domain.Supply, error)
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
	baseURL := h.acf.APIBaseURL() + "/supplies"

	public := r.Group(baseURL)
	{
		public.POST("", h.CreateSupply)
		public.GET("", h.ListSupplies)
		public.GET("/:id", h.GetSupply)
		public.PUT("/:id", h.UpdateSupply)
		public.DELETE("/:id", h.DeleteSupply)
	}
}

func (h *Handler) ListSupplies(c *gin.Context) {
	var filters SupplyFilters

	filters.ProjectID, _ = strconv.ParseInt(c.Query("project_id"), 10, 64)
	filters.CampaignID, _ = strconv.ParseInt(c.Query("campaign_id"), 10, 64)
	filters.FieldID, _ = strconv.ParseInt(c.Query("field_id"), 10, 64)
	filters.InvestorID, _ = strconv.ParseInt(c.Query("investor_id"), 10, 64)
	filters.EntryType = c.Query("entry_type")
	filters.Provider = c.Query("provider")
	filters.DeliveryNote = c.Query("delivery_note")
	filters.Search = c.Query("search")
	filters.Sort = c.Query("sort")
	filters.Order = c.DefaultQuery("order", "asc")
	filters.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "10"))
	filters.Offset, _ = strconv.Atoi(c.DefaultQuery("offset", "0"))

	supplies, err := h.ucs.ListSupplies(c.Request.Context(), filters)
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
