package businessinsights

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

// GinEnginePort abstrae el engine HTTP para testeo.
type GinEnginePort interface {
	GetRouter() *gin.Engine
}

// ConfigAPIPort expone la base URL del API (ej: "/api/v1").
type ConfigAPIPort interface {
	APIBaseURL() string
}

// MiddlewaresEnginePort expone middlewares de validacion (auth, etc).
type MiddlewaresEnginePort interface {
	GetValidation() []gin.HandlerFunc
}

// ListRepository expone la lectura de candidatos enriquecida con read_at por usuario.
type ListRepository interface {
	ListByTenantForUser(ctx context.Context, tenantID, userID string, opts ListOptions) ([]CandidateView, error)
	GetByIDForTenant(ctx context.Context, tenantID, candidateID, userID string) (CandidateView, error)
}

// Handler expone lectura y mutaciones de candidatos (business insights).
type Handler struct {
	repo ListRepository
	svc  *Service
	eng  GinEnginePort
	cfg  ConfigAPIPort
	mws  MiddlewaresEnginePort
}

func NewHandler(repo ListRepository, svc *Service, eng GinEnginePort, cfg ConfigAPIPort, mws MiddlewaresEnginePort) *Handler {
	return &Handler{repo: repo, svc: svc, eng: eng, cfg: cfg, mws: mws}
}

// Routes registra los endpoints bajo /api/v1/insights.
func (h *Handler) Routes() {
	base := h.cfg.APIBaseURL() + "/insights"
	group := h.eng.GetRouter().Group(base, h.mws.GetValidation()...)
	group.GET("", h.List)
	group.GET("/summary", h.Summary)
	group.GET("/:id/explain", h.Explain)
	group.POST("/:id/read", h.MarkRead)
	group.DELETE("/:id/read", h.MarkUnread)
	group.POST("/:id/resolve", h.Resolve)
	group.DELETE("/:id/resolve", h.Reopen)
}

// CandidateResponse es la forma serializada de un CandidateView.
type CandidateResponse struct {
	ID              string         `json:"id"`
	Kind            string         `json:"kind"`
	EventType       string         `json:"event_type"`
	EntityType      string         `json:"entity_type"`
	EntityID        string         `json:"entity_id"`
	Severity        string         `json:"severity"`
	Status          string         `json:"status"`
	Title           string         `json:"title"`
	Body            string         `json:"body"`
	Evidence        map[string]any `json:"evidence,omitempty"`
	OccurrenceCount int            `json:"occurrence_count"`
	FirstSeenAt     string         `json:"first_seen_at"`
	LastSeenAt      string         `json:"last_seen_at"`
	LastNotifiedAt  *string        `json:"last_notified_at,omitempty"`
	ResolvedAt      *string        `json:"resolved_at,omitempty"`
	ReadAt          *string        `json:"read_at,omitempty"`
}

// List devuelve los candidatos del tenant autenticado con la marca de lectura
// del usuario actual. Acepta ?limit= y ?include_resolved=true.
func (h *Handler) List(c *gin.Context) {
	orgID, err := sharedhandlers.ParseOrgID(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	userID, _ := sharedhandlers.ParseActor(c)
	opts := ListOptions{Limit: 100}
	if raw := c.Query("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			opts.Limit = v
		}
	}
	if strings.EqualFold(c.Query("include_resolved"), "true") {
		opts.IncludeResolved = true
	}
	views, err := h.repo.ListByTenantForUser(c.Request.Context(), orgID.String(), userID, opts)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	out := make([]CandidateResponse, 0, len(views))
	for _, v := range views {
		out = append(out, candidateResponse(v, orgID.String()))
	}
	c.JSON(http.StatusOK, gin.H{
		"items":    out,
		"evidence": baseEvidence("ponti.businessinsights.list", orgID.String()),
	})
}

