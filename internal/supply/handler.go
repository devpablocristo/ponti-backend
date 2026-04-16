package supply

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	ginmw "github.com/devpablocristo/core/http/gin/go"
	"github.com/gin-gonic/gin"

	"github.com/devpablocristo/core/errors/go/domainerr"
	providerdomain "github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
	supplyExcel "github.com/devpablocristo/ponti-backend/internal/supply/excel"
	createDto "github.com/devpablocristo/ponti-backend/internal/supply/handler/dto/create"
	getDto "github.com/devpablocristo/ponti-backend/internal/supply/handler/dto/get"
	listDto "github.com/devpablocristo/ponti-backend/internal/supply/handler/dto/list"
	updateDto "github.com/devpablocristo/ponti-backend/internal/supply/handler/dto/update"
	domain "github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
)

type UseCasesPort interface {
	CreateSupply(ctx context.Context, s *domain.Supply) (int64, error)
	CreatePendingSupply(ctx context.Context, projectID int64, name string) (*domain.Supply, bool, error)
	CreateSuppliesBulk(ctx context.Context, supplies []domain.Supply) error
	GetSupply(ctx context.Context, id int64) (*domain.Supply, error)
	GetSuppliesByIDs(ctx context.Context, ids []int64) (map[int64]domain.Supply, error)
	UpdateSupply(ctx context.Context, s *domain.Supply) error
	CompletePendingSupply(ctx context.Context, s *domain.Supply) error
	DeleteSupply(ctx context.Context, id int64) error
	CountWorkOrdersBySupplyID(ctx context.Context, supplyID int64) (int64, error)
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
	ValidateSupplyMovement(context.Context, *domain.SupplyMovement) error
	CreateSupplyMovementsStrict(context.Context, []*domain.SupplyMovement) ([]int64, error)
	ImportSupplyMovements(context.Context, []*domain.SupplyMovement) ([]int64, []SupplyMovementImportFailure, error)
	GetSupplyMovementByID(context.Context, int64) (*domain.SupplyMovement, error)
	UpdateSupplyMovement(context.Context, *domain.SupplyMovement) error
	GetProviders(context.Context) ([]providerdomain.Provider, error)
	ExportSupplyMovementsByProjectID(ctx context.Context, projectID int64) ([]byte, error)
	DeleteSupplyMovement(context.Context, int64, int64) error
	ArchiveSupply(context.Context, int64) error
	RestoreSupply(context.Context, int64) error
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

	supplies := r.Group(baseURL+"/supplies", h.mws.GetValidation()...)
	{
		supplies.POST("", h.CreateSupply)
		supplies.POST("/pending", h.CreatePendingSupply)
		supplies.POST("/bulk", h.CreateSuppliesBulk)
		supplies.GET("", h.ListSupplies)
		supplies.GET("/pending", h.ListPendingSupplies)
		supplies.GET("/export/all", h.ExportTableSupplies)
		supplies.PUT("/bulk", h.UpdateSuppliesBulk)
		supplies.GET("/:supply_id", h.GetSupply)
		supplies.PUT("/pending/:supply_id/complete", h.CompletePendingSupply)
		supplies.PUT("/:supply_id", h.UpdateSupply)
		supplies.DELETE("/:supply_id", h.DeleteSupply)
		supplies.POST("/:supply_id/archive", h.ArchiveSupply)
		supplies.POST("/:supply_id/restore", h.RestoreSupply)
		supplies.GET("/:supply_id/workorders-count", h.CountWorkOrdersBySupplyID)
	}

	supplyMovements := r.Group(baseURL + "/projects/:project_id/supply-movements")
	{
		supplyMovements.POST("", h.CreateSupplyMovement)
		supplyMovements.POST("/import", h.ImportSupplyMovements)
		supplyMovements.GET("", h.GetSupplyMovementsByProjectID)
		supplyMovements.GET("/export", h.ExportSupplyMovementsByProjectID)
		supplyMovements.GET("/providers", h.GetProviders)
		supplyMovements.PUT("/:supply_movement_id", h.UpdateSupplyMovementByID)
		supplyMovements.DELETE("/:supply_movement_id", h.DeleteSupplyMovement)
	}

	// Editor de stock: mismas rutas que supply-movements, semántica para vista Stock
	stockMovements := r.Group(baseURL + "/projects/:project_id/stock-movements")
	{
		stockMovements.POST("", h.CreateSupplyMovement)
		stockMovements.GET("", h.GetSupplyMovementsByProjectID)
		stockMovements.GET("/export", h.ExportSupplyMovementsByProjectID)
		stockMovements.GET("/providers", h.GetProviders)
		stockMovements.PUT("/:stock_movement_id", h.UpdateSupplyMovementByID)
		stockMovements.DELETE("/:stock_movement_id", h.DeleteSupplyMovement)
	}
}

