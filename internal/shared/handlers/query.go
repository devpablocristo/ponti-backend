package sharedhandlers

import (
	"github.com/gin-gonic/gin"

	ginmw "github.com/devpablocristo/core/http/gin/go"
)

// ParseOptionalInt64Query delega al helper estándar de core.
func ParseOptionalInt64Query(c *gin.Context, key string) (*int64, error) {
	return ginmw.ParseOptionalInt64Query(c, key)
}
