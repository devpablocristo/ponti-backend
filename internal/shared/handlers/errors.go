package sharedhandlers

import (
	"github.com/gin-gonic/gin"

	"github.com/devpablocristo/saas-core/shared/httperr"
)

// RespondError responde errores de dominio usando el helper estándar.
func RespondError(c *gin.Context, err error) {
	status, apiErr := httperr.Normalize(err)
	c.JSON(status, apiErr)
}
