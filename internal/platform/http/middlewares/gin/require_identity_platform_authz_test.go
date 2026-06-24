package pkgmwr

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newAuthTestDB crea una DB sqlite in-memory con el set mínimo de tablas auth
// que consume resolveMembership/loadMembershipPermissions.
func newAuthTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	stmts := []string{
		`CREATE TABLE auth_roles (id TEXT PRIMARY KEY, name TEXT NOT NULL);`,
		`CREATE TABLE auth_permissions (id TEXT PRIMARY KEY, name TEXT NOT NULL);`,
		`CREATE TABLE auth_role_permissions (role_id TEXT NOT NULL, permission_id TEXT NOT NULL);`,
		`CREATE TABLE auth_memberships (user_id TEXT NOT NULL, tenant_id TEXT NOT NULL, role_id TEXT NOT NULL, status TEXT NOT NULL);`,
	}
	for _, s := range stmts {
		require.NoError(t, db.Exec(s).Error)
	}
	return db
}

// TestResolveMembership_T1c valida el endurecimiento de selección de tenant:
// sin X-Tenant-Id, 1 membership se usa y >1 exige selección explícita
// (errTenantSelectionRequired); con header, se valida membership en ESE tenant
// sin fallback. Antes existía un Order("m.tenant_id ASC").Limit(1) que elegía
// un tenant arbitrario.
func TestResolveMembership_T1c(t *testing.T) {
	db := newAuthTestDB(t)
	ctx := context.Background()

	roleAdmin, roleViewer := uuid.New(), uuid.New()
	permRead, permWrite := uuid.New(), uuid.New()
	user1, user2 := uuid.New(), uuid.New()
	tenantA, tenantB, tenantC := uuid.New(), uuid.New(), uuid.New()

	require.NoError(t, db.Exec(`INSERT INTO auth_roles (id,name) VALUES (?,?),(?,?)`,
		roleAdmin, "admin", roleViewer, "viewer").Error)
	require.NoError(t, db.Exec(`INSERT INTO auth_permissions (id,name) VALUES (?,?),(?,?)`,
		permRead, "api.read", permWrite, "api.write").Error)
	require.NoError(t, db.Exec(`INSERT INTO auth_role_permissions (role_id,permission_id) VALUES (?,?),(?,?),(?,?)`,
		roleAdmin, permRead, roleAdmin, permWrite, roleViewer, permRead).Error)
	// user1: 2 memberships (tenantA=admin, tenantB=viewer); user2: 1 membership (tenantA=viewer).
	require.NoError(t, db.Exec(`INSERT INTO auth_memberships (user_id,tenant_id,role_id,status) VALUES (?,?,?,?),(?,?,?,?),(?,?,?,?)`,
		user1, tenantA, roleAdmin, "active",
		user1, tenantB, roleViewer, "active",
		user2, tenantA, roleViewer, "active").Error)

	t.Run("multi-membership sin header => errTenantSelectionRequired", func(t *testing.T) {
		_, err := resolveMembership(ctx, db, user1, "")
		assert.True(t, errors.Is(err, errTenantSelectionRequired))
	})

	t.Run("single-membership sin header => usa la única", func(t *testing.T) {
		m, err := resolveMembership(ctx, db, user2, "")
		require.NoError(t, err)
		assert.Equal(t, tenantA, m.TenantID)
		assert.Equal(t, "viewer", m.RoleName)
		_, hasRead := m.Permissions["api.read"]
		_, hasWrite := m.Permissions["api.write"]
		assert.True(t, hasRead)
		assert.False(t, hasWrite)
	})

	t.Run("header con membership válida => ese tenant (no el ASC)", func(t *testing.T) {
		m, err := resolveMembership(ctx, db, user1, tenantB.String())
		require.NoError(t, err)
		assert.Equal(t, tenantB, m.TenantID)
		assert.Equal(t, "viewer", m.RoleName)
	})

	t.Run("header sin membership => error, sin fallback", func(t *testing.T) {
		_, err := resolveMembership(ctx, db, user1, tenantC.String())
		require.Error(t, err)
		assert.False(t, errors.Is(err, errTenantSelectionRequired))
	})
}
