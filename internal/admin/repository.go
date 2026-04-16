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
		name = "viewer"
	}
	type row struct{ ID uuid.UUID }
	var out row
	if err := r.db.WithContext(ctx).
		Table("auth_roles").
		Select("id").
		Where("name = ?", name).
		Limit(1).
		Take(&out).Error; err != nil {
		return uuid.Nil, err
	}
	return out.ID, nil
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
