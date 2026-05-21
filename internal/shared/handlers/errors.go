package sharedhandlers

import (
	"github.com/gin-gonic/gin"

	ginmw "github.com/devpablocristo/platform/http/gin/go"
	"github.com/devpablocristo/platform/http/go/httperr"
	"github.com/devpablocristo/platform/observability/go"
)

// RespondError loggea el error con el logger request-scoped y delega la
// normalización HTTP a platform/http/gin/go.
func RespondError(c *gin.Context, err error) {
	if err != nil {
		status, apiErr := httperr.Normalize(err)
		observability.LoggerFromContext(c.Request.Context()).Error("http error",
			"event", "http_error",
			"error", err.Error(),
			"status", status,
			"code", apiErr.Code,
			"message", apiErr.Message,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"route", c.FullPath(),
		)
	}
	ginmw.Respond(c, err)
}

// ErrorMessage extrae el mensaje user-facing de un error normalizado.
func ErrorMessage(err error) string {
	_, apiErr := httperr.Normalize(err)
	return apiErr.Message
}