// Summary devuelve agregados read-only de insights para consumo de Axis.
func (h *Handler) Summary(c *gin.Context) {
	orgID, err := sharedhandlers.ParseOrgID(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	userID, _ := sharedhandlers.ParseActor(c)
	views, err := h.repo.ListByTenantForUser(c.Request.Context(), orgID.String(), userID, ListOptions{IncludeResolved: true, Limit: 200})
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	byStatus := map[string]int{}
	bySeverity := map[string]int{}
	byKind := map[string]int{}
	byEventType := map[string]int{}
	open := 0
	resolved := 0
	for _, v := range views {
		byStatus[v.Status]++
		bySeverity[v.Severity]++
		byKind[v.Kind]++
		byEventType[v.EventType]++
		if strings.EqualFold(v.Status, "resolved") {
			resolved++
		} else {
			open++
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"summary": gin.H{
			"total":         len(views),
			"open":          open,
			"resolved":      resolved,
			"by_status":     byStatus,
			"by_severity":   bySeverity,
			"by_kind":       byKind,
			"by_event_type": byEventType,
		},
		"evidence": baseEvidence("ponti.businessinsights.summary", orgID.String()),
	})
}

// Explain devuelve un insight puntual con evidencia y provenance.
func (h *Handler) Explain(c *gin.Context) {
	orgID, candidateID, userID, err := h.requireIDs(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	view, err := h.repo.GetByIDForTenant(c.Request.Context(), orgID.String(), candidateID, userID)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	evidence := baseEvidence("ponti.businessinsights.candidate:"+view.ID, orgID.String())
	evidence["first_seen"] = view.FirstSeenAt.Format(time.RFC3339)
	evidence["last_seen"] = view.LastSeenAt.Format(time.RFC3339)
	evidence["event_type"] = view.EventType
	evidence["entity"] = gin.H{
		"type": view.EntityType,
		"id":   view.EntityID,
	}
	c.JSON(http.StatusOK, gin.H{
		"insight":  candidateResponse(view, orgID.String()),
		"evidence": evidence,
	})
}

// MarkRead marca el candidato como leido para el usuario actual.
func (h *Handler) MarkRead(c *gin.Context) {
	orgID, candidateID, userID, err := h.requireIDs(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.svc.MarkRead(c.Request.Context(), orgID, candidateID, userID); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// MarkUnread desmarca la lectura del candidato para el usuario actual.
func (h *Handler) MarkUnread(c *gin.Context) {
	orgID, candidateID, userID, err := h.requireIDs(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.svc.MarkUnread(c.Request.Context(), orgID, candidateID, userID); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Resolve marca el candidato como resuelto (afecta a todos los usuarios del tenant).
func (h *Handler) Resolve(c *gin.Context) {
	orgID, candidateID, userID, err := h.requireIDs(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.svc.ResolveManual(c.Request.Context(), orgID, candidateID, userID); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Reopen reactiva un candidato resuelto y limpia las marcas de lectura.
func (h *Handler) Reopen(c *gin.Context) {
	orgID, candidateID, userID, err := h.requireIDs(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.svc.Reopen(c.Request.Context(), orgID, candidateID, userID); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) requireIDs(c *gin.Context) (uuid.UUID, string, string, error) {
	orgID, err := sharedhandlers.ParseOrgID(c)
	if err != nil {
		return uuid.Nil, "", "", err
	}
	userID, err := sharedhandlers.ParseActor(c)
	if err != nil {
		return uuid.Nil, "", "", err
	}
	candidateID := strings.TrimSpace(c.Param("id"))
	if candidateID == "" {
		return uuid.Nil, "", "", domainerr.Validation("candidate id is required")
	}
	return orgID, candidateID, userID, nil
}

func candidateResponse(v CandidateView, tenantID string) CandidateResponse {
	item := CandidateResponse{
		ID:              v.ID,
		Kind:            v.Kind,
		EventType:       v.EventType,
		EntityType:      v.EntityType,
		EntityID:        v.EntityID,
		Severity:        v.Severity,
		Status:          v.Status,
		Title:           v.Title,
		Body:            v.Body,
		Evidence:        mergeEvidence(v.Evidence, baseEvidence("ponti.businessinsights.candidate:"+v.ID, tenantID)),
		OccurrenceCount: v.OccurrenceCount,
		FirstSeenAt:     v.FirstSeenAt.Format(time.RFC3339),
		LastSeenAt:      v.LastSeenAt.Format(time.RFC3339),
	}
	if v.LastNotifiedAt != nil {
		ts := v.LastNotifiedAt.Format(time.RFC3339)
		item.LastNotifiedAt = &ts
	}
	if v.ResolvedAt != nil {
		ts := v.ResolvedAt.Format(time.RFC3339)
		item.ResolvedAt = &ts
	}
	if v.ReadAt != nil {
		ts := v.ReadAt.Format(time.RFC3339)
		item.ReadAt = &ts
	}
	return item
}

func baseEvidence(sourceRef, tenantID string) map[string]any {
	return map[string]any{
		"source_ref":   sourceRef,
		"captured_at":  time.Now().UTC().Format(time.RFC3339),
		"tenant_scope": tenantID,
		"workspace":    map[string]any{},
	}
}

func mergeEvidence(existing map[string]any, defaults map[string]any) map[string]any {
	out := make(map[string]any, len(defaults)+len(existing))
	for k, v := range defaults {
		out[k] = v
	}
	for k, v := range existing {
		out[k] = v
	}
	return out
}
