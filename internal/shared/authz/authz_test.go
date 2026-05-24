package authz

import (
	"context"
	"testing"

	"github.com/google/uuid"

	contextkeys "github.com/devpablocristo/platform/security/go/contextkeys"
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

func TestTenantFromContextAllowsTransitionMode(t *testing.T) {
	if _, ok := TenantFromContext(context.Background()); ok {
		t.Fatalf("empty context should not have tenant")
	}

	ctx := principalContext("tenant_viewer", []string{PermissionCustomersRead})
	if tenantID, ok := TenantFromContext(ctx); !ok || tenantID == uuid.Nil {
		t.Fatalf("expected tenant from authenticated context")
	}
}

// Nota: los tests de TenantWhere / MaybeTenantScope se movieron a platform —
// ver `platform/persistence/gorm/go/tenancy/tenancy_test.go`. Las isolation
// suites por módulo (`internal/*/repository_tenant_test.go`) siguen siendo
// la prueba E2E definitiva de aislamiento cross-tenant.
