package config

type Review struct {
	URL             string `envconfig:"REVIEW_URL" default:""`
	APIKey          string `envconfig:"REVIEW_API_KEY" default:""`
	SyncCooldownSec int    `envconfig:"REVIEW_SYNC_COOLDOWN_SEC" default:"60"`
}
