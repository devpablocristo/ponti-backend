package supply

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	providerdomain "github.com/alphacodinggroup/ponti-backend/internal/provider/usecases/domain"
	sharedhandlers "github.com/alphacodinggroup/ponti-backend/internal/shared/handlers"
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
		filter domain.SupplyFilter,
		page int,
		perPage int,
		mode string,
	) ([]domain.Supply, int64, error)
	UpdateSuppliesBulk(ctx context.Context, supplies []domain.Supply) error
	ExportTableSupplies(ctx context.Context, filter domain.SupplyFilter) ([]byte, error)
	GetEntriesSupplyMovementsByProjectID(ctx context.Context, projectID int64) ([]*domain.SupplyMovement, error)
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
		supplies.GET("/:supply_id", h.GetSupply)
		supplies.PUT("/:supply_id", h.UpdateSupply)
		supplies.DELETE("/:supply_id", h.DeleteSupply)
	}

	supplyMovements := r.Group(baseURL + "/projects/:project_id/supply-movements")
	{
		supplyMovements.POST("", h.CreateSupplyMovement)
		supplyMovements.GET("", h.GetSupplyMovementsByProjectID)
		supplyMovements.GET("/export", h.ExportSupplyMovementsByProjectID)
		supplyMovements.GET("/providers", h.GetProviders)
		supplyMovements.PUT("/:supply_movement_id", h.UpdateSupplyMovementById)
		supplyMovements.DELETE("/:supply_movement_id", h.DeleteSupplyMovement)
	}

	// Editor de stock: mismas rutas que supply-movements, semántica para vista Stock
	stockMovements := r.Group(baseURL + "/projects/:project_id/stock-movements")
	{
		stockMovements.POST("", h.CreateSupplyMovement)
		stockMovements.GET("", h.GetSupplyMovementsByProjectID)
		stockMovements.GET("/export", h.ExportSupplyMovementsByProjectID)
		stockMovements.GET("/providers", h.GetProviders)
		stockMovements.PUT("/:stock_movement_id", h.UpdateSupplyMovementById)
		stockMovements.DELETE("/:stock_movement_id", h.DeleteSupplyMovement)
	}
}

func (h *Handler) CreateSupply(c *gin.Context) {
	var req createDto.SupplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domErr := types.NewError(types.ErrBadRequest, "invalid request payload", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	newID, err := h.ucs.CreateSupply(c.Request.Context(), req.ToDomain())
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Supply created successfully", "id": newID})
}

func (h *Handler) CreateSuppliesBulk(c *gin.Context) {
	var req []createDto.SupplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domErr := types.NewError(types.ErrBadRequest, "invalid request payload", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	supplies := make([]domain.Supply, len(req))
	for i := range req {
		supplies[i] = *req[i].ToDomain()
	}
	if err := h.ucs.CreateSuppliesBulk(c.Request.Context(), supplies); err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusCreated, types.MessageResponse{Message: "Supplies created successfully"})
}

func (h *Handler) ListSupplies(c *gin.Context) {
	workspaceFilter, err := sharedhandlers.ParseWorkspaceFilter(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	mode := c.Query("mode")

	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 1000)

	filter := domain.SupplyFilter{
		CustomerID: workspaceFilter.CustomerID,
		ProjectID:  workspaceFilter.ProjectID,
		CampaignID: workspaceFilter.CampaignID,
		FieldID:    workspaceFilter.FieldID,
	}
	supplies, total, err := h.ucs.ListSuppliesPaginated(c.Request.Context(), filter, page, perPage, mode)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	resp := listDto.NewListSuppliesResponse(supplies, page, perPage, total)
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetSupply(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("supply_id"), "supply_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	supply, err := h.ucs.GetSupply(c.Request.Context(), id)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, createDto.FromDomain(supply))
}

func (h *Handler) UpdateSupply(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("supply_id"), "supply_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req createDto.SupplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domErr := types.NewError(types.ErrBadRequest, "invalid request payload", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	dom := req.ToDomain()
	dom.ID = id
	if err := h.ucs.UpdateSupply(c.Request.Context(), dom); err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) DeleteSupply(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("supply_id"), "supply_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.DeleteSupply(c.Request.Context(), id); err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) UpdateSuppliesBulk(c *gin.Context) {
	var req []createDto.SupplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domErr := types.NewError(types.ErrBadRequest, "invalid request payload", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	supplies := make([]domain.Supply, len(req))
	for i := range req {
		supplies[i] = *req[i].ToDomain()
	}
	if err := h.ucs.UpdateSuppliesBulk(c.Request.Context(), supplies); err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) ExportTableSupplies(c *gin.Context) {
	workspaceFilter, err := sharedhandlers.ParseWorkspaceFilter(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	filter := domain.SupplyFilter{
		CustomerID: workspaceFilter.CustomerID,
		ProjectID:  workspaceFilter.ProjectID,
		CampaignID: workspaceFilter.CampaignID,
		FieldID:    workspaceFilter.FieldID,
	}

	data, err := h.ucs.ExportTableSupplies(c.Request.Context(), filter)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
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

	userID, err := sharedhandlers.ParseUserID(c)
	if handleError(err, c) {
		return
	}

	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if handleError(err, c) {
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		domErr := types.NewError(types.ErrBadRequest, "invalid request payload", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	var supplyMovementsResponse []createDto.CreateSupplyMovementResponse

	for _, supplyMovement := range req.SupplyMovements {
		var supplyMovementResponse createDto.CreateSupplyMovementResponse

		err = supplyMovement.Validate()
		if err != nil {
			supplyMovementResponse = createDto.NewErrorCreateSupplyMovementResponse(err.Error())
		} else {
			supplyMovementId, err := h.ucs.CreateSupplyMovement(ctx, supplyMovement.ToDomain(projectID, &userID))
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

	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if handleError(err, c) {
		return
	}

	supplyMovements, err := h.ucs.GetEntriesSupplyMovementsByProjectID(ctx, projectID)
	if handleError(err, c) {
		return
	}

	c.JSON(http.StatusOK, getDto.NewGetEntrySupplyMovementsResponse(supplyMovements))
}

func (h *Handler) DeleteSupplyMovement(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if handleError(err, c) {
		return
	}

	supplyMovementId, err := sharedhandlers.ParseMovementIDParam(c)
	if handleError(err, c) {
		return
	}

	err = h.ucs.DeleteSupplyMovement(ctx, id, supplyMovementId)
	if handleError(err, c) {
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) UpdateSupplyMovementById(c *gin.Context) {
	ctx := c.Request.Context()
	var req updateDto.UpdateSupplyMovementEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domErr := types.NewError(types.ErrBadRequest, "invalid request payload", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if handleError(err, c) {
		return
	}

	supplyMovementId, err := sharedhandlers.ParseMovementIDParam(c)
	if handleError(err, c) {
		return
	}

	userID, err := sharedhandlers.ParseUserID(c)
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
	err = h.ucs.UpdateSupplyMovement(ctx, req.ToDomain(projectID, &userID, supplyMovement))

	if handleError(err, c) {
		return
	}

	c.Status(http.StatusNoContent)
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
	apiErr, status := types.NewAPIError(err)
	c.JSON(status, apiErr.ToResponse())
	return true
}

func (h *Handler) ExportSupplyMovementsByProjectID(c *gin.Context) {
	ctx := c.Request.Context()

	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if handleError(err, c) {
		return
	}

	data, err := h.ucs.ExportSupplyMovementsByProjectID(ctx, projectID)
	if handleError(err, c) {
		return
	}

	filename := supplyExcel.DefaultFilename

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
