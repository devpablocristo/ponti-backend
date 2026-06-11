package pkgmwr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/devpablocristo/platform/authn/go/jwks"
	"github.com/devpablocristo/platform/security/go/contextkeys"
	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/devpablocristo/platform/http/go/httperr"

	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
)

const (
	permissionAPIRead  = "api.read"
	permissionAPIWrite = "api.write"
)

type IdentityAuthConfig struct {
	Enabled       bool
	ProjectID     string
	Issuer        string
	Audience      string
	JWKSURL       string
	CacheTTL      time.Duration
	TenantHeader  string
	AutoProvision bool
	DefaultTenant string
	DefaultRole   string
}

type identityClaims struct {
	Subject string
	Email   string
}

type membershipResolved struct {
	TenantID    uuid.UUID
	RoleName    string
	Permissions map[string]struct{}
}

func extractClaimsFromMap(m map[string]any) identityClaims {
	var c identityClaims
	if sub, ok := m["sub"].(string); ok {
		c.Subject = sub
	}
	if email, ok := m["email"].(string); ok {
		c.Email = email
	}
	return c
}

func RequireIdentityPlatformAuthz(cfg IdentityAuthConfig, db *gorm.DB) gin.HandlerFunc {
	verifier := jwks.NewVerifier(cfg.JWKSURL)
	return func(c *gin.Context) {
		start := time.Now()
		permission := permissionForMethod(c.Request.Method)

		tokenStr := extractBearer(c.GetHeader("Authorization"))
		if tokenStr == "" {
			denyAuthRequest(c, "missing bearer token")
			logAuthDecision("", "", c.FullPath(), permission, "DENY", start)
			return
		}

		claimsMap, err := verifier.VerifyToken(c.Request.Context(), tokenStr)
		if err != nil {
			denyAuthRequest(c, "invalid token")
			logAuthDecision("", "", c.FullPath(), permission, "DENY", start)
			return
		}

		claims := extractClaimsFromMap(claimsMap)
		if claims.Subject == "" {
			denyAuthRequest(c, "token missing subject")
			logAuthDecision("", "", c.FullPath(), permission, "DENY", start)
			return
		}

		userID, err := ensureUserByIDPSub(c.Request.Context(), db, claims.Subject, claims.Email)
		if err != nil {
			domErr := domainerr.Forbidden("unable to resolve user")
			status, apiErr := httperr.Normalize(domErr)
			c.AbortWithStatusJSON(status, apiErr)
			logAuthDecision(claims.Subject, "", c.FullPath(), permission, "DENY", start)
			return
		}

		membership, err := resolveMembership(c.Request.Context(), db, userID, c.GetHeader(cfg.TenantHeader))
		if err != nil {
			// T1.c: usuario con >1 membership y sin X-Tenant-Id => exigir
			// selección explícita de tenant; NO elegir uno arbitrario.
			if errors.Is(err, errTenantSelectionRequired) {
				domErr := domainerr.Validation("tenant selection required: provide the X-Tenant-Id header")
				status, apiErr := httperr.Normalize(domErr)
				c.AbortWithStatusJSON(status, apiErr)
				logAuthDecision(claims.Subject, "", c.FullPath(), permission, "DENY", start)
				return
			}
			// PARTE IV: tenant suspendido o archivado.
			if errors.Is(err, errTenantInactive) {
				denyForbidden(c, "tenant is suspended or archived")
				logAuthDecision(claims.Subject, "", c.FullPath(), permission, "DENY", start)
				return
			}
			if cfg.AutoProvision {
				membership, err = ensureDefaultMembership(
					c.Request.Context(),
					db,
					userID,
					cfg.DefaultTenant,
					cfg.DefaultRole,
				)
			}
			if err != nil {
				domErr := domainerr.Forbidden("tenant membership required")
				status, apiErr := httperr.Normalize(domErr)
				c.AbortWithStatusJSON(status, apiErr)
				logAuthDecision(claims.Subject, "", c.FullPath(), permission, "DENY", start)
				return
			}
		}

		if _, ok := membership.Permissions[permission]; !ok {
			domErr := domainerr.Forbidden("insufficient permissions")
			status, apiErr := httperr.Normalize(domErr)
			c.AbortWithStatusJSON(status, apiErr)
			logAuthDecision(claims.Subject, membership.TenantID.String(), c.FullPath(), permission, "DENY", start)
			return
		}

		// Build scopes list from permissions.
		scopes := make([]string, 0, len(membership.Permissions))
		for p := range membership.Permissions {
			scopes = append(scopes, p)
		}

		// Inject core/saas/go context keys.
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, ctxkeys.Actor, claims.Subject)
		ctx = context.WithValue(ctx, ctxkeys.OrgID, membership.TenantID)
		ctx = context.WithValue(ctx, ctxkeys.Role, membership.RoleName)
		ctx = context.WithValue(ctx, ctxkeys.Scopes, scopes)
		c.Request = c.Request.WithContext(ctx)

		// Also set gin keys for convenience.
		c.Set(string(ctxkeys.Actor), claims.Subject)
		c.Set(string(ctxkeys.OrgID), membership.TenantID)
		c.Set(string(ctxkeys.Role), membership.RoleName)
		c.Set(string(ctxkeys.Scopes), scopes)

		logAuthDecision(claims.Subject, membership.TenantID.String(), c.FullPath(), permission, "ALLOW", start)
		c.Next()
	}
}