func (h *Handler) CreateSupply(c *gin.Context) {
	var req createDto.SupplyRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	newID, err := h.ucs.CreateSupply(c.Request.Context(), req.ToDomain())
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondCreated(c, newID)
}

func (h *Handler) CreatePendingSupply(c *gin.Context) {
	var req createDto.PendingSupplyRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}

	supply, created, err := h.ucs.CreatePendingSupply(c.Request.Context(), req.ProjectID, req.Name)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	statusCode := http.StatusOK
	if created {
		statusCode = http.StatusCreated
	}

	c.JSON(statusCode, createDto.PendingSupplyResponse{
		ID:        supply.ID,
		Name:      supply.Name,
		IsPending: supply.IsPending,
		Created:   created,
	})
}

func (h *Handler) CreateSuppliesBulk(c *gin.Context) {
	var req []createDto.SupplyRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	supplies := make([]domain.Supply, len(req))
	for i := range req {
		supplies[i] = *req[i].ToDomain()
	}
	if err := h.ucs.CreateSuppliesBulk(c.Request.Context(), supplies); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.Status(http.StatusCreated)
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
		sharedhandlers.RespondError(c, err)
		return
	}

	resp := listDto.NewListSuppliesResponse(supplies, page, perPage, total)
	sharedhandlers.RespondOK(c, resp)
}

func (h *Handler) ListPendingSupplies(c *gin.Context) {
	query := c.Request.URL.Query()
	query.Set("mode", "pending")
	c.Request.URL.RawQuery = query.Encode()
	h.ListSupplies(c)
}

func (h *Handler) GetSupply(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "supply_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	supply, err := h.ucs.GetSupply(c.Request.Context(), id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, createDto.FromDomain(supply))
}

