package pkggin

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func Bootstrap(port, apiVersion string, isTest bool) (*Server, error) {
	if gin.Mode() == gin.TestMode {
		return newTestServer()
	}

	if port == "" {
		port = os.Getenv("HTTP_SERVER_PORT")
	}
	if apiVersion == "" {
		apiVersion = os.Getenv("API_VERSION")
	}

	config := newConfig(
		port,
		apiVersion,
	)

	if err := config.Validate(); err != nil {
		return nil, err
	}

	Server, err := newServer(config)
	if err != nil {
		return nil, err
	}

	r := Server.GetRouter()

	api := r.Group(fmt.Sprintf("/api/%s", apiVersion))
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})

		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":    "healthy",
				"timestamp": time.Now(),
			})
		})
	}

	return Server, nil
}
