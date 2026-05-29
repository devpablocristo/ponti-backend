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

// ProvideConfigAI ...
func ProvideConfigAI(cfg *config.Config) *config.AI {
	return &cfg.AI
}

// ConfigSet ...
var ConfigSet = wire.NewSet(
	ProvideAllConfigs,
	ProvideConfigDB,
	ProvideConfigSuggester,
	ProvideConfigAPI,
	ProvideConfigAI,
)
