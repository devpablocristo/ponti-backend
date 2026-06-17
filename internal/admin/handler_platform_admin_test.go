package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/devpablocristo/platform/security/go/contextkeys"
)

func newTestContextWithActor(sub string) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	if sub != "" {
		c.Set(string(ctxkeys.Actor), sub)
	}
	return c
}

// TestIsPlatformAdmin_T1b valida el allowlist de platform-admin (env-allowlist
// interino) que gobierna las operaciones cross-tenant de /admin.
func TestIsPlatformAdmin_T1b(t *testing.T) {
	h := NewHandler(nil, nil, nil, nil, nil, []string{"sub-platform", " sub-spaces "})

	assert.True(t, h.isPlatformAdmin(newTestContextWithActor("sub-platform")))
	assert.True(t, h.isPlatformAdmin(newTestContextWithActor("sub-spaces")), "debe trimear espacios del allowlist")
	assert.False(t, h.isPlatformAdmin(newTestContextWithActor("sub-otro")))
	assert.False(t, h.isPlatformAdmin(newTestContextWithActor("")), "sin actor => no platform-admin")

	// Sin allowlist, nadie es platform-admin (ops cross-tenant cerradas por defecto).
	empty := NewHandler(nil, nil, nil, nil, nil, nil)
	assert.False(t, empty.isPlatformAdmin(newTestContextWithActor("sub-platform")))
}
