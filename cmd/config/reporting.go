package config

const (
	ReportingReadModeLegacy       = "legacy"
	ReportingReadModeActorsShadow = "actors_shadow"
	ReportingReadModeActorsLive   = "actors_live"
)

type Reporting struct {
	ReadMode string `envconfig:"REPORTING_READ_MODE" default:"legacy"`
}

func (r Reporting) IsActorsShadow() bool {
	return r.ReadMode == ReportingReadModeActorsShadow
}

func (r Reporting) IsActorsLive() bool {
	return r.ReadMode == ReportingReadModeActorsLive
}
