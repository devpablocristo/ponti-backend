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
	ListLotsByField(context.Context, int64) ([]domain.Lot, error)
	ListLotsByProject(context.Context, int64) ([]domain.Lot, error)
	ListLotsByProjectAndField(context.Context, int64, int64) ([]domain.Lot, error)
	ListLotsByProjectFieldAndCrop(context.Context, int64, int64, int64, string) ([]domain.Lot, error)
	GetLotKPIs(context.Context, int64, int64, int64, string) (*domain.LotKPIs, error)
	ListLotsTable(context.Context, int64, int64, int64, string, int, int) ([]domain.LotTable, int, float64, float64, error)
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
		public.GET("", h.ListLots)            // ÚNICO endpoint de lista
		public.GET("/kpis", h.GetLotKPIs)     // KPIs endpoint
		public.GET("/table", h.ListLotsTable) // Table endpoint
		public.GET("/:id", h.GetLot)
		public.PUT("/:id", h.UpdateLot)
		public.DELETE("/:id", h.DeleteLot)
	}
}

// --- Handlers ---

func (h *Handler) CreateLot(c *gin.Context) {
	var req dto.Lot
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}

	lotDomain, err := req.ToDomain()
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	newID, err := h.ucs.CreateLot(c.Request.Context(), lotDomain)
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
	case projectID > 0 && fieldID > 0 && currentCropID > 0:
		lots, err = h.ucs.ListLotsByProjectFieldAndCrop(c.Request.Context(), projectID, fieldID, currentCropID, "current")
	case projectID > 0 && fieldID > 0 && previousCropID > 0:
		lots, err = h.ucs.ListLotsByProjectFieldAndCrop(c.Request.Context(), projectID, fieldID, previousCropID, "previous")
	case projectID > 0 && fieldID > 0:
		lots, err = h.ucs.ListLotsByProjectAndField(c.Request.Context(), projectID, fieldID)
	case projectID > 0:
		lots, err = h.ucs.ListLotsByProject(c.Request.Context(), projectID)
	case fieldID > 0:
		lots, err = h.ucs.ListLotsByField(c.Request.Context(), fieldID)
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

	dom, err := req.ToDomain()
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	dom.ID = id
	if err := h.ucs.UpdateLot(c, dom); err != nil {
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
	if err := h.ucs.DeleteLot(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Lot deleted successfully"})
}

func (h *Handler) GetLotKPIs(c *gin.Context) {
	projectID, _ := strconv.ParseInt(c.Query("project_id"), 10, 64)
	fieldID, _ := strconv.ParseInt(c.Query("field_id"), 10, 64)
	cropID, _ := strconv.ParseInt(c.Query("crop_id"), 10, 64)
	cropType := c.DefaultQuery("crop_type", "current") // current | previous | both

	if projectID <= 0 {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "project_id is required"})
		return
	}

	kpis, err := h.ucs.GetLotKPIs(c, projectID, fieldID, cropID, cropType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.FromDomainKPIs(kpis))
}

func (h *Handler) ListLotsTable(c *gin.Context) {
	projectID, _ := strconv.ParseInt(c.Query("project_id"), 10, 64)
	fieldID, _ := strconv.ParseInt(c.Query("field_id"), 10, 64)
	cropID, _ := strconv.ParseInt(c.Query("crop_id"), 10, 64)
	cropType := c.DefaultQuery("crop_type", "current")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	rows, total, sumSowed, sumCost, err := h.ucs.ListLotsTable(c, projectID, fieldID, cropID, cropType, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	// Map domain → dto
	dtoRows := make([]dto.LotTable, len(rows))
	for i, row := range rows {
		dateMap := make(map[int]dto.LotDates)
		for _, date := range row.Dates {
			harvestDate := ""
			if date.HarvestDate != nil {
				harvestDate = date.HarvestDate.Format("2006-01-02")
			}
			dateMap[date.Sequence] = dto.LotDates{
				SowingDate:  date.SowingDate.Format("2006-01-02"),
				HarvestDate: harvestDate,
				Sequence:    date.Sequence,
			}
		}

		dates := make([]dto.LotDates, 3)
		for seq := 1; seq <= 3; seq++ {
			if d, ok := dateMap[seq]; ok {
				dates[seq-1] = d
			} else {
				dates[seq-1] = dto.LotDates{
					Sequence: seq,
				}
			}
		}

		dtoRows[i] = dto.LotTable{
			ID:             row.ID,
			ProjectName:    row.ProjectName,
			FieldName:      row.FieldName,
			LotName:        row.LotName,
			PreviousCropID: row.PreviousCropID,
			CurrentCropID:  row.CurrentCropID,
			PreviousCrop:   row.PreviousCrop,
			CurrentCrop:    row.CurrentCrop,
			Variety:        row.Variety,
			SowedArea:      row.SowedArea,
			Season:         row.Season,
			Dates:          dates,
			UpdatedAt:      row.UpdatedAt,
			CostPerHectare: row.CostPerHectare,
		}
	}
	c.JSON(http.StatusOK, dto.LotTableResponse{
		Rows:         dtoRows,
		Total:        total,
		SumSowedArea: sumSowed,
		SumCost:      sumCost,
	})
}
