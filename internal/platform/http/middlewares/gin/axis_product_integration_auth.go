package pkgmwr

import (
	"context"
	"crypto/subtle"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	ctxkeys "github.com/devpablocristo/platform/security/go/contextkeys"
)

const axisProductIntegrationActor = "axis-companion"

// BridgeAxisProductIntegrationAPIKey lets Axis use its product API key as a
// bearer token while Ponti still protects requests with the normal X_API_KEY
// middleware. It only applies to the product integration endpoints.
func BridgeAxisProductIntegrationAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		if axisProductIntegrationAuthorized(c) && axisProductIntegrationPath(c) {
			if apiKey := strings.TrimSpace(os.Getenv("X_API_KEY")); apiKey != "" {
				c.Request.Header.Set("X-API-KEY", apiKey)
			}
		}
	}
}

func authorizeAxisProductIntegration(c *gin.Context, cfg IdentityAuthConfig, permission string, start time.Time) bool {
	if !axisProductIntegrationAuthorized(c) {
		return false
	}
	if !axisProductIntegrationPath(c) {
		return false
	}

	var tenantID uuid.UUID
	if axisProductIntegrationRequiresTenant(c) {
		rawTenant := strings.TrimSpace(c.GetHeader(cfg.TenantHeader))
		parsed, err := uuid.Parse(rawTenant)
		if err != nil || parsed == uuid.Nil {
			denyForbidden(c, "axis product integration requires X-Tenant-Id")
			logAuthDecision(axisProductIntegrationActor, "", c.FullPath(), permission, "DENY(axis-product-integration)", start)
			return true
		}
		tenantID = parsed
	}

	ctx := c.Request.Context()
	ctx = context.WithValue(ctx, ctxkeys.Actor, axisProductIntegrationActor)
	ctx = context.WithValue(ctx, ctxkeys.Role, "service")
	ctx = context.WithValue(ctx, ctxkeys.Scopes, []string{permissionAPIRead})
	if tenantID != uuid.Nil {
		ctx = context.WithValue(ctx, ctxkeys.OrgID, tenantID)
	}
	c.Request = c.Request.WithContext(ctx)
	c.Set(string(ctxkeys.Actor), axisProductIntegrationActor)
	c.Set(string(ctxkeys.Role), "service")
	c.Set(string(ctxkeys.Scopes), []string{permissionAPIRead})
	if tenantID != uuid.Nil {
		c.Set(string(ctxkeys.OrgID), tenantID)
	}
	logAuthDecision(axisProductIntegrationActor, tenantID.String(), c.FullPath(), permission, "ALLOW(axis-product-integration)", start)
	return true
}

func axisProductIntegrationAuthorized(c *gin.Context) bool {
	key := strings.TrimSpace(os.Getenv("PONTI_AXIS_API_KEY"))
	if key == "" {
		return false
	}
	token := extractBearer(c.GetHeader("Authorization"))
	if token == "" {
		token = strings.TrimSpace(c.GetHeader("X-Ponti-Axis-Api-Key"))
	}
	return constantTimeEqual(token, key)
}

func axisProductIntegrationPath(c *gin.Context) bool {
	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}
	return path == "/api/v1/capabilities" ||
		path == "/api/v1/insights" ||
		path == "/api/v1/insights/summary" ||
		(strings.HasPrefix(path, "/api/v1/insights/") && strings.HasSuffix(path, "/explain")) ||
		(strings.HasPrefix(path, "/api/v1/ai/actions/") && strings.HasSuffix(path, "/prepare"))
}

func axisProductIntegrationRequiresTenant(c *gin.Context) bool {
	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}
	return path != "/api/v1/capabilities"
}

func constantTimeEqual(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
