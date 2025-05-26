package pkggin

import (
	"fmt"
)

type config struct {
	routerPort string
	apiVersion string
}

func newConfig(routerPort, ApiVersion string) *config {
	return &config{
		routerPort: routerPort,
		apiVersion: ApiVersion,
	}
}

func (c *config) GetRouterPort() string {
	return c.routerPort
}

func (c *config) Validate() error {
	if c.routerPort == "" {
		return fmt.Errorf("router port is not configured")
	}
	return nil
}
