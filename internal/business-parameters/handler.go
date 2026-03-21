// Package bparams expone endpoints para parametros de negocio.
package bparams

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/devpablocristo/core/backend/go/domainerr"
	"github.com/devpablocristo/ponti-backend/internal/business-parameters/handler/dto"
	domain "github.com/devpablocristo/ponti-backend/internal/business-parameters/usecases/domain"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
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

func (h *Handler) GetParameter(c *gin.Context) {
	key := c.Param("parameter_key")
	if key == "" {
		sharedhandlers.RespondError(c, domainerr.Validation("parameter_key is required"))
		return
	}
	param, err := h.ucs.GetParameter(c.Request.Context(), key)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.FromDomain(param))
}

func (h *Handler) GetParametersByCategory(c *gin.Context) {
	category := c.Param("category")
	if category == "" {
		sharedhandlers.RespondError(c, domainerr.Validation("category is required"))
		return
	}
	params, err := h.ucs.GetParametersByCategory(c.Request.Context(), category)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	response := make([]dto.BusinessParameterResponse, len(params))
	for i, param := range params {
		response[i] = dto.FromDomain(&param)
	}
	sharedhandlers.RespondOK(c, response)
}

func (h *Handler) GetAllParameters(c *gin.Context) {
	params, err := h.ucs.GetAllParameters(c.Request.Context())
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	response := make([]dto.BusinessParameterResponse, len(params))
	for i, param := range params {
		response[i] = dto.FromDomain(&param)
	}
	sharedhandlers.RespondOK(c, response)
}

func (h *Handler) CreateParameter(c *gin.Context) {
	var req dto.CreateBusinessParameterRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	param := req.ToDomain()
	userID, err := sharedmodels.ActorFromContext(c.Request.Context())
	if err == nil {
		param.CreatedBy = &userID
		param.UpdatedBy = &userID
	}
	id, err := h.ucs.CreateParameter(c.Request.Context(), param)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondCreated(c, id)
}

func (h *Handler) UpdateParameter(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("parameter_id"), "parameter_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req dto.UpdateBusinessParameterRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	param := req.ToDomain(id)
	userID, err := sharedmodels.ActorFromContext(c.Request.Context())
	if err == nil {
		param.UpdatedBy = &userID
	}
	if err := h.ucs.UpdateParameter(c.Request.Context(), param); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) DeleteParameter(c *gin.Context) {
	id, err := sharedhandlers.ParseParamID(c.Param("parameter_id"), "parameter_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.DeleteParameter(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}
