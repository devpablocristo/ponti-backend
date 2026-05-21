package pkgmwr

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/devpablocristo/platform/observability/go"
)

// Observability inyecta un request-ID y un logger request-scoped en el
// contexto, propagando el header X-Request-Id y emitiendo un access log JSON.
// Reusa los helpers de platform/observability/go para mantener consistencia
// con el resto del ecosistema.
func Observability(logger *slog.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = observability.NewJSONLogger("unknown")
	}
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader(observability.RequestIDHeader))
		if requestID == "" {
			requestID = newRequestID()
		}

		c.Writer.Header().Set(observability.RequestIDHeader, requestID)

		ctx := observability.ContextWithRequestID(c.Request.Context(), requestID)
		requestLogger := logger.With("request_id", requestID)
		ctx = observability.ContextWithLogger(ctx, requestLogger)
		c.Request = c.Request.WithContext(ctx)

		start := time.Now()
		c.Next()
		duration := time.Since(start)

		requestLogger.Info("http request completed",
			"event", "http_request_completed",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"route", c.FullPath(),
			"status", c.Writer.Status(),
			"duration_ms", duration.Milliseconds(),
			"remote_addr", c.ClientIP(),
		)
	}
}

func newRequestID() string {
	var buf [12]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(buf[:])
}
