package wire

import (
	"github.com/devpablocristo/ponti-backend/cmd/config"
	"github.com/google/wire"
)

// ProvideAllConfigs ...
func ProvideAllConfigs() (*config.Config, error) {
	return config.LoadConfig()
}

// ProvideConfigDB ...
func ProvideConfigDB(cfg *config.Config) *config.DB {
	return &cfg.DB
}

// ProvideConfigSuggester ...
func ProvideConfigSuggester(cfg *config.Config) *config.WordsSuggester {
	return &cfg.WordsSuggester
}

// ProvideConfigAPI ...
func ProvideConfigAPI(cfg *config.Config) *config.API {
	return &cfg.API
}

// ProvideConfigCompanion expone la sección Companion del config.
func ProvideConfigCompanion(cfg *config.Config) *config.Companion {
	return &cfg.Companion
}

// ConfigSet expone todos los providers de config para wire.
var ConfigSet = wire.NewSet(
	ProvideAllConfigs,
	ProvideConfigDB,
	ProvideConfigSuggester,
	ProvideConfigAPI,
	ProvideConfigCompanion,
)
