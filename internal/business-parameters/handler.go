package business_parameters

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	"github.com/alphacodinggroup/ponti-backend/internal/business-parameters/handler/dto"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/internal/shared/models"
)

type Handler struct {
	useCases *UseCases
}

// Interfaces para wire
type GinEnginePort interface{}
type ConfigAPIPort interface{}
type MiddlewaresEnginePort interface{}
type UseCasesPort interface{}

func NewHandler(useCases UseCasesPort, server GinEnginePort, cfg ConfigAPIPort, mws MiddlewaresEnginePort) *Handler {
	return &Handler{
		useCases: useCases.(*UseCases),
	}
}

// Routes configura las rutas del handler
func (h *Handler) Routes() {
	// Por ahora no implementamos rutas específicas
	// Las rutas se pueden agregar aquí cuando sea necesario
}

// GetParameter obtiene un parámetro por su clave
// @Summary Get business parameter by key
// @Description Get a business parameter by its key
// @Tags business-parameters
// @Accept json
// @Produce json
// @Param key path string true "Parameter key"
// @Success 200 {object} dto.BusinessParameterResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /business-parameters/{key} [get]
func (h *Handler) GetParameter(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "key parameter is required",
		})
		return
	}

	param, err := h.useCases.GetParameter(c.Request.Context(), key)
	if err != nil {
		if err.Error() == "business parameter not found" {
			c.JSON(http.StatusNotFound, types.ErrorResponse{
				Error: "business parameter not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "failed to get business parameter",
		})
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
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "category parameter is required",
		})
		return
	}

	params, err := h.useCases.GetParametersByCategory(c.Request.Context(), category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "failed to get business parameters by category",
		})
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
	params, err := h.useCases.GetAllParameters(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "failed to get all business parameters",
		})
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
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "invalid request body",
		})
		return
	}

	param := req.ToDomain()
	userID, err := sharedmodels.ConvertStringToID(c.Request.Context())
	if err == nil {
		param.CreatedBy = &userID
		param.UpdatedBy = &userID
	}

	id, err := h.useCases.CreateParameter(c.Request.Context(), param)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "failed to create business parameter",
		})
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
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "invalid id parameter",
		})
		return
	}

	var req dto.UpdateBusinessParameterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "invalid request body",
		})
		return
	}

	param := req.ToDomain(id)
	userID, err := sharedmodels.ConvertStringToID(c.Request.Context())
	if err == nil {
		param.UpdatedBy = &userID
	}

	err = h.useCases.UpdateParameter(c.Request.Context(), param)
	if err != nil {
		if err.Error() == "business parameter not found" {
			c.JSON(http.StatusNotFound, types.ErrorResponse{
				Error: "business parameter not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "failed to update business parameter",
		})
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
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "invalid id parameter",
		})
		return
	}

	err = h.useCases.DeleteParameter(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "business parameter not found" {
			c.JSON(http.StatusNotFound, types.ErrorResponse{
				Error: "business parameter not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "failed to delete business parameter",
		})
		return
	}

	c.Status(http.StatusNoContent)
}
