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

type inviteDTO struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedBy *string   `json:"created_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type inviteResolved struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	RoleID    uuid.UUID
	Email     string
	Status    string
	ExpiresAt time.Time
}

func (r *repo) createInvite(ctx context.Context, tenantID, roleID uuid.UUID, email, tokenHash, createdBy string, expiresAt time.Time) (uuid.UUID, error) {
	id := uuid.New()
	now := time.Now().UTC()
	payload := map[string]any{
		"id":         id,
		"tenant_id":  tenantID,
		"email":      strings.ToLower(strings.TrimSpace(email)),
		"role_id":    roleID,
		"token_hash": tokenHash,
		"status":     "pending",
		"expires_at": expiresAt,
		"created_by": createdBy,
		"created_at": now,
		"updated_at": now,
	}
	if err := r.db.WithContext(ctx).Table("tenant_invites").Create(payload).Error; err != nil {
		return uuid.Nil, domainerr.Internal("failed to create invite")
	}
	return id, nil
}

func (r *repo) listInvites(ctx context.Context, tenantID uuid.UUID) ([]inviteDTO, error) {
	var rows []inviteDTO
	if err := r.db.WithContext(ctx).
		Table("tenant_invites ti").
		Select("ti.id, ti.tenant_id, ti.email, r.name AS role, ti.status, ti.expires_at, ti.created_by, ti.created_at").
		Joins("JOIN auth_roles r ON r.id = ti.role_id").
		Where("ti.tenant_id = ?", tenantID).
		Order("ti.created_at DESC").
		Find(&rows).Error; err != nil {
		return nil, domainerr.Internal("failed to list invites")
	}
	return rows, nil
}

func (r *repo) getInviteByTokenHash(ctx context.Context, tokenHash string) (*inviteResolved, error) {
	var x inviteResolved
	if err := r.db.WithContext(ctx).
		Table("tenant_invites").
		Select("id, tenant_id, role_id, email, status, expires_at").
		Where("token_hash = ?", tokenHash).
		Limit(1).
		Take(&x).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerr.NotFound("invite not found")
		}
		return nil, domainerr.Internal("failed to load invite")
	}
	return &x, nil
}

func (r *repo) inviteTenantByID(ctx context.Context, inviteID uuid.UUID) (uuid.UUID, error) {
	type row struct {
		TenantID uuid.UUID
	}
	var x row
	if err := r.db.WithContext(ctx).
		Table("tenant_invites").
		Select("tenant_id").
		Where("id = ?", inviteID).
		Limit(1).
		Take(&x).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, domainerr.NotFound("invite not found")
		}
		return uuid.Nil, domainerr.Internal("failed to load invite")
	}
	return x.TenantID, nil
}

func (r *repo) revokeInvite(ctx context.Context, inviteID, tenantID uuid.UUID) error {
	res := r.db.WithContext(ctx).
		Table("tenant_invites").
		Where("id = ? AND tenant_id = ? AND status = 'pending'", inviteID, tenantID).
		Updates(map[string]any{"status": "revoked", "updated_at": time.Now().UTC()})
	if res.Error != nil {
		return domainerr.Internal("failed to revoke invite")
	}
	if res.RowsAffected == 0 {
		return domainerr.NotFound("pending invite not found")
	}
	return nil
}

// acceptInvite marca el invite como aceptado y crea la membership, atómicamente.
func (r *repo) acceptInvite(ctx context.Context, inv *inviteResolved, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		res := tx.Table("tenant_invites").
			Where("id = ? AND status = 'pending'", inv.ID).
			Updates(map[string]any{
				"status":      "accepted",
				"accepted_by": userID,
				"accepted_at": now,
				"updated_at":  now,
			})
		if res.Error != nil {
			return domainerr.Internal("failed to accept invite")
		}
		if res.RowsAffected == 0 {
			return domainerr.Conflict("invite is not pending")
		}
		return upsertMembershipTx(tx, userID, inv.TenantID, inv.RoleID)
	})
}

// upsertMembershipTx es la versión tx de upsertMembership (insert-or-update).
func upsertMembershipTx(tx *gorm.DB, userID, tenantID, roleID uuid.UUID) error {
	now := time.Now().UTC()
	payload := map[string]any{
		"user_id":    userID,
		"tenant_id":  tenantID,
		"role_id":    roleID,
		"status":     "active",
		"created_at": now,
		"updated_at": now,
	}
	if err := tx.Table("auth_memberships").Create(payload).Error; err != nil {
		return tx.Table("auth_memberships").
			Where("user_id = ? AND tenant_id = ?", userID, tenantID).
			Updates(map[string]any{
				"role_id":    roleID,
				"status":     "active",
				"updated_at": now,
			}).Error
	}
	return nil
}
