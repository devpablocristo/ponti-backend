package supply

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	providerdomain "github.com/alphacodinggroup/ponti-backend/internal/provider/usecase/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/internal/shared/models"
	supplyExcel "github.com/alphacodinggroup/ponti-backend/internal/supply/excel"
	createDto "github.com/alphacodinggroup/ponti-backend/internal/supply/handler/dto/create"
	getDto "github.com/alphacodinggroup/ponti-backend/internal/supply/handler/dto/get"
	listDto "github.com/alphacodinggroup/ponti-backend/internal/supply/handler/dto/list"
	updateDto "github.com/alphacodinggroup/ponti-backend/internal/supply/handler/dto/update"
	domain "github.com/alphacodinggroup/ponti-backend/internal/supply/usecases/domain"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

type UseCasesPort interface {
	CreateSupply(ctx context.Context, s *domain.Supply) (int64, error)
	CreateSuppliesBulk(ctx context.Context, supplies []domain.Supply) error
	GetSupply(ctx context.Context, id int64) (*domain.Supply, error)
	UpdateSupply(ctx context.Context, s *domain.Supply) error
	DeleteSupply(ctx context.Context, id int64) error
	ListSuppliesPaginated(
		ctx context.Context,
		projectID int64,
		campaignID int64,
		page int,
		perPage int,
		mode string,
	) ([]domain.Supply, int64, error)
	UpdateSuppliesBulk(ctx context.Context, supplies []domain.Supply) error
	ExportTableSupplies(ctx context.Context, projectID int64) ([]byte, error)
	GetEntriesSupplyMovementsByProjectID(ctx context.Context, projectId int64) ([]*domain.SupplyMovement, error)
	CreateSupplyMovement(context.Context, *domain.SupplyMovement) (int64, error)
	GetSupplyMovementByID(context.Context, int64) (*domain.SupplyMovement, error)
	UpdateSupplyMovement(context.Context, *domain.SupplyMovement) error
	GetProviders(context.Context) ([]providerdomain.Provider, error)
	ExportSupplyMovementsByProjectID(ctx context.Context, projectID int64) ([]byte, error)
	DeleteSupplyMovement(context.Context, int64, int64) error
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
	baseURL := h.acf.APIBaseURL()

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	supplies := r.Group(baseURL + "/supplies")
	{
		supplies.POST("", h.CreateSupply)
		supplies.POST("/bulk", h.CreateSuppliesBulk)
		supplies.GET("", h.ListSupplies)
		supplies.GET("/export/all", h.ExportTableSupplies)
		supplies.PUT("/bulk", h.UpdateSuppliesBulk)
		supplies.GET("/:id", h.GetSupply)
		supplies.PUT("/:id", h.UpdateSupply)
		supplies.DELETE("/:id", h.DeleteSupply)
	}

	supplyMovements := r.Group(baseURL + "/projects/:id/supply-movements")
	{
		supplyMovements.POST("", h.CreateSupplyMovement)
		supplyMovements.GET("", h.GetSupplyMovementsByProjectID)
		supplyMovements.GET("/export", h.ExportSupplyMovementsByProjectID)
		supplyMovements.GET("/providers", h.GetProviders)
		supplyMovements.PUT("/:idSupplyMovement", h.UpdateSupplyMovementById)
		supplyMovements.DELETE("/:idSupplyMovement", h.DeleteSupplyMovement)
	}
}

func (h *Handler) CreateSupply(c *gin.Context) {
	var req createDto.SupplyRequest
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

func (h *Handler) CreateSuppliesBulk(c *gin.Context) {
	var req []createDto.SupplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	supplies := make([]domain.Supply, len(req))
	for i := range req {
		supplies[i] = *req[i].ToDomain()
	}
	if err := h.ucs.CreateSuppliesBulk(c, supplies); err != nil {
		code := http.StatusInternalServerError
		if types.IsErrInvalidInput(err) {
			code = http.StatusBadRequest
		} else if types.IsConflict(err) {
			code = http.StatusConflict
		}
		c.JSON(code, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, types.MessageResponse{Message: "Supplies created successfully"})
}

func (h *Handler) ListSupplies(c *gin.Context) {
	projectID, _ := strconv.ParseInt(c.Query("project_id"), 10, 64)
	campaignID, _ := strconv.ParseInt(c.Query("campaign_id"), 10, 64)
	mode := c.Query("mode")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "1000"))

	supplies, total, err := h.ucs.ListSuppliesPaginated(c.Request.Context(), projectID, campaignID, page, perPage, mode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}

	resp := listDto.NewListSuppliesResponse(supplies, page, perPage, total)
	c.JSON(http.StatusOK, resp)
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
	c.JSON(http.StatusOK, createDto.FromDomain(supply))
}

func (h *Handler) UpdateSupply(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid supply id"})
		return
	}
	var req createDto.SupplyRequest
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
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Supply deleted successfully"})
}

