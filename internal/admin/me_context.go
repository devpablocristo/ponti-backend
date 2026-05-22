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
