package businessinsights

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

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

// ListRepository es el subset del repository que necesita el handler.
type ListRepository interface {
	ListByTenant(ctx context.Context, tenantID string, limit int) ([]CandidateRecord, error)
}

// Handler expone lectura de candidatos (business insights) del tenant.
type Handler struct {
	repo ListRepository
	eng  GinEnginePort
	cfg  ConfigAPIPort
	mws  MiddlewaresEnginePort
}

func NewHandler(repo ListRepository, eng GinEnginePort, cfg ConfigAPIPort, mws MiddlewaresEnginePort) *Handler {
	return &Handler{repo: repo, eng: eng, cfg: cfg, mws: mws}
}

// Routes registra GET /api/v1/insights.
func (h *Handler) Routes() {
	base := h.cfg.APIBaseURL() + "/insights"
	group := h.eng.GetRouter().Group(base, h.mws.GetValidation()...)
	group.GET("", h.List)
}

// CandidateResponse es la forma serializada de un CandidateRecord.
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
}

// List devuelve los candidatos del tenant autenticado (limit opcional por ?limit=).
func (h *Handler) List(c *gin.Context) {
	orgID, err := sharedhandlers.ParseOrgID(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	limit := 100
	if raw := c.Query("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}
	records, err := h.repo.ListByTenant(c.Request.Context(), orgID.String(), limit)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	out := make([]CandidateResponse, 0, len(records))
	for _, r := range records {
		item := CandidateResponse{
			ID:              r.ID,
			Kind:            r.Kind,
			EventType:       r.EventType,
			EntityType:      r.EntityType,
			EntityID:        r.EntityID,
			Severity:        r.Severity,
			Status:          r.Status,
			Title:           r.Title,
			Body:            r.Body,
			Evidence:        r.Evidence,
			OccurrenceCount: r.OccurrenceCount,
			FirstSeenAt:     r.FirstSeenAt.Format("2006-01-02T15:04:05Z07:00"),
			LastSeenAt:      r.LastSeenAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if r.LastNotifiedAt != nil {
			ts := r.LastNotifiedAt.Format("2006-01-02T15:04:05Z07:00")
			item.LastNotifiedAt = &ts
		}
		out = append(out, item)
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}
