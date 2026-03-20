package sharedhandlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/devpablocristo/saas-core/shared/domainerr"
)

// ParseOptionalInt64Query parsea un query param opcional a int64.
func ParseOptionalInt64Query(c *gin.Context, key string) (*int64, error) {
	raw := c.Query(key)
	if raw == "" {
		return nil, nil
	}
	val, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || val <= 0 {
		return nil, domainerr.Validation("invalid " + key)
	}
	return &val, nil
}
