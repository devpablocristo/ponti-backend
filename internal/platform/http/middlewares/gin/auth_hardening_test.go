package pkgmwr

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestValidateIdentityClaimsRejectsIssuerMismatch(t *testing.T) {
	err := validateIdentityClaims(IdentityAuthConfig{
		Issuer:   "https://securetoken.google.com/ponti",
		Audience: "ponti",
	}, identityClaims{
		Subject:  "user-1",
		Issuer:   "https://securetoken.google.com/other",
		Audience: []string{"ponti"},
	})
	if err == nil {
		t.Fatal("expected issuer mismatch")
	}
}

func TestValidateIdentityClaimsRejectsAudienceMismatch(t *testing.T) {
	err := validateIdentityClaims(IdentityAuthConfig{
		Issuer:   "https://securetoken.google.com/ponti",
		Audience: "ponti",
	}, identityClaims{
		Subject:  "user-1",
		Issuer:   "https://securetoken.google.com/ponti",
		Audience: []string{"other"},
	})
	if err == nil {
		t.Fatal("expected audience mismatch")
	}
}

func TestValidateIdentityClaimsAcceptsIssuerAndAudience(t *testing.T) {
	err := validateIdentityClaims(IdentityAuthConfig{
		Issuer:   "https://securetoken.google.com/ponti",
		Audience: "ponti",
	}, identityClaims{
		Subject:  "user-1",
		Issuer:   "https://securetoken.google.com/ponti",
		Audience: []string{"ponti"},
	})
	if err != nil {
		t.Fatalf("expected claims to be valid: %v", err)
	}
}

func TestRejectUnsafeLocalAuthzBlocksProductionEnv(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/secure", RejectUnsafeLocalAuthz("production"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}
}

func TestSensitiveHeadersAreRedacted(t *testing.T) {
	if got := redactHeaderValue("X-API-KEY", []string{"secret"}); len(got) != 1 || got[0] != "<redacted>" {
		t.Fatalf("expected API key to be redacted, got %#v", got)
	}
	if got := redactHeaderValue("X-Tenant-Id", []string{"tenant"}); len(got) != 1 || got[0] != "<redacted>" {
		t.Fatalf("expected tenant header to be redacted, got %#v", got)
	}
	if got := redactHeaderValue("Accept", []string{"application/json"}); len(got) != 1 || got[0] != "application/json" {
		t.Fatalf("expected ordinary header to pass through, got %#v", got)
	}
}