func ensureDefaultMembership(ctx context.Context, db *gorm.DB, userID uuid.UUID, tenantName, roleName string) (*membershipResolved, error) {
	if strings.TrimSpace(tenantName) == "" {
		tenantName = "default"
	}
	if strings.TrimSpace(roleName) == "" {
		roleName = "admin"
	}

	type tenantRow struct {
		ID uuid.UUID
	}
	type roleRow struct {
		ID uuid.UUID
	}

	var tenant tenantRow
	if err := db.WithContext(ctx).Table("auth_tenants").Select("id").Where("name = ?", tenantName).Limit(1).Take(&tenant).Error; err != nil {
		return nil, err
	}
	var role roleRow
	if err := db.WithContext(ctx).Table("auth_roles").Select("id").Where("name = ?", roleName).Limit(1).Take(&role).Error; err != nil {
		return nil, err
	}

	// Upsert membership
	now := time.Now().UTC()
	payload := map[string]any{
		"user_id":    userID,
		"tenant_id":  tenant.ID,
		"role_id":    role.ID,
		"status":     "active",
		"created_at": now,
		"updated_at": now,
	}
	if err := db.WithContext(ctx).Table("auth_memberships").Create(payload).Error; err != nil {
		// If duplicate, try update to ensure active status and role.
		_ = db.WithContext(ctx).
			Table("auth_memberships").
			Where("user_id = ? AND tenant_id = ?", userID, tenant.ID).
			Updates(map[string]any{
				"role_id":    role.ID,
				"status":     "active",
				"updated_at": now,
			}).Error
	}
	return resolveMembership(ctx, db, userID, tenant.ID.String())
}

// errTenantSelectionRequired indica que el usuario tiene >1 membership activa y
// no envió X-Tenant-Id: hay que pedir selección explícita de tenant en vez de
// elegir uno arbitrario (T1.c — antes: Order("m.tenant_id ASC").Limit(1)).
var errTenantSelectionRequired = errors.New("tenant selection required")

// errTenantInactive indica que el tenant resuelto está suspendido o archivado
// (PARTE IV). Se evalúa gated por TENANT_ENFORCEMENT.
var errTenantInactive = errors.New("tenant inactive")

// tenantActive devuelve true si el tenant existe, está 'active' y no archivado.
func tenantActive(ctx context.Context, db *gorm.DB, tenantID uuid.UUID) (bool, error) {
	var row struct {
		Status    string
		DeletedAt *time.Time
	}
	if err := db.WithContext(ctx).
		Table("auth_tenants").
		Select("status, deleted_at").
		Where("id = ?", tenantID).
		Limit(1).
		Take(&row).Error; err != nil {
		return false, err
	}
	return row.Status == "active" && row.DeletedAt == nil, nil
}

// denyForbidden corta el request con 403 (reutilizable dentro del paquete).
func denyForbidden(c *gin.Context, msg string) {
	domErr := domainerr.Forbidden(msg)
	status, apiErr := httperr.Normalize(domErr)
	c.AbortWithStatusJSON(status, apiErr)
}

func resolveMembership(ctx context.Context, db *gorm.DB, userID uuid.UUID, requestedTenant string) (*membershipResolved, error) {
	type membershipRow struct {
		TenantID uuid.UUID
		RoleID   uuid.UUID
		RoleName string
	}

	base := db.WithContext(ctx).
		Table("auth_memberships AS m").
		Select("m.tenant_id AS tenant_id, m.role_id AS role_id, r.name AS role_name").
		Joins("JOIN auth_roles r ON r.id = m.role_id").
		Where("m.user_id = ? AND m.status = 'active'", userID)

	requestedTenant = strings.TrimSpace(requestedTenant)
	if requestedTenant != "" {
		// Tenant explícito: validar membership en ESE tenant. Sin fallback:
		// si el usuario no la tiene, devolver el error (el caller deniega).
		tenantID, err := uuid.Parse(requestedTenant)
		if err != nil {
			return nil, err
		}
		var row membershipRow
		if err := base.Where("m.tenant_id = ?", tenantID).Take(&row).Error; err != nil {
			return nil, err
		}
		return loadMembershipPermissions(ctx, db, row.TenantID, row.RoleName, row.RoleID)
	}

	// Sin header de tenant: usar la única membership si hay exactamente una;
	// con más de una, exigir selección explícita (no elegir un tenant arbitrario).
	var rows []membershipRow
	if err := base.Limit(2).Find(&rows).Error; err != nil {
		return nil, err
	}
	switch len(rows) {
	case 0:
		return nil, gorm.ErrRecordNotFound
	case 1:
		return loadMembershipPermissions(ctx, db, rows[0].TenantID, rows[0].RoleName, rows[0].RoleID)
	default:
		return nil, errTenantSelectionRequired
	}
}

