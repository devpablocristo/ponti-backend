package sharedhandlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/devpablocristo/saas-core/shared/ctxkeys"

	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
	types "github.com/devpablocristo/ponti-backend/pkg/types"
)

// ParseActor extrae el actor (email/sub) desde el contexto.
func ParseActor(c *gin.Context) (string, error) {
	actor, err := sharedmodels.ActorFromContext(c.Request.Context())
	if err != nil {
		return "", types.NewError(types.ErrAuthorization, "invalid actor", err)
	}
	return actor, nil
}

// ParseOrgID extrae el org_id (uuid) desde el contexto.
func ParseOrgID(c *gin.Context) (uuid.UUID, error) {
	v := c.Request.Context().Value(ctxkeys.OrgID)
	if v == nil {
		return uuid.Nil, types.NewError(types.ErrAuthorization, "invalid org_id", fmt.Errorf("org_id missing in context"))
	}
	id, ok := v.(uuid.UUID)
	if !ok || id == uuid.Nil {
		return uuid.Nil, types.NewError(types.ErrAuthorization, "invalid org_id", fmt.Errorf("org_id is not a valid uuid"))
	}
	return id, nil
}

// ParseUserID es un alias temporal de ParseActor para backward compatibility.
// Retorna el actor como string. Los callers que necesitan int64 deben migrar.
// Deprecated: usar ParseActor.
func ParseUserID(c *gin.Context) (string, error) {
	return ParseActor(c)
}
