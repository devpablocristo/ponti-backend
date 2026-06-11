package governance

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/gin-gonic/gin"

	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

// maxCallbackBody limita el payload del webhook de Nexus.
const maxCallbackBody = 1 << 20

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

// Handler expone el callback de Nexus y el inbox de approvals.
type Handler struct {
	svc *Service
	gsv GinEnginePort
	acf ConfigAPIPort
	mws MiddlewaresEnginePort
}

func NewHandler(svc *Service, gsv GinEnginePort, acf ConfigAPIPort, mws MiddlewaresEnginePort) *Handler {
	return &Handler{svc: svc, gsv: gsv, acf: acf, mws: mws}
}

// Routes registra las rutas del módulo governance.
func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	base := h.acf.APIBaseURL()

	// Callback de Nexus: sin auth de identidad (Nexus no es un usuario Ponti);
	// la integridad se verifica por HMAC contra NEXUS_CALLBACK_TOKEN.
	r.POST(base+"/governance/callbacks/nexus", h.NexusCallback)

	approvals := r.Group(base+"/ai/approvals", h.mws.GetValidation()...)
	{
		approvals.GET("", h.ListApprovals)
		approvals.GET("/summary", h.Summary)
		approvals.GET("/:request_id", h.GetApproval)
		approvals.GET("/:request_id/evidence", h.Evidence)
		approvals.POST("/:request_id/approve", h.Approve)
		approvals.POST("/:request_id/reject", h.Reject)
	}
}

// NexusCallback (POST /governance/callbacks/nexus) procesa approval_pending /
// approval_resolved verificando X-Nexus-Callback-Signature.
func (h *Handler) NexusCallback(c *gin.Context) {
	if !h.svc.CallbackConfigured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "NEXUS_CALLBACK_TOKEN not configured"})
		return
	}
	payload, err := io.ReadAll(io.LimitReader(c.Request.Body, maxCallbackBody))
	if err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid request payload"))
		return
	}
	signature := c.GetHeader("X-Nexus-Callback-Signature")
	timestamp := c.GetHeader("X-Nexus-Callback-Timestamp")
	if !h.svc.VerifyCallbackSignature(timestamp, payload, signature) {
		sharedhandlers.RespondError(c, domainerr.Unauthorized("invalid callback signature"))
		return
	}
	if !h.svc.VerifyCallbackTimestamp(timestamp) {
		// Anti-replay: una firma válida con timestamp viejo es un callback
		// capturado; Nexus re-firma cada retry con timestamp fresco.
		sharedhandlers.RespondError(c, domainerr.Unauthorized("stale or invalid callback timestamp"))
		return
	}
	var event CallbackEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid callback payload"))
		return
	}
	if err := h.svc.HandleCallback(c.Request.Context(), event); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ListApprovals (GET /ai/approvals?status=pending|history&limit=50).
func (h *Handler) ListApprovals(c *gin.Context) {
	orgID, err := sharedhandlers.ParseOrgID(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	items, err := h.svc.ListApprovals(c.Request.Context(), orgID, c.Query("status"), parseLimitQuery(c, 50))
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// Summary (GET /ai/approvals/summary) — badge del inbox.
func (h *Handler) Summary(c *gin.Context) {
	orgID, err := sharedhandlers.ParseOrgID(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	summary, err := h.svc.Summary(c.Request.Context(), orgID)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, summary)
}

// GetApproval (GET /ai/approvals/:request_id).
func (h *Handler) GetApproval(c *gin.Context) {
	orgID, err := sharedhandlers.ParseOrgID(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	item, err := h.svc.GetApproval(c.Request.Context(), orgID, c.Param("request_id"))
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

// Evidence (GET /ai/approvals/:request_id/evidence). verified queda false
// hasta implementar la verificación ED25519 en Ola B.
func (h *Handler) Evidence(c *gin.Context) {
	orgID, err := sharedhandlers.ParseOrgID(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	pack, err := h.svc.Evidence(c.Request.Context(), orgID, c.Param("request_id"))
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"pack": pack, "verified": false})
}

type decisionRequest struct {
	Note string `json:"note,omitempty"`
}

// Approve (POST /ai/approvals/:request_id/approve).
func (h *Handler) Approve(c *gin.Context) {
	h.decide(c, "approve", "approved")
}

// Reject (POST /ai/approvals/:request_id/reject).
func (h *Handler) Reject(c *gin.Context) {
	h.decide(c, "reject", "rejected")
}

func (h *Handler) decide(c *gin.Context, action, resultStatus string) {
	orgID, err := sharedhandlers.ParseOrgID(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	actor, err := sharedhandlers.ParseActor(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req decisionRequest
	if c.Request.Body != nil && c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			sharedhandlers.RespondError(c, domainerr.Validation("invalid request payload"))
			return
		}
	}
	if err := h.svc.Decide(c.Request.Context(), orgID, c.Param("request_id"), actor, action, req.Note); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": resultStatus})
}

func parseLimitQuery(c *gin.Context, fallback int) int {
	raw := strings.TrimSpace(c.Query("limit"))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
