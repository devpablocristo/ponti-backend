package config

type Loader interface {
	GetAppConfig() AppConfig
	GetHrConfig() HrConfig
	GetAssessmentConfig() AssessmentConfig
	GetPepConfig() PepConfig
}
