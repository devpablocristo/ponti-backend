package authz

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	ctxkeys "github.com/devpablocristo/platform/security/go/contextkeys"
)

// TestPrincipalAndPermissions_U0 valida el paquete authz (U0): lectura del
// principal desde el contexto y los checks de permiso/tenant.
func TestPrincipalAndPermissions_U0(t *testing.T) {
	// Contexto vacío: sin principal.
	if _, ok := PrincipalFromContext(context.Background()); ok {
		t.Fatal("expected no principal in empty context")
	}

	org := uuid.New()
	ctx := context.Background()
	ctx = context.WithValue(ctx, ctxkeys.Actor, "sub-123")
	ctx = context.WithValue(ctx, ctxkeys.OrgID, org)
	ctx = context.WithValue(ctx, ctxkeys.Role, "admin")
	ctx = context.WithValue(ctx, ctxkeys.Scopes, []string{"api.read", "api.write"})

	p, ok := PrincipalFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "sub-123", p.Subject)
	assert.Equal(t, org, p.TenantID)
	assert.Equal(t, "admin", p.Role)
	assert.Equal(t, []string{"api.read", "api.write"}, p.Permissions)

	assert.True(t, HasPermission(ctx, "api.read"))
	assert.False(t, HasPermission(ctx, "api.delete"))
	assert.NoError(t, RequirePermission(ctx, "api.write"))
	assert.Error(t, RequirePermission(ctx, "api.delete"))

	tid, err := RequireTenant(ctx)
	assert.NoError(t, err)
	assert.Equal(t, org, tid)

	// Principal sin tenant => RequireTenant falla.
	noTenant := context.WithValue(context.Background(), ctxkeys.Actor, "s")
	_, err = RequireTenant(noTenant)
	assert.Error(t, err)

	// Sin principal => HasPermission false, RequirePermission error.
	assert.False(t, HasPermission(context.Background(), "api.read"))
	assert.Error(t, RequirePermission(context.Background(), "api.read"))
}
