package pkgmwr

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
			subject = "1"
		}
		// Normalize context user ID to numeric string, matching the rest of the backend.
		// If the subject is an IDP UID, resolve/create a local user and use its numeric ID.
		resolvedUserID := int64(1)
		if parsed, err := strconv.ParseInt(subject, 10, 64); err == nil {
			resolvedUserID = parsed
		} else if db != nil {
			if ensuredID, ensureErr := ensureUserByIDPSub(c.Request.Context(), db, subject, email); ensureErr == nil {
				resolvedUserID = ensuredID
			}
		}
		subject = strconv.FormatInt(resolvedUserID, 10)

		tenantID := strings.TrimSpace(c.GetHeader(cfg.TenantHeader))
		if tenantID == "" {
			tenantID = "1"
		} else {
			// Normalize to numeric string if possible
			if _, err := strconv.ParseInt(tenantID, 10, 64); err != nil {
				tenantID = "1"
			}
		}

		roles := []string{"admin"}

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, ContextUserIDKey, subject)
		ctx = context.WithValue(ctx, ContextTenantIDKey, tenantID)
		ctx = context.WithValue(ctx, ContextRolesKey, roles)
		c.Request = c.Request.WithContext(ctx)

		c.Set(ContextUserID, subject)
		c.Set(ContextTenantID, tenantID)
		c.Set(ContextRoles, roles)

		// Log as allow; real authorization is handled in usecases.
		logAuthDecision(subject, tenantID, c.FullPath(), permission, "ALLOW(local)", start)

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

