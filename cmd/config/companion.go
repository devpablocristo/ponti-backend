package config

// Companion agrupa la configuración del cliente HTTP de Companion (servicio AI
// horizontal en axis/). Defaults razonables; sólo `BaseURL` + `JWTSecret` son
// obligatorios.
type Companion struct {
	BaseURL     string `envconfig:"COMPANION_BASE_URL"`
	JWTSecret   string `envconfig:"COMPANION_INTERNAL_JWT_SECRET"`
	JWTIssuer   string `envconfig:"COMPANION_INTERNAL_JWT_ISSUER" default:"ponti-backend"`
	JWTAudience string `envconfig:"COMPANION_INTERNAL_JWT_AUDIENCE" default:"companion"`
	JWTTTLSec   int    `envconfig:"COMPANION_INTERNAL_JWT_TTL_SEC" default:"300"`
	TimeoutMS   int    `envconfig:"COMPANION_TIMEOUT_MS" default:"30000"`
	MaxRetries  int    `envconfig:"COMPANION_MAX_RETRIES" default:"2"`
}

// Nexus agrupa la configuración del cliente de Nexus (motor de decisiones en
// axis/nexus). Mismo patrón que Companion: HS256 con secret compartido.
type Nexus struct {
	BaseURL     string `envconfig:"NEXUS_BASE_URL"`
	JWTSecret   string `envconfig:"NEXUS_INTERNAL_JWT_SECRET"`
	JWTIssuer   string `envconfig:"NEXUS_INTERNAL_JWT_ISSUER" default:"ponti-backend"`
	JWTAudience string `envconfig:"NEXUS_INTERNAL_JWT_AUDIENCE" default:"nexus"`
	JWTTTLSec   int    `envconfig:"NEXUS_INTERNAL_JWT_TTL_SEC" default:"300"`
	TimeoutMS   int    `envconfig:"NEXUS_TIMEOUT_MS" default:"10000"`
	MaxRetries  int    `envconfig:"NEXUS_MAX_RETRIES" default:"2"`
}
