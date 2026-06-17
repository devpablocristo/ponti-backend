package config

type Auth struct {
	Enabled              bool   `envconfig:"AUTH_ENABLED" default:"true"`
	IdentityProjectID    string `envconfig:"IDENTITY_PROJECT_ID" default:""`
	IdentityIssuer       string `envconfig:"IDENTITY_ISSUER" default:""`
	IdentityAudience     string `envconfig:"IDENTITY_AUDIENCE" default:""`
	IdentityJWKSURL      string `envconfig:"IDENTITY_JWKS_URL" default:"https://www.googleapis.com/service_accounts/v1/jwk/securetoken@system.gserviceaccount.com"`
	IdentityJWKSCacheTTL int    `envconfig:"IDENTITY_JWKS_CACHE_TTL_SECONDS" default:"300"`
	TenantHeader         string `envconfig:"TENANT_HEADER_NAME" default:"X-Tenant-Id"`
	// T1.a (seguridad): auto-provisioning de membership DESACTIVADO por defecto.
	// Antes default:"true" + DefaultRole "admin" => cualquier JWT válido sin
	// membership era auto-onboardeado como admin del tenant "default" (fuga
	// cross-tenant). Si se habilita para dev, NUNCA usar rol "admin" por defecto.
	AutoProvision        bool   `envconfig:"AUTH_AUTO_PROVISION_MEMBERSHIP" default:"false"`
	DefaultTenantName    string `envconfig:"AUTH_DEFAULT_TENANT_NAME" default:"default"`
	DefaultRoleName      string `envconfig:"AUTH_DEFAULT_ROLE_NAME" default:"viewer"`

	// T1.b (seguridad): subjects (idp_sub del JWT) habilitados como platform-admin
	// (operadores globales). Solo ellos pueden crear tenants y gestionar memberships
	// cross-tenant vía /admin. Vacío = nadie => esas ops cross-tenant quedan cerradas.
	// Interino sin esquema; se reemplaza por is_platform_admin (users) en T2.
	PlatformAdminSubjects []string `envconfig:"AUTH_PLATFORM_ADMIN_SUBJECTS"`
}
