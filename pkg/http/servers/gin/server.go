package pkggin

import (
	"context"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	instance  *Server
	once      sync.Once
	initError error
)

type Config interface {
	GetRouterPort() string
	Validate() error
}

type Server struct {
	router *gin.Engine
	config Config
}

func newServer(config Config) (*Server, error) {
	once.Do(func() {
		err := config.Validate()
		if err != nil {
			initError = err
			return
		}

		r := gin.New()
		instance = &Server{
			config: config,
			router: r,
		}
	})
	return instance, initError
}

func newTestServer() (*Server, error) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	testConfig := &config{
		routerPort: "8080",
		apiVersion: "v1",
	}

	return &Server{
		router: r,
		config: testConfig,
	}, nil
}

// RunServer lanza el servidor en el puerto configurado.
func (s *Server) RunServer(ctx context.Context) error {
	// Ejemplo de "Run" bloqueante:
	return s.router.Run(":" + s.config.GetRouterPort())
}

// GetRouter expone el router para poder añadir rutas, middlewares, etc.
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

// WrapH sirve para anidar un http.Handler dentro de Gin.
func (s *Server) WrapH(h http.Handler) gin.HandlerFunc {
	return gin.WrapH(h)
}
