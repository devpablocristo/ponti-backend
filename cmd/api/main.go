package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/devpablocristo/platform/observability/go"

	"github.com/devpablocristo/ponti-backend/internal/shared/lifecycle"
	wire "github.com/devpablocristo/ponti-backend/wire"
)

func main() {
	logger := observability.NewJSONLogger("ponti-backend")
	slog.SetDefault(logger)

	metrics := observability.NewMetrics(observability.DefaultMetricsConfig("ponti_backend"))
	lifecycle.RegisterMetrics(metrics.Registry(), "ponti_backend")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tracerShutdown, err := observability.NewTracerProvider(ctx, tracingConfigFromEnv())
	if err != nil {
		// No abortamos por fallo de tracing: logueamos y seguimos con
		// TracerProvider no-op para no romper el servicio si el collector
		// está caído.
		logger.Warn("tracing init failed; continuing without tracing", "error", err.Error())
	}
	defer func() {
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()
		if err := tracerShutdown(shutdownCtx); err != nil {
			logger.Warn("tracing shutdown error", "error", err.Error())
		}
	}()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		<-sigChan
		logger.Info("received termination signal", "event", "shutdown_signal")
		cancel()
	}()

	deps, err := wire.Initialize()
	if err != nil {
		logger.Error("initializing dependencies failed", "error", err.Error())
		os.Exit(1)
	}

	if err := runHTTPServer(ctx, logger, metrics, deps); err != nil {
		logger.Error("running HTTP server failed", "error", err.Error())
		os.Exit(1)
	}

	logger.Info("application terminated successfully", "event", "shutdown_complete")
}

// tracingConfigFromEnv lee la configuración del TracerProvider desde env vars.
// Valores soportados:
//   - OTEL_EXPORTER: "otlp" | "stdout" | "none" (default "none")
//   - OTEL_OTLP_ENDPOINT: host:port (default "localhost:4318")
//   - OTEL_OTLP_INSECURE: "true" para HTTP (default "true")
//   - OTEL_SAMPLE_RATIO: 0.0 - 1.0 (default 1.0)
//   - ENVIRONMENT: dev / staging / prod (default "local")
//   - SERVICE_VERSION: pasa a service.version (default "0.0.0")
func tracingConfigFromEnv() observability.TracingConfig {
	cfg := observability.TracingConfig{
		ServiceName:    "ponti-backend",
		ServiceVersion: strings.TrimSpace(os.Getenv("SERVICE_VERSION")),
		Environment:    strings.TrimSpace(os.Getenv("ENVIRONMENT")),
		Exporter:       strings.TrimSpace(os.Getenv("OTEL_EXPORTER")),
		OTLPEndpoint:   strings.TrimSpace(os.Getenv("OTEL_OTLP_ENDPOINT")),
		OTLPInsecure:   true,
	}
	if v := strings.TrimSpace(os.Getenv("OTEL_OTLP_INSECURE")); v != "" {
		if parsed, err := strconv.ParseBool(v); err == nil {
			cfg.OTLPInsecure = parsed
		}
	}
	if v := strings.TrimSpace(os.Getenv("OTEL_SAMPLE_RATIO")); v != "" {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.SampleRatio = parsed
		}
	}
	return cfg
}
