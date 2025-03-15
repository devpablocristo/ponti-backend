package assessment

import (
	"net/http"

	dto "github.com/alphacodinggroup/euxcel-backend/internal/assessment/handler/dto"
	types "github.com/alphacodinggroup/euxcel-backend/pkg/types"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GenerateLink(c *gin.Context) {
	assessmentID := c.Param("id")

	linkId, err := h.ucs.GenerateLink(c.Request.Context(), assessmentID)
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	c.JSON(http.StatusCreated, dto.GenerateLinkResponse{
		Message: "Assessment link successfully generated",
		LinkID:  linkId,
	})
}

func (h *Handler) SendLink(c *gin.Context) {
	assessmentLinkID := c.Param("id")
	if err := h.ucs.SendLink(c.Request.Context(), assessmentLinkID); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	c.JSON(http.StatusCreated, dto.SendLinkResponse{
		Message: "Unique link successfully sent",
		Link:    assessmentLinkID,
	})
}
