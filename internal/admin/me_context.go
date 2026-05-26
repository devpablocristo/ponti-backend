package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/devpablocristo/platform/security/go/contextkeys"

	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

func (h *Handler) registerMeContextRoute() {
	group := h.gsv.GetRouter().Group(h.acf.APIBaseURL(), h.mws.GetValidation()...)
	group.GET("/me/context", h.GetMeContext)
}

// GetMeContext godoc
// @Summary      Bootstrap del FE: usuario + tenants + roles + permisos
// @Description  Endpoint de arranque del frontend. Lee el JWT del header Authorization, identifica al usuario local, y devuelve los tenants donde tiene membership con sus roles + permisos. Si X-Tenant-ID está presente, marca ese tenant como `is_current=true`.
// @Tags         admin
// @Produce      json
// @Success      200  {object}  MeContext
// @Failure      401  {object}  map[string]string  "authentication context required"
// @Failure      403  {object}  map[string]string  "local user not found / tenant context required"
// @Failure      500  {object}  map[string]string  "internal"
// @Security     BearerAuth
// @Router       /me/context [get]
func (h *Handler) GetMeContext(c *gin.Context) {
	actor, _ := c.Request.Context().Value(ctxkeys.Actor).(string)
	currentTenantID, _ := c.Request.Context().Value(ctxkeys.OrgID).(uuid.UUID)

	out, err := h.uc.GetMeContext(c.Request.Context(), actor, currentTenantID)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, out)
}
