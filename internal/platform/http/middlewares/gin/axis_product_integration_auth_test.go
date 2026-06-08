package pkgmwr

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	ctxkeys "github.com/devpablocristo/platform/security/go/contextkeys"
)

func TestAxisProductIntegrationAuth_AllowsInsightsWithTenant(t *testing.T) {
	t.Setenv("X_API_KEY", "backend-key")
	t.Setenv("PONTI_AXIS_API_KEY", "axis-key")
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mws := NewDefaultMiddlewares(BuildConfig{Auth: IdentityAuthConfig{Enabled: true, TenantHeader: "X-Tenant-Id"}})
	handlers := append(mws.GetValidation(), func(c *gin.Context) {
		if c.Request.Context().Value(ctxkeys.Actor) != axisProductIntegrationActor {
			t.Fatalf("axis actor was not injected")
		}
		if c.Request.Context().Value(ctxkeys.OrgID) == nil {
			t.Fatalf("tenant was not injected")
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	router.GET("/api/v1/insights/summary", handlers...)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/insights/summary", nil)
	req.Header.Set("Authorization", "Bearer axis-key")
	req.Header.Set("X-Tenant-Id", uuid.NewString())
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
}

func TestAxisProductIntegrationAuth_RequiresTenantForInsights(t *testing.T) {
	t.Setenv("X_API_KEY", "backend-key")
	t.Setenv("PONTI_AXIS_API_KEY", "axis-key")
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mws := NewDefaultMiddlewares(BuildConfig{Auth: IdentityAuthConfig{Enabled: true, TenantHeader: "X-Tenant-Id"}})
	handlers := append(mws.GetValidation(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	router.GET("/api/v1/insights/summary", handlers...)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/insights/summary", nil)
	req.Header.Set("Authorization", "Bearer axis-key")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
}

func TestAxisProductIntegrationAuth_AllowsCapabilitiesWithoutTenant(t *testing.T) {
	t.Setenv("X_API_KEY", "backend-key")
	t.Setenv("PONTI_AXIS_API_KEY", "axis-key")
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mws := NewDefaultMiddlewares(BuildConfig{Auth: IdentityAuthConfig{Enabled: true, TenantHeader: "X-Tenant-Id"}})
	handlers := append(mws.GetValidation(), func(c *gin.Context) {
		if c.Request.Context().Value(ctxkeys.Actor) != axisProductIntegrationActor {
			t.Fatalf("axis actor was not injected")
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	router.GET("/api/v1/capabilities", handlers...)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities", nil)
	req.Header.Set("Authorization", "Bearer axis-key")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
}

func TestAxisProductIntegrationAuth_AllowsAIDraftPrepareWithTenant(t *testing.T) {
	t.Setenv("X_API_KEY", "backend-key")
	t.Setenv("PONTI_AXIS_API_KEY", "axis-key")
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mws := NewDefaultMiddlewares(BuildConfig{Auth: IdentityAuthConfig{Enabled: true, TenantHeader: "X-Tenant-Id"}})
	handlers := append(mws.GetValidation(), func(c *gin.Context) {
		if c.Request.Context().Value(ctxkeys.Actor) != axisProductIntegrationActor {
			t.Fatalf("axis actor was not injected")
		}
		if c.Request.Context().Value(ctxkeys.OrgID) == nil {
			t.Fatalf("tenant was not injected")
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	router.POST("/api/v1/ai/actions/workorder-draft/prepare", handlers...)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/actions/workorder-draft/prepare", nil)
	req.Header.Set("Authorization", "Bearer axis-key")
	req.Header.Set("X-Tenant-Id", uuid.NewString())
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
}
