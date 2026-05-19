package pkgmwr

import (
	"net/http"
	"strings"
	"time"

	coreginmw "github.com/devpablocristo/core/http/gin/go"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Middlewares struct {
	global     []gin.HandlerFunc
	validation []gin.HandlerFunc
}

type BuildConfig struct {
	DB   *gorm.DB
	Auth IdentityAuthConfig
}

func NewDefaultMiddlewares(cfg BuildConfig) *Middlewares {
	global := []gin.HandlerFunc{
		ErrorHandling(),
		RequestAndResponseLogger(HttpLoggingOptions{
			LogLevel:       "info",
			IncludeHeaders: true,
			IncludeBody:    false,
			ExcludedPaths:  []string{"/health", "/ping", "/swagger/spec", "/swagger/ui/index.html"},
		}),
	}
	validation := []gin.HandlerFunc{coreginmw.RequireAPIKeyFromEnv("X_API_KEY")}
	if cfg.Auth.Enabled {
		validation = append(validation, RequireIdentityPlatformAuthz(cfg.Auth, cfg.DB))
	} else {
		if isLocalLikeEnvironment(cfg.Auth.Environment) {
			validation = append(validation, RequireLocalDevAuthz(cfg.Auth, cfg.DB))
		} else {
			validation = append(validation, RejectUnsafeLocalAuthz(cfg.Auth.Environment))
		}
	}
	if cfg.Auth.CacheTTL <= 0 {
		cfg.Auth.CacheTTL = 5 * time.Minute
	}

	return &Middlewares{
		global:     global,
		validation: validation,
	}
}

func isLocalLikeEnvironment(env string) bool {
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "", "local", "localhost", "dev", "development", "test", "testing":
		return true
	default:
		return false
	}
}

func RejectUnsafeLocalAuthz(env string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
			"message": "AUTH_ENABLED=false is allowed only in local/test environments",
			"env":     strings.TrimSpace(env),
		})
	}
}
func (m *Middlewares) GetGlobal() []gin.HandlerFunc     { return m.global }
func (m *Middlewares) GetValidation() []gin.HandlerFunc { return m.validation }
