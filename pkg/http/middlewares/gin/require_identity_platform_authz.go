package pkgmwr

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	pkgtypes "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

const (
	permissionAPIRead  = "api.read"
	permissionAPIWrite = "api.write"
)

type IdentityAuthConfig struct {
	Enabled      bool
	ProjectID    string
	Issuer       string
	Audience     string
	JWKSURL      string
	CacheTTL     time.Duration
	TenantHeader string
	AutoProvision bool
	DefaultTenant string
	DefaultRole   string
}

type jwksCache struct {
	mu      sync.RWMutex
	keys    map[string]*rsa.PublicKey
	expires time.Time
}

type identityVerifier struct {
	cfg    IdentityAuthConfig
	client *http.Client
	cache  *jwksCache
}

type identityClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

type membershipResolved struct {
	TenantID    int64
	RoleName    string
	Permissions map[string]struct{}
}

func newIdentityVerifier(cfg IdentityAuthConfig) *identityVerifier {
	return &identityVerifier{
		cfg: cfg,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		cache: &jwksCache{
			keys: map[string]*rsa.PublicKey{},
		},
	}
}

func (v *identityVerifier) verify(rawToken string) (*identityClaims, error) {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithAudience(v.cfg.Audience),
		jwt.WithIssuer(v.cfg.Issuer),
	)
	claims := &identityClaims{}

	token, err := parser.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (any, error) {
		kid, _ := token.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("token missing kid")
		}
		key, keyErr := v.getKeyByKID(kid)
		if keyErr != nil {
			return nil, keyErr
		}
		return key, nil
	})
	if err != nil {
		return nil, err
	}
	if token == nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (v *identityVerifier) getKeyByKID(kid string) (*rsa.PublicKey, error) {
	v.cache.mu.RLock()
	key, ok := v.cache.keys[kid]
	notExpired := time.Now().Before(v.cache.expires)
	v.cache.mu.RUnlock()
	if ok && notExpired {
		return key, nil
	}

	v.cache.mu.Lock()
	defer v.cache.mu.Unlock()

	// double-check after acquiring write lock
	if key, ok := v.cache.keys[kid]; ok && time.Now().Before(v.cache.expires) {
		return key, nil
	}

	keys, ttl, err := v.fetchJWKS()
	if err != nil {
		return nil, err
	}
	v.cache.keys = keys
	v.cache.expires = time.Now().Add(ttl)

	key, ok = v.cache.keys[kid]
	if !ok {
		return nil, fmt.Errorf("kid %s not found in jwks", kid)
	}
	return key, nil
}

