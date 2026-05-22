package admin

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
)

// Repository encapsula todas las consultas de admin contra la DB. Se construye
// desde wire vía NewRepository y se consume a través del puerto en usecases.
type Repository struct {
	db *gorm.DB
}

// NewRepository devuelve un Repository listo para inyectar en UseCases.
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// LocalUser representa la fila de la tabla `users` que mapea al sub de IDP.
type LocalUser struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
	IDPSub   string    `json:"idp_sub"`
	IDPEmail string    `json:"idp_email"`
}

// Tenant es una fila de auth_tenants expuesta hacia afuera.
type Tenant struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// UserMembership es la vista agregada user+tenant+role para listados de admin.
type UserMembership struct {
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
	IDPSub   string    `json:"idp_sub" gorm:"column:idp_sub"`
	TenantID uuid.UUID `json:"tenant_id"`
	Tenant   string    `json:"tenant"`
	Role     string    `json:"role"`
}

// TenantInvite representa una invitación pendiente o aceptada.
type TenantInvite struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   uuid.UUID  `json:"tenant_id"`
	Email      string     `json:"email"`
	RoleID     uuid.UUID  `json:"role_id"`
	Role       string     `json:"role"`
	ExpiresAt  time.Time  `json:"expires_at"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	InvitedBy  *uuid.UUID `json:"invited_by,omitempty"`
	AcceptedBy *uuid.UUID `json:"accepted_by,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// MeMembership es una fila simple de membership para construir /me/context.
type MeMembership struct {
	TenantID uuid.UUID `json:"tenant_id"`
	Name     string    `json:"name"`
	RoleID   uuid.UUID `json:"-"`
	RoleName string    `json:"role"`
}

// RolePermission asocia un role con uno de sus permission names.
type RolePermission struct {
	RoleID uuid.UUID
	Name   string
}

// EnsureLocalUserByIDPSub busca el local user que mapea al idp_sub. Si no
// existe lo inserta con email/username derivados del idp_sub. Idempotente.
func (r *Repository) EnsureLocalUserByIDPSub(ctx context.Context, idpSub, email string) (*LocalUser, error) {
	idpSub = strings.TrimSpace(idpSub)
	email = strings.TrimSpace(email)
	if idpSub == "" {
		return nil, domainerr.Validation("missing idp_sub")
	}

	type row struct {
		ID       uuid.UUID
		Email    string
		Username string
		IDPSub   string `gorm:"column:idp_sub"`
		IDPEmail string `gorm:"column:idp_email"`
	}
	var existing row
	err := r.db.WithContext(ctx).
		Table("users").
		Select("id, email, username, idp_sub, idp_email").
		Where("idp_sub = ?", idpSub).
		Limit(1).
		Take(&existing).Error
	if err == nil {
		return &LocalUser{ID: existing.ID, Email: existing.Email, Username: existing.Username, IDPSub: existing.IDPSub, IDPEmail: existing.IDPEmail}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if email == "" {
		email = idpSub + "@idp.local"
	}
	username := strings.SplitN(email, "@", 2)[0]
	if username == "" {
		username = "idp_" + idpSub
	}

	now := time.Now().UTC()
	insert := map[string]any{
		"email":          email,
		"username":       username,
		"password":       "",
		"token_hash":     "",
		"refresh_tokens": "{}",
		"active":         true,
		"is_verified":    true,
		"idp_sub":        idpSub,
		"idp_email":      email,
		"created_at":     now,
		"updated_at":     now,
	}
	if err := r.db.WithContext(ctx).Table("users").Create(insert).Error; err != nil {
		// concurrent insert -> retry read
		var retry row
		if err2 := r.db.WithContext(ctx).
			Table("users").
			Select("id, email, username, idp_sub, idp_email").
			Where("idp_sub = ?", idpSub).
			Limit(1).
			Take(&retry).Error; err2 == nil {
			return &LocalUser{ID: retry.ID, Email: retry.Email, Username: retry.Username, IDPSub: retry.IDPSub, IDPEmail: retry.IDPEmail}, nil
		}
		return nil, err
	}

	var created row
	if err := r.db.WithContext(ctx).
		Table("users").
		Select("id, email, username, idp_sub, idp_email").
		Where("idp_sub = ?", idpSub).
		Limit(1).
		Take(&created).Error; err != nil {
		return nil, err
	}
	return &LocalUser{ID: created.ID, Email: created.Email, Username: created.Username, IDPSub: created.IDPSub, IDPEmail: created.IDPEmail}, nil
}

// GetLocalUserByIDPSub solo lee — no crea. Devuelve NotFound si no existe.
// Reemplaza la función local currentLocalUserID que el handler usaba antes.
func (r *Repository) GetLocalUserByIDPSub(ctx context.Context, idpSub string) (*LocalUser, error) {
	idpSub = strings.TrimSpace(idpSub)
	if idpSub == "" {
		return nil, domainerr.Validation("missing idp_sub")
	}
	type row struct {
		ID       uuid.UUID
		Email    string
		Username string
		IDPSub   string `gorm:"column:idp_sub"`
		IDPEmail string `gorm:"column:idp_email"`
	}
	var out row
	err := r.db.WithContext(ctx).
		Table("users").
		Select("id, email, username, idp_sub, idp_email").
		Where("idp_sub = ?", idpSub).
		Limit(1).
		Take(&out).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerr.NotFound("local user not found")
		}
		return nil, err
	}
	return &LocalUser{ID: out.ID, Email: out.Email, Username: out.Username, IDPSub: out.IDPSub, IDPEmail: out.IDPEmail}, nil
}

