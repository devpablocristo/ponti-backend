// Package admin exposes admin endpoints to manage Identity Platform users and RBAC memberships.
package admin

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/devpablocristo/core/errors/go/domainerr"
	"github.com/devpablocristo/core/security/go/contextkeys"

	"github.com/devpablocristo/ponti-backend/internal/shared/authz"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"

	"github.com/devpablocristo/ponti-backend/internal/admin/idp"
)

type GinEnginePort interface {
	GetRouter() *gin.Engine
	RunServer(ctx context.Context) error
}

type ConfigAPIPort interface {
	APIVersion() string
	APIBaseURL() string
}

type MiddlewaresEnginePort interface {
	GetGlobal() []gin.HandlerFunc
	GetValidation() []gin.HandlerFunc
	GetProtected() []gin.HandlerFunc
}

type Handler struct {
	db  *gorm.DB
	idp idp.AdminClient
	gsv GinEnginePort
	acf ConfigAPIPort
	mws MiddlewaresEnginePort
}

func NewHandler(db *gorm.DB, idpAdmin idp.AdminClient, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *Handler {
	return &Handler{db: db, idp: idpAdmin, gsv: s, acf: c, mws: m}
}

func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/admin"

	admin := r.Group(baseURL, h.mws.GetValidation()...)
	{
		admin.GET("/tenants", h.ListTenants)
		admin.POST("/tenants", h.CreateTenant)

		admin.GET("/users", h.ListUsers)
		admin.POST("/users", h.CreateUser)

		admin.POST("/memberships", h.UpsertMembership)
		admin.POST("/tenants/:tenant_id/invites", h.CreateInvite)
		admin.POST("/invites/accept", h.AcceptInvite)
		admin.POST("/tenants/:tenant_id/memberships/:membership_id/role", h.UpdateMembershipRole)
		admin.POST("/tenants/:tenant_id/memberships/:membership_id/archive", h.ArchiveMembership)
	}
}

func requireAdmin(c *gin.Context) bool {
	role, _ := c.Request.Context().Value(ctxkeys.Role).(string)
	if !isAdminLikeRole(role) {
		sharedhandlers.RespondError(c, domainerr.Forbidden("admin role required"))
		return false
	}
	return true
}

func requireAdminPermission(c *gin.Context, permission string) bool {
	if authz.HasPermission(c.Request.Context(), permission) {
		return true
	}
	sharedhandlers.RespondError(c, domainerr.Forbidden("insufficient permissions"))
	return false
}

func isAdminLikeRole(role string) bool {
	switch strings.TrimSpace(role) {
	case "admin", "saas_superadmin", "tenant_owner", "tenant_admin":
		return true
	default:
		return false
	}
}

type createTenantReq struct {
	Name string `json:"name"`
}

func (h *Handler) ListTenants(c *gin.Context) {
	if !requireAdminPermission(c, authz.PermissionAdminTenants) {
		return
	}
	rp := newRepo(h.db)
	items, err := rp.listTenants(c.Request.Context())
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": items})
}

func (h *Handler) CreateTenant(c *gin.Context) {
	if !requireAdminPermission(c, authz.PermissionAdminTenants) {
		return
	}
	var req createTenantReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid request payload"))
		return
	}
	rp := newRepo(h.db)
	id, err := rp.ensureTenantByName(c.Request.Context(), req.Name)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": gin.H{"id": id}})
}

type createUserReq struct {
	Email         string `json:"email"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	TenantName    string `json:"tenant_name"`
	RoleName      string `json:"role_name"`
	SendResetLink bool   `json:"send_reset_link"`
}

type createUserResp struct {
	User      *localUser `json:"user"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	RoleName  string     `json:"role_name"`
	ResetLink string     `json:"reset_link,omitempty"`
}

func usernameToEmail(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	if strings.Contains(v, "@") {
		return v
	}
	return v + "@ponti.local"
}

func (h *Handler) CreateUser(c *gin.Context) {
	if !requireAdminPermission(c, authz.PermissionAdminUsers) {
		return
	}
	var req createUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid request payload"))
		return
	}

	email := usernameToEmail(req.Email)
	if email == "" {
		email = usernameToEmail(req.Username)
	}
	password := strings.TrimSpace(req.Password)
	if email == "" || password == "" {
		sharedhandlers.RespondError(c, domainerr.Validation("email and password required"))
		return
	}

	ctx := c.Request.Context()

	// Create user in Identity Platform or fetch UID if already exists.
	uid, err := h.idp.CreateUser(ctx, email, password)
	if err != nil {
		// If the user already exists, allow attaching membership by looking up UID.
		if strings.Contains(strings.ToLower(err.Error()), "email") && strings.Contains(strings.ToLower(err.Error()), "exists") {
			uid, err = h.idp.GetUserUIDByEmail(ctx, email)
		}
	}
	if err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("unable to create identity user"))
		return
	}

	rp := newRepo(h.db)
	u, err := rp.ensureLocalUserByIDPSub(ctx, uid, email)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	tenantID, err := rp.ensureTenantByName(ctx, req.TenantName)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	roleID, err := rp.roleIDByName(ctx, req.RoleName)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := rp.upsertMembership(ctx, u.ID, tenantID, roleID); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	resp := createUserResp{
		User:     u,
		TenantID: tenantID,
		RoleName: strings.TrimSpace(req.RoleName),
	}
	if resp.RoleName == "" {
		resp.RoleName = "viewer"
	}
	if req.SendResetLink {
		if link, linkErr := h.idp.GeneratePasswordResetLink(ctx, email); linkErr == nil {
			resp.ResetLink = link
		}
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": resp})
}

