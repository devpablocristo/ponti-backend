package ai

import (
	"context"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// InsightTrigger dispara cómputo de insights de forma asíncrona y throttleada.
// Después de una mutación exitosa en el backend, notifica a ponti-ai para que
// recalcule insights del proyecto afectado.
type InsightTrigger struct {
	baseURL    string
	serviceKey string
	httpClient *http.Client
	throttle   sync.Map      // projectID -> time.Time
	cooldown   time.Duration // mínimo entre disparos por proyecto
}

// NewInsightTrigger crea un trigger con cooldown configurable.
// Si cooldownSec <= 0, usa 300s (5 minutos) como default.
func NewInsightTrigger(baseURL, serviceKey string, timeoutMS, cooldownSec int) *InsightTrigger {
	if cooldownSec <= 0 {
		cooldownSec = 300
	}
	if timeoutMS <= 0 {
		timeoutMS = 10000
	}
	return &InsightTrigger{
		baseURL:    strings.TrimRight(baseURL, "/"),
		serviceKey: serviceKey,
		httpClient: &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond},
		cooldown:   time.Duration(cooldownSec) * time.Second,
	}
}

// NotifyDataChange dispara un cómputo de insights async si el proyecto no fue
// computado recientemente (throttle por cooldown).
func (t *InsightTrigger) NotifyDataChange(projectID, userID string) {
	if t.baseURL == "" || t.serviceKey == "" {
		return // AI service no configurado
	}

	// Throttle: saltar si se disparó recientemente para este proyecto
	now := time.Now()
	if last, ok := t.throttle.Load(projectID); ok {
		if now.Sub(last.(time.Time)) < t.cooldown {
			return
		}
	}
	t.throttle.Store(projectID, now)

	// Fire and forget
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.baseURL+"/v1/insights/compute", nil)
		if err != nil {
			log.Printf("[ai-trigger] error creando request para proyecto %s: %v", projectID, err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-SERVICE-KEY", t.serviceKey)
		req.Header.Set("X-USER-ID", userID)
		req.Header.Set("X-PROJECT-ID", projectID)

		resp, err := t.httpClient.Do(req)
		if err != nil {
			log.Printf("[ai-trigger] error computando insights para proyecto %s: %v", projectID, err)
			return
		}
		defer func() { _ = resp.Body.Close() }()
		_, _ = io.Copy(io.Discard, resp.Body)

		if resp.StatusCode >= 300 {
			log.Printf("[ai-trigger] insights compute retornó %d para proyecto %s", resp.StatusCode, projectID)
		} else {
			log.Printf("[ai-trigger] insights computados para proyecto %s (status %d)", projectID, resp.StatusCode)
		}
	}()
}

// InsightTriggerMiddleware crea un middleware Gin que dispara cómputo de insights
// después de mutaciones exitosas (POST/PUT/PATCH/DELETE con respuesta 2xx).
//
// Extrae project_id del path param `:project_id`. Si no existe, no dispara.
// Ignora rutas de AI para evitar triggers recursivos.
func InsightTriggerMiddleware(trigger *InsightTrigger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // Ejecutar handler primero

		// Solo disparar en mutaciones exitosas
		method := c.Request.Method
		if method != http.MethodPost && method != http.MethodPut &&
			method != http.MethodPatch && method != http.MethodDelete {
			return
		}

		status := c.Writer.Status()
		if status < 200 || status >= 300 {
			return
		}

		// Ignorar rutas de AI para evitar recursión
		if strings.Contains(c.Request.URL.Path, "/ai/") {
			return
		}

		// Intentar obtener project_id del path
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
