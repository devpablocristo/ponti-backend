package sharedhandlers

import (
	"github.com/gin-gonic/gin"

	sharedmodels "github.com/alphacodinggroup/ponti-backend/internal/shared/models"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

// ParseUserID extrae y valida el user_id desde el contexto.
func ParseUserID(c *gin.Context) (int64, error) {
	userID, err := sharedmodels.ConvertStringToID(c)
	if err != nil {
		return 0, types.NewError(types.ErrAuthorization, "invalid user_id", err)
	}
	return userID, nil
}
