package wire

import (
	config "github.com/devpablocristo/monorepo/projects/qh/internal/config"
)

func ProvideConfigLoader() (config.Loader, error) {
	return config.NewConfigLoader()
}
