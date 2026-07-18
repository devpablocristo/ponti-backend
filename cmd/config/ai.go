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

	InsightLowStockEnabled   bool    `envconfig:"AI_INSIGHT_LOW_STOCK_ENABLED" default:"false"`
	InsightLowStockThreshold float64 `envconfig:"AI_INSIGHT_LOW_STOCK_THRESHOLD"` // fallback si el tenant no define ai.insights.low_stock_threshold_units
}
