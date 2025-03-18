package wire

import (
	config "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/config"
)

func ProvideConfigLoader() (config.Loader, error) {
	return config.NewConfigLoader()
}
