package lot

import (
	"context"
	"net/http"
	"strconv"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	dto "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/handler/dto"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
	"github.com/gin-gonic/gin"
)

type UseCasesPort interface {
	CreateLot(context.Context, *domain.Lot) (int64, error)
	GetLot(context.Context, int64) (*domain.Lot, error)
	UpdateLot(context.Context, *domain.Lot) error
	DeleteLot(context.Context, int64) error
	ListLots(context.Context, int64) ([]domain.Lot, error)
	ListLotsByProject(context.Context, int64) ([]domain.Lot, error)
	ListLotsByProjectAndField(context.Context, int64, int64) ([]domain.Lot, error)
	ListLotsByProjectFieldAndCrop(context.Context, int64, int64, int64, string) ([]domain.Lot, error)
}

type GinEnginePort interface {
	GetRouter() *gin.Engine
	RunServer(context.Context) error
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
	baseURL := h.acf.APIBaseURL() + "/lots"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	public := r.Group(baseURL)
	{
		public.POST("", h.CreateLot)
		public.GET("", h.ListLots)
		public.GET("/:id", h.GetLot)
		public.PUT("/:id", h.UpdateLot)
		public.DELETE("/:id", h.DeleteLot)
	}
}

func (h *Handler) CreateLot(c *gin.Context) {
	var req dto.Lot
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	newID, err := h.ucs.CreateLot(c.Request.Context(), req.ToDomain())
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.CreateLotResponse{Message: "Lot created successfully", ID: newID})
}

func (h *Handler) ListLots(c *gin.Context) {
	projectID, _ := strconv.ParseInt(c.Query("project_id"), 10, 64)
	fieldID, _ := strconv.ParseInt(c.Query("field_id"), 10, 64)
	currentCropID, _ := strconv.ParseInt(c.Query("current_crop_id"), 10, 64)
	previousCropID, _ := strconv.ParseInt(c.Query("previous_crop_id"), 10, 64)

	var (
		lots []domain.Lot
		err  error
	)

	switch {
	// Caso 3: proyecto, campo y cultivo
	case projectID > 0 && fieldID > 0 && currentCropID > 0:
		lots, err = h.ucs.ListLotsByProjectFieldAndCrop(c.Request.Context(), projectID, fieldID, currentCropID, "current")
	case projectID > 0 && fieldID > 0 && previousCropID > 0:
		lots, err = h.ucs.ListLotsByProjectFieldAndCrop(c.Request.Context(), projectID, fieldID, previousCropID, "previous")

	// Caso 2: proyecto y campo
	case projectID > 0 && fieldID > 0:
		lots, err = h.ucs.ListLotsByProjectAndField(c.Request.Context(), projectID, fieldID)

	// Caso 1: solo proyecto
	case projectID > 0:
		lots, err = h.ucs.ListLotsByProject(c.Request.Context(), projectID)

	default:
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Missing required parameters"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	out := make([]dto.Lot, len(lots))
	for i := range lots {
		out[i] = *dto.FromDomain(&lots[i])
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) GetLot(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid lot id"})
		return
	}
	lot, err := h.ucs.GetLot(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.FromDomain(lot))
}

func (h *Handler) UpdateLot(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid lot id"})
		return
	}
	var req dto.Lot
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	dom := req.ToDomain()
	dom.ID = id
	if err := h.ucs.UpdateLot(c.Request.Context(), dom); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Lot updated successfully"})
}

func (h *Handler) DeleteLot(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid lot id"})
		return
	}
	if err := h.ucs.DeleteLot(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Lot deleted successfully"})
}
