package authz

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	contextkeys "github.com/devpablocristo/core/security/go/contextkeys"
)

func principalContext(role string, permissions []string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "user-1")
	ctx = context.WithValue(ctx, contextkeys.OrgID, uuid.New())
	ctx = context.WithValue(ctx, contextkeys.Role, role)
	ctx = context.WithValue(ctx, contextkeys.Scopes, permissions)
	return ctx
}

func TestTenantOwnerIsNotGlobalWildcard(t *testing.T) {
	ctx := principalContext("tenant_owner", []string{PermissionAdminMemberships})

	if HasPermission(ctx, PermissionAdminTenants) {
		t.Fatalf("tenant_owner must not pass global tenant admin without explicit permission")
	}
	if !HasPermission(ctx, PermissionAdminMemberships) {
		t.Fatalf("tenant_owner should pass explicitly granted tenant membership permission")
	}
}

func TestSaaSSuperadminIsGlobalWildcard(t *testing.T) {
	ctx := principalContext("saas_superadmin", nil)

	if !HasPermission(ctx, PermissionAdminTenants) {
		t.Fatalf("saas_superadmin should pass global tenant admin")
	}
}

func TestTenantWhereBuildsScopedClause(t *testing.T) {
	ctx := principalContext("tenant_viewer", []string{PermissionCustomersRead})

	clause, args, err := TenantWhere(ctx, "customers")
	if err != nil {
		t.Fatalf("expected tenant clause: %v", err)
	}
	if clause != "customers.tenant_id = ?" {
		t.Fatalf("unexpected tenant clause %q", clause)
	}
	if len(args) != 1 {
		t.Fatalf("expected one tenant arg, got %d", len(args))
	}
}

func TestTenantWhereRequiresTenant(t *testing.T) {
	_, _, err := TenantWhere(context.Background(), "customers")
	if err == nil {
		t.Fatalf("expected missing tenant to fail")
	}
}

func TestTenantFromContextAllowsTransitionMode(t *testing.T) {
	if _, ok := TenantFromContext(context.Background()); ok {
		t.Fatalf("empty context should not have tenant")
	}

	ctx := principalContext("tenant_viewer", []string{PermissionCustomersRead})
	if tenantID, ok := TenantFromContext(ctx); !ok || tenantID == uuid.Nil {
		t.Fatalf("expected tenant from authenticated context")
	}
}

func TestMaybeTenantScopeRequiresTenantWhenStrictModeIsEnabled(t *testing.T) {
	t.Setenv("TENANT_STRICT_MODE", "true")

	db := &gorm.DB{}
	scoped := MaybeTenantScope(context.Background(), db, "customers")
	if scoped == nil || scoped.Error == nil {
		t.Fatalf("expected missing tenant to add a strict-mode error")
	}
}

func TestMaybeTenantScopeAllowsMissingTenantInTransitionMode(t *testing.T) {
	t.Setenv("TENANT_STRICT_MODE", "false")

	db := &gorm.DB{}
	scoped := MaybeTenantScope(context.Background(), db, "customers")
	if scoped == nil {
		t.Fatalf("expected db to be returned")
		return
	}
	if scoped.Error != nil {
		t.Fatalf("expected transition mode to allow missing tenant, got %v", scoped.Error)
	}
}