func (h *Handler) UpdateSupply(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "supply_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req createDto.SupplyRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	dom := req.ToDomain()
	dom.ID = id
	if req.IsPartialPrice == nil {
		currentSupply, err := h.ucs.GetSupply(c.Request.Context(), id)
		if err != nil {
			sharedhandlers.RespondError(c, err)
			return
		}
		dom.IsPartialPrice = currentSupply.IsPartialPrice
	}
	if err := h.ucs.UpdateSupply(c.Request.Context(), dom); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) CompletePendingSupply(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("supply_id"), "supply_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	var req createDto.SupplyRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}

	dom := req.ToDomain()
	dom.ID = id
	if req.IsPartialPrice == nil {
		currentSupply, err := h.ucs.GetSupply(c.Request.Context(), id)
		if err != nil {
			sharedhandlers.RespondError(c, err)
			return
		}
		dom.IsPartialPrice = currentSupply.IsPartialPrice
	}

	if err := h.ucs.CompletePendingSupply(c.Request.Context(), dom); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) DeleteSupply(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "supply_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.DeleteSupply(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) ArchiveSupply(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "supply_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.ArchiveSupply(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) RestoreSupply(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "supply_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.RestoreSupply(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) CountWorkOrdersBySupplyID(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "supply_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	count, err := h.ucs.CountWorkOrdersBySupplyID(c.Request.Context(), id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, gin.H{"count": count})
}

func (h *Handler) UpdateSuppliesBulk(c *gin.Context) {
	var req []createDto.SupplyRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}

	idsToResolve := make([]int64, 0, len(req))
	seen := make(map[int64]struct{}, len(req))
	for i := range req {
		if req[i].IsPartialPrice != nil || req[i].ID == 0 {
			continue
		}
		if _, exists := seen[req[i].ID]; exists {
			continue
		}
		seen[req[i].ID] = struct{}{}
		idsToResolve = append(idsToResolve, req[i].ID)
	}

	currentSuppliesByID := map[int64]domain.Supply{}
	if len(idsToResolve) > 0 {
		resolved, err := h.ucs.GetSuppliesByIDs(c.Request.Context(), idsToResolve)
		if err != nil {
			sharedhandlers.RespondError(c, err)
			return
		}
		currentSuppliesByID = resolved
	}

	supplies := make([]domain.Supply, len(req))
	for i := range req {
		supply := req[i].ToDomain()
		if req[i].IsPartialPrice == nil && supply.ID != 0 {
			currentSupply, ok := currentSuppliesByID[supply.ID]
			if !ok {
				sharedhandlers.RespondError(c, domainerr.New(domainerr.KindNotFound, fmt.Sprintf("supply %d not found", supply.ID)))
				return
			}
			supply.IsPartialPrice = currentSupply.IsPartialPrice
		}
		supplies[i] = *supply
	}
	if err := h.ucs.UpdateSuppliesBulk(c.Request.Context(), supplies); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
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
		sharedhandlers.RespondError(c, err)
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

	userID, err := sharedhandlers.ParseActor(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}

	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode == "" {
		mode = "strict"
	}
	if mode != "partial" && mode != "strict" {
		sharedhandlers.RespondError(c, domainerr.Validation("mode must be one of [partial, strict]"))
		return
	}

	warnings := make([]string, 0)
	for _, item := range req.SupplyMovements {
		if item.MovementType == domain.INTERNAL_MOVEMENT && mode == "partial" {
			mode = "strict"
			warnings = append(warnings, "partial mode was overridden to strict for internal movements")
			break
		}
	}

	total := len(req.SupplyMovements)
	supplyMovementsResponse := make([]createDto.CreateSupplyMovementResponse, 0, total)
	successes := make([]createDto.SupplyMovementSuccess, 0, total)
	failures := make([]createDto.SupplyMovementFailure, 0)
	skipped := make([]createDto.SupplyMovementSkipped, 0)
	domainMovements := make([]*domain.SupplyMovement, 0, total)
	validIndexes := make([]int, 0, total)
	requestReturnSupplyKeys := make(map[string]int)

	for i, item := range req.SupplyMovements {
		if err := item.Validate(); err != nil {
			message := sharedhandlers.ErrorMessage(err)
			failures = append(failures, createDto.SupplyMovementFailure{
				Index:    i,
				RowIndex: i + 2,
				SupplyID: item.SupplyID,
				Code:     "validation_error",
				Message:  message,
			})
			supplyMovementsResponse = append(supplyMovementsResponse, createDto.NewErrorCreateSupplyMovementResponse(message))
			continue
		}

		if item.MovementType == domain.RETURN_MOVEMENT {
			returnSupplyKey := fmt.Sprintf("%s|%s|%d", item.MovementType, strings.TrimSpace(item.Reference), item.SupplyID)
			if _, exists := requestReturnSupplyKeys[returnSupplyKey]; exists {
				message := fmt.Sprintf(
					"El remito de devolución %s ya contiene el insumo %d dentro del request",
					strings.TrimSpace(item.Reference),
					item.SupplyID,
				)
				failures = append(failures, createDto.SupplyMovementFailure{
					Index:    i,
					RowIndex: i + 2,
					SupplyID: item.SupplyID,
					Code:     "duplicate_request",
					Message:  message,
				})
				supplyMovementsResponse = append(supplyMovementsResponse, createDto.NewErrorCreateSupplyMovementResponse(message))
				continue
			}
			requestReturnSupplyKeys[returnSupplyKey] = i
		}

		domainMovements = append(domainMovements, item.ToDomain(projectID, &userID))
		validIndexes = append(validIndexes, i)
		supplyMovementsResponse = append(supplyMovementsResponse, createDto.CreateSupplyMovementResponse{})
	}

	if mode == "strict" {
		if len(failures) == 0 && len(domainMovements) > 0 {
			prevalidationFailedIndexes := make(map[int]bool)
			for i, movement := range domainMovements {
				if err := h.ucs.ValidateSupplyMovement(ctx, movement); err != nil {
					itemIndex := validIndexes[i]
					message := sharedhandlers.ErrorMessage(err)
					prevalidationFailedIndexes[itemIndex] = true
					failures = append(failures, createDto.SupplyMovementFailure{
						Index:    itemIndex,
						RowIndex: itemIndex + 2,
						SupplyID: req.SupplyMovements[itemIndex].SupplyID,
						Code:     "validation_error",
						Message:  message,
					})
					supplyMovementsResponse[itemIndex] = createDto.NewErrorCreateSupplyMovementResponse(message)
				}
			}

			if len(prevalidationFailedIndexes) > 0 {
				for _, itemIndex := range validIndexes {
					if prevalidationFailedIndexes[itemIndex] {
						continue
					}
					skipped = append(skipped, createDto.SupplyMovementSkipped{
						Index:    itemIndex,
						SupplyID: req.SupplyMovements[itemIndex].SupplyID,
						Reason:   "No ejecutado por rollback estricto",
					})
					supplyMovementsResponse[itemIndex] = createDto.CreateSupplyMovementResponse{
						IsSaved: false,
					}
				}
				warnings = append(warnings, "Hay insumos válidos que no se ejecutaron por rollback estricto")
			} else {
				ids, err := h.ucs.CreateSupplyMovementsStrict(ctx, domainMovements)
				if err != nil {
					failedValidPos := -1
					msg := sharedhandlers.ErrorMessage(err)
					if strings.HasPrefix(msg, "item ") {
						parts := strings.SplitN(msg, ": ", 2)
						if len(parts) == 2 {
							prefix := strings.TrimPrefix(parts[0], "item ")
							if pos, convErr := strconv.Atoi(strings.TrimSpace(prefix)); convErr == nil {
								failedValidPos = pos
							}
							msg = parts[1]
						}
					}
					for _, itemIndex := range validIndexes {
						if failedValidPos >= 0 && failedValidPos < len(validIndexes) && itemIndex == validIndexes[failedValidPos] {
							failures = append(failures, createDto.SupplyMovementFailure{
								Index:    itemIndex,
								RowIndex: itemIndex + 2,
								SupplyID: req.SupplyMovements[itemIndex].SupplyID,
								Code:     "apply_error",
								Message:  msg,
							})
							supplyMovementsResponse[itemIndex] = createDto.NewErrorCreateSupplyMovementResponse(msg)
							continue
						}
						skipped = append(skipped, createDto.SupplyMovementSkipped{
							Index:    itemIndex,
							SupplyID: req.SupplyMovements[itemIndex].SupplyID,
							Reason:   "No ejecutado por rollback estricto",
						})
						supplyMovementsResponse[itemIndex] = createDto.CreateSupplyMovementResponse{
							IsSaved: false,
						}
					}
					if len(skipped) > 0 {
						warnings = append(warnings, "Hay insumos válidos que no se ejecutaron por rollback estricto")
					}
				} else {
					for i, movementID := range ids {
						itemIndex := validIndexes[i]
						successes = append(successes, createDto.SupplyMovementSuccess{
							Index:            itemIndex,
							SupplyID:         req.SupplyMovements[itemIndex].SupplyID,
							SupplyMovementID: movementID,
						})
						supplyMovementsResponse[itemIndex] = createDto.NewSuccessfulCreateSupplyMovementResponse(movementID)
					}
				}
			}
		} else if len(failures) > 0 {
			msg := "No ejecutado por rollback estricto"
			for _, itemIndex := range validIndexes {
				skipped = append(skipped, createDto.SupplyMovementSkipped{
					Index:    itemIndex,
					SupplyID: req.SupplyMovements[itemIndex].SupplyID,
					Reason:   msg,
				})
				supplyMovementsResponse[itemIndex] = createDto.CreateSupplyMovementResponse{
					IsSaved: false,
				}
			}
			warnings = append(warnings, "Hay insumos válidos que no se ejecutaron por rollback estricto")
		}
	} else {
		validPos := 0
		for i := range req.SupplyMovements {
			if supplyMovementsResponse[i].IsSaved || supplyMovementsResponse[i].ErrorDetail != "" {
				continue
			}
			supplyMovementID, err := h.ucs.CreateSupplyMovement(ctx, domainMovements[validPos])
			if err != nil {
				message := sharedhandlers.ErrorMessage(err)
				failures = append(failures, createDto.SupplyMovementFailure{
					Index:    i,
					RowIndex: i + 2,
					SupplyID: req.SupplyMovements[i].SupplyID,
					Code:     "apply_error",
					Message:  message,
				})
				supplyMovementsResponse[i] = createDto.NewErrorCreateSupplyMovementResponse(message)
			} else {
				successes = append(successes, createDto.SupplyMovementSuccess{
					Index:            i,
					SupplyID:         req.SupplyMovements[i].SupplyID,
					SupplyMovementID: supplyMovementID,
				})
				supplyMovementsResponse[i] = createDto.NewSuccessfulCreateSupplyMovementResponse(supplyMovementID)
			}
			validPos++
		}
	}

	response := createDto.CreateSupplyMovementBulkResponse{
		Success:         len(failures) == 0,
		Mode:            mode,
		Total:           total,
		Applied:         len(successes),
		Failed:          len(failures),
		Successes:       successes,
		Failures:        failures,
		Skipped:         skipped,
		Warnings:        warnings,
		SupplyMovements: supplyMovementsResponse,
	}

	// Siempre responder 200 para que frontend lea failures detallados
	// y no caiga en mensaje genérico por códigos HTTP de error.
	c.JSON(http.StatusOK, response)
}