func (h *Handler) ListUsers(c *gin.Context) {
	if !requireAdminPermission(c, authz.PermissionAdminUsers) {
		return
	}
	orgID, err := sharedhandlers.ParseOrgID(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	rp := newRepo(h.db)
	rows, err := rp.listUsersForTenant(c.Request.Context(), orgID)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": rows})
}

type upsertMembershipReq struct {
	Email      string `json:"email"`
	Username   string `json:"username"`
	TenantName string `json:"tenant_name"`
	RoleName   string `json:"role_name"`
}

func (h *Handler) UpsertMembership(c *gin.Context) {
	if !requireAdminPermission(c, authz.PermissionAdminMemberships) {
		return
	}
	var req upsertMembershipReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid request payload"))
		return
	}

	email := usernameToEmail(req.Email)
	if email == "" {
		email = usernameToEmail(req.Username)
	}
	if email == "" {
		sharedhandlers.RespondError(c, domainerr.Validation("email required"))
		return
	}

	ctx := c.Request.Context()
	uid, err := h.idp.GetUserUIDByEmail(ctx, email)
	if err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("identity user not found"))
		return
	}

	rp := newRepo(h.db)
	u, err := rp.ensureLocalUserByIDPSub(ctx, uid, email)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	tenantID, err := rp.ensureTenantByName(ctx, req.TenantName)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	roleID, err := rp.roleIDByName(ctx, req.RoleName)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := rp.upsertMembership(ctx, u.ID, tenantID, roleID); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"user_id": u.ID, "tenant_id": tenantID}})
}

type createInviteReq struct {
	Email     string `json:"email"`
	RoleName  string `json:"role_name"`
	ExpiresIn string `json:"expires_in"`
}

func (h *Handler) CreateInvite(c *gin.Context) {
	if !requireAdminPermission(c, authz.PermissionAdminMemberships) {
		return
	}
	tenantID, err := uuid.Parse(strings.TrimSpace(c.Param("tenant_id")))
	if err != nil || tenantID == uuid.Nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid tenant_id"))
		return
	}
	var req createInviteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid request payload"))
		return
	}
	email := usernameToEmail(req.Email)
	if email == "" {
		sharedhandlers.RespondError(c, domainerr.Validation("email required"))
		return
	}
	roleName := strings.TrimSpace(req.RoleName)
	if roleName == "" {
		roleName = "tenant_viewer"
	}
	expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour)
	if req.ExpiresIn != "" {
		if d, parseErr := time.ParseDuration(req.ExpiresIn); parseErr == nil && d > 0 {
			expiresAt = time.Now().UTC().Add(d)
		}
	}
	token, err := newInviteToken()
	if err != nil {
		sharedhandlers.RespondError(c, domainerr.Internal("unable to create invite token"))
		return
	}
	rp := newRepo(h.db)
	roleID, err := rp.roleIDByName(c.Request.Context(), roleName)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	invitedBy, _ := currentLocalUserID(c, h.db)
	invite, err := rp.createInvite(c.Request.Context(), tenantID, email, roleID, hashInviteToken(token), expiresAt, invitedBy)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": gin.H{
		"invite": invite,
		"token":  token,
	}})
}

type acceptInviteReq struct {
	Token string `json:"token"`
}

func (h *Handler) AcceptInvite(c *gin.Context) {
	var req acceptInviteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid request payload"))
		return
	}
	userID, err := currentLocalUserID(c, h.db)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	rp := newRepo(h.db)
	invite, err := rp.acceptInvite(c.Request.Context(), hashInviteToken(req.Token), userID)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := rp.upsertMembership(c.Request.Context(), userID, invite.TenantID, invite.RoleID); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": invite})
}

type updateMembershipRoleReq struct {
	RoleName string `json:"role_name"`
}

func (h *Handler) UpdateMembershipRole(c *gin.Context) {
	if !requireAdminPermission(c, authz.PermissionAdminMemberships) {
		return
	}
	tenantID, membershipID, ok := parseTenantAndMembership(c)
	if !ok {
		return
	}
	var req updateMembershipRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid request payload"))
		return
	}
	rp := newRepo(h.db)
	roleID, err := rp.roleIDByName(c.Request.Context(), req.RoleName)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := rp.updateMembershipRole(c.Request.Context(), tenantID, membershipID, roleID); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) ArchiveMembership(c *gin.Context) {
	if !requireAdminPermission(c, authz.PermissionAdminMemberships) {
		return
	}
	tenantID, membershipID, ok := parseTenantAndMembership(c)
	if !ok {
		return
	}
	rp := newRepo(h.db)
	if err := rp.archiveMembership(c.Request.Context(), tenantID, membershipID); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func parseTenantAndMembership(c *gin.Context) (uuid.UUID, uuid.UUID, bool) {
	tenantID, err := uuid.Parse(strings.TrimSpace(c.Param("tenant_id")))
	if err != nil || tenantID == uuid.Nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid tenant_id"))
		return uuid.Nil, uuid.Nil, false
	}
	membershipID, err := uuid.Parse(strings.TrimSpace(c.Param("membership_id")))
	if err != nil || membershipID == uuid.Nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid membership_id"))
		return uuid.Nil, uuid.Nil, false
	}
	return tenantID, membershipID, true
}

func newInviteToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func hashInviteToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return fmt.Sprintf("%x", sum[:])
}

func currentLocalUserID(c *gin.Context, db *gorm.DB) (uuid.UUID, error) {
	actor, err := sharedhandlers.ParseActor(c)
	if err != nil {
		return uuid.Nil, err
	}
	type row struct{ ID uuid.UUID }
	var out row
	if err := db.WithContext(c.Request.Context()).
		Table("users").
		Select("id").
		Where("idp_sub = ?", actor).
		Limit(1).
		Take(&out).Error; err != nil {
		return uuid.Nil, domainerr.Forbidden("local user not found")
	}
	return out.ID, nil
}
