package campaign

import (
	"net/http"

	ginmw "github.com/devpablocristo/platform/http/gin/go"
	"github.com/gin-gonic/gin"

	dto "github.com/devpablocristo/ponti-backend/internal/campaign/handler/dto"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

// CreateCampaign (POST /campaigns).
func (h *Handler) CreateCampaign(c *gin.Context) {
	var req dto.CreateCampaignRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	id, err := h.ucs.CreateCampaign(c.Request.Context(), req.ToDomain())
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondCreated(c, id)
}

// GetArchivedCampaigns (GET /campaigns/archived).
func (h *Handler) GetArchivedCampaigns(c *gin.Context) {
	items, err := h.ucs.GetArchivedCampaigns(c.Request.Context())
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewCampaignsDetailResponse(items))
}

// GetCampaign (GET /campaigns/:campaign_id).
func (h *Handler) GetCampaign(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "campaign_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	camp, err := h.ucs.GetCampaign(c.Request.Context(), id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.CampaignFromDomain(camp))
}

// UpdateCampaign (PUT /campaigns/:campaign_id).
func (h *Handler) UpdateCampaign(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "campaign_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req dto.UpdateCampaignRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	if err := h.ucs.UpdateCampaign(c.Request.Context(), req.ToDomain(id)); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

// DeleteCampaign (DELETE /campaigns/:campaign_id) — hard delete.
func (h *Handler) DeleteCampaign(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "campaign_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.DeleteCampaign(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

// ArchiveCampaign (POST /campaigns/:campaign_id/archive).
func (h *Handler) ArchiveCampaign(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "campaign_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.ArchiveCampaign(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

// RestoreCampaign (POST /campaigns/:campaign_id/restore).
func (h *Handler) RestoreCampaign(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "campaign_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.RestoreCampaign(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}
