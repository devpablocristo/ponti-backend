// Package ai expone endpoints HTTP que proxyean al AI Copilot Service.
package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/devpablocristo/core/saas/go/shared/domainerr"
	dto "github.com/devpablocristo/ponti-backend/internal/ai/handler/dto"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

type UseCasesPort interface {
	ComputeInsights(ctx context.Context, userID, projectID string) (int, []byte, error)
	GetInsights(ctx context.Context, userID, projectID, entityType, entityID string) (int, []byte, error)
	GetSummary(ctx context.Context, userID, projectID string) (int, []byte, error)
	RecordAction(ctx context.Context, userID, projectID, insightID string, body any) (int, []byte, error)
}

type GinEnginePort interface {
	GetRouter() *gin.Engine
}

type ConfigAPIPort interface {
	APIVersion() string
	APIBaseURL() string
}

type MiddlewaresEnginePort interface {
	GetValidation() []gin.HandlerFunc
}

// Handler encapsula dependencias del handler HTTP de AI.
type Handler struct {
	ucs UseCasesPort
	gsv GinEnginePort
	acf ConfigAPIPort
	mws MiddlewaresEnginePort
}

// NewHandler crea un handler de AI.
func NewHandler(u UseCasesPort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		acf: c,
		mws: m,
	}
}

// Routes registra las rutas del módulo AI.
func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/ai"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	public := r.Group(baseURL)
	{
		public.POST("/insights/compute", h.ComputeInsights)
		public.GET("/insights/summary", h.GetSummary)
		public.GET("/insights/:entity_type/:entity_id", h.GetInsights)
		public.POST("/insights/:insight_id/actions", h.RecordAction)
	}
}

func (h *Handler) ComputeInsights(c *gin.Context) {
	userID, projectID, err := extractIDs(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	status, body, err := h.ucs.ComputeInsights(c.Request.Context(), userID, projectID)
	h.respondProxy(c, status, body, err)
}

func (h *Handler) GetSummary(c *gin.Context) {
	userID, projectID, err := extractIDs(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	status, body, err := h.ucs.GetSummary(c.Request.Context(), userID, projectID)
	h.respondProxy(c, status, body, err)
}

func (h *Handler) GetInsights(c *gin.Context) {
	entityType := c.Param("entity_type")
	entityID := c.Param("entity_id")
	userID, projectID, err := extractIDs(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	status, body, err := h.ucs.GetInsights(c.Request.Context(), userID, projectID, entityType, entityID)
	h.respondProxy(c, status, body, err)
}

func (h *Handler) RecordAction(c *gin.Context) {
	var req dto.ActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid request payload"))
		return
	}
	insightID := c.Param("insight_id")
	userID, projectID, err := extractIDs(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	status, body, err := h.ucs.RecordAction(c.Request.Context(), userID, projectID, insightID, req)
	h.respondProxy(c, status, body, err)
}

func (h *Handler) respondProxy(c *gin.Context, status int, body []byte, err error) {
	if err != nil {
		sharedhandlers.RespondError(c, domainerr.Internal("ai service unavailable"))
		return
	}
	if status >= http.StatusBadRequest {
		c.Data(status, "application/json", body)
		return
	}
	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		c.Data(status, "application/json", body)
		return
	}
	c.JSON(status, payload)
}

func extractIDs(c *gin.Context) (string, string, error) {
	userID := strings.TrimSpace(c.GetHeader("X-USER-ID"))
	projectID := strings.TrimSpace(c.GetHeader("X-PROJECT-ID"))
	if userID == "" {
		return "", "", domainerr.Validation("The field 'user_id' is required")
	}
	if projectID == "" {
		return "", "", domainerr.Validation("The field 'project_id' is required")
	}
	return userID, projectID, nil
}