func (h *Handler) UpdateSuppliesBulk(c *gin.Context) {
	var req []createDto.SupplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	supplies := make([]domain.Supply, len(req))
	for i := range req {
		supplies[i] = *req[i].ToDomain()
	}
	if err := h.ucs.UpdateSuppliesBulk(c.Request.Context(), supplies); err != nil {
		code := http.StatusInternalServerError
		if types.IsErrInvalidInput(err) {
			code = http.StatusBadRequest
		}
		c.JSON(code, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Supplies updated successfully"})
}

func (h *Handler) ExportTableSupplies(c *gin.Context) {
	projectID, _ := strconv.ParseInt(c.Query("project_id"), 10, 64)

	data, err := h.ucs.ExportTableSupplies(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}

	filename := supplyExcel.DefaultFilename

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

func (h *Handler) CreateSupplyMovement(c *gin.Context) {
	ctx := c.Request.Context()
	var req createDto.CreateSupplyMovementRequestBulk

	userID, err := sharedmodels.ConvertStringToID(c)
	if handleError(err, c) {
		return
	}

	projectIdStr := c.Param("id")

	projectId, err := strconv.ParseInt(projectIdStr, 10, 64)
	if handleError(err, c) {
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}

	var supplyMovementsResponse []createDto.CreateSupplyMovementResponse

	for _, supplyMovement := range req.SupplyMovements {
		var supplyMovementResponse createDto.CreateSupplyMovementResponse

		err = supplyMovement.Validate()
		if err != nil {
			supplyMovementResponse = createDto.NewErrorCreateSupplyMovementResponse(err.Error())
		} else {
			supplyMovementId, err := h.ucs.CreateSupplyMovement(ctx, supplyMovement.ToDomain(projectId, &userID))
			if err != nil {
				supplyMovementResponse = createDto.NewErrorCreateSupplyMovementResponse(err.Error())
			} else {
				supplyMovementResponse = createDto.NewSuccessfulCreateSupplyMovementResponse(supplyMovementId)
			}
		}

		supplyMovementsResponse = append(supplyMovementsResponse, supplyMovementResponse)
	}

	c.JSON(http.StatusMultiStatus, createDto.CreateSupplyMovementBulkResponse{
		SupplyMovements: supplyMovementsResponse,
	})
}

func (h *Handler) GetSupplyMovementsByProjectID(c *gin.Context) {
	ctx := c.Request.Context()

	projectIdStr := c.Param("id")
	projectId, err := strconv.ParseInt(projectIdStr, 10, 64)
	if handleError(err, c) {
		return
	}

	supplyMovements, err := h.ucs.GetEntriesSupplyMovementsByProjectID(ctx, projectId)
	if handleError(err, c) {
		return
	}

	c.JSON(http.StatusOK, getDto.NewGetEntrySupplyMovementsResponse(supplyMovements))
}

func (h *Handler) DeleteSupplyMovement(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if handleError(err, c) {
		return
	}

	supplyMovementStr := c.Param("idSupplyMovement")
	supplyMovementId, err := strconv.ParseInt(supplyMovementStr, 10, 64)
	if handleError(err, c) {
		return
	}

	err = h.ucs.DeleteSupplyMovement(ctx, id, supplyMovementId)
	if handleError(err, c) {
		return
	}

	c.JSON(http.StatusOK, types.MessageResponse{Message: "supply movement deleted successfully"})
}

func (h *Handler) UpdateSupplyMovementById(c *gin.Context) {
	ctx := c.Request.Context()
	var req updateDto.UpdateSupplyMovementEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}

	projectIdStr := c.Param("id")
	projectId, err := strconv.ParseInt(projectIdStr, 10, 64)
	if handleError(err, c) {
		return
	}

	supplyMovementStr := c.Param("idSupplyMovement")
	supplyMovementId, err := strconv.ParseInt(supplyMovementStr, 10, 64)
	if handleError(err, c) {
		return
	}

	userID, err := sharedmodels.ConvertStringToID(c)
	if handleError(err, c) {
		return
	}

	supplyMovement, err := h.ucs.GetSupplyMovementByID(ctx, supplyMovementId)
	if handleError(err, c) {
		return
	}

	if err = req.Validate(); handleError(err, c) {
		return
	}
	err = h.ucs.UpdateSupplyMovement(ctx, req.ToDomain(projectId, &userID, supplyMovement))

	if handleError(err, c) {
		return
	}

	c.JSON(http.StatusOK, types.MessageResponse{Message: "supply movement updated successfully"})
}

func (h *Handler) GetProviders(c *gin.Context) {
	ctx := c.Request.Context()

	providers, err := h.ucs.GetProviders(ctx)
	if handleError(err, c) {
		return
	}

	c.JSON(http.StatusOK, getDto.NewGetProvidersResponse(providers))
}

func handleError(err error, c *gin.Context) bool {
	if err == nil {
		return false
	}
	apiErr, _ := types.NewAPIError(err)
	c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
	return true
}

func (h *Handler) ExportSupplyMovementsByProjectID(c *gin.Context) {
	ctx := c.Request.Context()

	projectIdStr := c.Param("id")
	projectId, err := strconv.ParseInt(projectIdStr, 10, 64)
	if handleError(err, c) {
		return
	}

	data, err := h.ucs.ExportSupplyMovementsByProjectID(ctx, projectId)
	if handleError(err, c) {
		return
	}

	filename := supplyExcel.DefaultFilename

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
