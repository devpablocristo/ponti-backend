package config

type Governance struct {
	URL             string `envconfig:"GOVERNANCE_URL" default:""`
	APIKey          string `envconfig:"GOVERNANCE_API_KEY" default:""`
	SyncCooldownSec int    `envconfig:"GOVERNANCE_SYNC_COOLDOWN_SEC" default:"60"`
}
