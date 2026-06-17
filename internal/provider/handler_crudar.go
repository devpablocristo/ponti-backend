package provider

import (
	"net/http"

	ginmw "github.com/devpablocristo/platform/http/gin/go"
	"github.com/gin-gonic/gin"

	"github.com/devpablocristo/ponti-backend/internal/provider/handler/dto"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

// CreateProvider (POST /providers).
func (h *Handler) CreateProvider(c *gin.Context) {
	var req dto.CreateProviderRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	id, err := h.ucs.CreateProvider(c.Request.Context(), req.ToDomain())
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondCreated(c, id)
}

// GetArchivedProviders (GET /providers/archived).
func (h *Handler) GetArchivedProviders(c *gin.Context) {
	items, err := h.ucs.GetArchivedProviders(c.Request.Context())
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewProvidersDetailResponse(items))
}

// GetProvider (GET /providers/:provider_id).
func (h *Handler) GetProvider(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "provider_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	p, err := h.ucs.GetProvider(c.Request.Context(), id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.ProviderFromDomain(p))
}

// UpdateProvider (PUT /providers/:provider_id).
func (h *Handler) UpdateProvider(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "provider_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req dto.UpdateProviderRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	if err := h.ucs.UpdateProvider(c.Request.Context(), req.ToDomain(id)); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

// DeleteProvider (DELETE /providers/:provider_id) — hard delete.
func (h *Handler) DeleteProvider(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "provider_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.DeleteProvider(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

// ArchiveProvider (POST /providers/:provider_id/archive).
func (h *Handler) ArchiveProvider(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "provider_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.ArchiveProvider(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

// RestoreProvider (POST /providers/:provider_id/restore).
func (h *Handler) RestoreProvider(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "provider_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.RestoreProvider(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}
