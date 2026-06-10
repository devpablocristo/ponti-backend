package config

type AI struct {
	ServiceURL             string `envconfig:"AI_SERVICE_URL"`
	ServiceKey             string `envconfig:"AI_SERVICE_KEY"`
	TimeoutMS              int    `envconfig:"AI_SERVICE_TIMEOUT_MS"`
	ComputeThrottleSec     int    `envconfig:"AI_COMPUTE_THROTTLE_SEC"` // cooldown entre auto-compute por proyecto (default 300s)
	Provider               string `envconfig:"AI_PROVIDER" default:"legacy"`
	AxisEnabled            bool   `envconfig:"AI_AXIS_ENABLED" default:"false"`
	AxisCompanionURL       string `envconfig:"AXIS_COMPANION_BASE_URL"`
	AxisCompanionKey       string `envconfig:"AXIS_COMPANION_API_KEY"`
	AxisCompanionTimeoutMS int    `envconfig:"AXIS_COMPANION_TIMEOUT_MS" default:"45000"`
	AxisProductSurface     string `envconfig:"AXIS_PRODUCT_SURFACE" default:"ponti"`
}
