package sharedhandlers

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/devpablocristo/core/saas/go/shared/domainerr"
	"github.com/devpablocristo/core/saas/go/shared/httperr"
)

// RespondError responde errores de dominio usando el helper estándar.
func RespondError(c *gin.Context, err error) {
	status, apiErr := httperr.Normalize(err)
	c.JSON(status, apiErr)
}

// ErrorMessage extracts the human-readable message from a domainerr.Error,
// falling back to err.Error() for other error types.
func ErrorMessage(err error) string {
	var de domainerr.Error
	if errors.As(err, &de) {
		return de.Message()
	}
	return err.Error()
}
