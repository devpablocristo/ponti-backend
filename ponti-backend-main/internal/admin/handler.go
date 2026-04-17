// Package admin exposes admin endpoints to manage Identity Platform users and RBAC memberships.
package admin

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	sharedhandlers "github.com/alphacodinggroup/ponti-backend/internal/shared/handlers"
	pkgmwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pkgtypes "github.com/alphacodinggroup/ponti-backend/pkg/types"

	"github.com/alphacodinggroup/ponti-backend/internal/admin/idp"
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

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	admin := r.Group(baseURL)
	{
		admin.GET("/tenants", h.ListTenants)
		admin.POST("/tenants", h.CreateTenant)

		admin.GET("/users", h.ListUsers)
		admin.POST("/users", h.CreateUser)

		admin.POST("/memberships", h.UpsertMembership)
	}
}

func requireAdmin(c *gin.Context) bool {
	raw, ok := c.Get(pkgmwr.ContextRoles)
	if !ok {
		sharedhandlers.RespondError(c, pkgtypes.NewError(pkgtypes.ErrAuthorization, "admin role required", nil))
		return false
	}
	roles, ok := raw.([]string)
	if !ok || len(roles) == 0 || roles[0] != "admin" {
		sharedhandlers.RespondError(c, pkgtypes.NewError(pkgtypes.ErrAuthorization, "admin role required", nil))
		return false
	}
	return true
}

type createTenantReq struct {
	Name string `json:"name"`
}

func (h *Handler) ListTenants(c *gin.Context) {
	if !requireAdmin(c) {
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
	if !requireAdmin(c) {
		return
	}
	var req createTenantReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedhandlers.RespondError(c, pkgtypes.NewError(pkgtypes.ErrBadRequest, "invalid request payload", err))
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
	Email          string `json:"email"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	TenantName     string `json:"tenant_name"`
	RoleName       string `json:"role_name"`
	SendResetLink  bool   `json:"send_reset_link"`
}

type createUserResp struct {
	User      *localUser `json:"user"`
	TenantID  int64      `json:"tenant_id"`
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
	if !requireAdmin(c) {
		return
	}
	var req createUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedhandlers.RespondError(c, pkgtypes.NewError(pkgtypes.ErrBadRequest, "invalid request payload", err))
		return
	}

	email := usernameToEmail(req.Email)
	if email == "" {
		email = usernameToEmail(req.Username)
	}
	password := strings.TrimSpace(req.Password)
	if email == "" || password == "" {
		sharedhandlers.RespondError(c, pkgtypes.NewError(pkgtypes.ErrBadRequest, "email and password required", nil))
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
		sharedhandlers.RespondError(c, pkgtypes.NewError(pkgtypes.ErrBadRequest, "unable to create identity user", err))
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
	if !requireAdmin(c) {
		return
	}
	tenantID, err := sharedhandlers.ParseTenantID(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	rp := newRepo(h.db)
	rows, err := rp.listUsersForTenant(c.Request.Context(), tenantID)
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
	if !requireAdmin(c) {
		return
	}
	var req upsertMembershipReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedhandlers.RespondError(c, pkgtypes.NewError(pkgtypes.ErrBadRequest, "invalid request payload", err))
		return
	}

	email := usernameToEmail(req.Email)
	if email == "" {
		email = usernameToEmail(req.Username)
	}
	if email == "" {
		sharedhandlers.RespondError(c, pkgtypes.NewError(pkgtypes.ErrBadRequest, "email required", nil))
		return
	}

	ctx := c.Request.Context()
	uid, err := h.idp.GetUserUIDByEmail(ctx, email)
	if err != nil {
		sharedhandlers.RespondError(c, pkgtypes.NewError(pkgtypes.ErrBadRequest, "identity user not found", err))
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

