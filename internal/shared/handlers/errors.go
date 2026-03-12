package sharedhandlers

import (
	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

// RespondError responde errores de dominio usando el helper estándar.
func RespondError(c *gin.Context, err error) {
	apiErr, status := types.NewAPIError(err)
	c.JSON(status, apiErr.ToResponse())
}

