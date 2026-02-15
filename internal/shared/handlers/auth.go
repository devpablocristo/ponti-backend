package sharedhandlers

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	sharedmodels "github.com/alphacodinggroup/ponti-backend/internal/shared/models"
	pkgmwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

// ParseUserID extrae y valida el user_id desde el contexto.
func ParseUserID(c *gin.Context) (int64, error) {
	userID, err := sharedmodels.ConvertStringToID(c.Request.Context())
	if err != nil {
		return 0, types.NewError(types.ErrAuthorization, "invalid user_id", err)
	}
	return userID, nil
}

func ParseTenantID(c *gin.Context) (int64, error) {
	raw := c.Request.Context().Value(pkgmwr.ContextTenantIDKey)
	str, ok := raw.(string)
	if !ok || str == "" {
		return 0, types.NewError(types.ErrAuthorization, "invalid tenant_id", fmt.Errorf("tenant_id missing in context"))
	}

	id, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, types.NewError(types.ErrAuthorization, "invalid tenant_id", err)
	}
	return id, nil
}
