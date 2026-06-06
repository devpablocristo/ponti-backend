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

type repo struct {
	db *gorm.DB
}

func newRepo(db *gorm.DB) *repo { return &repo{db: db} }

// === U3: /me/context ===

type meTenant struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Role        string    `json:"role"`
	Permissions []string  `json:"permissions"`
	IsCurrent   bool      `json:"is_current"`
}

type meContext struct {
	User            *localUser `json:"user"`
	CurrentTenantID *uuid.UUID `json:"current_tenant_id"`
	Tenants         []meTenant `json:"tenants"`
}

// getMeContext arma el contexto de identidad: el usuario + sus tenants activos
// (con rol y permisos) + el tenant actual resuelto.
func (r *repo) getMeContext(ctx context.Context, idpSub, requestedTenant string) (*meContext, error) {
	user, err := r.ensureLocalUserByIDPSub(ctx, idpSub, "")
	if err != nil {
		return nil, err
	}

	type mrow struct {
		TenantID uuid.UUID
		Name     string
		RoleID   uuid.UUID
		RoleName string
	}
	var rows []mrow
	if err := r.db.WithContext(ctx).
		Table("auth_memberships AS m").
		Select("m.tenant_id AS tenant_id, t.name AS name, m.role_id AS role_id, r.name AS role_name").
		Joins("JOIN auth_tenants t ON t.id = m.tenant_id").
		Joins("JOIN auth_roles r ON r.id = m.role_id").
		Where("m.user_id = ? AND m.status = 'active'", user.ID).
		Where("t.deleted_at IS NULL").
		Order("t.name ASC").
		Find(&rows).Error; err != nil {
		return nil, domainerr.Internal("failed to list memberships")
	}

	// Tenant actual: header X-Tenant-Id si es una membership válida; si no y hay
	// exactamente una membership, esa; si no, null (el FE debe elegir).
	var current *uuid.UUID
	if requestedTenant != "" {
		if reqID, perr := uuid.Parse(requestedTenant); perr == nil {
			for _, row := range rows {
				if row.TenantID == reqID {
					id := reqID
					current = &id
					break
				}
			}
		}
	}
	if current == nil && len(rows) == 1 {
		id := rows[0].TenantID
		current = &id
	}

	tenants := make([]meTenant, 0, len(rows))
	for _, row := range rows {
		perms, err := r.listRolePermissions(ctx, row.RoleID)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, meTenant{
			ID:          row.TenantID,
			Name:        row.Name,
			Role:        row.RoleName,
			Permissions: perms,
			IsCurrent:   current != nil && *current == row.TenantID,
		})
	}

	return &meContext{User: user, CurrentTenantID: current, Tenants: tenants}, nil
}

// isPlatformAdminBySub devuelve users.is_platform_admin para el idp_sub (U1).
// false si el usuario no existe o no es platform-admin.
func (r *repo) isPlatformAdminBySub(ctx context.Context, idpSub string) bool {
	type row struct {
		IsPlatformAdmin bool `gorm:"column:is_platform_admin"`
	}
	var rr row
	if err := r.db.WithContext(ctx).
		Table("users").
		Select("is_platform_admin").
		Where("idp_sub = ?", idpSub).
		Limit(1).
		Take(&rr).Error; err != nil {
		return false
	}
	return rr.IsPlatformAdmin
}

func (r *repo) listRolePermissions(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	type prow struct{ Name string }
	var prows []prow
	if err := r.db.WithContext(ctx).
		Table("auth_role_permissions rp").
		Select("p.name").
		Joins("JOIN auth_permissions p ON p.id = rp.permission_id").
		Where("rp.role_id = ?", roleID).
		Order("p.name ASC").
		Find(&prows).Error; err != nil {
		return nil, domainerr.Internal("failed to list role permissions")
	}
	out := make([]string, 0, len(prows))
	for _, p := range prows {
		out = append(out, p.Name)
	}
	return out, nil
}

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

// --- Tenant lifecycle (CRUDAR platform-admin, PARTE IV) ---

type tenantDetail struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

func (r *repo) getTenant(ctx context.Context, id uuid.UUID) (*tenantDetail, error) {
	var t tenantDetail
	if err := r.db.WithContext(ctx).
		Table("auth_tenants").
		Select("id, name, status, created_at, updated_at, deleted_at").
		Where("id = ?", id).
		Limit(1).
		Take(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerr.New(domainerr.KindNotFound, "tenant not found")
		}
		return nil, domainerr.Internal("failed to get tenant")
	}
	return &t, nil
}

func (r *repo) listTenantsByArchived(ctx context.Context, archived bool) ([]tenantDetail, error) {
	q := r.db.WithContext(ctx).
		Table("auth_tenants").
		Select("id, name, status, created_at, updated_at, deleted_at")
	if archived {
		q = q.Where("deleted_at IS NOT NULL")
	} else {
		q = q.Where("deleted_at IS NULL")
	}
	var rows []tenantDetail
	if err := q.Order("name ASC").Find(&rows).Error; err != nil {
		return nil, domainerr.Internal("failed to list tenants")
	}
	return rows, nil
}

func (r *repo) updateTenantName(ctx context.Context, id uuid.UUID, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return domainerr.Validation("tenant name required")
	}
	res := r.db.WithContext(ctx).Table("auth_tenants").
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{"name": name, "updated_at": time.Now().UTC()})
	if res.Error != nil {
		return domainerr.Internal("failed to update tenant")
	}
	if res.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, "tenant not found")
	}
	return nil
}

func (r *repo) setTenantStatus(ctx context.Context, id uuid.UUID, status string) error {
	res := r.db.WithContext(ctx).Table("auth_tenants").
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{"status": status, "updated_at": time.Now().UTC()})
	if res.Error != nil {
		return domainerr.Internal("failed to update tenant status")
	}
	if res.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, "tenant not found")
	}
	return nil
}

func (r *repo) archiveTenant(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).Table("auth_tenants").
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{"deleted_at": now, "updated_at": now})
	if res.Error != nil {
		return domainerr.Internal("failed to archive tenant")
	}
	if res.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, "tenant not found or already archived")
	}
	return nil
}

func (r *repo) restoreTenant(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Table("auth_tenants").
		Where("id = ? AND deleted_at IS NOT NULL", id).
		Updates(map[string]any{"deleted_at": nil, "updated_at": time.Now().UTC()})
	if res.Error != nil {
		return domainerr.Internal("failed to restore tenant")
	}
	if res.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, "tenant not found or not archived")
	}
	return nil
}

// hardDeleteTenant borra físicamente un tenant. Requiere que esté archivado
// primero (offboard tras retención). Si quedan FKs (memberships), la DB rechaza.
func (r *repo) hardDeleteTenant(ctx context.Context, id uuid.UUID) error {
	t, err := r.getTenant(ctx, id)
	if err != nil {
		return err
	}
	if t.DeletedAt == nil {
		return domainerr.Conflict("tenant must be archived before hard delete")
	}
	if err := r.db.WithContext(ctx).Exec("DELETE FROM auth_tenants WHERE id = ?", id).Error; err != nil {
		return domainerr.Conflict("cannot hard delete tenant (still referenced); remove memberships/data first")
	}
	return nil
}
