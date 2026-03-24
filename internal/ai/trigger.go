package ai

import (
	"context"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/devpablocristo/core/backend/go/httpclient"
	"github.com/gin-gonic/gin"
)

// InsightTrigger dispara cómputo de insights de forma asíncrona y throttleada.
type InsightTrigger struct {
	caller   *httpclient.Caller
	throttle sync.Map
	cooldown time.Duration
	sem      chan struct{}
}

// NewInsightTrigger crea un trigger con cooldown configurable.
func NewInsightTrigger(baseURL, serviceKey string, timeoutMS, cooldownSec int) *InsightTrigger {
	if cooldownSec <= 0 {
		cooldownSec = 300
	}
	if timeoutMS <= 0 {
		timeoutMS = 10000
	}
	h := make(http.Header)
	h.Set("X-SERVICE-KEY", strings.TrimSpace(serviceKey))
	return &InsightTrigger{
		caller: &httpclient.Caller{
			BaseURL: strings.TrimRight(baseURL, "/"),
			Header:  h,
			HTTP:    &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond},
		},
		cooldown: time.Duration(cooldownSec) * time.Second,
		sem:      make(chan struct{}, 32),
	}
}

// NotifyDataChange dispara un cómputo de insights async con throttle por proyecto.
func (t *InsightTrigger) NotifyDataChange(projectID, userID string) {
	if t.caller.BaseURL == "" {
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

		st, _, err := t.caller.DoJSON(ctx, http.MethodPost, "/v1/insights/compute", nil,
			httpclient.WithHeader("X-USER-ID", userID),
			httpclient.WithHeader("X-PROJECT-ID", projectID),
		)
		if err != nil {
			log.Printf("[ai-trigger] error computando insights para proyecto %s: %v", projectID, err)
			return
		}
		if st >= 300 {
			log.Printf("[ai-trigger] insights compute retornó %d para proyecto %s", st, projectID)
		} else {
			log.Printf("[ai-trigger] insights computados para proyecto %s (status %d)", projectID, st)
		}
	}()
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

		trigger.NotifyDataChange(projectID, userID)
	}
}
