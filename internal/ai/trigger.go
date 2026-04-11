package ai

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/devpablocristo/core/security/go/contextkeys"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// InsightTrigger dispara cómputo de insights de forma asíncrona y throttleada.
type InsightTrigger struct {
	client   *Client
	syncer   InsightSummarySyncer
	throttle sync.Map
	cooldown time.Duration
	sem      chan struct{}
}

type InsightSummarySyncer interface {
	SyncFromSummary(ctx context.Context, orgID uuid.UUID, projectID int64, actor string, raw []byte) error
}

// NewInsightTrigger crea un trigger con cooldown configurable.
func NewInsightTrigger(client *Client, syncer InsightSummarySyncer, cooldownSec int) *InsightTrigger {
	if cooldownSec <= 0 {
		cooldownSec = 300
	}
	return &InsightTrigger{
		client:   client,
		syncer:   syncer,
		cooldown: time.Duration(cooldownSec) * time.Second,
		sem:      make(chan struct{}, 32),
	}
}

// NotifyDataChange dispara un cómputo de insights async con throttle por proyecto.
func (t *InsightTrigger) NotifyDataChange(projectID, userID string, orgID uuid.UUID) {
	if t.client == nil || t.client.caller == nil || t.client.caller.BaseURL == "" {
		return
	}

	now := time.Now()
	if last, ok := t.throttle.Load(projectID); ok {
		if now.Sub(last.(time.Time)) < t.cooldown {
			return
		}
	}
	t.throttle.Store(projectID, now)

	select {
	case t.sem <- struct{}{}:
	default:
		log.Printf("[ai-trigger] semáforo lleno, descartando trigger para proyecto %s", projectID)
		return
	}
	go func() {
		defer func() { <-t.sem }()
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		st, _, err := t.client.Do(ctx, http.MethodPost, "/v1/insights/compute", nil, userID, projectID)
		if err != nil {
			log.Printf("[ai-trigger] error computando insights para proyecto %s: %v", projectID, err)
			return
		}
		if st >= 300 {
			log.Printf("[ai-trigger] insights compute retornó %d para proyecto %s", st, projectID)
		} else {
			log.Printf("[ai-trigger] insights computados para proyecto %s (status %d)", projectID, st)
			t.syncSummary(ctx, orgID, projectID, userID)
		}
	}()
}

func (t *InsightTrigger) syncSummary(ctx context.Context, orgID uuid.UUID, projectID, userID string) {
	if t.syncer == nil || orgID == uuid.Nil {
		return
	}
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return
	}
	projectNum, err := strconv.ParseInt(projectID, 10, 64)
	if err != nil {
		// Project IDs in Ponti are numeric strings; fall back to silent skip if parsing fails.
		return
	}
	_, raw, err := t.client.Do(ctx, http.MethodGet, "/v1/insights/summary", nil, userID, projectID)
	if err != nil {
		log.Printf("[ai-trigger] error obteniendo summary para proyecto %s: %v", projectID, err)
		return
	}
	if err := t.syncer.SyncFromSummary(ctx, orgID, projectNum, userID, raw); err != nil {
		log.Printf("[ai-trigger] error sincronizando notificaciones del proyecto %s: %v", projectID, err)
	}
}

// InsightTriggerMiddleware crea un middleware Gin que dispara cómputo de insights
// después de mutaciones exitosas (POST/PUT/PATCH/DELETE con respuesta 2xx).
func InsightTriggerMiddleware(trigger *InsightTrigger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		method := c.Request.Method
		if method != http.MethodPost && method != http.MethodPut &&
			method != http.MethodPatch && method != http.MethodDelete {
			return
		}

		status := c.Writer.Status()
		if status < 200 || status >= 300 {
			return
		}

		if strings.Contains(c.Request.URL.Path, "/ai/") {
			return
		}

		projectID := c.Param("project_id")
		if projectID == "" {
			return
		}

		userID := c.GetHeader("X-USER-ID")
		if userID == "" {
			userID = "system"
		}

		orgID, _ := c.Request.Context().Value(ctxkeys.OrgID).(uuid.UUID)
		trigger.NotifyDataChange(projectID, userID, orgID)
	}
}