func (v *identityVerifier) fetchJWKS() (map[string]*rsa.PublicKey, time.Duration, error) {
	req, err := http.NewRequest(http.MethodGet, v.cfg.JWKSURL, nil)
	if err != nil {
		return nil, 0, err
	}
	resp, err := v.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("jwks returned status %d", resp.StatusCode)
	}

	var payload struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			Alg string `json:"alg"`
			Use string `json:"use"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, 0, err
	}

	keys := make(map[string]*rsa.PublicKey, len(payload.Keys))
	for _, k := range payload.Keys {
		if k.Kty != "RSA" || k.N == "" || k.E == "" || k.Kid == "" {
			continue
		}
		pub, err := parseRSAPublicKeyFromJWK(k.N, k.E)
		if err != nil {
			continue
		}
		keys[k.Kid] = pub
	}
	if len(keys) == 0 {
		return nil, 0, errors.New("jwks did not contain usable rsa keys")
	}

	ttl := v.cfg.CacheTTL
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return keys, ttl, nil
}

func parseRSAPublicKeyFromJWK(nStr, eStr string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, err
	}
	n := new(big.Int).SetBytes(nBytes)
	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}
	if e == 0 {
		return nil, errors.New("invalid exponent")
	}
	return &rsa.PublicKey{N: n, E: e}, nil
}

func RequireIdentityPlatformAuthz(cfg IdentityAuthConfig, db *gorm.DB) gin.HandlerFunc {
	verifier := newIdentityVerifier(cfg)
	return func(c *gin.Context) {
		start := time.Now()
		permission := permissionForMethod(c.Request.Method)

		tokenStr := extractBearer(c.GetHeader("Authorization"))
		if tokenStr == "" {
			denyAuthRequest(c, "missing bearer token")
			logAuthDecision("", "", c.FullPath(), permission, "DENY", start)
			return
		}

		claims, err := verifier.verify(tokenStr)
		if err != nil {
			denyAuthRequest(c, "invalid token")
			logAuthDecision("", "", c.FullPath(), permission, "DENY", start)
			return
		}
		if claims.Subject == "" {
			denyAuthRequest(c, "token missing subject")
			logAuthDecision("", "", c.FullPath(), permission, "DENY", start)
			return
		}

		userID, err := ensureUserByIDPSub(c.Request.Context(), db, claims.Subject, claims.Email)
		if err != nil {
			domErr := pkgtypes.NewError(pkgtypes.ErrAuthorization, "unable to resolve user", err)
			apiErr, status := pkgtypes.NewAPIError(domErr)
			c.AbortWithStatusJSON(status, apiErr.ToResponse())
			logAuthDecision(claims.Subject, "", c.FullPath(), permission, "DENY", start)
			return
		}

		membership, err := resolveMembership(c.Request.Context(), db, userID, c.GetHeader(cfg.TenantHeader))
		if err != nil {
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
				domErr := pkgtypes.NewError(pkgtypes.ErrAuthorization, "tenant membership required", err)
				apiErr, _ := pkgtypes.NewAPIError(domErr)
				c.AbortWithStatusJSON(http.StatusForbidden, apiErr.ToResponse())
				logAuthDecision(claims.Subject, "", c.FullPath(), permission, "DENY", start)
				return
			}
		}

		if _, ok := membership.Permissions[permission]; !ok {
			domErr := pkgtypes.NewError(pkgtypes.ErrAuthorization, "insufficient permissions", nil)
			apiErr, _ := pkgtypes.NewAPIError(domErr)
			c.AbortWithStatusJSON(http.StatusForbidden, apiErr.ToResponse())
			logAuthDecision(claims.Subject, strconv.FormatInt(membership.TenantID, 10), c.FullPath(), permission, "DENY", start)
			return
		}

		userIDStr := strconv.FormatInt(userID, 10)
		tenantIDStr := strconv.FormatInt(membership.TenantID, 10)
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, ContextUserIDKey, userIDStr)
		ctx = context.WithValue(ctx, ContextUserEmailKey, claims.Email)
		ctx = context.WithValue(ctx, ContextTenantIDKey, tenantIDStr)
		ctx = context.WithValue(ctx, ContextRolesKey, []string{membership.RoleName})
		c.Request = c.Request.WithContext(ctx)
		c.Set(ContextUserID, userIDStr)
		c.Set(ContextUserEmail, claims.Email)
		c.Set(ContextTenantID, tenantIDStr)
		c.Set(ContextRoles, []string{membership.RoleName})

		logAuthDecision(claims.Subject, tenantIDStr, c.FullPath(), permission, "ALLOW", start)
		c.Next()
	}
}

func ensureDefaultMembership(ctx context.Context, db *gorm.DB, userID int64, tenantName, roleName string) (*membershipResolved, error) {
	if strings.TrimSpace(tenantName) == "" {
		tenantName = "default"
	}
	if strings.TrimSpace(roleName) == "" {
		roleName = "admin"
	}

	type tenantRow struct {
		ID int64
	}
	type roleRow struct {
		ID int64
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
	return resolveMembership(ctx, db, userID, strconv.FormatInt(tenant.ID, 10))
}

func resolveMembership(ctx context.Context, db *gorm.DB, userID int64, requestedTenant string) (*membershipResolved, error) {
	type membershipRow struct {
		TenantID int64
		RoleID   int64
		RoleName string
	}

	query := db.WithContext(ctx).
		Table("auth_memberships AS m").
		Select("m.tenant_id AS tenant_id, m.role_id AS role_id, r.name AS role_name").
		Joins("JOIN auth_roles r ON r.id = m.role_id").
		Where("m.user_id = ? AND m.status = 'active'", userID)

	if strings.TrimSpace(requestedTenant) != "" {
		tenantID, err := strconv.ParseInt(strings.TrimSpace(requestedTenant), 10, 64)
		if err != nil {
			return nil, err
		}
		query = query.Where("m.tenant_id = ?", tenantID)
	}

	var row membershipRow
	if err := query.Order("m.tenant_id ASC").Limit(1).Take(&row).Error; err != nil {
		return nil, err
	}

	type permRow struct {
		Name string
	}
	var perms []permRow
	if err := db.WithContext(ctx).
		Table("auth_role_permissions rp").
		Select("p.name").
		Joins("JOIN auth_permissions p ON p.id = rp.permission_id").
		Where("rp.role_id = ?", row.RoleID).
		Find(&perms).Error; err != nil {
		return nil, err
	}

	permSet := make(map[string]struct{}, len(perms))
	for _, p := range perms {
		permSet[p.Name] = struct{}{}
	}
	return &membershipResolved{
		TenantID:    row.TenantID,
		RoleName:    row.RoleName,
		Permissions: permSet,
	}, nil
}

func ensureUserByIDPSub(ctx context.Context, db *gorm.DB, sub, email string) (int64, error) {
	type userRow struct {
		ID int64
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
		return 0, err
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
		return 0, err
	}

	if err := db.WithContext(ctx).
		Table("users").
		Select("id").
		Where("idp_sub = ?", sub).
		Limit(1).
		Take(&existing).Error; err != nil {
		return 0, err
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
	domErr := pkgtypes.NewError(pkgtypes.ErrAuthentication, details, nil)
	apiErr, status := pkgtypes.NewAPIError(domErr)
	c.AbortWithStatusJSON(status, apiErr.ToResponse())
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
