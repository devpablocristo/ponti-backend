package wire

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
	"github.com/google/wire"
)

// Provee toda la config
func ProvideAllConfigs() (*config.AllConfigs, error) {
	return config.LoadConfig()
}

// Extrae DB de la config
func ProvideConfigDB(cfg *config.AllConfigs) *config.DB {
	return &cfg.DB
}

// Extrae WordsSuggester de la config
func ProvideConfigSuggester(cfg *config.AllConfigs) *config.WordsSuggester {
	return &cfg.WordsSuggester
}

// ProvideConfigAPI extrae cfg.API y satisface todos los ConfigAPIPort de los dominios.
func ProvideConfigAPI(cfg *config.AllConfigs) *config.API {
	return &cfg.API
}

var ConfigSet = wire.NewSet(
	ProvideAllConfigs,
	ProvideConfigDB,
	ProvideConfigSuggester,
	ProvideConfigAPI,
)
