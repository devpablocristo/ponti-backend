// Package ai expone endpoints HTTP que proxyean al copilot conversacional
// de Ponti AI (`POST /v1/chat`, `POST /v1/chat/stream`, conversaciones).
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/devpablocristo/ponti-backend/internal/shared/authz"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

type UseCasesPort interface {
	Chat(ctx context.Context, userID, tenantID, projectID string, body any) (int, []byte, error)
	ListChatConversations(ctx context.Context, userID, tenantID, projectID string, limit int) (int, []byte, error)
	GetChatConversation(ctx context.Context, userID, tenantID, projectID, conversationID string) (int, []byte, error)
	ChatStream(ctx context.Context, userID, tenantID, projectID string, body io.Reader, w http.ResponseWriter) error
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

type Handler struct {
	ucs           UseCasesPort
	gsv           GinEnginePort
	acf           ConfigAPIPort
	mws           MiddlewaresEnginePort
	db            *gorm.DB
	aiTenantScope bool
}

func NewHandler(u UseCasesPort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort, db *gorm.DB, aiTenantScope bool) *Handler {
	return &Handler{ucs: u, gsv: s, acf: c, mws: m, db: db, aiTenantScope: aiTenantScope}
}

func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/ai"

	public := r.Group(baseURL, h.mws.GetValidation()...)
	{
		public.POST("/chat", h.Chat)
		public.POST("/chat/stream", h.ChatStream)
		public.GET("/chat/conversations", h.ListChatConversations)
		public.GET("/chat/conversations/:conversation_id", h.GetChatConversation)
	}
}

func (h *Handler) Chat(c *gin.Context) {
	var body map[string]any
	if err := c.ShouldBindJSON(&body); err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid request payload"))
		return
	}
	userID, tenantID, projectID, err := h.extractIDs(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	status, raw, err := h.ucs.Chat(c.Request.Context(), userID, tenantID, projectID, body)
	h.respondProxy(c, status, raw, err)
}

func (h *Handler) ChatStream(c *gin.Context) {
	if c.Request.Body == nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid request payload"))
		return
	}
	userID, tenantID, projectID, err := h.extractIDs(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	// Leer el body completo antes de proxear: pasar c.Request.Body directo a http.Client
	// puede producir deadlock (el cliente saliente espera leer el body mientras el server
	// Gin aún no entrega bytes al handler / viceversa) y el cliente ve 0 bytes hasta timeout.
	const maxChatStreamBody = 1 << 20
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxChatStreamBody+1))
	if err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid request payload"))
		return
	}
	if len(body) > maxChatStreamBody {
		sharedhandlers.RespondError(c, domainerr.Validation("request body too large"))
		return
	}
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	if err := h.ucs.ChatStream(c.Request.Context(), userID, tenantID, projectID, bytes.NewReader(body), c.Writer); err != nil && !c.Writer.Written() {
		slog.ErrorContext(c.Request.Context(), "chat stream upstream failed", "error", err.Error())
		sharedhandlers.RespondError(c, err)
	}
}

func (h *Handler) ListChatConversations(c *gin.Context) {
	userID, tenantID, projectID, err := h.extractIDs(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	limit := 50
	if q := strings.TrimSpace(c.Query("limit")); q != "" {
		if n, convErr := strconv.Atoi(q); convErr == nil && n > 0 {
			limit = n
		}
	}
	status, raw, err := h.ucs.ListChatConversations(c.Request.Context(), userID, tenantID, projectID, limit)
	h.respondProxy(c, status, raw, err)
}

func (h *Handler) GetChatConversation(c *gin.Context) {
	conversationID := strings.TrimSpace(c.Param("conversation_id"))
	if conversationID == "" {
		sharedhandlers.RespondError(c, domainerr.Validation("conversation_id is required"))
		return
	}
	userID, tenantID, projectID, err := h.extractIDs(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	status, raw, err := h.ucs.GetChatConversation(c.Request.Context(), userID, tenantID, projectID, conversationID)
	h.respondProxy(c, status, raw, err)
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

func (h *Handler) extractIDs(c *gin.Context) (string, string, string, error) {
	// El permiso explícito `ai.use` (modelo legacy de ponti-ai) se removió en
	// el cutover a Companion. Hoy basta con que el usuario haya pasado el
	// middleware de autenticación + tenant scope (api.read/api.write). Si en
	// el futuro hace falta gating fino por user, agregar de nuevo este check.
	principal, err := authz.PrincipalFromContext(c.Request.Context())
	if err != nil {
		return "", "", "", err
	}
	projectID := strings.TrimSpace(c.GetHeader("X-PROJECT-ID"))
	if projectID == "" {
		return "", "", "", domainerr.Validation("The field 'project_id' is required")
	}
	if err := h.validateProjectTenant(c.Request.Context(), principal.TenantID.String(), projectID); err != nil {
		return "", "", "", err
	}
	return principal.Actor, principal.TenantID.String(), projectID, nil
}

func (h *Handler) validateProjectTenant(ctx context.Context, tenantID string, projectID string) error {
	if !h.aiTenantScope {
		return nil
	}
	if h.db == nil {
		return domainerr.Forbidden("tenant-scoped AI is not configured")
	}
	type row struct{ One int }
	var out row
	err := h.db.WithContext(ctx).
		Table("projects").
		Select("1 AS one").
		Where("CAST(id AS TEXT) = ? AND CAST(tenant_id AS TEXT) = ? AND deleted_at IS NULL", projectID, tenantID).
		Limit(1).
		Scan(&out).Error
	if err != nil {
		return domainerr.Forbidden("project is not available for this tenant")
	}
	if out.One != 1 {
		return domainerr.Forbidden("project is not available for this tenant")
	}
	return nil
}
