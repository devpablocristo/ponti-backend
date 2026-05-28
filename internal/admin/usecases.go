package admin

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/devpablocristo/platform/errors/go/domainerr"

	"github.com/devpablocristo/ponti-backend/internal/admin/idp"
)

// RepositoryPort define lo que UseCases necesita del Repository. Permite
// mockear en tests sin levantar DB real.
type RepositoryPort interface {
	EnsureLocalUserByIDPSub(ctx context.Context, idpSub, email string) (*LocalUser, error)
	GetLocalUserByIDPSub(ctx context.Context, idpSub string) (*LocalUser, error)
	EnsureTenantByName(ctx context.Context, name string) (uuid.UUID, error)
	RoleIDByName(ctx context.Context, name string) (uuid.UUID, error)
	UpsertMembership(ctx context.Context, userID, tenantID, roleID uuid.UUID) error
	ListTenants(ctx context.Context) ([]Tenant, error)
	ListUsersForTenant(ctx context.Context, tenantID uuid.UUID) ([]UserMembership, error)
	CreateInvite(ctx context.Context, tenantID uuid.UUID, email string, roleID uuid.UUID, tokenHash string, expiresAt time.Time, invitedBy uuid.UUID) (*TenantInvite, error)
	AcceptInvite(ctx context.Context, tokenHash string, userID uuid.UUID) (*TenantInvite, error)
	UpdateMembershipRole(ctx context.Context, tenantID, membershipID, roleID uuid.UUID) error
	ArchiveMembership(ctx context.Context, tenantID, membershipID uuid.UUID) error
	ListMembershipsForUser(ctx context.Context, userID uuid.UUID) ([]MeMembership, error)
	ListPermissionsByRoleIDs(ctx context.Context, roleIDs []uuid.UUID) ([]RolePermission, error)
}

// IDPClient es el subset del cliente de Identity Platform que UseCases usa.
// Se alias-ea idp.AdminClient para evitar leak del nombre completo al port.
type IDPClient = idp.AdminClient

// UseCases agrupa la lógica de admin que orquesta repository + IDP. El
// handler HTTP es delgado y solo mapea request/response.
type UseCases struct {
	repo RepositoryPort
	idp  IDPClient
}

// NewUseCases construye un UseCases. Wire pasa el Repository concreto y el
// AdminClient del IDP.
func NewUseCases(repo RepositoryPort, idpAdmin IDPClient) *UseCases {
	return &UseCases{repo: repo, idp: idpAdmin}
}

// CreateUserInput es el payload normalizado para CreateUser.
type CreateUserInput struct {
	Email         string
	Username      string
	Password      string
	TenantName    string
	RoleName      string
	SendResetLink bool
}

