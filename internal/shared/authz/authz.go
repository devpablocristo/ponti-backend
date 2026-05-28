package authz

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/devpablocristo/platform/security/go/contextkeys"
	platformtenant "github.com/devpablocristo/platform/security/go/tenant"
)

const (
	PermissionAPIRead  = "api.read"
	PermissionAPIWrite = "api.write"

	PermissionAdminTenants     = "admin.tenants"
	PermissionAdminUsers       = "admin.users"
	PermissionAdminMemberships = "admin.memberships"
	PermissionAIUse            = "ai.use"

	PermissionCustomersRead     = "customers.read"
	PermissionCustomersWrite    = "customers.write"
	PermissionCustomersArchive  = "customers.archive"
	PermissionProjectsRead      = "projects.read"
	PermissionProjectsWrite     = "projects.write"
	PermissionProjectsArchive   = "projects.archive"
	PermissionLotsRead          = "lots.read"
	PermissionLotsWrite         = "lots.write"
	PermissionLotsArchive       = "lots.archive"
	PermissionWorkOrdersRead    = "workorders.read"
	PermissionWorkOrdersWrite   = "workorders.write"
	PermissionWorkOrdersArchive = "workorders.archive"
	PermissionLaborsRead        = "labors.read"
	PermissionLaborsWrite       = "labors.write"
	PermissionLaborsArchive     = "labors.archive"
	PermissionSuppliesRead      = "supplies.read"
	PermissionSuppliesWrite     = "supplies.write"
	PermissionSuppliesArchive   = "supplies.archive"
	PermissionStockRead         = "stock.read"
	PermissionStockWrite        = "stock.write"
	PermissionStockArchive      = "stock.archive"
	PermissionActorsRead        = "actors.read"
	PermissionActorsWrite       = "actors.write"
	PermissionActorsArchive     = "actors.archive"
	PermissionActorsMerge       = "actors.merge"
	PermissionExportsRun        = "exports.run"
	PermissionImportsRun        = "imports.run"
)

type Principal struct {
	Actor       string
	TenantID    uuid.UUID
	Role        string
	Permissions []string
}

func PrincipalFromContext(ctx context.Context) (Principal, error) {
	if ctx == nil {
		return Principal{}, domainerr.Forbidden("authentication context required")
	}

	actor, _ := ctx.Value(ctxkeys.Actor).(string)
	role, _ := ctx.Value(ctxkeys.Role).(string)
	tenantID, _ := ctx.Value(ctxkeys.OrgID).(uuid.UUID)

	var permissions []string
	switch raw := ctx.Value(ctxkeys.Scopes).(type) {
	case []string:
		permissions = append(permissions, raw...)
	case []any:
		for _, item := range raw {
			if s, ok := item.(string); ok {
				permissions = append(permissions, s)
			}
		}
	}

	if strings.TrimSpace(actor) == "" {
		return Principal{}, domainerr.Forbidden("invalid actor")
	}
	if tenantID == uuid.Nil {
		return Principal{}, domainerr.TenantMissing()
	}

	return Principal{
		Actor:       strings.TrimSpace(actor),
		TenantID:    tenantID,
		Role:        strings.TrimSpace(role),
		Permissions: permissions,
	}, nil
}

func RequireTenant(ctx context.Context) (uuid.UUID, error) {
	principal, err := PrincipalFromContext(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return principal.TenantID, nil
}

func TenantFromContext(ctx context.Context) (uuid.UUID, bool) {
	if ctx == nil {
		return uuid.Nil, false
	}
	if tenantID, ok := ctx.Value(ctxkeys.OrgID).(uuid.UUID); ok && tenantID != uuid.Nil {
		return tenantID, true
	}
	return uuid.Nil, false
}

func OptionalTenantOrStrict(ctx context.Context) (uuid.UUID, bool, error) {
	tenantID, ok := TenantFromContext(ctx)
	if !ok && TenantStrictModeEnabled() {
		return uuid.Nil, false, domainerr.TenantMissing()
	}
	return tenantID, ok, nil
}

// TenantStrictModeEnabled delega en platform/security/go/tenant. Es el
// único forwarder que queda — los demás callers de scoping migraron a
// `platform/persistence/gorm/go/tenancy.Scope` directo (Fase 7).
func TenantStrictModeEnabled() bool {
	return platformtenant.StrictModeEnabled()
}

func HasPermission(ctx context.Context, permission string) bool {
	permission = strings.TrimSpace(permission)
	if permission == "" {
		return false
	}
	principal, err := PrincipalFromContext(ctx)
	if err != nil {
		return false
	}
	return principal.HasPermission(permission)
}

func RequirePermission(ctx context.Context, permission string) error {
	permission = strings.TrimSpace(permission)
	if permission == "" {
		return domainerr.Forbidden("permission required")
	}
	principal, err := PrincipalFromContext(ctx)
	if err != nil {
		return err
	}
	if !principal.HasPermission(permission) {
		return domainerr.Forbidden("insufficient permissions")
	}
	return nil
}

func (p Principal) HasPermission(permission string) bool {
	permission = strings.TrimSpace(permission)
	if permission == "" {
		return false
	}
	for _, item := range p.Permissions {
		if item == permission {
			return true
		}
	}
	return strings.EqualFold(p.Role, "saas_superadmin")
}
