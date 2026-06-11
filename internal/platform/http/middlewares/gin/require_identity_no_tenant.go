package pkgmwr

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/devpablocristo/platform/authn/go/jwks"
	ctxkeys "github.com/devpablocristo/platform/security/go/contextkeys"
)

// RequireIdentityNoTenant autentica el request (JWT vía JWKS + ensureUserByIDPSub)
// pero NO exige selección de tenant. Para endpoints tenant-agnósticos como
// GET /me/context, cuyo propósito es listar los tenants del usuario para que el FE
// elija uno — un usuario con >1 membership y sin X-Tenant-Id NO debe ser rechazado
// acá (el middleware con tenant sí lo rechazaría). Inyecta solo ctxkeys.Actor
// (idp_sub); la resolución de tenant queda a cargo del handler.
func RequireIdentityNoTenant(cfg IdentityAuthConfig, db *gorm.DB) gin.HandlerFunc {
	verifier := jwks.NewVerifier(cfg.JWKSURL)
	return func(c *gin.Context) {
		start := time.Now()

		tokenStr := extractBearer(c.GetHeader("Authorization"))
		if tokenStr == "" {
			denyAuthRequest(c, "missing bearer token")
			logAuthDecision("", "", c.FullPath(), "identity", "DENY", start)
			return
		}

		claimsMap, err := verifier.VerifyToken(c.Request.Context(), tokenStr)
		if err != nil {
			denyAuthRequest(c, "invalid token")
			logAuthDecision("", "", c.FullPath(), "identity", "DENY", start)
			return
		}

		claims := extractClaimsFromMap(claimsMap)
		if claims.Subject == "" {
			denyAuthRequest(c, "token missing subject")
			logAuthDecision("", "", c.FullPath(), "identity", "DENY", start)
			return
		}

		// Asegura la fila de usuario (idempotente) para que /me/context funcione
		// incluso en el primer request del usuario.
		if _, err := ensureUserByIDPSub(c.Request.Context(), db, claims.Subject, claims.Email); err != nil {
			denyForbidden(c, "unable to resolve user")
			logAuthDecision(claims.Subject, "", c.FullPath(), "identity", "DENY", start)
			return
		}

		ctx := context.WithValue(c.Request.Context(), ctxkeys.Actor, claims.Subject)
		c.Request = c.Request.WithContext(ctx)
		c.Set(string(ctxkeys.Actor), claims.Subject)

		logAuthDecision(claims.Subject, "", c.FullPath(), "identity", "ALLOW", start)
		c.Next()
	}
}