func (h *Handler) ImportSupplyMovements(c *gin.Context) {
	ctx := c.Request.Context()
	var req createDto.CreateSupplyMovementRequestBulk

	userID, err := sharedhandlers.ParseActor(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}

	total := len(req.SupplyMovements)
	const maxImportItems = 500
	if total > maxImportItems {
		sharedhandlers.RespondError(c, domainerr.New(domainerr.KindValidation, fmt.Sprintf("el máximo de items por importación es %d, se recibieron %d", maxImportItems, total)))
		return
	}

	domainMovements := make([]*domain.SupplyMovement, 0, total)
	failures := make([]createDto.SupplyMovementFailure, 0)
	supplyMovementsResponse := make([]createDto.CreateSupplyMovementResponse, total)

	for i, item := range req.SupplyMovements {
		if err := item.Validate(); err != nil {
			message := sharedhandlers.ErrorMessage(err)
			failures = append(failures, createDto.SupplyMovementFailure{
				Index:    i,
				RowIndex: i + 2,
				SupplyID: item.SupplyID,
				Code:     "validation_error",
				Message:  message,
			})
			supplyMovementsResponse[i] = createDto.NewErrorCreateSupplyMovementResponse(message)
			continue
		}
		domainMovements = append(domainMovements, item.ToDomain(projectID, &userID))
	}

	if len(failures) == 0 {
		ids, importFailures, err := h.ucs.ImportSupplyMovements(ctx, domainMovements)
		if err != nil {
			sharedhandlers.RespondError(c, err)
			return
		}

		if len(importFailures) > 0 {
			for _, failure := range importFailures {
				failures = append(failures, createDto.SupplyMovementFailure{
					Index:           failure.Index,
					RowIndex:        failure.RowIndex,
					SupplyID:        failure.SupplyID,
					SupplyName:      failure.SupplyName,
					ReferenceNumber: failure.ReferenceNumber,
					Code:            failure.Code,
					Message:         failure.Message,
				})
				if failure.Index >= 0 && failure.Index < len(supplyMovementsResponse) {
					supplyMovementsResponse[failure.Index] = createDto.NewErrorCreateSupplyMovementResponse(failure.Message)
				}
			}
		} else {
			successes := make([]createDto.SupplyMovementSuccess, 0, len(ids))
			for i, movementID := range ids {
				successes = append(successes, createDto.SupplyMovementSuccess{
					Index:            i,
					SupplyID:         req.SupplyMovements[i].SupplyID,
					SupplyMovementID: movementID,
				})
				supplyMovementsResponse[i] = createDto.NewSuccessfulCreateSupplyMovementResponse(movementID)
			}
			c.JSON(http.StatusOK, createDto.CreateSupplyMovementBulkResponse{
				Success:         true,
				Mode:            "strict",
				Total:           total,
				Applied:         len(ids),
				Failed:          0,
				Successes:       successes,
				Failures:        nil,
				Skipped:         nil,
				Warnings:        nil,
				SupplyMovements: supplyMovementsResponse,
			})
			return
		}
	}

	for i := range supplyMovementsResponse {
		if supplyMovementsResponse[i].IsSaved || supplyMovementsResponse[i].ErrorDetail != "" {
			continue
		}
		supplyMovementsResponse[i] = createDto.CreateSupplyMovementResponse{IsSaved: false}
	}

	c.JSON(http.StatusBadRequest, createDto.CreateSupplyMovementBulkResponse{
		Success:         false,
		Mode:            "strict",
		Total:           total,
		Applied:         0,
		Failed:          len(failures),
		Successes:       nil,
		Failures:        failures,
		Skipped:         nil,
		Warnings:        []string{"No se guardó ningún movimiento porque la importación es atómica"},
		SupplyMovements: supplyMovementsResponse,
	})
}

