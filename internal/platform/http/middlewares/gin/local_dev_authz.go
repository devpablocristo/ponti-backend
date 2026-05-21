package pkgmwr

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/devpablocristo/platform/http/go/httperr"
	"github.com/devpablocristo/platform/security/go/contextkeys"
)

// RequireLocalDevAuthz is a lightweight auth middleware intended for local development.
//
// - It does NOT validate JWT signatures.
// - It relies on X-USER-ID (or the JWT "sub" claim if present) to populate context.
// - It always assigns the "admin" role by default.
//
// Enable it by setting AUTH_ENABLED=false in the backend.
func RequireLocalDevAuthz(cfg IdentityAuthConfig, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		permission := permissionForMethod(c.Request.Method)

		// Try to extract subject from a (possibly fake) JWT so the BFF can just pass Authorization.
		subject := ""
		email := ""
		if tokenStr := extractBearer(c.GetHeader("Authorization")); tokenStr != "" {
			if payload := decodeTokenPayload(tokenStr); payload != nil {
				if v, ok := payload["sub"]; ok {
					subject = strings.TrimSpace(toString(v))
				}
				if v, ok := payload["email"]; ok {
					email = strings.TrimSpace(toString(v))
				}
			}
		}

		if subject == "" {
			subject = strings.TrimSpace(c.GetHeader(HeaderUserID))
		}
		if subject == "" {
			subject = "local-dev-user"
		}

		tenantHeader := strings.TrimSpace(c.GetHeader(cfg.TenantHeader))
		var tenantID uuid.UUID
		if tenantHeader == "" {
			if cfg.RequireTenantHeader && !allowsImplicitTenant(c) {
				denyLocalDevAuthzRequest(c, "tenant header required")
				logAuthDecision(c.Request.Context(), subject, "", c.FullPath(), permission, "DENY(local)", start)
				return
			}
		} else {
			parsed, err := uuid.Parse(tenantHeader)
			if err != nil {
				denyLocalDevAuthzRequest(c, "invalid tenant header")
				logAuthDecision(c.Request.Context(), subject, "", c.FullPath(), permission, "DENY(local)", start)
				return
			}
			tenantID = parsed
		}

		// Resolve the user in the DB if possible to get a valid UUID user.
		var resolvedUserID uuid.UUID
		if db != nil {
			if ensuredID, ensureErr := ensureUserByIDPSub(c.Request.Context(), db, subject, email); ensureErr == nil {
				resolvedUserID = ensuredID
			}
		}

		// If no tenant UUID was provided, try to look up the "default" tenant for
		// legacy local-dev compatibility. Strict tenant-scoped routes are denied
		// above before reaching this fallback.
		if tenantID == uuid.Nil && db != nil {
			type tRow struct{ ID uuid.UUID }
			var t tRow
			if err := db.WithContext(c.Request.Context()).Table("auth_tenants").Select("id").Where("name = 'default'").Limit(1).Take(&t).Error; err == nil {
				tenantID = t.ID
			}
		}

		role := "admin"
		if resolvedUserID != uuid.Nil && tenantID != uuid.Nil && db != nil {
			if membership, err := ensureMembershipForTenantID(c.Request.Context(), db, resolvedUserID, tenantID, cfg.DefaultRole); err == nil {
				role = membership.RoleName
				tenantID = membership.TenantID
			}
		}

		// Inject core/saas/go context keys.
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, ctxkeys.Actor, subject)
		ctx = context.WithValue(ctx, ctxkeys.OrgID, tenantID)
		ctx = context.WithValue(ctx, ctxkeys.Role, role)
		ctx = context.WithValue(ctx, ctxkeys.Scopes, []string{permissionAPIRead, permissionAPIWrite})
		c.Request = c.Request.WithContext(ctx)

		// Also set gin keys for convenience.
		c.Set(string(ctxkeys.Actor), subject)
		c.Set(string(ctxkeys.OrgID), tenantID)
		c.Set(string(ctxkeys.Role), role)
		c.Set(string(ctxkeys.Scopes), []string{permissionAPIRead, permissionAPIWrite})

		// Log as allow; real authorization is handled in usecases.
		logAuthDecision(c.Request.Context(), subject, tenantID.String(), c.FullPath(), permission, "ALLOW(local)", start)

		// Use resolvedUserID if needed elsewhere via gin key.
		_ = resolvedUserID

		// Allow
		c.Next()
	}
}

func denyLocalDevAuthzRequest(c *gin.Context, details string) {
	domErr := domainerr.Forbidden(details)
	status, apiErr := httperr.Normalize(domErr)
	c.AbortWithStatusJSON(status, apiErr)
}

func decodeTokenPayload(token string) map[string]any {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil
	}
	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(payloadJSON, &out); err != nil {
		return nil
	}
	return out
}

func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return ""
	}
}