// EnsureTenantByName devuelve el ID del tenant. Si no existe lo crea.
func (r *Repository) EnsureTenantByName(ctx context.Context, name string) (uuid.UUID, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "default"
	}
	type row struct{ ID uuid.UUID }
	var existing row
	err := r.db.WithContext(ctx).
		Table("auth_tenants").
		Select("id").
		Where("name = ?", name).
		Limit(1).
		Take(&existing).Error
	if err == nil {
		return existing.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return uuid.Nil, err
	}

	now := time.Now().UTC()
	if err := r.db.WithContext(ctx).Table("auth_tenants").Create(map[string]any{
		"name":       name,
		"created_at": now,
		"updated_at": now,
	}).Error; err != nil {
		// race -> retry read
		if err2 := r.db.WithContext(ctx).
			Table("auth_tenants").
			Select("id").
			Where("name = ?", name).
			Limit(1).
			Take(&existing).Error; err2 == nil {
			return existing.ID, nil
		}
		return uuid.Nil, err
	}
	if err := r.db.WithContext(ctx).
		Table("auth_tenants").
		Select("id").
		Where("name = ?", name).
		Limit(1).
		Take(&existing).Error; err != nil {
		return uuid.Nil, err
	}
	return existing.ID, nil
}

// RoleIDByName resuelve un role name a id. Soporta aliases legacy (viewer ↔
// tenant_viewer, manager ↔ tenant_manager, admin ↔ tenant_admin) durante el
// período de transición.
func (r *Repository) RoleIDByName(ctx context.Context, name string) (uuid.UUID, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "tenant_viewer"
	}
	candidates := []string{name}
	switch name {
	case "tenant_viewer":
		candidates = append(candidates, "viewer")
	case "tenant_manager":
		candidates = append(candidates, "manager")
	case "tenant_admin":
		candidates = append(candidates, "admin")
	case "viewer":
		candidates = append(candidates, "tenant_viewer")
	case "manager":
		candidates = append(candidates, "tenant_manager")
	case "admin":
		candidates = append(candidates, "tenant_admin")
	}
	type row struct{ ID uuid.UUID }
	for _, candidate := range candidates {
		var out row
		err := r.db.WithContext(ctx).
			Table("auth_roles").
			Select("id").
			Where("name = ?", candidate).
			Limit(1).
			Take(&out).Error
		if err == nil {
			return out.ID, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, err
		}
	}
	return uuid.Nil, domainerr.Validation("role not found")
}

// UpsertMembership inserta o actualiza la membership user×tenant con el rol
// dado. Mantiene status=active y bumpea updated_at.
func (r *Repository) UpsertMembership(ctx context.Context, userID, tenantID, roleID uuid.UUID) error {
	now := time.Now().UTC()
	payload := map[string]any{
		"user_id":    userID,
		"tenant_id":  tenantID,
		"role_id":    roleID,
		"status":     "active",
		"created_at": now,
		"updated_at": now,
	}
	if err := r.db.WithContext(ctx).Table("auth_memberships").Create(payload).Error; err != nil {
		// likely duplicate -> update
		return r.db.WithContext(ctx).
			Table("auth_memberships").
			Where("user_id = ? AND tenant_id = ?", userID, tenantID).
			Updates(map[string]any{
				"role_id":    roleID,
				"status":     "active",
				"updated_at": now,
			}).Error
	}
	return nil
}

