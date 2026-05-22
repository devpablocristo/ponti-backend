// Package admin exposes admin endpoints to manage Identity Platform users and RBAC memberships.
package admin

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/devpablocristo/platform/errors/go/domainerr"

	"github.com/devpablocristo/ponti-backend/internal/shared/authz"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
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
}

// UseCasesPort es el contrato que el Handler consume. Lo implementa
// *UseCases pero queda como interface para facilitar tests con mocks.
type UseCasesPort interface {
	ListTenants(ctx context.Context) ([]Tenant, error)
	CreateTenant(ctx context.Context, name string) (uuid.UUID, error)
	CreateUser(ctx context.Context, in CreateUserInput) (*CreateUserOutput, error)
	ListUsers(ctx context.Context, tenantID uuid.UUID) ([]UserMembership, error)
	UpsertMembership(ctx context.Context, in UpsertMembershipInput) (userID uuid.UUID, tenantID uuid.UUID, err error)
	CreateInvite(ctx context.Context, in CreateInviteInput) (*CreateInviteOutput, error)
	AcceptInvite(ctx context.Context, token, actorSub string) (*TenantInvite, error)
	UpdateMembershipRole(ctx context.Context, tenantID, membershipID uuid.UUID, roleName string) error
	ArchiveMembership(ctx context.Context, tenantID, membershipID uuid.UUID) error
	GetMeContext(ctx context.Context, actorSub string, currentTenantID uuid.UUID) (*MeContext, error)
}

// Handler delgada: solo mapea HTTP request/response y delega en UseCases.
type Handler struct {
	uc  UseCasesPort
	gsv GinEnginePort
	acf ConfigAPIPort
	mws MiddlewaresEnginePort
}

func NewHandler(uc UseCasesPort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *Handler {
	return &Handler{uc: uc, gsv: s, acf: c, mws: m}
}

func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/admin"

	h.registerMeContextRoute()

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

func requireAdminPermission(c *gin.Context, permission string) bool {
	if authz.HasPermission(c.Request.Context(), permission) {
		return true
	}
	sharedhandlers.RespondError(c, domainerr.Forbidden("insufficient permissions"))
	return false
}

type createTenantReq struct {
	Name string `json:"name"`
}

func (h *Handler) ListTenants(c *gin.Context) {
	if !requireAdminPermission(c, authz.PermissionAdminTenants) {
		return
	}
	items, err := h.uc.ListTenants(c.Request.Context())
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, items)
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
	id, err := h.uc.CreateTenant(c.Request.Context(), req.Name)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, gin.H{"id": id})
}

type createUserReq struct {
	Email         string `json:"email"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	TenantName    string `json:"tenant_name"`
	RoleName      string `json:"role_name"`
	SendResetLink bool   `json:"send_reset_link"`
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
	out, err := h.uc.CreateUser(c.Request.Context(), CreateUserInput{
		Email:         req.Email,
		Username:      req.Username,
		Password:      req.Password,
		TenantName:    req.TenantName,
		RoleName:      req.RoleName,
		SendResetLink: req.SendResetLink,
	})
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, out)
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
	rows, err := h.uc.ListUsers(c.Request.Context(), orgID)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, rows)
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
	userID, tenantID, err := h.uc.UpsertMembership(c.Request.Context(), UpsertMembershipInput{
		Email:      req.Email,
		Username:   req.Username,
		TenantName: req.TenantName,
		RoleName:   req.RoleName,
	})
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, gin.H{"user_id": userID, "tenant_id": tenantID})
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
	actorSub, _ := sharedhandlers.ParseActor(c)
	out, err := h.uc.CreateInvite(c.Request.Context(), CreateInviteInput{
		TenantID:  tenantID,
		Email:     req.Email,
		RoleName:  req.RoleName,
		ExpiresIn: req.ExpiresIn,
		ActorSub:  actorSub,
	})
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, out)
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
	actorSub, err := sharedhandlers.ParseActor(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	invite, err := h.uc.AcceptInvite(c.Request.Context(), req.Token, actorSub)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, invite)
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
	if err := h.uc.UpdateMembershipRole(c.Request.Context(), tenantID, membershipID, req.RoleName); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) ArchiveMembership(c *gin.Context) {
	if !requireAdminPermission(c, authz.PermissionAdminMemberships) {
		return
	}
	tenantID, membershipID, ok := parseTenantAndMembership(c)
	if !ok {
		return
	}
	if err := h.uc.ArchiveMembership(c.Request.Context(), tenantID, membershipID); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
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
