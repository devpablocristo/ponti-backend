package sharedhandlers

import (
	"strconv"
	"strings"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

// ParseParamID parsea un ID de path y valida que sea > 0.
func ParseParamID(raw string, param string) (int64, error) {
	value := strings.TrimSpace(raw)
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, types.NewError(types.ErrInvalidID, "invalid "+param, err)
	}
	return id, nil
}
