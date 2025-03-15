package wire

import (
	config "github.com/alphacodinggroup/euxcel-backend/internal/config"
)

func ProvideConfigLoader() (config.Loader, error) {
	return config.NewConfigLoader()
}
