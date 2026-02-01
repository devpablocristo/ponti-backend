package config

type AI struct {
	ServiceURL string `envconfig:"AI_SERVICE_URL"`
	ServiceKey string `envconfig:"AI_SERVICE_KEY"`
	TimeoutMS  int    `envconfig:"AI_SERVICE_TIMEOUT_MS"`
}