func (h *Handler) GetSupplyMovementsByProjectID(c *gin.Context) {
	ctx := c.Request.Context()

	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	supplyMovements, err := h.ucs.GetEntriesSupplyMovementsByProjectID(ctx, projectID)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	sharedhandlers.RespondOK(c, getDto.NewGetEntrySupplyMovementsResponse(supplyMovements))
}

func (h *Handler) DeleteSupplyMovement(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	supplyMovementID, err := sharedhandlers.ParseMovementIDParam(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	err = h.ucs.DeleteSupplyMovement(ctx, id, supplyMovementID)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) UpdateSupplyMovementByID(c *gin.Context) {
	ctx := c.Request.Context()
	var req updateDto.UpdateSupplyMovementEntryRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}

	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	supplyMovementID, err := sharedhandlers.ParseMovementIDParam(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	userID, err := sharedhandlers.ParseActor(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	supplyMovement, err := h.ucs.GetSupplyMovementByID(ctx, supplyMovementID)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	if err = req.Validate(); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err = h.ucs.UpdateSupplyMovement(ctx, req.ToDomain(projectID, &userID, supplyMovement)); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) GetProviders(c *gin.Context) {
	ctx := c.Request.Context()

	providers, err := h.ucs.GetProviders(ctx)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	sharedhandlers.RespondOK(c, getDto.NewGetProvidersResponse(providers))
}

func (h *Handler) ExportSupplyMovementsByProjectID(c *gin.Context) {
	ctx := c.Request.Context()

	projectID, err := sharedhandlers.ParseProjectIDParam(c, "project_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	data, err := h.ucs.ExportSupplyMovementsByProjectID(ctx, projectID)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	filename := supplyExcel.DefaultFilename

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
