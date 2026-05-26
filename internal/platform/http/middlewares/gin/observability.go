package pkgmwr

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/devpablocristo/platform/observability/go"
)

const tracerName = "github.com/devpablocristo/ponti-backend/internal/platform/http/middlewares/gin"

// Observability inyecta un request-ID y un logger request-scoped en el
// contexto, propagando el header X-Request-Id y emitiendo un access log JSON.
// También abre un span OTel por request (no-op si el TracerProvider global no
// está configurado) y enriquece el logger con trace_id/span_id cuando aplica.
func Observability(logger *slog.Logger) gin.HandlerFunc {
	return observabilityHandler(logger, nil)
}

// ObservabilityWithMetrics extiende Observability registrando RED metrics
// Prometheus (counter, errors, duration histogram) por request usando el
// helper Metrics de platform/observability/go.
func ObservabilityWithMetrics(logger *slog.Logger, metrics *observability.Metrics) gin.HandlerFunc {
	return observabilityHandler(logger, metrics)
}

func observabilityHandler(logger *slog.Logger, metrics *observability.Metrics) gin.HandlerFunc {
	if logger == nil {
		logger = observability.NewJSONLogger("unknown")
	}
	tracer := otel.Tracer(tracerName)
	propagator := otel.GetTextMapPropagator()
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader(observability.RequestIDHeader))
		if requestID == "" {
			requestID = newRequestID()
		}

		c.Writer.Header().Set(observability.RequestIDHeader, requestID)

		// Propagación trace context entrante (W3C traceparent + baggage).
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))
		ctx = observability.ContextWithRequestID(ctx, requestID)

		spanName := c.FullPath()
		if spanName == "" {
			spanName = c.Request.Method + " " + c.Request.URL.Path
		} else {
			spanName = c.Request.Method + " " + spanName
		}
		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPRequestMethodKey.String(c.Request.Method),
				semconv.URLPath(c.Request.URL.Path),
				attribute.String("http.route", c.FullPath()),
				attribute.String("http.request_id", requestID),
				semconv.UserAgentOriginal(c.Request.UserAgent()),
				attribute.String("net.peer.ip", c.ClientIP()),
			),
		)
		defer span.End()

		requestLogger := logger.With("request_id", requestID)
		if sc := span.SpanContext(); sc.IsValid() {
			requestLogger = requestLogger.With(
				"trace_id", sc.TraceID().String(),
				"span_id", sc.SpanID().String(),
			)
		}
		ctx = observability.ContextWithLogger(ctx, requestLogger)
		c.Request = c.Request.WithContext(ctx)

		start := time.Now()
		c.Next()
		duration := time.Since(start)

		status := c.Writer.Status()
		route := c.FullPath()

		span.SetAttributes(semconv.HTTPResponseStatusCode(status))
		if status >= 500 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", status))
		}

		// Releer el logger del context para captar enriquecimientos hechos
		// por middlewares posteriores (auth → user_id/tenant_id/role).
		// LoggerFromContext devuelve un logger no-op si el context no tiene
		// uno, así que es seguro incluso para rutas sin auth middleware.
		finalLogger := observability.LoggerFromContext(c.Request.Context())
		finalLogger.Info("http request completed",
			"event", "http_request_completed",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"route", route,
			"status", status,
			"duration_ms", duration.Milliseconds(),
			"remote_addr", c.ClientIP(),
		)

		if metrics != nil {
			// platform/observability/go.routeLabel lee r.Pattern (formato
			// "METHOD route") para etiquetar las métricas. Gin no lo setea,
			// así que lo poblamos con FullPath para evitar que toda la
			// cardinalidad caiga en "unmatched". FullPath queda vacío para
			// rutas no matcheadas (404), en cuyo caso routeLabel devuelve
			// "unmatched" intencionalmente.
			if route != "" {
				c.Request.Pattern = c.Request.Method + " " + route
			}
			metrics.ObserveHTTPRequest(c.Request, status, duration)
		}
	}
}

func newRequestID() string {
	var buf [12]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(buf[:])
}
