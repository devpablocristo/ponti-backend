package sharedhandlers

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/devpablocristo/core/errors/go/domainerr"
)

// ParseParamID parsea un ID de path y valida que sea > 0.
func ParseParamID(raw string, param string) (int64, error) {
	value := strings.TrimSpace(raw)
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, domainerr.Validation("invalid " + param)
	}
	return id, nil
}

// ParseMovementIDParam obtiene el ID de movimiento desde supply_movement_id o stock_movement_id.
func ParseMovementIDParam(c *gin.Context) (int64, error) {
	raw := c.Param("supply_movement_id")
	if raw == "" {
		raw = c.Param("stock_movement_id")
	}
	return ParseParamID(raw, "movement_id")
}
