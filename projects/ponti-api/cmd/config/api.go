package config

type API struct {
	Version string `envconfig:"API_VERSION" default:"v1" validate:"required"`
	URL     string // construida a partir de APIVersion
}

func (c *API) APIVersion() string {
	return c.Version
}

func (c *API) BaseURL() string {
	return c.URL
}
