package admin

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newMeContextTestRepo(t *testing.T) (*repo, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	for _, s := range []string{
		`CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT, username TEXT, idp_sub TEXT, is_platform_admin INTEGER NOT NULL DEFAULT 0);`,
		`CREATE TABLE auth_tenants (id TEXT PRIMARY KEY, name TEXT NOT NULL, status TEXT DEFAULT 'active', deleted_at DATETIME);`,
		`CREATE TABLE auth_roles (id TEXT PRIMARY KEY, name TEXT NOT NULL);`,
		`CREATE TABLE auth_permissions (id TEXT PRIMARY KEY, name TEXT NOT NULL);`,
		`CREATE TABLE auth_role_permissions (role_id TEXT, permission_id TEXT);`,
		`CREATE TABLE auth_memberships (user_id TEXT, tenant_id TEXT, role_id TEXT, status TEXT, created_at DATETIME, updated_at DATETIME);`,
		`CREATE TABLE tenant_invites (id TEXT PRIMARY KEY, tenant_id TEXT, email TEXT, role_id TEXT, token_hash TEXT, status TEXT DEFAULT 'pending', expires_at DATETIME, created_by TEXT, accepted_by TEXT, accepted_at DATETIME, created_at DATETIME, updated_at DATETIME);`,
	} {
		require.NoError(t, db.Exec(s).Error)
	}
	return newRepo(db), db
}

// TestGetMeContext_U3 valida /me/context: lista solo tenants activos del usuario
// (excluye archivados), resuelve el tenant actual y trae rol + permisos.
func TestGetMeContext_U3(t *testing.T) {
	r, db := newMeContextTestRepo(t)
	ctx := context.Background()

	userID := uuid.New()
	t1, t2, t3 := uuid.New(), uuid.New(), uuid.New() // acme, beta, gamma(archivado)
	roleAdmin := uuid.New()
	permR, permW := uuid.New(), uuid.New()

	require.NoError(t, db.Exec(`INSERT INTO users (id,email,username,idp_sub) VALUES (?,?,?,?)`, userID, "u@x.com", "u", "sub-1").Error)
	require.NoError(t, db.Exec(`INSERT INTO auth_tenants (id,name,status) VALUES (?,?,'active')`, t1, "acme").Error)
	require.NoError(t, db.Exec(`INSERT INTO auth_tenants (id,name,status) VALUES (?,?,'active')`, t2, "beta").Error)
	require.NoError(t, db.Exec(`INSERT INTO auth_tenants (id,name,status,deleted_at) VALUES (?,?,'active',?)`, t3, "gamma", time.Now()).Error)
	require.NoError(t, db.Exec(`INSERT INTO auth_roles (id,name) VALUES (?,?)`, roleAdmin, "admin").Error)
	require.NoError(t, db.Exec(`INSERT INTO auth_permissions (id,name) VALUES (?,?)`, permR, "api.read").Error)
	require.NoError(t, db.Exec(`INSERT INTO auth_permissions (id,name) VALUES (?,?)`, permW, "api.write").Error)
	require.NoError(t, db.Exec(`INSERT INTO auth_role_permissions (role_id,permission_id) VALUES (?,?)`, roleAdmin, permR).Error)
	require.NoError(t, db.Exec(`INSERT INTO auth_role_permissions (role_id,permission_id) VALUES (?,?)`, roleAdmin, permW).Error)
	require.NoError(t, db.Exec(`INSERT INTO auth_memberships (user_id,tenant_id,role_id,status) VALUES (?,?,?,'active')`, userID, t1, roleAdmin).Error)
	require.NoError(t, db.Exec(`INSERT INTO auth_memberships (user_id,tenant_id,role_id,status) VALUES (?,?,?,'active')`, userID, t2, roleAdmin).Error)
	require.NoError(t, db.Exec(`INSERT INTO auth_memberships (user_id,tenant_id,role_id,status) VALUES (?,?,?,'active')`, userID, t3, roleAdmin).Error)

	// Con X-Tenant-Id válido (t1) => current = t1; gamma archivado excluido.
	out, err := r.getMeContext(ctx, "sub-1", t1.String())
	require.NoError(t, err)
	require.NotNil(t, out.User)
	assert.Equal(t, "u@x.com", out.User.Email)
	require.NotNil(t, out.CurrentTenantID)
	assert.Equal(t, t1, *out.CurrentTenantID)
	assert.Len(t, out.Tenants, 2) // acme + beta (gamma archivado fuera)

	var acme *meTenant
	for i := range out.Tenants {
		if out.Tenants[i].Name == "acme" {
			acme = &out.Tenants[i]
		}
	}
	require.NotNil(t, acme)
	assert.True(t, acme.IsCurrent)
	assert.Equal(t, "admin", acme.Role)
	assert.ElementsMatch(t, []string{"api.read", "api.write"}, acme.Permissions)

	// Sin header y con >1 membership => current = nil (el FE debe elegir).
	out2, err := r.getMeContext(ctx, "sub-1", "")
	require.NoError(t, err)
	assert.Nil(t, out2.CurrentTenantID)

	// Header de un tenant donde NO tiene membership => current = nil.
	out3, err := r.getMeContext(ctx, "sub-1", uuid.New().String())
	require.NoError(t, err)
	assert.Nil(t, out3.CurrentTenantID)
}

// TestIsPlatformAdminBySub_U1 valida la fuente persistente de platform-admin.
func TestIsPlatformAdminBySub_U1(t *testing.T) {
	r, db := newMeContextTestRepo(t)
	ctx := context.Background()
	require.NoError(t, db.Exec(`INSERT INTO users (id,email,username,idp_sub,is_platform_admin) VALUES (?,?,?,?,1)`, uuid.New(), "pa@x.com", "pa", "sub-pa").Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id,email,username,idp_sub,is_platform_admin) VALUES (?,?,?,?,0)`, uuid.New(), "n@x.com", "n", "sub-n").Error)

	assert.True(t, r.isPlatformAdminBySub(ctx, "sub-pa"))
	assert.False(t, r.isPlatformAdminBySub(ctx, "sub-n"))
	assert.False(t, r.isPlatformAdminBySub(ctx, "sub-missing"))
}
