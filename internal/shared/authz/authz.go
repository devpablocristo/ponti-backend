package authz

import (
	"context"
	"os"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/devpablocristo/platform/security/go/contextkeys"
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
		return Principal{}, domainerr.Forbidden("tenant context required")
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
		return uuid.Nil, false, domainerr.Forbidden("tenant context required")
	}
	return tenantID, ok, nil
}

func TenantWhere(ctx context.Context, columnOrAlias string) (string, []any, error) {
	tenantID, err := RequireTenant(ctx)
	if err != nil {
		return "", nil, err
	}
	column := normalizeTenantColumn(columnOrAlias)
	return column + " = ?", []any{tenantID}, nil
}

func TenantScope(ctx context.Context, db *gorm.DB, columnOrAlias string) (*gorm.DB, error) {
	if db == nil {
		return nil, domainerr.Internal("database connection required")
	}
	clause, args, err := TenantWhere(ctx, columnOrAlias)
	if err != nil {
		return nil, err
	}
	return db.Where(clause, args...), nil
}

func MaybeTenantScope(ctx context.Context, db *gorm.DB, columnOrAlias string) *gorm.DB {
	if db == nil {
		return db
	}
	tenantID, ok := TenantFromContext(ctx)
	if !ok {
		if TenantStrictModeEnabled() {
			err := domainerr.Forbidden("tenant context required")
			if db.Config == nil {
				db.Error = err
			} else {
				_ = db.AddError(err)
			}
		}
		return db
	}
	column := normalizeTenantColumn(columnOrAlias)
	return db.Where(column+" = ?", tenantID)
}

func TenantStrictModeEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("TENANT_STRICT_MODE"))) {
	case "1", "t", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func normalizeTenantColumn(columnOrAlias string) string {
	value := strings.TrimSpace(columnOrAlias)
	if value == "" {
		return "tenant_id"
	}
	if strings.Contains(value, ".") || strings.Contains(value, "(") || strings.Contains(value, " ") {
		return value
	}
	return value + ".tenant_id"
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
