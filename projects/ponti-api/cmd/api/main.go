package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	env "github.com/alphacodinggroup/ponti-backend/pkg/config/env"
	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
	wire "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/wire"
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

	// Load configuration
	cfgSet, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("unable to load config: %v", err)
	}

	// Initialize dependencies using Wire
	deps, err := wire.Initialize(cfgSet)
	if err != nil {
		log.Fatalf("Error initializing dependencies: %s", err)
	}

	// Set environment
	currentEnv := env.GetFromString(cfgSet.General.Environment)
	switch currentEnv {
	case env.Local, env.Dev:
		if err := RunGormMigrations(ctx, deps.GormRepo); err != nil {
			log.Fatalf("Failed to run Gorm migrations: %v", err)
		}
	}

	// Run the HTTP server
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		if err := RunHttpServer(ctx, deps); err != nil {
			log.Fatalf("Error running HTTP server: %v", err)
		}
	}()

	wg.Wait()

	log.Println("Application terminated successfully.")
}
