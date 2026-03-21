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

	"github.com/devpablocristo/core/saas/go/shared/ctxkeys"
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

		// Resolve the user in the DB if possible to get a valid UUID user.
		var resolvedUserID uuid.UUID
		if db != nil {
			if ensuredID, ensureErr := ensureUserByIDPSub(c.Request.Context(), db, subject, email); ensureErr == nil {
				resolvedUserID = ensuredID
			}
		}

		// Resolve tenant ID from header; parse as UUID or default.
		tenantHeader := strings.TrimSpace(c.GetHeader(cfg.TenantHeader))
		var tenantID uuid.UUID
		if tenantHeader != "" {
			if parsed, err := uuid.Parse(tenantHeader); err == nil {
				tenantID = parsed
			}
		}
		// If no valid tenant UUID was provided, try to look up the "default" tenant.
		if tenantID == uuid.Nil && db != nil {
			type tRow struct{ ID uuid.UUID }
			var t tRow
			if err := db.WithContext(c.Request.Context()).Table("auth_tenants").Select("id").Where("name = 'default'").Limit(1).Take(&t).Error; err == nil {
				tenantID = t.ID
			}
		}

		role := "admin"

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
		logAuthDecision(subject, tenantID.String(), c.FullPath(), permission, "ALLOW(local)", start)

		// Use resolvedUserID if needed elsewhere via gin key.
		_ = resolvedUserID

		// Allow
		c.Next()
	}
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
