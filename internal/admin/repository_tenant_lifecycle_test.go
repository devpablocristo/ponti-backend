package admin

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newAdminTestRepo(t *testing.T) (*repo, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE auth_tenants (
		id TEXT PRIMARY KEY,
		legacy_id INTEGER,
		name TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'active',
		deleted_at DATETIME,
		created_at DATETIME,
		updated_at DATETIME
	);`).Error)
	return newRepo(db), db
}

// TestTenantLifecycle_PARTEIV valida el CRUDAR de tenant (PARTE IV):
// get / list(active/archived) / suspend-activate / archive-restore / hard delete.
func TestTenantLifecycle_PARTEIV(t *testing.T) {
	r, db := newAdminTestRepo(t)
	ctx := context.Background()
	id := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO auth_tenants (id, name, status) VALUES (?, 'acme', 'active')`, id).Error)

	// get
	tt, err := r.getTenant(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "acme", tt.Name)
	assert.Equal(t, "active", tt.Status)

	// suspend / activate
	require.NoError(t, r.setTenantStatus(ctx, id, "suspended"))
	tt, _ = r.getTenant(ctx, id)
	assert.Equal(t, "suspended", tt.Status)
	require.NoError(t, r.setTenantStatus(ctx, id, "active"))

	// listados: activo lo incluye, archivado no
	active, err := r.listTenantsByArchived(ctx, false)
	require.NoError(t, err)
	assert.Len(t, active, 1)
	archived, err := r.listTenantsByArchived(ctx, true)
	require.NoError(t, err)
	assert.Len(t, archived, 0)

	// archive
	require.NoError(t, r.archiveTenant(ctx, id))
	tt, _ = r.getTenant(ctx, id)
	assert.NotNil(t, tt.DeletedAt)
	active, _ = r.listTenantsByArchived(ctx, false)
	assert.Len(t, active, 0)
	archived, _ = r.listTenantsByArchived(ctx, true)
	assert.Len(t, archived, 1)

	// archivar de nuevo => error (ya archivado)
	assert.Error(t, r.archiveTenant(ctx, id))

	// restore
	require.NoError(t, r.restoreTenant(ctx, id))
	tt, _ = r.getTenant(ctx, id)
	assert.Nil(t, tt.DeletedAt)

	// hard delete requiere archivado primero
	assert.Error(t, r.hardDeleteTenant(ctx, id))
	require.NoError(t, r.archiveTenant(ctx, id))
	require.NoError(t, r.hardDeleteTenant(ctx, id))
	_, err = r.getTenant(ctx, id)
	assert.Error(t, err) // ya no existe
}