// ListTenants devuelve todos los tenants (auth_tenants).
func (r *Repository) ListTenants(ctx context.Context) ([]Tenant, error) {
	var rows []Tenant
	if err := r.db.WithContext(ctx).
		Table("auth_tenants").
		Select("id, name").
		Order("id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// ListUsersForTenant devuelve los usuarios activos de un tenant con su rol.
func (r *Repository) ListUsersForTenant(ctx context.Context, tenantID uuid.UUID) ([]UserMembership, error) {
	var rows []UserMembership
	if err := r.db.WithContext(ctx).
		Table("auth_memberships m").
		Select("u.id as user_id, u.email as email, u.username as username, u.idp_sub as idp_sub, t.id as tenant_id, t.name as tenant, r.name as role").
		Joins("JOIN users u ON u.id = m.user_id").
		Joins("JOIN auth_tenants t ON t.id = m.tenant_id").
		Joins("JOIN auth_roles r ON r.id = m.role_id").
		Where("m.tenant_id = ? AND m.status = 'active'", tenantID).
		Order("u.id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// CreateInvite persiste una invitación con su hash de token.
func (r *Repository) CreateInvite(ctx context.Context, tenantID uuid.UUID, email string, roleID uuid.UUID, tokenHash string, expiresAt time.Time, invitedBy uuid.UUID) (*TenantInvite, error) {
	email = strings.TrimSpace(email)
	if tenantID == uuid.Nil || roleID == uuid.Nil || email == "" || strings.TrimSpace(tokenHash) == "" {
		return nil, domainerr.Validation("invalid invite")
	}
	now := time.Now().UTC()
	payload := map[string]any{
		"tenant_id":  tenantID,
		"email":      email,
		"role_id":    roleID,
		"token_hash": strings.TrimSpace(tokenHash),
		"expires_at": expiresAt.UTC(),
		"created_at": now,
		"updated_at": now,
	}
	if invitedBy != uuid.Nil {
		payload["invited_by"] = invitedBy
	}
	if err := r.db.WithContext(ctx).Table("tenant_invites").Create(payload).Error; err != nil {
		return nil, err
	}
	return r.findInviteByTokenHash(ctx, tokenHash)
}

// AcceptInvite valida y marca la invitación como aceptada por userID.
func (r *Repository) AcceptInvite(ctx context.Context, tokenHash string, userID uuid.UUID) (*TenantInvite, error) {
	tokenHash = strings.TrimSpace(tokenHash)
	if tokenHash == "" || userID == uuid.Nil {
		return nil, domainerr.Validation("invalid invite token")
	}
	now := time.Now().UTC()
	var inviteID uuid.UUID
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		type row struct{ ID uuid.UUID }
		var out row
		if err := tx.
			Table("tenant_invites").
			Select("id").
			Where("token_hash = ? AND accepted_at IS NULL AND revoked_at IS NULL AND expires_at > ?", tokenHash, now).
			Limit(1).
			Take(&out).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.Validation("invite is invalid or expired")
			}
			return err
		}
		inviteID = out.ID
		return tx.Table("tenant_invites").
			Where("id = ?", inviteID).
			Updates(map[string]any{
				"accepted_at": now,
				"accepted_by": userID,
				"updated_at":  now,
			}).Error
	})
	if err != nil {
		return nil, err
	}
	return r.findInviteByID(ctx, inviteID)
}

func (r *Repository) findInviteByTokenHash(ctx context.Context, tokenHash string) (*TenantInvite, error) {
	var out TenantInvite
	if err := r.db.WithContext(ctx).
		Table("tenant_invites i").
		Select("i.id, i.tenant_id, i.email, i.role_id, r.name AS role, i.expires_at, i.accepted_at, i.revoked_at, i.invited_by, i.accepted_by, i.created_at").
		Joins("JOIN auth_roles r ON r.id = i.role_id").
		Where("i.token_hash = ?", strings.TrimSpace(tokenHash)).
		Limit(1).
		Take(&out).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *Repository) findInviteByID(ctx context.Context, id uuid.UUID) (*TenantInvite, error) {
	var out TenantInvite
	if err := r.db.WithContext(ctx).
		Table("tenant_invites i").
		Select("i.id, i.tenant_id, i.email, i.role_id, r.name AS role, i.expires_at, i.accepted_at, i.revoked_at, i.invited_by, i.accepted_by, i.created_at").
		Joins("JOIN auth_roles r ON r.id = i.role_id").
		Where("i.id = ?", id).
		Limit(1).
		Take(&out).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateMembershipRole cambia el rol de un membership preservando el
// invariante de tenant_owner (debe quedar al menos uno activo).
func (r *Repository) UpdateMembershipRole(ctx context.Context, tenantID, membershipID, roleID uuid.UUID) error {
	if tenantID == uuid.Nil || membershipID == uuid.Nil || roleID == uuid.Nil {
		return domainerr.Validation("invalid membership")
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updatedAt := time.Now().UTC()
		if err := r.ensureOwnerInvariantForRoleChange(ctx, tx, tenantID, membershipID, roleID); err != nil {
			return err
		}
		res := tx.Table("auth_memberships").
			Where("id = ? AND tenant_id = ? AND status = 'active'", membershipID, tenantID).
			Updates(map[string]any{"role_id": roleID, "updated_at": updatedAt})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return domainerr.NotFound("membership not found")
		}
		return nil
	})
}

