package actors

import (
	ginmw "github.com/devpablocristo/platform/http/gin/go"
	"github.com/gin-gonic/gin"

	dto "github.com/devpablocristo/ponti-backend/internal/actors/handler/dto"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

// ListActors (GET /actors?status=active|archived|all&page=&per_page=).
func (h *Handler) ListActors(c *gin.Context) {
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 100)
	status := c.DefaultQuery("status", "active")
	items, total, err := h.ucs.List(c.Request.Context(), status, page, perPage)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.NewListActorsResponse(items, page, perPage, total))
}

// GetActor (GET /actors/:actor_id).
func (h *Handler) GetActor(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "actor_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	a, err := h.ucs.Get(c.Request.Context(), id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.ActorFromDomain(a))
}

// UpdateActor (PUT /actors/:actor_id).
func (h *Handler) UpdateActor(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "actor_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req dto.UpdateActorRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	if err := h.ucs.Update(c.Request.Context(), req.ToDomain(id)); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

// DeleteActor (DELETE /actors/:actor_id) — hard delete.
func (h *Handler) DeleteActor(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "actor_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.Delete(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

// ArchiveActor (POST /actors/:actor_id/archive) — soft delete.
func (h *Handler) ArchiveActor(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "actor_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.Archive(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

// SetActorRoles (PUT /actors/:actor_id/roles) — reemplaza el conjunto de roles.
func (h *Handler) SetActorRoles(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "actor_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req dto.SetRolesRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	if err := h.ucs.SetRoles(c.Request.Context(), id, req.Roles); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

// SetActorTaxID (PUT /actors/:actor_id/tax-id) — corrige la clave fiscal sin re-crear el
// actor. 409 si otra identidad activa ya tiene ese CUIT/DNI.
func (h *Handler) SetActorTaxID(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "actor_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req dto.SetTaxIDRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	if err := h.ucs.SetTaxID(c.Request.Context(), id, req.TaxID); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

// RestoreActor (POST /actors/:actor_id/restore).
func (h *Handler) RestoreActor(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "actor_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.ucs.Restore(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}