// CreateUserOutput es la respuesta de CreateUser.
type CreateUserOutput struct {
	User      *LocalUser `json:"user"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	RoleName  string     `json:"role_name"`
	ResetLink string     `json:"reset_link,omitempty"`
}

// ListTenants devuelve todos los tenants registrados.
func (uc *UseCases) ListTenants(ctx context.Context) ([]Tenant, error) {
	return uc.repo.ListTenants(ctx)
}

// CreateTenant inserta el tenant (idempotente por nombre) y devuelve su ID.
func (uc *UseCases) CreateTenant(ctx context.Context, name string) (uuid.UUID, error) {
	return uc.repo.EnsureTenantByName(ctx, name)
}

// CreateUser crea (o resuelve) el usuario en IDP, lo sincroniza con la tabla
// local users, y le crea membership en el tenant con el rol indicado.
func (uc *UseCases) CreateUser(ctx context.Context, in CreateUserInput) (*CreateUserOutput, error) {
	email := UsernameToEmail(in.Email)
	if email == "" {
		email = UsernameToEmail(in.Username)
	}
	password := strings.TrimSpace(in.Password)
	if email == "" || password == "" {
		return nil, domainerr.Validation("email and password required")
	}

	uid, err := uc.idp.CreateUser(ctx, email, password)
	if err != nil {
		// Si ya existe, intentamos lookup por email para attachar membership.
		if strings.Contains(strings.ToLower(err.Error()), "email") && strings.Contains(strings.ToLower(err.Error()), "exists") {
			uid, err = uc.idp.GetUserUIDByEmail(ctx, email)
		}
	}
	if err != nil {
		return nil, domainerr.Validation("unable to create identity user")
	}

	u, err := uc.repo.EnsureLocalUserByIDPSub(ctx, uid, email)
	if err != nil {
		return nil, err
	}
	tenantID, err := uc.repo.EnsureTenantByName(ctx, in.TenantName)
	if err != nil {
		return nil, err
	}
	roleID, err := uc.repo.RoleIDByName(ctx, in.RoleName)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.UpsertMembership(ctx, u.ID, tenantID, roleID); err != nil {
		return nil, err
	}

	roleName := strings.TrimSpace(in.RoleName)
	if roleName == "" {
		roleName = "viewer"
	}
	out := &CreateUserOutput{
		User:     u,
		TenantID: tenantID,
		RoleName: roleName,
	}
	if in.SendResetLink {
		if link, linkErr := uc.idp.GeneratePasswordResetLink(ctx, email); linkErr == nil {
			out.ResetLink = link
		}
	}
	return out, nil
}

// ListUsers devuelve los usuarios activos del tenant.
func (uc *UseCases) ListUsers(ctx context.Context, tenantID uuid.UUID) ([]UserMembership, error) {
	return uc.repo.ListUsersForTenant(ctx, tenantID)
}

// UpsertMembershipInput es el payload de UpsertMembership.
type UpsertMembershipInput struct {
	Email      string
	Username   string
	TenantName string
	RoleName   string
}

// UpsertMembership resuelve el usuario por email en IDP, lo asegura en la
// tabla local, y le crea/actualiza membership en el tenant.
func (uc *UseCases) UpsertMembership(ctx context.Context, in UpsertMembershipInput) (userID uuid.UUID, tenantID uuid.UUID, err error) {
	email := UsernameToEmail(in.Email)
	if email == "" {
		email = UsernameToEmail(in.Username)
	}
	if email == "" {
		return uuid.Nil, uuid.Nil, domainerr.Validation("email required")
	}
	uid, err := uc.idp.GetUserUIDByEmail(ctx, email)
	if err != nil {
		return uuid.Nil, uuid.Nil, domainerr.Validation("identity user not found")
	}
	u, err := uc.repo.EnsureLocalUserByIDPSub(ctx, uid, email)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	tenantID, err = uc.repo.EnsureTenantByName(ctx, in.TenantName)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	roleID, err := uc.repo.RoleIDByName(ctx, in.RoleName)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	if err := uc.repo.UpsertMembership(ctx, u.ID, tenantID, roleID); err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return u.ID, tenantID, nil
}

// CreateInviteInput es el payload de CreateInvite.
type CreateInviteInput struct {
	TenantID  uuid.UUID
	Email     string
	RoleName  string
	ExpiresIn string // formato time.ParseDuration; default 7 días
	ActorSub  string // idp_sub del invitador (para audit trail)
}

// CreateInviteOutput devuelve la invitación + el token plano (solo se ve aquí).
type CreateInviteOutput struct {
	Invite *TenantInvite `json:"invite"`
	Token  string        `json:"token"`
}

// CreateInvite genera el token, lo hashea, y persiste la invitación. El
// token plano se devuelve solo en este momento (se entrega al usuario por
// canal seguro).
func (uc *UseCases) CreateInvite(ctx context.Context, in CreateInviteInput) (*CreateInviteOutput, error) {
	if in.TenantID == uuid.Nil {
		return nil, domainerr.Validation("invalid tenant_id")
	}
	email := UsernameToEmail(in.Email)
	if email == "" {
		return nil, domainerr.Validation("email required")
	}
	roleName := strings.TrimSpace(in.RoleName)
	if roleName == "" {
		roleName = "tenant_viewer"
	}
	now := time.Now().UTC()
	expiresAt := now.Add(7 * 24 * time.Hour)
	if in.ExpiresIn != "" {
		if d, err := time.ParseDuration(in.ExpiresIn); err == nil && d > 0 {
			expiresAt = now.Add(d)
		}
	}
	token, err := newInviteToken()
	if err != nil {
		return nil, domainerr.Internal("unable to create invite token")
	}
	roleID, err := uc.repo.RoleIDByName(ctx, roleName)
	if err != nil {
		return nil, err
	}

	// El invitedBy es opcional; si no podemos resolverlo seguimos sin él.
	var invitedBy uuid.UUID
	if actor := strings.TrimSpace(in.ActorSub); actor != "" {
		if u, lookupErr := uc.repo.GetLocalUserByIDPSub(ctx, actor); lookupErr == nil {
			invitedBy = u.ID
		}
	}

	invite, err := uc.repo.CreateInvite(ctx, in.TenantID, email, roleID, hashInviteToken(token), expiresAt, invitedBy)
	if err != nil {
		return nil, err
	}
	return &CreateInviteOutput{Invite: invite, Token: token}, nil
}

// AcceptInvite valida el token y crea/actualiza la membership del invitado.
func (uc *UseCases) AcceptInvite(ctx context.Context, token, actorSub string) (*TenantInvite, error) {
	actor := strings.TrimSpace(actorSub)
	if actor == "" {
		return nil, domainerr.Unauthorized("authentication context required")
	}
	u, err := uc.repo.GetLocalUserByIDPSub(ctx, actor)
	if err != nil {
		return nil, domainerr.Forbidden("local user not found")
	}
	invite, err := uc.repo.AcceptInvite(ctx, hashInviteToken(token), u.ID)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.UpsertMembership(ctx, u.ID, invite.TenantID, invite.RoleID); err != nil {
		return nil, err
	}
	return invite, nil
}

// UpdateMembershipRole resuelve role name → id y actualiza la membership.
func (uc *UseCases) UpdateMembershipRole(ctx context.Context, tenantID, membershipID uuid.UUID, roleName string) error {
	roleID, err := uc.repo.RoleIDByName(ctx, roleName)
	if err != nil {
		return err
	}
	return uc.repo.UpdateMembershipRole(ctx, tenantID, membershipID, roleID)
}

// ArchiveMembership desactiva la membership manteniendo el invariante de
// tenant_owner.
func (uc *UseCases) ArchiveMembership(ctx context.Context, tenantID, membershipID uuid.UUID) error {
	return uc.repo.ArchiveMembership(ctx, tenantID, membershipID)
}

// MeContext es la respuesta agregada de /me/context: usuario + tenants con
// sus roles + permisos. Diseñada para el bootstrap del FE.
type MeContext struct {
	User            MeUser     `json:"user"`
	CurrentTenantID uuid.UUID  `json:"current_tenant_id"`
	Tenants         []MeTenant `json:"tenants"`
}

// MeUser es la sección "user" del MeContext.
type MeUser struct {
	ID       uuid.UUID `json:"id"`
	IDPSub   string    `json:"idp_sub"`
	IDPEmail string    `json:"idp_email"`
	Email    string    `json:"email"`
}

// MeTenant es cada elemento del array "tenants" en MeContext.
type MeTenant struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Role        string    `json:"role"`
	Permissions []string  `json:"permissions"`
	IsCurrent   bool      `json:"is_current"`
}

// GetMeContext compone la respuesta /me/context para el usuario autenticado.
// actorSub viene del context (ctxkeys.Actor) — es el idp_sub. currentTenantID
// viene de ctxkeys.OrgID y se usa para marcar el tenant activo.
func (uc *UseCases) GetMeContext(ctx context.Context, actorSub string, currentTenantID uuid.UUID) (*MeContext, error) {
	actor := strings.TrimSpace(actorSub)
	if actor == "" {
		return nil, domainerr.Unauthorized("authentication context required")
	}
	user, err := uc.repo.GetLocalUserByIDPSub(ctx, actor)
	if err != nil {
		return nil, domainerr.Forbidden("local user not found")
	}
	memberships, err := uc.repo.ListMembershipsForUser(ctx, user.ID)
	if err != nil {
		return nil, domainerr.Internal("unable to load memberships")
	}
	roleIDs := make([]uuid.UUID, 0, len(memberships))
	for _, m := range memberships {
		roleIDs = append(roleIDs, m.RoleID)
	}
	perms, err := uc.repo.ListPermissionsByRoleIDs(ctx, roleIDs)
	if err != nil {
		return nil, domainerr.Internal("unable to load permissions")
	}
	permsByRole := map[uuid.UUID][]string{}
	for _, p := range perms {
		permsByRole[p.RoleID] = append(permsByRole[p.RoleID], p.Name)
	}
	tenants := make([]MeTenant, 0, len(memberships))
	for _, m := range memberships {
		tenants = append(tenants, MeTenant{
			ID:          m.TenantID,
			Name:        m.Name,
			Role:        m.RoleName,
			Permissions: permsByRole[m.RoleID],
			IsCurrent:   m.TenantID == currentTenantID,
		})
	}
	return &MeContext{
		User: MeUser{
			ID:       user.ID,
			IDPSub:   user.IDPSub,
			IDPEmail: user.IDPEmail,
			Email:    user.Email,
		},
		CurrentTenantID: currentTenantID,
		Tenants:         tenants,
	}, nil
}

// UsernameToEmail normaliza un username/email — si ya es email lo deja, si
// no le sufija el dominio local del IDP. Exportado por si el handler quiere
// usarlo para preview/validation (hoy lo usa solo el usecase).
func UsernameToEmail(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	if strings.Contains(v, "@") {
		return v
	}
	return v + "@ponti.local"
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
