package labor

import (
	"context"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	utils "github.com/alphacodinggroup/ponti-backend/pkg/utils"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/handler/dto"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/usecases/domain"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UseCasesPort interface {
	CreateLabor(context.Context, *domain.Labor) (int64, error)
	ListLabor(context.Context, int, int, int64) ([]domain.ListedLabor, int64, error)
	DeleteLabor(context.Context, int64) error
	UpdateLabor(context.Context, *domain.Labor) error
	ListLaborCategoriesByTypeId(context.Context, int64) ([]domain.LaborCategory, error)
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

// Handler encapsulates all dependencies for the LeaseType HTTP handler.
type Handler struct {
	ucs  UseCasesPort
	gsv  GinEnginePort
	acf  ConfigAPIPort
	mws  MiddlewaresEnginePort
	ucps project.UseCasesPort
}

func NewHandler(u UseCasesPort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort, up project.UseCasesPort) *Handler {
	return &Handler{
		ucs:  u,
		gsv:  s,
		acf:  c,
		mws:  m,
		ucps: up,
	}
}

func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/projects/:id/labors"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	public := r.Group(baseURL)
	{
		public.POST("", h.CreateLabor)
		public.GET("", h.ListLabor)
		public.DELETE("/:idLabor", h.DeleteLabor)
		public.PUT("/:idLabor", h.UpdateLabor)
		public.GET("/labor-categories/:typeId", h.ListLaborCategories)
	}
}

func (h *Handler) CreateLabor(c *gin.Context) {
	var req dto.LaborList
	projectIdStr := c.Param("id")

	projectId, err := strconv.ParseInt(projectIdStr, 10, 64)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	userID, err := sharedmodels.ConvertStringToID(c)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	_, err = h.ucps.GetProject(ctx, projectId)

	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	if err = utils.ValidateRequest(c, &req); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	var labors []dto.CreateLabor

	for _, labor := range req.Labors {
		var laborResponse dto.CreateLabor

		laborId, err := h.ucs.CreateLabor(ctx, labor.ToDomain(projectId, userID))
		if err != nil {
			laborResponse = dto.CreateLabor{
				LaborID:     0,
				LaborName:   labor.Name,
				IsSaved:     false,
				ErrorDetail: err.Error(),
			}
		} else {
			laborResponse = dto.CreateLabor{
				LaborID:   laborId,
				LaborName: labor.Name,
				IsSaved:   true,
			}
		}

		labors = append(labors, laborResponse)
	}

	c.JSON(http.StatusMultiStatus, dto.CreateLaborsResponse{
		Message: "labors created",
		Labors:  labors,
	})
}

func (h *Handler) ListLabor(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "100"))

	projectIdStr := c.Param("id")

	projectId, err := strconv.ParseInt(projectIdStr, 10, 64)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	items, total, err := h.ucs.ListLabor(c.Request.Context(), page, perPage, projectId)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	resp := dto.NewListLaborsResponse(items, page, perPage, total)
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) UpdateLabor(c *gin.Context) {
	projectIdStr := c.Param("id")

	projectId, err := strconv.ParseInt(projectIdStr, 10, 64)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	userID, err := sharedmodels.ConvertStringToID(c)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	id, err := strconv.ParseInt(c.Param("idLabor"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid labor id"})
		return
	}
	var req dto.Labor
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid payload"})
		return
	}
	req.ID = id
	if err := h.ucs.UpdateLabor(c.Request.Context(), req.ToDomain(projectId, userID)); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Labor updated successfully"})
}

func (h *Handler) DeleteLabor(c *gin.Context) {
	projectIdStr := c.Param("id")

	_, err := strconv.ParseInt(projectIdStr, 10, 64)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	id, err := strconv.ParseInt(c.Param("idLabor"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid labor id"})
		return
	}
	if err := h.ucs.DeleteLabor(c.Request.Context(), id); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Labor deleted successfully"})
}

func (h *Handler) ListLaborCategories(c *gin.Context) {
	projectIdStr := c.Param("id")

	_, err := strconv.ParseInt(projectIdStr, 10, 64)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	id, err := strconv.ParseInt(c.Param("typeId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid labor type id"})
		return
	}

	laborCategories, err := h.ucs.ListLaborCategoriesByTypeId(c, id)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	resp := dto.NewLaborCategoriesListResponse(laborCategories)
	c.JSON(http.StatusOK, resp)

}
