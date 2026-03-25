package sharedhandlers

import (
	"github.com/gin-gonic/gin"

	ginmw "github.com/devpablocristo/core/http/go/gin"
)

// ParsePaginationParams devuelve page y perPage con defaults seguros.
// Delega a ginmw.ParsePageQuery que acepta "per_page" y "page_size" como fallback.
func ParsePaginationParams(c *gin.Context, defaultPage, defaultPerPage int) (int, int) {
	return ginmw.ParsePageQuery(c, defaultPage, defaultPerPage)
}
