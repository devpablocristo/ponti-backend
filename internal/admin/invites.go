package admin

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/devpablocristo/platform/errors/go/domainerr"

	authz "github.com/devpablocristo/ponti-backend/internal/shared/authz"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

const inviteTTL = 7 * 24 * time.Hour

type createInviteReq struct {
	Email    string `json:"email"`
	RoleName string `json:"role_name"`
}

type acceptInviteReq struct {
	Token string `json:"token"`
}

func newInviteToken() (token, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	token = hex.EncodeToString(b)
	return token, hashToken(token), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// canManageTenant: platform-admin, o admin/tenant_owner del tenant ACTIVO (que debe
// coincidir con el tenant del path). Usa el principal de authz (U0).
func (h *Handler) canManageTenant(c *gin.Context, tenantID uuid.UUID) bool {
	if h.isPlatformAdmin(c) {
		return true
	}
	p, ok := authz.PrincipalFromContext(c.Request.Context())
	if !ok || p.TenantID != tenantID {
		return false
	}
	// U2 dual-check: permiso fino invites:write con fallback (transición) a los roles
	// admin/tenant_owner del tenant activo.
	return authz.HasPermissionOrRole(c.Request.Context(), "invites:write", "admin", "tenant_owner")
}

// CreateInvite (U4): un admin/tenant_owner (o platform-admin) invita un email a su
// tenant con un rol. Devuelve el token EN CLARO una sola vez (no se persiste).
func (h *Handler) CreateInvite(c *gin.Context) {
	tenantID, err := uuid.Parse(strings.TrimSpace(c.Param("tenant_id")))
	if err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid tenant_id"))
		return
	}
	if !h.canManageTenant(c, tenantID) {
		sharedhandlers.RespondError(c, domainerr.Forbidden("not allowed to manage this tenant"))
		return
	}
	var req createInviteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid request payload"))
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" {
		sharedhandlers.RespondError(c, domainerr.Validation("email required"))
		return
	}
	ctx := c.Request.Context()
	rp := newRepo(h.db)
	roleID, err := rp.roleIDByName(ctx, req.RoleName)
	if err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid role_name"))
		return
	}
	token, tokenHash, err := newInviteToken()
	if err != nil {
		sharedhandlers.RespondError(c, domainerr.Internal("failed to generate token"))
		return
	}
	createdBy := ""
	if p, ok := authz.PrincipalFromContext(ctx); ok {
		createdBy = p.Subject
	}
	expiresAt := time.Now().UTC().Add(inviteTTL)
	id, err := rp.createInvite(ctx, tenantID, roleID, email, tokenHash, createdBy, expiresAt)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": gin.H{
		"invite_id":  id,
		"tenant_id":  tenantID,
		"email":      email,
		"token":      token, // one-time; solo se devuelve acá
		"expires_at": expiresAt,
	}})
}

// ListInvites (U4): lista los invites de un tenant.
func (h *Handler) ListInvites(c *gin.Context) {
	tenantID, err := uuid.Parse(strings.TrimSpace(c.Param("tenant_id")))
	if err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid tenant_id"))
		return
	}
	if !h.canManageTenant(c, tenantID) {
		sharedhandlers.RespondError(c, domainerr.Forbidden("not allowed to manage this tenant"))
		return
	}
	items, err := newRepo(h.db).listInvites(c.Request.Context(), tenantID)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": items})
}

// RevokeInvite (U4): revoca un invite pendiente.
func (h *Handler) RevokeInvite(c *gin.Context) {
	inviteID, err := uuid.Parse(strings.TrimSpace(c.Param("invite_id")))
	if err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid invite_id"))
		return
	}
	ctx := c.Request.Context()
	rp := newRepo(h.db)
	tenantID, err := rp.inviteTenantByID(ctx, inviteID)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if !h.canManageTenant(c, tenantID) {
		sharedhandlers.RespondError(c, domainerr.Forbidden("not allowed to manage this tenant"))
		return
	}
	if err := rp.revokeInvite(ctx, inviteID, tenantID); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// AcceptInvite (U4): el usuario autenticado canjea un token y obtiene la membership.
// Ruta tenant-agnóstica (GetIdentity): el invitado todavía no tiene membership.
func (h *Handler) AcceptInvite(c *gin.Context) {
	p, ok := authz.PrincipalFromContext(c.Request.Context())
	if !ok || strings.TrimSpace(p.Subject) == "" {
		sharedhandlers.RespondError(c, domainerr.Unauthorized("not authenticated"))
		return
	}
	var req acceptInviteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid request payload"))
		return
	}
	token := strings.TrimSpace(req.Token)
	if token == "" {
		sharedhandlers.RespondError(c, domainerr.Validation("token required"))
		return
	}
	ctx := c.Request.Context()
	rp := newRepo(h.db)
	inv, err := rp.getInviteByTokenHash(ctx, hashToken(token))
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if inv.Status != "pending" {
		sharedhandlers.RespondError(c, domainerr.Conflict("invite is not pending"))
		return
	}
	if time.Now().UTC().After(inv.ExpiresAt) {
		sharedhandlers.RespondError(c, domainerr.Validation("invite expired"))
		return
	}
	u, err := rp.ensureLocalUserByIDPSub(ctx, p.Subject, "")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := rp.acceptInvite(ctx, inv, u.ID); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"tenant_id": inv.TenantID}})
}
