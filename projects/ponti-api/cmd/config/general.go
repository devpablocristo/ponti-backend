package config

type General struct {
	AppName     string `envconfig:"APP_NAME" default:"ponti-api" validate:"required"`
	AppVersion  string `envconfig:"APP_VERSION" default:"1.0" validate:"required"`
	Environment string `envconfig:"APP_ENV" default:"local" validate:"oneof=local dev stage prod"`
	AppRoot     string `envconfig:"APP_ROOT" default:"/app"`
	MaxRetries  int    `envconfig:"APP_MAX_RETRIES" default:"5" validate:"gte=0"`
}
