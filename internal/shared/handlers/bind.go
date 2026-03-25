package sharedhandlers

import (
	"github.com/gin-gonic/gin"

	ginmw "github.com/devpablocristo/core/http/go/gin"
)

// BindJSON delega al helper estándar de core.
func BindJSON(c *gin.Context, req any) error {
	return ginmw.BindJSON(c, req)
}

// HumanizeBindError delega al helper estándar de core.
func HumanizeBindError(err error) string {
	return ginmw.HumanizeBindError(err)
}
