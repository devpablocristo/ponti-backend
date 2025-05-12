package wire

import (
	config "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/cmd/config"
)

func ProvideConfigLoader() (config.Loader, error) {
	return config.NewConfigLoader()
}
