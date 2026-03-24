package sharedhandlers

import (
	"github.com/gin-gonic/gin"

	ginmw "github.com/devpablocristo/core/backend/gin/go"
	"github.com/devpablocristo/core/backend/go/httperr"
)

// RespondError delega al helper estándar de core.
func RespondError(c *gin.Context, err error) {
	ginmw.Respond(c, err)
}

// ErrorMessage extrae el mensaje user-facing de un error normalizado.
func ErrorMessage(err error) string {
	_, apiErr := httperr.Normalize(err)
	return apiErr.Message
}
