package admin

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInvites_U4 valida el flujo de invites a nivel repo: crear, listar, resolver por
// token, aceptar (crea membership), re-aceptar (conflicto) y revocar.
func TestInvites_U4(t *testing.T) {
	r, db := newMeContextTestRepo(t)
	ctx := context.Background()

	tenant := uuid.New()
	roleViewer := uuid.New()
	userID := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO auth_tenants (id,name,status) VALUES (?,?,'active')`, tenant, "acme").Error)
	require.NoError(t, db.Exec(`INSERT INTO auth_roles (id,name) VALUES (?,?)`, roleViewer, "viewer").Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id,email,username,idp_sub) VALUES (?,?,?,?)`, userID, "invitee@x.com", "invitee", "sub-invitee").Error)

	// token hashing: el mismo token da el mismo hash; distinto token, distinto hash.
	assert.Equal(t, hashToken("abc"), hashToken("abc"))
	assert.NotEqual(t, hashToken("abc"), hashToken("xyz"))

	// create
	th := hashToken("the-token")
	invID, err := r.createInvite(ctx, tenant, roleViewer, "invitee@x.com", th, "sub-inviter", time.Now().Add(time.Hour))
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, invID)

	// list
	list, err := r.listInvites(ctx, tenant)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "viewer", list[0].Role)
	assert.Equal(t, "pending", list[0].Status)
	assert.Equal(t, "invitee@x.com", list[0].Email)

	// resolve by token hash
	inv, err := r.getInviteByTokenHash(ctx, th)
	require.NoError(t, err)
	assert.Equal(t, "pending", inv.Status)
	assert.Equal(t, tenant, inv.TenantID)
	assert.Equal(t, roleViewer, inv.RoleID)

	// token desconocido => NotFound
	_, err = r.getInviteByTokenHash(ctx, hashToken("nope"))
	assert.Error(t, err)

	// accept => membership activa creada + invite ya no pending
	require.NoError(t, r.acceptInvite(ctx, inv, userID))
	var memCount int64
	require.NoError(t, db.Raw(`SELECT count(*) FROM auth_memberships WHERE user_id=? AND tenant_id=? AND role_id=? AND status='active'`, userID, tenant, roleViewer).Scan(&memCount).Error)
	assert.Equal(t, int64(1), memCount)

	// re-aceptar el mismo invite => conflicto (ya no pending)
	assert.Error(t, r.acceptInvite(ctx, inv, userID))

	// revoke de un invite pendiente fresco
	invID2, err := r.createInvite(ctx, tenant, roleViewer, "x2@x.com", hashToken("tok2"), "sub", time.Now().Add(time.Hour))
	require.NoError(t, err)
	require.NoError(t, r.revokeInvite(ctx, invID2, tenant))
	// revocar de nuevo => no hay pending => error
	assert.Error(t, r.revokeInvite(ctx, invID2, tenant))
}
