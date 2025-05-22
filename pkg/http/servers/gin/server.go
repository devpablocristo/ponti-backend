package pkggin

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// con sigleton
var (
	instance  Server
	once      sync.Once
	initError error
)

type server struct {
	router *gin.Engine
	config Config
}

func newServer(config Config) (Server, error) {
	once.Do(func() {
		err := config.Validate()
		if err != nil {
			initError = err
			return
		}

		r := gin.New()
		instance = &server{
			config: config,
			router: r,
		}
	})
	return instance, initError
}

// sin singleton
// type server struct {
// 	router *gin.Engine
// 	config Config
// }

// func newServer(cfg Config) (Server, error) {
// 	r := gin.Default()
// 	return &server{
// 		router: r,
// 		config: cfg,
// 	}, nil
// }

func newTestServer() (Server, error) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	testConfig := &config{
		routerPort: "8080",
		apiVersion: "v1",
	}

	return &server{
		router: r,
		config: testConfig,
	}, nil
}

func (s *server) RunServer(ctx context.Context) error {
	httpServer := &http.Server{
		Addr:    ":" + s.config.GetRouterPort(),
		Handler: s.router,
	}

	errChan := make(chan error, 1)

	go func() {
		log.Printf("Starting server on port %s...\n", s.config.GetRouterPort())
		errChan <- httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		log.Println("Context canceled. Shutting down HTTP server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("HTTP server shutdown failed: %w", err)
		}
		return nil

	case err := <-errChan:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("HTTP server error: %w", err)
		}
		return nil
	}
}

func (s *server) GetRouter() *gin.Engine {
	return s.router
}

func (s *server) GetApiVersion() string {
	return s.config.GetApiVersion()
}

func (s *server) WrapH(h http.Handler) gin.HandlerFunc {
	return gin.WrapH(h)
}
