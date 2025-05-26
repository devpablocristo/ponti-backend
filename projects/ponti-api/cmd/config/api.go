package config

type API struct {
	APIVersion string `envconfig:"API_VERSION" default:"v1" validate:"required"`
	BaseURL    string // construida a partir de APIVersion
}

func (c *App) APIVersion() string {
	return c.API.APIVersion
}

func (c *App) BaseURL() string {
	return c.API.BaseURL
}
