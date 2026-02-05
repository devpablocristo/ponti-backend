// Package bparams expone endpoints para parametros de negocio.
package bparams

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	"github.com/alphacodinggroup/ponti-backend/internal/business-parameters/handler/dto"
	domain "github.com/alphacodinggroup/ponti-backend/internal/business-parameters/usecases/domain"
	sharedhandlers "github.com/alphacodinggroup/ponti-backend/internal/shared/handlers"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/internal/shared/models"
)

type UseCasesPort interface {
	GetParameter(context.Context, string) (*domain.BusinessParameter, error)
	GetParametersByCategory(context.Context, string) ([]domain.BusinessParameter, error)
	GetAllParameters(context.Context) ([]domain.BusinessParameter, error)
	CreateParameter(context.Context, *domain.BusinessParameter) (int64, error)
	UpdateParameter(context.Context, *domain.BusinessParameter) error
	DeleteParameter(context.Context, int64) error
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

// Routes configura las rutas del handler
func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/business-parameters"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	group := r.Group(baseURL)
	{
		group.GET("", h.GetAllParameters)
		group.GET("/category/:category", h.GetParametersByCategory)
		group.GET("/:parameter_key", h.GetParameter)
		group.POST("", h.CreateParameter)
		group.PUT("/:parameter_id", h.UpdateParameter)
		group.DELETE("/:parameter_id", h.DeleteParameter)
	}
}

// GetParameter obtiene un parámetro por su clave
// @Summary Get business parameter by key
// @Description Get a business parameter by its key
// @Tags business-parameters
// @Accept json
// @Produce json
// @Param parameter_key path string true "Parameter key"
// @Success 200 {object} dto.BusinessParameterResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /business-parameters/{key} [get]
func (h *Handler) GetParameter(c *gin.Context) {
	key := c.Param("parameter_key")
	if key == "" {
		domErr := types.NewError(types.ErrBadRequest, "parameter_key is required", nil)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	param, err := h.ucs.GetParameter(c.Request.Context(), key)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	c.JSON(http.StatusOK, dto.FromDomain(param))
}

// GetParametersByCategory obtiene parámetros por categoría
// @Summary Get business parameters by category
// @Description Get all business parameters filtered by category
// @Tags business-parameters
// @Accept json
// @Produce json
// @Param category path string true "Parameter category"
// @Success 200 {array} dto.BusinessParameterResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /business-parameters/category/{category} [get]
func (h *Handler) GetParametersByCategory(c *gin.Context) {
	category := c.Param("category")
	if category == "" {
		domErr := types.NewError(types.ErrBadRequest, "category is required", nil)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	params, err := h.ucs.GetParametersByCategory(c.Request.Context(), category)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	response := make([]dto.BusinessParameterResponse, len(params))
	for i, param := range params {
		response[i] = dto.FromDomain(&param)
	}

	c.JSON(http.StatusOK, response)
}

// GetAllParameters obtiene todos los parámetros
// @Summary Get all business parameters
// @Description Get all business parameters
// @Tags business-parameters
// @Accept json
// @Produce json
// @Success 200 {array} dto.BusinessParameterResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /business-parameters [get]
func (h *Handler) GetAllParameters(c *gin.Context) {
	params, err := h.ucs.GetAllParameters(c.Request.Context())
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	response := make([]dto.BusinessParameterResponse, len(params))
	for i, param := range params {
		response[i] = dto.FromDomain(&param)
	}

	c.JSON(http.StatusOK, response)
}

// CreateParameter crea un nuevo parámetro
// @Summary Create business parameter
// @Description Create a new business parameter
// @Tags business-parameters
// @Accept json
// @Produce json
// @Param request body dto.CreateBusinessParameterRequest true "Create business parameter request"
// @Success 201 {object} dto.BusinessParameterResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /business-parameters [post]
func (h *Handler) CreateParameter(c *gin.Context) {
	var req dto.CreateBusinessParameterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domErr := types.NewError(types.ErrBadRequest, "invalid request payload", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	param := req.ToDomain()
	userID, err := sharedmodels.ConvertStringToID(c.Request.Context())
	if err == nil {
		param.CreatedBy = &userID
		param.UpdatedBy = &userID
	}

	id, err := h.ucs.CreateParameter(c.Request.Context(), param)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	param.ID = id
	c.JSON(http.StatusCreated, dto.FromDomain(param))
}

// UpdateParameter actualiza un parámetro existente
// @Summary Update business parameter
// @Description Update an existing business parameter
// @Tags business-parameters
// @Accept json
// @Produce json
// @Param id path int true "Parameter ID"
// @Param request body dto.UpdateBusinessParameterRequest true "Update business parameter request"
// @Success 200 {object} dto.BusinessParameterResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /business-parameters/{id} [put]
func (h *Handler) UpdateParameter(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("parameter_id"), "parameter_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	var req dto.UpdateBusinessParameterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domErr := types.NewError(types.ErrBadRequest, "invalid request payload", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	param := req.ToDomain(id)
	userID, err := sharedmodels.ConvertStringToID(c.Request.Context())
	if err == nil {
		param.UpdatedBy = &userID
	}

	err = h.ucs.UpdateParameter(c.Request.Context(), param)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	c.JSON(http.StatusOK, dto.FromDomain(param))
}

// DeleteParameter elimina un parámetro
// @Summary Delete business parameter
// @Description Delete a business parameter
// @Tags business-parameters
// @Accept json
// @Produce json
// @Param id path int true "Parameter ID"
// @Success 204
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /business-parameters/{id} [delete]
func (h *Handler) DeleteParameter(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("parameter_id"), "parameter_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	err = h.ucs.DeleteParameter(c.Request.Context(), id)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	c.Status(http.StatusNoContent)
}
