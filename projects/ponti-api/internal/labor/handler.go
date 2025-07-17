package labor

import (
	"context"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	utils "github.com/alphacodinggroup/ponti-backend/pkg/utils"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/handler/dto"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/usecases/domain"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UseCasesPort interface {
	CreateLabor(context.Context, *domain.Labor) (int64, error)
	ListLabor(context.Context, int, int) ([]domain.ListedLabor, int64, error)
	deleteLabor(context.Context, int64) error
	UpdateLabor(context.Context, *domain.Labor) error
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
	baseURL := h.acf.APIBaseURL() + "/labor"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	public := r.Group(baseURL)
	{
		public.POST("projects/:idProject/labors", h.CreateLabor) // Create an investor
		public.GET("projects/:idProject/labors", h.ListLabor)    // List all investors
	}
}

func (h *Handler) CreateLabor(c *gin.Context) {
	var req dto.LaborList
	projectIdStr := c.Param("idProject")

	projectId, err := strconv.ParseInt(projectIdStr, 10, 64)
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

		laborId, err := h.ucs.CreateLabor(ctx, labor.ToDomain(projectId))
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

	items, total, err := h.ucs.ListLabor(c.Request.Context(), page, perPage)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	resp := dto.NewListLaborsResponse(items, page, perPage, total)
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) UpdateLabor(c *gin.Context) {
	projectIdStr := c.Param("idProject")

	projectId, err := strconv.ParseInt(projectIdStr, 10, 64)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
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
	if err := h.ucs.UpdateLabor(c.Request.Context(), req.ToDomain(projectId)); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Labor updated successfully"})
}

func (h *Handler) DeleteCustomer(c *gin.Context) {
	projectIdStr := c.Param("idProject")

	_, err := strconv.ParseInt(projectIdStr, 10, 64)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid customer id"})
		return
	}
	if err := h.ucs.deleteLabor(c.Request.Context(), id); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Labor deleted successfully"})
}
