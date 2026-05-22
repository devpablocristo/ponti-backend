package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/devpablocristo/platform/observability/go"

	wire "github.com/devpablocristo/ponti-backend/wire"
)

func main() {
	logger := observability.NewJSONLogger("ponti-backend")
	slog.SetDefault(logger)

	metrics := observability.NewMetrics(observability.DefaultMetricsConfig("ponti_backend"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
