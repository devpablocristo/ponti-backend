package config

type Nexus struct {
	BaseURL               string `envconfig:"NEXUS_BASE_URL"`
	APIKey                string `envconfig:"NEXUS_API_KEY"`
	TimeoutMS             int    `envconfig:"NEXUS_TIMEOUT_MS" default:"30000"`
	CallbackToken         string `envconfig:"NEXUS_CALLBACK_TOKEN"`          // HMAC compartido para verificar X-Nexus-Callback-Signature
	AttestationHMACSecret string `envconfig:"NEXUS_ATTESTATION_HMAC_SECRET"` // opcional: firma de attestations (Ola B)
	GovernedWritesEnabled bool   `envconfig:"AI_GOVERNED_WRITES_ENABLED" default:"false"`
	VerifyNexus           bool   `envconfig:"GOVERNANCE_VERIFY_NEXUS" default:"false"`
}
