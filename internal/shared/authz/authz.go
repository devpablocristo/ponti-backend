// Package authz centraliza la lectura del principal autenticado y los checks de
// autorización por permiso para el request actual.
//
// Es un paquete HOJA (U0 del Pilar 2): solo depende de la librería de contextkeys
// de plataforma, uuid y stdlib — NO importa ningún internal/<dominio>. Los
// middlewares de auth inyectan las ctxkeys (Actor/OrgID/Role/Scopes) y este
// paquete las consume de forma uniforme, reemplazando los reads dispersos.
package authz

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	ctxkeys "github.com/devpablocristo/platform/security/go/contextkeys"
)

// Principal es el sujeto autenticado del request, según lo dejaron los middlewares
// de auth en el contexto.
type Principal struct {
	Subject     string    // idp_sub del usuario autenticado
	TenantID    uuid.UUID // tenant activo (OrgID); uuid.Nil si no hay
	Role        string    // role_name de la membership activa
	Permissions []string  // permisos/scopes del rol en el tenant activo
}

// PrincipalFromContext arma el Principal a partir de las ctxkeys inyectadas por el
// middleware de auth. ok=false si no hay sujeto autenticado en el contexto.
func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	var p Principal
	if ctx == nil {
		return p, false
	}
	if sub, ok := ctx.Value(ctxkeys.Actor).(string); ok {
		p.Subject = sub
	}
	if org, ok := ctx.Value(ctxkeys.OrgID).(uuid.UUID); ok {
		p.TenantID = org
	}
	if role, ok := ctx.Value(ctxkeys.Role).(string); ok {
		p.Role = role
	}
	if scopes, ok := ctx.Value(ctxkeys.Scopes).([]string); ok {
		p.Permissions = scopes
	}
	return p, p.Subject != ""
}

// HasPermission indica si el principal del contexto tiene el permiso dado.
func HasPermission(ctx context.Context, permission string) bool {
	p, ok := PrincipalFromContext(ctx)
	if !ok {
		return false
	}
	for _, perm := range p.Permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// RequirePermission devuelve un error Forbidden si el principal no tiene el permiso.
func RequirePermission(ctx context.Context, permission string) error {
	if !HasPermission(ctx, permission) {
		return domainerr.Forbidden("insufficient permissions: " + permission)
	}
	return nil
}

// HasPermissionOrRole es el dual-check de transición (U2): evalúa el permiso FINO y, si
// el principal no lo tiene pero sí cumple alguno de los roles gruesos de fallback,
// permite y loguea `fallback_to_coarse`. Cuando todos los roles tengan el permiso fino
// el fallback deja de dispararse (fallback_to_coarse=0) y se puede retirar (U5).
func HasPermissionOrRole(ctx context.Context, permission string, fallbackRoles ...string) bool {
	if HasPermission(ctx, permission) {
		return true
	}
	p, ok := PrincipalFromContext(ctx)
	if !ok {
		return false
	}
	for _, r := range fallbackRoles {
		if r != "" && p.Role == r {
			slog.WarnContext(ctx, "authz fallback_to_coarse", "permission", permission, "role", p.Role, "subject", p.Subject)
			return true
		}
	}
	return false
}

// RequirePermissionOrRole es HasPermissionOrRole devolviendo Forbidden si falla.
func RequirePermissionOrRole(ctx context.Context, permission string, fallbackRoles ...string) error {
	if HasPermissionOrRole(ctx, permission, fallbackRoles...) {
		return nil
	}
	return domainerr.Forbidden("insufficient permissions: " + permission)
}

// RequireTenant devuelve el tenant activo del contexto, o un error si no hay.
func RequireTenant(ctx context.Context) (uuid.UUID, error) {
	p, ok := PrincipalFromContext(ctx)
	if !ok || p.TenantID == uuid.Nil {
		return uuid.Nil, domainerr.Forbidden("tenant context required")
	}
	return p.TenantID, nil
}