// ArchiveMembership marca status=archived preservando el invariante de owners.
func (r *Repository) ArchiveMembership(ctx context.Context, tenantID, membershipID uuid.UUID) error {
	if tenantID == uuid.Nil || membershipID == uuid.Nil {
		return domainerr.Validation("invalid membership")
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updatedAt := time.Now().UTC()
		if err := r.ensureOwnerInvariantForArchive(ctx, tx, tenantID, membershipID); err != nil {
			return err
		}
		res := tx.Table("auth_memberships").
			Where("id = ? AND tenant_id = ? AND status = 'active'", membershipID, tenantID).
			Updates(map[string]any{"status": "archived", "updated_at": updatedAt})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return domainerr.NotFound("membership not found")
		}
		return nil
	})
}

// ListMembershipsForUser devuelve los tenants activos del usuario con su rol.
// Consumido por GetMeContext.
func (r *Repository) ListMembershipsForUser(ctx context.Context, userID uuid.UUID) ([]MeMembership, error) {
	var rows []MeMembership
	if err := r.db.WithContext(ctx).
		Table("auth_memberships AS m").
		Select("m.tenant_id, t.name, m.role_id, r.name AS role_name").
		Joins("JOIN auth_tenants t ON t.id = m.tenant_id").
		Joins("JOIN auth_roles r ON r.id = m.role_id").
		Where("m.user_id = ? AND m.status = 'active'", userID).
		Order("t.name ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// ListPermissionsByRoleIDs devuelve los permisos planos para los roles dados.
// Consumido por GetMeContext para armar el mapa role→[]permission.
func (r *Repository) ListPermissionsByRoleIDs(ctx context.Context, roleIDs []uuid.UUID) ([]RolePermission, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}
	var rows []RolePermission
	if err := r.db.WithContext(ctx).
		Table("auth_role_permissions rp").
		Select("rp.role_id, p.name").
		Joins("JOIN auth_permissions p ON p.id = rp.permission_id").
		Where("rp.role_id IN ?", roleIDs).
		Order("p.name ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) ensureOwnerInvariantForRoleChange(ctx context.Context, tx *gorm.DB, tenantID, membershipID, nextRoleID uuid.UUID) error {
	currentRole, err := r.membershipRoleName(ctx, tx, tenantID, membershipID)
	if err != nil {
		return err
	}
	nextRole, err := r.roleNameByID(ctx, tx, nextRoleID)
	if err != nil {
		return err
	}
	if currentRole == "tenant_owner" && nextRole != "tenant_owner" {
		return r.requireAnotherActiveOwner(ctx, tx, tenantID, membershipID)
	}
	return nil
}

func (r *Repository) ensureOwnerInvariantForArchive(ctx context.Context, tx *gorm.DB, tenantID, membershipID uuid.UUID) error {
	currentRole, err := r.membershipRoleName(ctx, tx, tenantID, membershipID)
	if err != nil {
		return err
	}
	if currentRole == "tenant_owner" {
		return r.requireAnotherActiveOwner(ctx, tx, tenantID, membershipID)
	}
	return nil
}

func (r *Repository) membershipRoleName(ctx context.Context, tx *gorm.DB, tenantID, membershipID uuid.UUID) (string, error) {
	type row struct{ Name string }
	var out row
	err := tx.WithContext(ctx).
		Table("auth_memberships m").
		Select("r.name").
		Joins("JOIN auth_roles r ON r.id = m.role_id").
		Where("m.id = ? AND m.tenant_id = ? AND m.status = 'active'", membershipID, tenantID).
		Limit(1).
		Take(&out).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", domainerr.NotFound("membership not found")
	}
	if err != nil {
		return "", err
	}
	return out.Name, nil
}

func (r *Repository) roleNameByID(ctx context.Context, tx *gorm.DB, roleID uuid.UUID) (string, error) {
	type row struct{ Name string }
	var out row
	err := tx.WithContext(ctx).
		Table("auth_roles").
		Select("name").
		Where("id = ?", roleID).
		Limit(1).
		Take(&out).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", domainerr.Validation("role not found")
	}
	if err != nil {
		return "", err
	}
	return out.Name, nil
}

func (r *Repository) requireAnotherActiveOwner(ctx context.Context, tx *gorm.DB, tenantID, excludedMembershipID uuid.UUID) error {
	var count int64
	if err := tx.WithContext(ctx).
		Table("auth_memberships m").
		Joins("JOIN auth_roles r ON r.id = m.role_id").
		Where("m.tenant_id = ? AND m.status = 'active' AND r.name = 'tenant_owner' AND m.id <> ?", tenantID, excludedMembershipID).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return domainerr.Validation("tenant must keep at least one active owner")
	}
	return nil
}