func loadMembershipPermissions(ctx context.Context, db *gorm.DB, tenantID uuid.UUID, roleName string, roleID uuid.UUID) (*membershipResolved, error) {
	type permRow struct {
		Name string
	}
	var perms []permRow
	if err := db.WithContext(ctx).
		Table("auth_role_permissions rp").
		Select("p.name").
		Joins("JOIN auth_permissions p ON p.id = rp.permission_id").
		Where("rp.role_id = ?", roleID).
		Find(&perms).Error; err != nil {
		return nil, err
	}

	permSet := make(map[string]struct{}, len(perms))
	for _, p := range perms {
		permSet[p.Name] = struct{}{}
	}

	// PARTE IV: tenant suspendido/archivado => denegar (gated por TENANT_ENFORCEMENT;
	// requiere migración 000233 aplicada).
	if sharedmodels.TenantEnforcementEnabled() {
		active, err := tenantActive(ctx, db, tenantID)
		if err != nil {
			return nil, err
		}
		if !active {
			return nil, errTenantInactive
		}
	}

	return &membershipResolved{
		TenantID:    tenantID,
		RoleName:    roleName,
		Permissions: permSet,
	}, nil
}

func ensureUserByIDPSub(ctx context.Context, db *gorm.DB, sub, email string) (uuid.UUID, error) {
	type userRow struct {
		ID uuid.UUID
	}
	var existing userRow
	if err := db.WithContext(ctx).
		Table("users").
		Select("id").
		Where("idp_sub = ?", sub).
		Limit(1).
		Take(&existing).Error; err == nil {
		return existing.ID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return uuid.Nil, err
	}

	if email == "" {
		email = fmt.Sprintf("%s@idp.local", sanitizeSubject(sub))
	}
	username := fmt.Sprintf("idp_%s", sanitizeSubject(sub))
	if len(username) > 100 {
		username = username[:100]
	}

	now := time.Now().UTC()
	insert := map[string]any{
		"email":          email,
		"username":       username,
		"password":       "",
		"token_hash":     "",
		"refresh_tokens": "{}",
		"active":         true,
		"is_verified":    true,
		"idp_sub":        sub,
		"idp_email":      email,
		"created_at":     now,
		"updated_at":     now,
	}
	if err := db.WithContext(ctx).Table("users").Create(insert).Error; err != nil {
		// Carrera concurrente: reintentar lectura.
		if err2 := db.WithContext(ctx).
			Table("users").
			Select("id").
			Where("idp_sub = ?", sub).
			Limit(1).
			Take(&existing).Error; err2 == nil {
			return existing.ID, nil
		}
		return uuid.Nil, err
	}

	if err := db.WithContext(ctx).
		Table("users").
		Select("id").
		Where("idp_sub = ?", sub).
		Limit(1).
		Take(&existing).Error; err != nil {
		return uuid.Nil, err
	}
	return existing.ID, nil
}

func sanitizeSubject(sub string) string {
	replacer := strings.NewReplacer("|", "_", ":", "_", "/", "_", "@", "_", " ", "_")
	s := replacer.Replace(sub)
	if s == "" {
		return "unknown"
	}
	return s
}

func extractBearer(authHeader string) string {
	authHeader = strings.TrimSpace(authHeader)
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func permissionForMethod(method string) string {
	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return permissionAPIRead
	default:
		return permissionAPIWrite
	}
}

func denyAuthRequest(c *gin.Context, details string) {
	domErr := domainerr.Unauthorized(details)
	status, apiErr := httperr.Normalize(domErr)
	c.AbortWithStatusJSON(status, apiErr)
}

func logAuthDecision(sub, tenantID, route, permission, result string, start time.Time) {
	payload := map[string]any{
		"user_id":             sub,
		"tenant_id":           tenantID,
		"route":               route,
		"required_permission": permission,
		"result":              result,
		"timestamp":           time.Now().UTC().Format(time.RFC3339),
		"latency_ms":          time.Since(start).Milliseconds(),
	}
	raw, _ := json.Marshal(payload)
	log.Printf("authz_audit=%s", string(raw))
}
