package config

type General struct {
	Name        string `envconfig:"APP_NAME" default:"ponti-api" validate:"required"`
	Version     string `envconfig:"APP_VERSION" default:"1.0" validate:"required"`
	Environment string `envconfig:"DEPLOY_ENV" default:"local" validate:"oneof=local test dev staging prod"`
	Root        string `envconfig:"APP_ROOT" default:"/app"`
	MaxRetries  int    `envconfig:"APP_MAX_RETRIES" default:"5" validate:"gte=0"`
}
