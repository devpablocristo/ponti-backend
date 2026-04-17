package sharedhandlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// ParsePaginationParams devuelve page y perPage con defaults seguros.
// Acepta "per_page" como parámetro principal y "page_size" como fallback
// para compatibilidad.
func ParsePaginationParams(c *gin.Context, defaultPage, defaultPerPage int) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(defaultPage)))

	perPageStr := c.Query("per_page")
	if perPageStr == "" {
		perPageStr = c.Query("page_size")
	}

	perPage := defaultPerPage
	if perPageStr != "" {
		if v, err := strconv.Atoi(perPageStr); err == nil {
			perPage = v
		}
	}

	if page < 1 {
		page = defaultPage
	}
	if perPage < 1 {
		perPage = defaultPerPage
	}
	return page, perPage
}
