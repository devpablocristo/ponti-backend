package sharedhandlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// ParsePaginationParams devuelve page y perPage con defaults seguros.
func ParsePaginationParams(c *gin.Context, defaultPage, defaultPerPage int) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(defaultPage)))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", strconv.Itoa(defaultPerPage)))
	if page < 1 {
		page = defaultPage
	}
	if perPage < 1 {
		perPage = defaultPerPage
	}
	return page, perPage
}
