// Package ai expone endpoints HTTP que proxyean al copilot conversacional
// de Ponti AI (`POST /v1/chat`, `POST /v1/chat/stream`, conversaciones).
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

type UseCasesPort interface {
	Chat(ctx context.Context, userID, projectID string, body any) (int, []byte, error)
	ListChatConversations(ctx context.Context, userID, projectID string, limit int) (int, []byte, error)
	GetChatConversation(ctx context.Context, userID, projectID, conversationID string) (int, []byte, error)
	ChatStream(ctx context.Context, userID, projectID string, body io.Reader, w http.ResponseWriter) error
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

// GovernedActionsConfig agrupa los flags de enforcement de las acciones
// gobernadas (subset de config.Nexus).
type GovernedActionsConfig struct {
	// VerifyNexus refleja GOVERNANCE_VERIFY_NEXUS: con true las acciones
	// gobernadas llamadas por Axis exigen un X-Nexus-Request-ID verificado.
	VerifyNexus bool
}

type Handler struct {
	ucs         UseCasesPort
	decisions   *DecisionService
	actions     *ActionExecutor
	verifier    ActionVerifierPort
	governedCfg GovernedActionsConfig
	gsv         GinEnginePort
	acf         ConfigAPIPort
	mws         MiddlewaresEnginePort
}

func NewHandler(u UseCasesPort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *Handler {
	return &Handler{ucs: u, gsv: s, acf: c, mws: m}
}

// SetGovernedActions conecta el executor de writes gobernados y el verifier de
// Nexus después de wire (mismo patrón que SetDecisionService: las dependencias
// se arman en bootstrap).
func (h *Handler) SetGovernedActions(actions *ActionExecutor, verifier ActionVerifierPort, cfg GovernedActionsConfig) {
	h.actions = actions
	h.verifier = verifier
	h.governedCfg = cfg
}

func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/ai"

	capabilities := r.Group(h.acf.APIBaseURL(), h.mws.GetValidation()...)
	{
		capabilities.GET("/capabilities", h.Capabilities)
	}

	public := r.Group(baseURL, h.mws.GetValidation()...)
	{
		public.POST("/chat", h.Chat)
		public.POST("/chat/stream", h.ChatStream)
		public.GET("/chat/conversations", h.ListChatConversations)
		public.GET("/chat/conversations/:conversation_id", h.GetChatConversation)
		public.POST("/decision-runs", h.CreateDecisionRun)
		public.GET("/decision-runs", h.ListDecisionRuns)
		public.GET("/decision-cards", h.ListDecisionCards)
		public.POST("/decision-cards/external", h.ImportExternalDecisionCard)
		public.PATCH("/decision-cards/:id", h.PatchDecisionCard)
		public.POST("/decision-cards/:id/actions/:action_id", h.ExecuteDecisionCardAction)
		public.POST("/actions/insight-resolve/prepare", h.PrepareInsightResolve)
		public.POST("/actions/workorder-draft/prepare", h.PrepareWorkOrderDraft)
		public.POST("/actions/stock-adjustment/prepare", h.PrepareStockAdjustment)
		public.POST("/actions/insight-resolution/draft", h.DraftInsightResolution)
		public.POST("/actions/stock-count/draft", h.DraftStockCount)
	}
}

func (h *Handler) Capabilities(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": pontiCapabilities()})
}

func (h *Handler) Chat(c *gin.Context) {
	var body map[string]any
	if err := c.ShouldBindJSON(&body); err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid request payload"))
		return
	}
	userID, projectID, err := extractIDs(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	status, raw, err := h.ucs.Chat(c.Request.Context(), userID, projectID, body)
	h.respondProxy(c, status, raw, err)
}

func (h *Handler) ChatStream(c *gin.Context) {
	if c.Request.Body == nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid request payload"))
		return
	}
	userID, projectID, err := extractIDs(c)
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
	if err := h.ucs.ChatStream(c.Request.Context(), userID, projectID, bytes.NewReader(body), c.Writer); err != nil && !c.Writer.Written() {
		sharedhandlers.RespondError(c, domainerr.Internal("ai service unavailable"))
	}
}

func (h *Handler) ListChatConversations(c *gin.Context) {
	userID, projectID, err := extractIDs(c)
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
	status, raw, err := h.ucs.ListChatConversations(c.Request.Context(), userID, projectID, limit)
	h.respondProxy(c, status, raw, err)
}

func (h *Handler) GetChatConversation(c *gin.Context) {
	conversationID := strings.TrimSpace(c.Param("conversation_id"))
	if conversationID == "" {
		sharedhandlers.RespondError(c, domainerr.Validation("conversation_id is required"))
		return
	}
	userID, projectID, err := extractIDs(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	status, raw, err := h.ucs.GetChatConversation(c.Request.Context(), userID, projectID, conversationID)
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
