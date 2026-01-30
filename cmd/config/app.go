package config

type Service struct {
	Name       string `envconfig:"SERVICE_NAME" default:"ponti-api" validate:"required"`
	Version    string `envconfig:"SERVICE_VERSION" default:"1.0" validate:"required"`
	MaxRetries int    `envconfig:"SERVICE_MAX_RETRIES" default:"5" validate:"gte=0,lte=1000"`
}
