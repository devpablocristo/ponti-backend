package config

type Auth struct {
	Enabled              bool   `envconfig:"AUTH_ENABLED" default:"true"`
	IdentityProjectID    string `envconfig:"IDENTITY_PROJECT_ID" default:""`
	IdentityIssuer       string `envconfig:"IDENTITY_ISSUER" default:""`
	IdentityAudience     string `envconfig:"IDENTITY_AUDIENCE" default:""`
	IdentityJWKSURL      string `envconfig:"IDENTITY_JWKS_URL" default:"https://www.googleapis.com/service_accounts/v1/jwk/securetoken@system.gserviceaccount.com"`
	IdentityJWKSCacheTTL int    `envconfig:"IDENTITY_JWKS_CACHE_TTL_SECONDS" default:"300"`
	TenantHeader         string `envconfig:"TENANT_HEADER_NAME" default:"X-Tenant-Id"`
	RequireTenantHeader  bool   `envconfig:"AUTH_REQUIRE_TENANT_HEADER" default:"true"`
	AutoProvision        bool   `envconfig:"AUTH_AUTO_PROVISION_MEMBERSHIP" default:"false"`
	DefaultTenantName    string `envconfig:"AUTH_DEFAULT_TENANT_NAME" default:"default"`
	DefaultRoleName      string `envconfig:"AUTH_DEFAULT_ROLE_NAME" default:"admin"`
}
