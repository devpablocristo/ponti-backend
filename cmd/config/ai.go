package config

type AI struct {
	ServiceURL         string `envconfig:"AI_SERVICE_URL"`
	ServiceKey         string `envconfig:"AI_SERVICE_KEY"`
	TimeoutMS          int    `envconfig:"AI_SERVICE_TIMEOUT_MS"`
	ComputeThrottleSec int    `envconfig:"AI_COMPUTE_THROTTLE_SEC"` // cooldown entre auto-compute por proyecto (default 300s)
}
