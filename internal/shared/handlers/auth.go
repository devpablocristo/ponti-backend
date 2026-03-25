package sharedhandlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/devpablocristo/core/security/go/contextkeys"
	"github.com/devpablocristo/core/errors/go/domainerr"

	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
)

// ParseActor extrae el actor (email/sub) desde el contexto.
func ParseActor(c *gin.Context) (string, error) {
	actor, err := sharedmodels.ActorFromContext(c.Request.Context())
	if err != nil {
		return "", domainerr.Forbidden("invalid actor")
	}
	return actor, nil
}

// ParseOrgID extrae el org_id (uuid) desde el contexto.
func ParseOrgID(c *gin.Context) (uuid.UUID, error) {
	v := c.Request.Context().Value(ctxkeys.OrgID)
	if v == nil {
		return uuid.Nil, domainerr.Forbidden(fmt.Sprintf("invalid org_id: %s", "org_id missing in context"))
	}
	id, ok := v.(uuid.UUID)
	if !ok || id == uuid.Nil {
		return uuid.Nil, domainerr.Forbidden(fmt.Sprintf("invalid org_id: %s", "org_id is not a valid uuid"))
	}
	return id, nil
}

// ParseUserID es un alias temporal de ParseActor para backward compatibility.
// Retorna el actor como string. Los callers que necesitan int64 deben migrar.
// Deprecated: usar ParseActor.
func ParseUserID(c *gin.Context) (string, error) {
	return ParseActor(c)
}
