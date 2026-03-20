package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	wire "github.com/devpablocristo/ponti-backend/wire"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		<-sigChan
		log.Println("Received termination signal. Shutting down the application...")
		cancel()
	}()

	deps, err := wire.Initialize()
	if err != nil {
		log.Fatalf("Error initializing dependencies: %s", err)
	}

	// Ejecutar el servidor HTTP.
	if err := runHTTPServer(ctx, deps); err != nil {
		log.Fatalf("Error running HTTP server: %v", err)
	}

	log.Println("Application terminated successfully.")
}
