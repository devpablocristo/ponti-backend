package admin

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/devpablocristo/core/errors/go/domainerr"
)

type repo struct {
	db *gorm.DB
}

func newRepo(db *gorm.DB) *repo { return &repo{db: db} }

type localUser struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
	IDPSub   string    `json:"idp_sub"`
}

func (r *repo) ensureLocalUserByIDPSub(ctx context.Context, idpSub, email string) (*localUser, error) {
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
	}
	var existing row
	err := r.db.WithContext(ctx).
		Table("users").
		Select("id, email, username, idp_sub").
		Where("idp_sub = ?", idpSub).
		Limit(1).
		Take(&existing).Error
	if err == nil {
		return &localUser{ID: existing.ID, Email: existing.Email, Username: existing.Username, IDPSub: existing.IDPSub}, nil
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
			Select("id, email, username, idp_sub").
			Where("idp_sub = ?", idpSub).
			Limit(1).
			Take(&retry).Error; err2 == nil {
			return &localUser{ID: retry.ID, Email: retry.Email, Username: retry.Username, IDPSub: retry.IDPSub}, nil
		}
		return nil, err
	}

	var created row
	if err := r.db.WithContext(ctx).
		Table("users").
		Select("id, email, username, idp_sub").
		Where("idp_sub = ?", idpSub).
		Limit(1).
		Take(&created).Error; err != nil {
		return nil, err
	}
	return &localUser{ID: created.ID, Email: created.Email, Username: created.Username, IDPSub: created.IDPSub}, nil
}

func (r *repo) ensureTenantByName(ctx context.Context, name string) (uuid.UUID, error) {
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

func (r *repo) roleIDByName(ctx context.Context, name string) (uuid.UUID, error) {
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

func (r *repo) upsertMembership(ctx context.Context, userID, tenantID, roleID uuid.UUID) error {
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

type tenantDTO struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func (r *repo) listTenants(ctx context.Context) ([]tenantDTO, error) {
	var rows []tenantDTO
	if err := r.db.WithContext(ctx).
		Table("auth_tenants").
		Select("id, name").
		Order("id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

type userMembershipDTO struct {
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
	IDPSub   string    `json:"idp_sub" gorm:"column:idp_sub"`
	TenantID uuid.UUID `json:"tenant_id"`
	Tenant   string    `json:"tenant"`
	Role     string    `json:"role"`
}

func (r *repo) listUsersForTenant(ctx context.Context, tenantID uuid.UUID) ([]userMembershipDTO, error) {
	var rows []userMembershipDTO
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

type tenantInviteDTO struct {
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

func (r *repo) createInvite(ctx context.Context, tenantID uuid.UUID, email string, roleID uuid.UUID, tokenHash string, expiresAt time.Time, invitedBy uuid.UUID) (*tenantInviteDTO, error) {
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

func (r *repo) acceptInvite(ctx context.Context, tokenHash string, userID uuid.UUID) (*tenantInviteDTO, error) {
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

func (r *repo) findInviteByTokenHash(ctx context.Context, tokenHash string) (*tenantInviteDTO, error) {
	var out tenantInviteDTO
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

func (r *repo) findInviteByID(ctx context.Context, id uuid.UUID) (*tenantInviteDTO, error) {
	var out tenantInviteDTO
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

func (r *repo) updateMembershipRole(ctx context.Context, tenantID, membershipID, roleID uuid.UUID) error {
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

func (r *repo) archiveMembership(ctx context.Context, tenantID, membershipID uuid.UUID) error {
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

func (r *repo) ensureOwnerInvariantForRoleChange(ctx context.Context, tx *gorm.DB, tenantID, membershipID, nextRoleID uuid.UUID) error {
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

func (r *repo) ensureOwnerInvariantForArchive(ctx context.Context, tx *gorm.DB, tenantID, membershipID uuid.UUID) error {
	currentRole, err := r.membershipRoleName(ctx, tx, tenantID, membershipID)
	if err != nil {
		return err
	}
	if currentRole == "tenant_owner" {
		return r.requireAnotherActiveOwner(ctx, tx, tenantID, membershipID)
	}
	return nil
}

func (r *repo) membershipRoleName(ctx context.Context, tx *gorm.DB, tenantID, membershipID uuid.UUID) (string, error) {
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

func (r *repo) roleNameByID(ctx context.Context, tx *gorm.DB, roleID uuid.UUID) (string, error) {
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

func (r *repo) requireAnotherActiveOwner(ctx context.Context, tx *gorm.DB, tenantID, excludedMembershipID uuid.UUID) error {
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
