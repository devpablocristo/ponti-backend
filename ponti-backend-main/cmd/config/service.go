package config

type Service struct {
	Name       string `envconfig:"SERVICE_NAME" default:"ponti-api" validate:"required"`
	Version    string `envconfig:"SERVICE_VERSION" default:"1.0" validate:"required"`
	GitSHA     string `envconfig:"SERVICE_GIT_SHA" default:""`
	BuildTime  string `envconfig:"SERVICE_BUILD_TIME" default:""`
	MaxRetries int    `envconfig:"SERVICE_MAX_RETRIES" default:"5" validate:"gte=0,lte=1000"`
}
