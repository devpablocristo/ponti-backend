package config

type Security struct {
	TenantStrictMode bool `envconfig:"TENANT_STRICT_MODE" default:"false"`
	DomainPoliciesV2 bool `envconfig:"DOMAIN_POLICIES_V2" default:"false"`
	AITenantScope    bool `envconfig:"AI_TENANT_SCOPE" default:"false"`
}
